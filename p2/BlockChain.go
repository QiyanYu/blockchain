package p2

import (
	"encoding/json"
	"strings"
)

//BlockChain struct
type BlockChain struct {
	Chain  map[int32][]Block `json:"chain"`
	Length int32             `json:"length"`
}

//Get takes a height as the argument, return the list of blocks
func (blockChain *BlockChain) Get(height int32) []Block {
	return blockChain.Chain[height]
}

//Insert takes a block as the argument, insert it into blockchain
func (blockChain *BlockChain) Insert(block *Block) {
	height := block.HeaderValue.Height
	hashValue := block.HeaderValue.Hash
	if len(blockChain.Chain[height]) > 0 {
		for i := range blockChain.Chain[height] {
			if blockChain.Chain[height][i].HeaderValue.Hash == hashValue {
				return
			}
		}
	}
	blockChain.Chain[height] = append(blockChain.Chain[height], *block)
}

//BlockChainEncodeToJSON iterates over all the blocks, generate blocks JSONString
func (blockChain *BlockChain) BlockChainEncodeToJSON() (string, error) {
	var sb strings.Builder
	sb.WriteString("[")
	blockIndex := 1
	for _, value := range blockChain.Chain {
		if blockIndex != 1 {
			sb.WriteString(",")
		}
		for i := range value {
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

		mptValue := arr[i]["mpt"].(map[string]interface{})
		insertMpt(&block, mptValue)
		blockChain.Insert(&block)
	}
	return err
}

func (blockChain *BlockChain) BlockChainInitial() {
	blockChain.Chain = make(map[int32][]Block)
}
