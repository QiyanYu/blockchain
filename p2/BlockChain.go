package p2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/crypto/sha3"
)

//BlockChain struct
type BlockChain struct {
	Chain  map[int32][]Block `json:"chain"`
	Length int32             `json:"length"`
}

//NewBlockChain called in SyncBlockChain
func NewBlockChain() BlockChain {
	var bc BlockChain
	bc.BlockChainInitial()
	return bc
}

// Show to help demonstrate the BlockChain
func (blockChain *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range blockChain.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range blockChain.Chain[int32(id)] {
			hashs = append(hashs, block.HeaderValue.Hash+"<="+block.HeaderValue.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

//Get takes a height as the argument, return the list of blocks
func (blockChain *BlockChain) Get(height int32) ([]Block, bool) {
	// currentHeight := len(blockChain.Chain)
	if blockChain.Length < height {
		return nil, false
	}
	return blockChain.Chain[height], true
}

//GetLatestBlocks return the list of blocks of height "BlockChain.Length"
func (blockChain *BlockChain) GetLatestBlocks() []Block {
	return blockChain.Chain[blockChain.Length]
}

//GetParentBlock takes a block as the parameter, and returns its parent block
func (blockChain *BlockChain) GetParentBlock(block Block) (Block, bool) {
	parentHash := block.HeaderValue.ParentHash
	parentHeight := block.HeaderValue.Height - 1
	blockList, isExisted := blockChain.Get(parentHeight)
	if isExisted {
		for _, block := range blockList {
			if block.HeaderValue.Hash == parentHash {
				return block, true
			}
		}
	}

	return Block{}, false
}

//Insert takes a block as the argument, insert it into blockchain
func (blockChain *BlockChain) Insert(block *Block) {
	height := block.HeaderValue.Height
	hashValue := block.HeaderValue.Hash
	if blockChain.Length > 0 {
		for i := range blockChain.Chain[height] {
			if blockChain.Chain[height][i].HeaderValue.Hash == hashValue {
				return
			}
		}
	}
	blockChain.Chain[height] = append(blockChain.Chain[height], *block)
	if height > blockChain.Length {
		blockChain.Length = height
	}
}

//BlockChainEncodeToJSON iterates over all the blocks, generate blocks JSONString
func (blockChain *BlockChain) BlockChainEncodeToJSON() (string, error) {
	var sb strings.Builder
	sb.WriteString("[")
	blockIndex := 1
	for _, value := range blockChain.Chain {
		for i := range value {
			if blockIndex != 1 {
				sb.WriteString(",")
			}
			blockJSONStr := value[i].BlockEncodeToJSON()
			sb.WriteString(blockJSONStr)
			blockIndex++
		}
	}
	sb.WriteString("]")
	return sb.String(), nil
}

//BlockChainDecodeFromJSON takes JSON string as input, get block instance back and insert into the blockchain
func (blockChain *BlockChain) BlockChainDecodeFromJSON(JSONString string) error {
	var arr []map[string]interface{}
	err := json.Unmarshal([]byte(JSONString), &arr)
	for i := range arr {
		block := Block{}
		block.HeaderValue.Height = int32(arr[i]["height"].(float64))
		block.HeaderValue.Hash = arr[i]["hash"].(string)
		block.HeaderValue.ParentHash = arr[i]["parentHash"].(string)
		block.HeaderValue.Size = int32(arr[i]["size"].(float64))
		block.HeaderValue.Timestamp = int64(arr[i]["timeStamp"].(float64))
		block.HeaderValue.Nonce = arr[i]["nonce"].(string)
		mptValue := arr[i]["mpt"].(map[string]interface{})
		insertMpt(&block, mptValue)
		blockChain.Insert(&block)
	}
	return err
}

//BlockChainInitial initial the blockchain
func (blockChain *BlockChain) BlockChainInitial() {
	blockChain.Chain = make(map[int32][]Block)
}
