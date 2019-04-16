package data

import (
	"sync"

	"../../p1"
	"../../p2"
)

//SyncBlockChain struct
type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

//NewBlockChain initial syncBlockChain
func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}

//Get cast the blockchain Get
func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Get(height)
}

//GetBlock get specific block
func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	blockList, isExisted := sbc.Get(height)
	if isExisted {
		for _, block := range blockList {
			if block.HeaderValue.Hash == hash {
				return block, true
			}
		}
	}
	return p2.Block{}, false
}

//Insert cast the blockchain Insert
func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(&block)
	sbc.mux.Unlock()
}

//CheckParentHash check whether existed parent block
func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	parentHash := insertBlock.HeaderValue.ParentHash
	parentHeight := insertBlock.HeaderValue.Height - 1
	_, isExisted := sbc.GetBlock(parentHeight, parentHash)
	return isExisted
}

//UpdateEntireBlockChain update own blockchain
func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJSON string) {
	sbc.mux.Lock()
	err := sbc.bc.BlockChainDecodeFromJSON(blockChainJSON)
	if err != nil {
		panic(err)
	}
	sbc.mux.Unlock()
}

//BlockChainToJSON cast blockchain BlockChainEncodeToJSON
func (sbc *SyncBlockChain) BlockChainToJSON() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.BlockChainEncodeToJSON()
}

//GenBlock return a new block
func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie, nonce string) p2.Block {
	sbc.mux.Lock()
	// fmt.Println("first map 11111111111111111: ", sbc.bc.Chain)
	parentHeight := sbc.bc.Length
	height := parentHeight + 1
	parentHashStr := sbc.bc.Chain[parentHeight][0].HeaderValue.Hash
	newBlock := p2.Block{}
	newBlock.BlockInitial(height, parentHashStr, &mpt, nonce)
	sbc.bc.Insert(&newBlock)
	// fmt.Println("first map 22222222222222222: ", sbc.bc.Chain)
	defer sbc.mux.Unlock()
	return newBlock
}

//Show help to demonstrate syncBlockChain
func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}

//GetLatestBlocks sync version
func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

//GetParentBlock sync version
func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(block)
}
