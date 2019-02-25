package p1

//package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Flag_value struct {
	Encoded_prefix []uint8
	Value          string
}

type Node struct {
	Node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	Branch_value [17]string
	Flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	Db             map[string]Node
	Root           string
	InsertedRecord map[string]string
}

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	if key == "" {
		return "", nil
	}
	var path = getHexArray(key)
	var value = mpt.getHelper(mpt.Root, path)
	if value == "" {
		return "", errors.New("path_not_found")
	} else {
		return value, nil
	}
}

func (mpt *MerklePatriciaTrie) getHelper(nodeHash string, path []uint8) string {
	var node = mpt.Db[nodeHash]
	var nodeType = node.Node_type
	if nodeType == 0 {
		return ""
	} else if nodeType == 1 {
		if getBranchCommonPath(node.Branch_value, path) {
			if path[0] == uint8(16) {
				return node.Branch_value[16]
			} else {
				return mpt.getHelper(node.Branch_value[path[0]], path[1:])
			}
		} else {
			return ""
		}
	} else if nodeType == 2 {
		var encodeValue = node.Flag_value.Encoded_prefix
		var decodeValue = compact_decode(encodeValue)
		var nodeValue = node.Flag_value.Value
		var isLeaf = isLeafNode(encodeValue)
		var nodePath []uint8
		if isLeaf {
			nodePath = append(decodeValue, uint8(16)) //since it is the leaf node, add 16 back
		} else {
			nodePath = decodeValue
		}
		var commonPath = getExtLeafCommonPath(nodePath, path)
		var restPath = getRestPath(path, commonPath)
		var restNibble = getRestNibble(nodePath, commonPath)
		var cpLen = len(commonPath)
		var rpLen = len(restPath)
		var rnLen = len(restNibble)
		if isLeaf {
			if cpLen != 0 && rpLen == 0 && rnLen == 0 {
				return nodeValue
			} else {
				return ""
			}
		} else {
			if cpLen != 0 && rnLen == 0 {
				return mpt.getHelper(nodeValue, restPath)
			} else {
				return ""
			}
		}
	}
	return ""
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	mpt.InsertedRecord[key] = new_value
	if mpt.Root == "" {
		mpt.Db = make(map[string]Node)
		var rootNode Node
		rootNode.Node_type = 0
		mpt.Db[rootNode.hash_node()] = rootNode
		mpt.Root = rootNode.hash_node()
	}
	var path = getHexArray(key)
	mpt.Root = mpt.insertHelper(mpt.Root, path, new_value)
}
func (mpt *MerklePatriciaTrie) insertHelper(nodeHash string, path []uint8, value string) string {
	var node = mpt.Db[nodeHash]
	var nodeType = node.Node_type
	var nodeKey = node.hash_node()
	if nodeType == 0 { //insert into Null
		delete(mpt.Db, nodeKey)
		var rootNode Node
		rootNode.Node_type = 2
		rootNode.Flag_value.Encoded_prefix = compact_encode(path)
		rootNode.Flag_value.Value = value
		var hashValue = rootNode.hash_node()
		mpt.Db[hashValue] = rootNode
		return hashValue
	} else if nodeType == 1 { //insert into Branch Node
		if path[0] == uint8(16) { //if insert into branch node value, just update the value
			delete(mpt.Db, nodeKey)
			node.Branch_value[16] = value
			mpt.Db[node.hash_node()] = node
			return node.hash_node()
		}
		if getBranchCommonPath(node.Branch_value, path) { //exist common path
			var commonPath = path[0]
			var nextNodeHash = node.Branch_value[commonPath]
			node.Branch_value[commonPath] = mpt.insertHelper(nextNodeHash, path[1:], value)
			var nodeHashValue = node.hash_node()
			delete(mpt.Db, nodeKey)
			mpt.Db[nodeHashValue] = node
			return nodeHashValue
		} else { // don't exist common path
			var restPath = path[0]
			var newLeafNode Node
			newLeafNode.Node_type = 2
			newLeafNode.Flag_value.Value = value
			newLeafNode.Flag_value.Encoded_prefix = compact_encode(path[1:])
			delete(mpt.Db, nodeKey)
			mpt.Db[newLeafNode.hash_node()] = newLeafNode
			node.Branch_value[restPath] = newLeafNode.hash_node()
			var nodeHashValue = node.hash_node()
			mpt.Db[nodeHashValue] = node
			return nodeHashValue
		}
	} else if nodeType == 2 { //insert into extension node or leaf node
		var encodeValue = node.Flag_value.Encoded_prefix
		var nodeValue = node.Flag_value.Value
		var isLeaf = isLeafNode(encodeValue)
		if isLeaf { //insert into leaf node
			var nodePath = append(compact_decode(encodeValue), uint8(16)) //since it is the leaf node, add 16 back
			var commonPath = getExtLeafCommonPath(nodePath, path)
			var restPath = getRestPath(path, commonPath)
			var restNibble = getRestNibble(nodePath, commonPath)
			var cpLen = len(commonPath)
			var rpLen = len(restPath)
			var rnLen = len(restNibble)
			if cpLen != 0 && rpLen == 0 && rnLen == 0 { //update the leaf node value
				node.Flag_value.Value = value
				delete(mpt.Db, nodeKey)
				mpt.Db[node.hash_node()] = node
				return node.hash_node()
			}
			if cpLen != 0 && rpLen != 0 && rnLen != 0 { //has common path so 1)new extension node 2) new branch node 3)insert these two nodes
				delete(mpt.Db, nodeKey)
				var newExtNode Node
				newExtNode.Node_type = 2
				newExtNode.Flag_value.Encoded_prefix = compact_encode(commonPath)
				var newBranchNode Node
				newBranchNode.Node_type = 1
				mpt.Db[newBranchNode.hash_node()] = newBranchNode
				var newBranchNodeHash = newBranchNode.hash_node()
				newBranchNodeHash = mpt.insertHelper(newBranchNodeHash, restPath, value)
				newBranchNodeHash = mpt.insertHelper(newBranchNodeHash, restNibble, nodeValue)
				newExtNode.Flag_value.Value = newBranchNodeHash
				mpt.Db[newExtNode.hash_node()] = newExtNode
				return newExtNode.hash_node()
			}
			if cpLen == 0 && rpLen != 0 && rnLen != 0 { //doesn't have common path so 1)new branch node 2) insert two nodes into branch node
				var newBranchNode Node
				newBranchNode.Node_type = 1
				var newBranchNodeHash = newBranchNode.hash_node()
				mpt.Db[newBranchNodeHash] = newBranchNode
				newBranchNodeHash = mpt.insertHelper(newBranchNodeHash, restPath, value)
				newBranchNodeHash = mpt.insertHelper(newBranchNodeHash, restNibble, nodeValue)
				return newBranchNodeHash
			}
		} else { //insert into extension node
			var nodePath = compact_decode(encodeValue)
			var commonPath = getExtLeafCommonPath(nodePath, path)
			var restPath = getRestPath(path, commonPath)
			var restNibble = getRestNibble(nodePath, commonPath)
			var cpLen = len(commonPath)
			var rpLen = len(restPath)
			var rnLen = len(restNibble)
			if cpLen != 0 && rpLen != 0 && rnLen != 0 { // 1）new extension node 2)new branch node 3)insert two paths into branch node
				delete(mpt.Db, nodeKey)
				var newExtNode Node
				newExtNode.Node_type = 2
				newExtNode.Flag_value.Encoded_prefix = compact_encode(commonPath)
				var newBranchNode Node
				newBranchNode.Node_type = 1
				if rpLen > 1 {
					var newLeafNode Node
					newLeafNode.Node_type = 2
					newLeafNode.Flag_value.Value = value
					newLeafNode.Flag_value.Encoded_prefix = compact_encode(restPath[1:])
					mpt.Db[newLeafNode.hash_node()] = newLeafNode
					newBranchNode.Branch_value[restPath[0]] = newLeafNode.hash_node()
				} else {
					newBranchNode.Branch_value[16] = value
				}
				if rnLen > 1 {
					var newNextExtNode Node
					newNextExtNode.Node_type = 2
					newNextExtNode.Flag_value.Value = nodeValue
					newNextExtNode.Flag_value.Encoded_prefix = compact_encode(restNibble[1:])
					mpt.Db[newNextExtNode.hash_node()] = newNextExtNode
					newBranchNode.Branch_value[restNibble[0]] = newNextExtNode.hash_node()
				} else {
					newBranchNode.Branch_value[restNibble[0]] = nodeValue
				}
				mpt.Db[newBranchNode.hash_node()] = newBranchNode
				newExtNode.Flag_value.Value = newBranchNode.hash_node()
				mpt.Db[newExtNode.hash_node()] = newExtNode
				return newExtNode.hash_node()
			} else if cpLen != 0 && rpLen != 0 && rnLen == 0 { //directly insert rest path into next node
				//var nextNode = mpt.db[nodeValue]
				node.Flag_value.Value = mpt.insertHelper(nodeValue, restPath, value)
				delete(mpt.Db, nodeValue)
				mpt.Db[node.hash_node()] = node
				return node.hash_node()
			} else if cpLen == 0 && rpLen != 0 && rnLen == 1 { //1)new branch 2)insert branch
				var newBranchNode Node
				newBranchNode.Node_type = 1
				if rpLen == 1 { //insert 16
					newBranchNode.Branch_value[16] = value
				} else {
					var newLeafNode Node
					newLeafNode.Node_type = 2
					newLeafNode.Flag_value.Value = value
					newLeafNode.Flag_value.Encoded_prefix = compact_encode(restPath[1:])
					newBranchNode.Branch_value[restPath[0]] = newLeafNode.hash_node()
					mpt.Db[newLeafNode.hash_node()] = newLeafNode
				}
				newBranchNode.Branch_value[restNibble[0]] = nodeValue
				delete(mpt.Db, nodeKey)
				mpt.Db[newBranchNode.hash_node()] = newBranchNode
				return newBranchNode.hash_node()
			} else if cpLen == 0 && rpLen != 0 && rnLen > 1 { //1)new branch node 2)new extension node
				var newBranchNode Node
				newBranchNode.Node_type = 1
				if rpLen == 1 { //insert 16
					newBranchNode.Branch_value[16] = value
				} else {
					var newLeafNode Node
					newLeafNode.Node_type = 2
					newLeafNode.Flag_value.Value = value
					newLeafNode.Flag_value.Encoded_prefix = compact_encode(restPath[1:])
					newBranchNode.Branch_value[restPath[0]] = newLeafNode.hash_node()
					mpt.Db[newLeafNode.hash_node()] = newLeafNode
				}
				var newExtNode Node
				newExtNode.Node_type = 2
				newExtNode.Flag_value.Encoded_prefix = compact_encode(restNibble[1:])
				newExtNode.Flag_value.Value = nodeValue
				delete(mpt.Db, nodeKey)
				mpt.Db[newExtNode.hash_node()] = newExtNode
				newBranchNode.Branch_value[restNibble[0]] = newExtNode.hash_node()
				mpt.Db[newBranchNode.hash_node()] = newBranchNode
				return newBranchNode.hash_node()
			}
		}
	}
	return ""
}

func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	delete(mpt.InsertedRecord, key)
	var path = getHexArray(key)
	isSuc, hashValue := mpt.deleteHelper(mpt.Root, path)
	if isSuc {
		mpt.Root = hashValue
		return "", nil
	} else {
		return "", errors.New("path_not_found")
	}
}

func getArrayInBranchValue(branchValue [17]string) []int {
	var result []int
	for i := range branchValue {
		if branchValue[i] != "" {
			result = append(result, i)
		}
	}
	return result
}

func (mpt *MerklePatriciaTrie) deleteHelper(nodeHash string, path []uint8) (bool, string) {
	var node = mpt.Db[nodeHash]
	var nodeType = node.Node_type
	var nodeKey = node.hash_node()
	if nodeType == 0 { // delete at Null node
		return false, ""
	} else if nodeType == 1 { // delete at Branch node
		if getBranchCommonPath(node.Branch_value, path) {
			if path[0] == uint8(16) { // delete branch node value at 16
				node.Branch_value[16] = ""
			} else { // has common path but not delete the branch value, into recursion
				var isSuc bool
				var nextKey string
				isSuc, nextKey = mpt.deleteHelper(node.Branch_value[path[0]], path[1:])
				if isSuc { //delete successfully
					node.Branch_value[path[0]] = nextKey
				} else { //not found
					return false, ""
				}
			}
			//check if there is only one value in branch node remaining
			if len(getArrayInBranchValue(node.Branch_value)) > 1 { //do not need to balance
				delete(mpt.Db, nodeKey)
				mpt.Db[node.hash_node()] = node
				return true, node.hash_node()
			} else if getArrayInBranchValue(node.Branch_value)[0] == 16 { // it is a leaf node
				var returnNode Node
				returnNode.Node_type = 2
				returnNode.Flag_value.Value = node.Branch_value[16]
				returnNode.Flag_value.Encoded_prefix = compact_encode([]uint8{uint8(16)})
				delete(mpt.Db, nodeKey)
				mpt.Db[returnNode.hash_node()] = returnNode
				return true, returnNode.hash_node()
			} else {
				var returnNode Node
				var branchIndex = getArrayInBranchValue(node.Branch_value)[0]
				var nextNodeKey = node.Branch_value[branchIndex]
				var nextNode = mpt.Db[nextNodeKey]
				var nextNodeType = nextNode.Node_type
				if nextNodeType == 2 { // if next node is extension or leaf node, combine them
					returnNode.Node_type = 2
					returnNode.Flag_value.Value = nextNode.Flag_value.Value
					var nextNodeEncodeValue = nextNode.Flag_value.Encoded_prefix
					var nextNodeDecodeValue = compact_decode(nextNodeEncodeValue)
					var newPathValue = append([]uint8{uint8(branchIndex)}, nextNodeDecodeValue...)
					if isLeafNode(nextNodeEncodeValue) {
						newPathValue = append(newPathValue, uint8(16))
					}
					returnNode.Flag_value.Encoded_prefix = compact_encode(newPathValue)
					delete(mpt.Db, nodeKey)
					delete(mpt.Db, nextNodeKey)
					mpt.Db[returnNode.hash_node()] = returnNode
					return true, returnNode.hash_node()
				} else { // if next node is branch node, return it as extension node
					returnNode.Node_type = 2
					returnNode.Flag_value.Value = nextNodeKey
					returnNode.Flag_value.Encoded_prefix = compact_encode([]uint8{uint8(branchIndex)})
					mpt.Db[returnNode.hash_node()] = returnNode
					delete(mpt.Db, nodeKey)
					return true, returnNode.hash_node()
				}
			}
		} else { // not found
			return false, ""
		}
	} else if nodeType == 2 { //delete at leaf or extension node
		var encodeValue = node.Flag_value.Encoded_prefix
		var nodeValue = node.Flag_value.Value
		var isLeaf = isLeafNode(encodeValue)
		if isLeaf {
			var nodePath = append(compact_decode(encodeValue), uint8(16)) //since it is the leaf node, add 16 back
			var commonPath = getExtLeafCommonPath(nodePath, path)
			var restPath = getRestPath(path, commonPath)
			var restNibble = getRestNibble(nodePath, commonPath)
			var cpLen = len(commonPath)
			var rpLen = len(restPath)
			var rnLen = len(restNibble)
			if cpLen != 0 && rpLen == 0 && rnLen == 0 {
				delete(mpt.Db, nodeKey)
				return true, ""
			} else {
				return false, ""
			}
		} else { // exstension node
			var nodePath = compact_decode(encodeValue)
			var commonPath = getExtLeafCommonPath(nodePath, path)
			var restPath = getRestPath(path, commonPath)
			var restNibble = getRestNibble(nodePath, commonPath)
			var cpLen = len(commonPath)
			var rpLen = len(restPath)
			var rnLen = len(restNibble)
			if cpLen != 0 && rpLen != 0 && rnLen == 0 {
				var isSuc bool
				var nextKey string
				isSuc, nextKey = mpt.deleteHelper(nodeValue, restPath)
				if isSuc {
					var nextReturnNode = mpt.Db[nextKey]
					var nextReturnNodeType = nextReturnNode.Node_type
					if nextReturnNodeType == 2 { //combine the return node and this extension node
						var newNode Node
						newNode.Node_type = 2
						var nextReturnNodeEncodeValue = nextReturnNode.Flag_value.Encoded_prefix
						var newValue = append(compact_decode(encodeValue), compact_decode(nextReturnNodeEncodeValue)...)
						if isLeafNode(nextReturnNodeEncodeValue) {
							newValue = append(newValue, uint8(16))
						}
						var newEncodeValue = compact_encode(newValue)
						newNode.Flag_value.Encoded_prefix = newEncodeValue
						newNode.Flag_value.Value = nextReturnNode.Flag_value.Value
						mpt.Db[newNode.hash_node()] = newNode
						delete(mpt.Db, nodeKey)
						return true, newNode.hash_node()
					} else { // just connect the next node
						node.Flag_value.Value = nextKey
						mpt.Db[node.hash_node()] = node
						delete(mpt.Db, nodeKey)
						return true, node.hash_node()
					}
				} else {
					return false, ""
				}
			} else {
				return false, ""
			}
		}
	}
	return false, ""
}

func getBranchCommonPath(branchValue [17]string, path []uint8) bool {
	var n = path[0]
	if branchValue[n] == "" {
		return false
	} else {
		return true
	}
}

func getExtLeafCommonPath(nodePath []uint8, insertPath []uint8) []uint8 {
	commonPath := []uint8{}
	var loopTimes int
	if len(nodePath) > len(insertPath) {
		loopTimes = len(insertPath)
	} else {
		loopTimes = len(nodePath)
	}
	for i := 0; i < loopTimes; i++ {
		if nodePath[i] == insertPath[i] {
			commonPath = append(commonPath, nodePath[i])
		} else {
			return commonPath
		}
	}
	return commonPath
}

func getRestPath(insertPath []uint8, commonPath []uint8) []uint8 {
	return insertPath[len(commonPath):]
}

func getRestNibble(nodePath []uint8, commonPath []uint8) []uint8 {
	return nodePath[len(commonPath):]
}

func compact_encode(hex_array []uint8) []uint8 {
	var term int
	var lenArray = len(hex_array)
	if hex_array[lenArray-1] == 16 {
		term = 1
		hex_array = hex_array[:lenArray-1]
	} else {
		term = 0
	}
	var oddLen = len(hex_array) % 2
	var flag = 2*term + oddLen
	var tempArray []uint8
	if oddLen == 1 {
		tempArray = append([]uint8{uint8(flag)}, hex_array...)
	} else {
		tempArray = append([]uint8{uint8(flag), 0}, hex_array...)
	}
	var hpArray []uint8
	for i := 0; i < len(tempArray); {
		hpArray = append(hpArray, tempArray[i]*16+tempArray[i+1])
		i = i + 2
	}
	return hpArray
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	var hexArray []uint8
	for i := 0; i < len(encoded_arr); i++ {
		n := encoded_arr[i]
		hexArray = append(hexArray, uint8(n/16))
		hexArray = append(hexArray, uint8(n%16))
	}
	if hexArray[0] == 1 || hexArray[0] == 3 {
		hexArray = hexArray[1:]
	} else {
		hexArray = hexArray[2:]
	}
	return hexArray
}

func isLeafNode(encodedArray []uint8) bool {
	var hexArray []uint8
	for i := 0; i < len(encodedArray); i++ {
		n := encodedArray[i]
		hexArray = append(hexArray, uint8(n/16))
		hexArray = append(hexArray, uint8(n%16))
	}
	//	if hexArray[0] == 0 || hexArray[0] == 1 {
	//		return false
	//	}
	if hexArray[0] == 2 || hexArray[0] == 3 {
		return true
	}
	return false
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func (node *Node) hash_node() string {
	var str string
	switch node.Node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.Branch_value {
			str += v
		}
	case 2:
		str = node.Flag_value.Value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func getHexArray(key string) []uint8 {
	rawArray := []uint8(key)
	n := len(rawArray)
	var hexArray []uint8
	for i := 0; i < n; i++ {
		num := rawArray[i]
		n1 := uint8(num / 16)
		n2 := uint8(num % 16)
		hexArray = append(hexArray, n1, n2)
	}
	return append(hexArray, 16)
}

func (node *Node) String() string {
	str := "empty string"
	switch node.Node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.Branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.Branch_value[16])
	case 2:
		Encoded_prefix := node.Flag_value.Encoded_prefix
		node_name := "Leaf"
		if is_ext_node(Encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(Encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.Flag_value.Value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.Db = make(map[string]Node)
	mpt.Root = ""
	mpt.InsertedRecord = make(map[string]string)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.Db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.Db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}
