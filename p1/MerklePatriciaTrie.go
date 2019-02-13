//package p1
package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Flag_value struct {
	encoded_prefix []uint8
	value          string
}

type Node struct {
	node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	db   map[string]Node
	root string
}

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	if key == "" {
		return "", nil
	}
	var path = getHexArray(key)
	var value = mpt.getHelper(mpt.db[mpt.root], path)
	if value == "" {
		return "", errors.New("path_not_found")
	} else {
		return value, nil
	}
}

func (mpt *MerklePatriciaTrie) getHelper(node Node, path []uint8) string {
	var nodeType = node.node_type
	if nodeType == 0 {
		return ""
	} else if nodeType == 1 {
		if getBranchCommonPath(node.branch_value, path) {
			if path[0] == uint8(16) {
				return node.branch_value[16]
			} else {
				return mpt.getHelper(mpt.db[node.branch_value[path[0]]], path[1:])
			}
		} else {
			return ""
		}
	} else if nodeType == 2 {
		var encodeValue = node.flag_value.encoded_prefix
		var decodeValue = compact_decode(encodeValue)
		var nodeValue = node.flag_value.value
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
		if cpLen != 0 && rpLen == 0 && rnLen == 0 {
			if isLeaf {
				return nodeValue
			} else {
				var nextNode = mpt.db[nodeValue]
				return mpt.getHelper(nextNode, restPath)
			}
		} else {
			return ""
		}
	}
	return ""
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	if mpt.root == "" {
		mpt.db = make(map[string]Node)
		var rootNode Node
		rootNode.node_type = 0
		mpt.db[rootNode.hash_node()] = rootNode
		mpt.root = rootNode.hash_node()
	}
	var path = getHexArray(key)
	var rootNode = mpt.db[mpt.root]
	mpt.root = mpt.insertHelper(rootNode, path, new_value)
}
func (mpt *MerklePatriciaTrie) insertHelper(node Node, path []uint8, value string) string {
	var nodeType = node.node_type
	var nodeKey = node.hash_node()
	if nodeType == 0 { //insert into Null
		var rootNode Node
		rootNode.node_type = 2
		rootNode.flag_value.encoded_prefix = compact_encode(path)
		rootNode.flag_value.value = value
		var hashValue = rootNode.hash_node()
		mpt.db[hashValue] = rootNode
		delete(mpt.db, nodeKey)
		return hashValue
	} else if nodeType == 1 { //insert into Branch Node
		if path[0] == uint8(16) { //if insert into branch node value, just update the value
			node.branch_value[16] = value
			mpt.db[node.hash_node()] = node
			delete(mpt.db, nodeKey)
			return node.hash_node()
		}
		if getBranchCommonPath(node.branch_value, path) { //exist common path
			var commonPath = path[0]
			var nextNode = mpt.db[node.branch_value[commonPath]]
			node.branch_value[commonPath] = mpt.insertHelper(nextNode, path[1:], value)
			var nodeHashValue = node.hash_node()
			mpt.db[nodeHashValue] = node
			delete(mpt.db, nodeKey)
			return nodeHashValue
		} else { // don't exist common path
			var restPath = path[0]
			var newLeafNode Node
			newLeafNode.node_type = 2
			newLeafNode.flag_value.value = value
			newLeafNode.flag_value.encoded_prefix = compact_encode(path[1:])
			mpt.db[newLeafNode.hash_node()] = newLeafNode
			node.branch_value[restPath] = newLeafNode.hash_node()
			var nodeHashValue = node.hash_node()
			mpt.db[nodeHashValue] = node
			delete(mpt.db, nodeKey)
			return nodeHashValue
		}
	} else if nodeType == 2 { //insert into extension node or leaf node
		var encodeValue = node.flag_value.encoded_prefix
		var nodeValue = node.flag_value.value
		var isLeaf = encodeValue[0] == uint8(2) || encodeValue[0] == uint8(3)
		if isLeaf { //insert into leaf node
			var nodePath = append(compact_decode(encodeValue), uint8(16)) //since it is the leaf node, add 16 back
			var commonPath = getExtLeafCommonPath(nodePath, path)
			var restPath = getRestPath(path, commonPath)
			var restNibble = getRestNibble(nodePath, commonPath)
			var cpLen = len(commonPath)
			var rpLen = len(restPath)
			var rnLen = len(restNibble)
			if cpLen != 0 && rpLen == 0 && rnLen == 0 { //update the leaf node value
				nodeValue = value
				mpt.db[node.hash_node()] = node
				delete(mpt.db, nodeKey)
				return node.hash_node()
			}
			if cpLen != 0 && rpLen != 0 && rnLen != 0 { //has common path so 1)new extension node 2) new branch node 3)insert these two nodes
				var newExtNode Node
				newExtNode.node_type = 2
				newExtNode.flag_value.encoded_prefix = compact_encode(commonPath)
				var newBranchNode Node
				mpt.insertHelper(newBranchNode, restPath, value)
				newExtNode.flag_value.value = mpt.insertHelper(newBranchNode, restNibble, nodeValue)
				mpt.db[newExtNode.hash_node()] = newExtNode
				delete(mpt.db, nodeKey)
				return newExtNode.hash_node()
			}
			if cpLen == 0 && rpLen != 0 && rnLen != 0 { //doesn't have common path so 1)new branch node 2) insert two nodes into branch node
				var newBranchNode Node
				newBranchNode.node_type = 1
				newBranchNode.branch_value[restPath[0]] = mpt.insertHelper(newBranchNode, restPath, value)
				newBranchNode.branch_value[restNibble[0]] = mpt.insertHelper(newBranchNode, restNibble, nodeValue)
				mpt.db[newBranchNode.hash_node()] = newBranchNode
				delete(mpt.db, nodeKey)
				return newBranchNode.hash_node()
			}
		} else { //insert into extension node
			var nodePath = compact_decode(encodeValue)
			var commonPath = getExtLeafCommonPath(nodePath, path)
			var restPath = getRestPath(path, commonPath)
			var restNibble = getRestNibble(path, commonPath)
			var cpLen = len(commonPath)
			var rpLen = len(restPath)
			var rnLen = len(restNibble)
			if cpLen != 0 && rpLen != 0 && rnLen != 0 { // 1）new extension node 2)new branch node 3)insert two paths into branch node
				var newExtNode Node
				newExtNode.node_type = 2
				newExtNode.flag_value.encoded_prefix = compact_encode(commonPath)
				var newBranchNode Node
				newBranchNode.node_type = 1
				mpt.insertHelper(newBranchNode, restPath, value)
				newExtNode.flag_value.value = mpt.insertHelper(newBranchNode, restNibble, nodeValue)
				mpt.db[newExtNode.hash_node()] = newExtNode
				delete(mpt.db, nodeKey)
				return newExtNode.hash_node()
			} else if cpLen != 0 && rpLen != 0 && rnLen != 0 { //directly insert rest path into next node
				var nextNode = mpt.db[nodeValue]
				nodeValue = mpt.insertHelper(nextNode, restPath, value)
				mpt.db[node.hash_node()] = node
				delete(mpt.db, nodeValue)
				return node.hash_node()
			} else if cpLen == 0 && rpLen != 0 && rnLen == 1 { //1)new branch 2)insert branch
				var newBranchNode Node
				newBranchNode.node_type = 1
				if rpLen == 1 { //insert 16
					newBranchNode.branch_value[16] = value
				} else {
					var newLeafNode Node
					newLeafNode.node_type = 2
					newLeafNode.flag_value.value = value
					newLeafNode.flag_value.encoded_prefix = compact_encode(restPath[1:])
					newBranchNode.branch_value[restPath[0]] = newLeafNode.hash_node()
					mpt.db[newLeafNode.hash_node()] = newLeafNode
				}
				newBranchNode.branch_value[restNibble[0]] = nodeValue
				mpt.db[newBranchNode.hash_node()] = newBranchNode
				delete(mpt.db, nodeKey)
				return newBranchNode.hash_node()
			} else if cpLen == 0 && rpLen != 0 && rnLen > 1 { //1)new branch node 2)new extension node
				var newBranchNode Node
				newBranchNode.node_type = 1
				if rpLen == 1 { //insert 16
					newBranchNode.branch_value[16] = value
				} else {
					var newLeafNode Node
					newLeafNode.node_type = 2
					newLeafNode.flag_value.value = value
					newLeafNode.flag_value.encoded_prefix = compact_encode(restPath[1:])
					newBranchNode.branch_value[restPath[0]] = newLeafNode.hash_node()
					mpt.db[newLeafNode.hash_node()] = newLeafNode
				}
				var newExtNode Node
				newExtNode.node_type = 1
				newExtNode.flag_value.encoded_prefix = compact_encode(restNibble[1:])
				newExtNode.flag_value.value = nodeValue
				mpt.db[newExtNode.hash_node()] = newExtNode
				newBranchNode.branch_value[restNibble[0]] = newExtNode.hash_node()
				mpt.db[newBranchNode.hash_node()] = newBranchNode
				delete(mpt.db, nodeKey)
				return newBranchNode.hash_node()
			}
		}
	}
	return ""
}

func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	// TODO
	var path = getHexArray(key)
	var isSuc bool
	isSuc, _ = mpt.deleteHelper(mpt.db[mpt.root], path)
	if isSuc {
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

func (mpt *MerklePatriciaTrie) deleteHelper(node Node, path []uint8) (bool, string) {
	var nodeType = node.node_type
	var nodeKey = node.hash_node()
	if nodeType == 0 { // delete at Null node
		return false, ""
	} else if nodeType == 1 { // delete at Branch node
		if getBranchCommonPath(node.branch_value, path) {
			if path[0] == uint8(16) { // delete branch node value at 16
				node.branch_value[16] = ""
			} else { // has common path but not delete the branch value, into recursion
				var isSuc bool
				var nextKey string
				isSuc, nextKey = mpt.deleteHelper(mpt.db[node.branch_value[path[0]]], path[1:])
				if isSuc { //delete successfully
					node.branch_value[path[0]] = nextKey
				} else { //not found
					return false, ""
				}
			}
			//check if there is only one value in branch node remaining
			if len(getArrayInBranchValue(node.branch_value)) > 1 { //do not need to balance
				mpt.db[node.hash_node()] = node
				delete(mpt.db, nodeKey)
				return true, node.hash_node()
			} else if getArrayInBranchValue(node.branch_value)[0] == 16 { // it is a leaf node
				var returnNode Node
				returnNode.node_type = 2
				returnNode.flag_value.value = node.branch_value[16]
				returnNode.flag_value.encoded_prefix = compact_encode([]uint8{uint8(16)})
				mpt.db[returnNode.hash_node()] = returnNode
				delete(mpt.db, nodeKey)
				return true, returnNode.hash_node()
			} else {
				var returnNode Node
				var branchIndex = getArrayInBranchValue(node.branch_value)[0]
				var nextNodeKey = node.branch_value[branchIndex]
				var nextNode = mpt.db[nextNodeKey]
				var nextNodeType = nextNode.node_type
				if nextNodeType == 2 { // if next node is extension or leaf node, combine them
					returnNode.node_type = 2
					returnNode.flag_value.value = nextNode.flag_value.value
					var nextNodeEncodeValue = nextNode.flag_value.encoded_prefix
					var newValue = append([]uint8{uint8(branchIndex)}, compact_decode(nextNode.flag_value.encoded_prefix)...)
					if nextNodeEncodeValue[0] == uint8(2) || nextNodeEncodeValue[0] == uint8(3) {
						newValue = append(newValue, uint8(16))
					}
					returnNode.flag_value.encoded_prefix = compact_encode(newValue)
					mpt.db[returnNode.hash_node()] = returnNode
					delete(mpt.db, nodeKey)
					delete(mpt.db, nextNodeKey)
					return true, returnNode.hash_node()
				} else { // if next node is branch node, return it as extension node
					returnNode.node_type = 2
					returnNode.flag_value.value = nextNodeKey
					returnNode.flag_value.encoded_prefix = compact_encode([]uint8{uint8(branchIndex)})
					mpt.db[returnNode.hash_node()] = returnNode
					delete(mpt.db, nodeKey)
					return true, returnNode.hash_node()
				}
			}
		} else { // not found
			return false, ""
		}
	} else if nodeType == 2 { //delete at leaf or extension node
		var encodeValue = node.flag_value.encoded_prefix
		var nodeValue = node.flag_value.value
		var isLeaf = encodeValue[0] == uint8(2) || encodeValue[0] == uint8(3)
		if isLeaf {
			var nodePath = append(compact_decode(encodeValue), uint8(16)) //since it is the leaf node, add 16 back
			var commonPath = getExtLeafCommonPath(nodePath, path)
			var restPath = getRestPath(path, commonPath)
			var restNibble = getRestNibble(nodePath, commonPath)
			var cpLen = len(commonPath)
			var rpLen = len(restPath)
			var rnLen = len(restNibble)
			if cpLen != 0 && rpLen == 0 && rnLen == 0 {
				delete(mpt.db, nodeKey)
				return true, ""
			} else {
				return false, ""
			}
		} else { // exstension node
			var nodePath = compact_decode(encodeValue)
			var commonPath = getExtLeafCommonPath(nodePath, path)
			var restPath = getRestPath(path, commonPath)
			var restNibble = getRestNibble(path, commonPath)
			var cpLen = len(commonPath)
			var rpLen = len(restPath)
			var rnLen = len(restNibble)
			if cpLen != 0 && rpLen == 0 && rnLen == 0 {
				var isSuc bool
				var nextKey string
				isSuc, nextKey = mpt.deleteHelper(mpt.db[nodeValue], restPath)
				if isSuc {
					var nextReturnNode = mpt.db[nextKey]
					var nextReturnNodeType = nextReturnNode.node_type
					if nextReturnNodeType == 2 { //combine the return node and this extension node
						var newNode Node
						newNode.node_type = 2
						var nextReturnNodeEncodeValue = nextReturnNode.flag_value.encoded_prefix
						var newValue = append(compact_decode(encodeValue), compact_decode(nextReturnNodeEncodeValue)...)
						if nextReturnNodeEncodeValue[0] == uint8(2) || nextReturnNodeEncodeValue[0] == uint8(3) {
							newValue = append(newValue, uint8(16))
						}
						var newEncodeValue = compact_encode(newValue)
						newNode.flag_value.encoded_prefix = newEncodeValue
						newNode.flag_value.value = nextReturnNode.flag_value.value
						mpt.db[newNode.hash_node()] = newNode
						delete(mpt.db, nodeKey)
						return true, newNode.hash_node()
					} else { // just connect the next node
						node.flag_value.value = nextKey
						mpt.db[node.hash_node()] = node
						delete(mpt.db, nodeKey)
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
	}
	if hexArray[0] == 0 || hexArray[0] == 2 {
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
	if hexArray[0] == 0 || hexArray[0] == 1 {
		return false
	}
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
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func getHexArray(key string) []uint8 {
	rawArray := []uint8(key)
	n := len(rawArray)
	//fmt.Println(rawArray)
	var hexArray []uint8
	for i := 0; i < n; i++ {
		num := rawArray[i]
		n1 := uint8(num / 16)
		n2 := uint8(num % 16)
		hexArray = append(hexArray, n1, n2)
	}
	return append(hexArray, 16)
}

func main() {
	pa := []uint8{2, 7, 9}
	pp := []uint8{2}
	//fmt.Println(compact_encode(pa)[0] == 7)

	//var s [17]string
	//s[2] = "adf"
	//var ss uint8
	//ss = 3
	fmt.Println(getRestPath(pa, pp))
	//fmt.Println(compact_decode(compact_encode(pa))[0] == 7)
	//fmt.Println(compact_decode(pa))
	//fmt.Println(getHexArray("do"))
	//fmt.Println(compact_encode(getHexArray("do")))

}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.db = make(map[string]Node)
	mpt.root = ""
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.root)
	for hash := range mpt.db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.db[hash]))
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
