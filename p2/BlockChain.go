package p2

import (
	"encoding/json"
	"fmt"
	"strings"
)

//BlockChain struct
type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
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

//EncodeToJSON iterates over all the blocks, generate blocks JSONString
func (blockChain *BlockChain) EncodeToJSON() string {
	var sb strings.Builder
	sb.WriteString("[")
	blockIndex := 1
	for _, value := range blockChain.Chain {
		if blockIndex != 1 {
			sb.WriteString(",")
		}
		for i := range value {
			blockJSONStr := value[i].EncodeToJSON()
			sb.WriteString(blockJSONStr)
			blockIndex++
		}
	}
	sb.WriteString("]")
	return sb.String()
}

//DecodeFromJSON takes JSON string as input, get block instance back and insert into the blockchain
func (blockChain *BlockChain) DecodeFromJSON(JSONString string) {
	var blocks []Block
	json.Unmarshal([]byte(JSONString), &blocks)
	fmt.Println(blocks)
}

func (blockChain *BlockChain) Initial() {
	blockChain.Chain = make(map[int32][]Block)
}
