package p2

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

func (blockChain *BlockChain) Get(height int32) []Block {
	return blockChain.Chain[height]
}

func (blockChain *BlockChain) Insert(block Block) {
	height := block.HeaderValue.Height
	hashValue := block.HeaderValue.Hash
	if len(blockChain.Chain[height]) > 0 {
		for i := range blockChain.Chain[height] {
			if blockChain.Chain[height][i].HeaderValue.Hash == hashValue {
				return
			}
		}
	}
	blockChain.Chain[height] = append(blockChain.Chain[height], block)
}

func EncodeToJSON() {

}
