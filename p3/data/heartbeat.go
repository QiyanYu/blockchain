package data

//HeartBeatData heatbeat data strut
type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	ID          int32  `json:"id"`
	BlockJSON   string `json:"blockJson"`
	PeerMapJSON string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

//NewHeartBeatData initial the heartbeat data
func NewHeartBeatData(ifNewBlock bool, id int32, blockJSON string, peerMapJSON string, addr string) HeartBeatData {
	heartBeatData := HeartBeatData{IfNewBlock: ifNewBlock, ID: id, BlockJSON: blockJSON, PeerMapJSON: peerMapJSON, Addr: addr, Hops: 3}
	return heartBeatData
}

// //PrepareHeartBeatData first create a new instance of HeartBeatData, then decide whether or not you will create a new block and send the new block to other peers
// func PrepareHeartBeatData(sbc *SyncBlockChain, selfID int32, peerMap string, addr string) HeartBeatData {
// 	// isNewBlock := randBool()
// 	var heartbeatData HeartBeatData
// 	// if isNewBlock {
// 	// 	rand.Seed(time.Now().UTC().UnixNano())
// 	// 	mpt := p1.MerklePatriciaTrie{Root: "new block", Db: make(map[string]p1.Node), InsertedRecord: make(map[string]string)}
// 	// 	block := sbc.GenBlock(mpt)
// 	// 	blockJSON := block.BlockEncodeToJSON()
// 	// 	heartbeatData = NewHeartBeatData(true, selfID, blockJSON, peerMap, addr)
// 	// } else {
// 	heartbeatData = NewHeartBeatData(false, selfID, "", peerMap, addr)
// 	// }
// 	return heartbeatData
// }

// func randBool() bool {
// 	return rand.Float32() < 0.5
// }
