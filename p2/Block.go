package p2

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/sha3"

	"../p1"
)

//Header  block header struct
type Header struct {
	Height     int32  `json:"height"`
	Timestamp  int64  `json:"timeStamp"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parentHash"`
	Size       int32  `json:"size"`
	Nonce      string `json:"nonce"`
}

// Block struct
type Block struct {
	HeaderValue Header                `json:"header"`
	Value       p1.MerklePatriciaTrie `json:"mpt"`
}

//BlockInitial the block
func (block *Block) BlockInitial(height int32, parentHash string, value *p1.MerklePatriciaTrie, nonce string) {
	block.HeaderValue.Timestamp = time.Now().Unix()
	block.HeaderValue.Height = height
	block.HeaderValue.ParentHash = parentHash
	block.HeaderValue.Size = getValueLen(value)
	block.HeaderValue.Nonce = nonce
	hashStr := string(height) + string(block.HeaderValue.Timestamp) + parentHash + block.Value.Root + string(block.HeaderValue.Size)
	fmt.Printf("hashStr: %s\n", hashStr)
	fmt.Printf("timeStamp: %v\n", block.HeaderValue.Timestamp)
	sum := sha3.Sum256([]byte(hashStr))
	block.HeaderValue.Hash = hex.EncodeToString(sum[:])
	block.Value = *value
}

func getValueLen(value *p1.MerklePatriciaTrie) int32 {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		panic(err)
	}
	return int32(len(buf.Bytes()))
}

//BlockDecodeFromJSON takes a sting that represents the JSON value of a block as an input, decodes back to a block instance
func BlockDecodeFromJSON(jsonString string) Block {
	block := Block{}
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &dat); err != nil {
		panic(err)
	}
	block.HeaderValue.Height = int32(dat["height"].(float64))
	block.HeaderValue.Hash = dat["hash"].(string)
	block.HeaderValue.ParentHash = dat["parentHash"].(string)
	block.HeaderValue.Size = int32(dat["size"].(float64))
	block.HeaderValue.Timestamp = int64(dat["timeStamp"].(float64))
	block.HeaderValue.Nonce = dat["nonce"].(string)
	mptValue := dat["mpt"].(map[string]interface{})
	insertMpt(&block, mptValue)

	return block
}

func insertMpt(block *Block, mptValue map[string]interface{}) {
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	for key, value := range mptValue {
		strValue := value.(string)
		mpt.Insert(key, strValue)
	}
	block.Value = mpt
}

//BlockEncodeToJSON encodes a block instance into a JSON format string
func (block *Block) BlockEncodeToJSON() string {
	insertedRecord := block.Value.InsertedRecord
	cacheContent := make(map[string]interface{})
	cacheContent["height"] = block.HeaderValue.Height
	cacheContent["timeStamp"] = block.HeaderValue.Timestamp
	cacheContent["hash"] = block.HeaderValue.Hash
	cacheContent["parentHash"] = block.HeaderValue.ParentHash
	cacheContent["size"] = block.HeaderValue.Size
	cacheContent["nonce"] = block.HeaderValue.Nonce
	cacheContent["mpt"] = insertedRecord
	str, err := json.Marshal(cacheContent)
	if err != nil {
		panic(err)
	}
	return string(str)
}

//BlockString return the block string version
func (block *Block) BlockString() string {
	rs := fmt.Sprintf("height: %v, timestamp: %v, hash: %s, parentHash: %s, size: %v \n",
		block.HeaderValue.Height, block.HeaderValue.Timestamp, block.HeaderValue.Hash,
		block.HeaderValue.ParentHash, block.HeaderValue.Size)
	return rs

}
