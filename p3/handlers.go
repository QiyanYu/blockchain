package p3

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"

	"../p1"
	"../p2"
	"./data"
)

// var taServer = "http://localhost:6688"
// var registerServer = taServer + "/peer"
// var bcDownloadServer = taServer + "/upload"

//FirstNodeID first node id
var FirstNodeID int32
var firstNodeAddr string

//SelfID assign self id
var SelfID int32
var selfAddr string

//SBC self sync blockchain
var SBC data.SyncBlockChain

//Peers self peerlist
var Peers data.PeerList

var ifStarted bool

func init() {
	// This function will be executed before everything else.
	// Do some initialization here.
	rand.Seed(time.Now().UTC().UnixNano())
	SBC = data.NewBlockChain()
	Peers = data.NewPeerList(SelfID, 2)
}

//Start Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	selfAddr = "http://localhost:" + strconv.Itoa(int(SelfID))
	firstNodeAddr = "http://localhost:" + strconv.Itoa(int(FirstNodeID))
	if !ifStarted {
		Register()
		if selfAddr != firstNodeAddr { // download blockchain form first node
			Download()
			Peers.Add(firstNodeAddr, FirstNodeID)
		} else { //initial first block
			block := p2.Block{}
			mpt := p1.MerklePatriciaTrie{}
			mpt.Initial()
			mpt.Insert("initial", "first node")
			block.BlockInitial(0, "", &mpt, "")
			SBC.Insert(block)
		}
		go StartHeartBeat()
		go StartTryingNonces()
	}
	ifStarted = true
	fmt.Fprintf(w, "Start Node Id: %v", SelfID)
}

//Show Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register actually assign the self id into peerlist
func Register() {
	Peers.Register(SelfID)
}

// Download blockchain from first node
func Download() {
	resp, err := http.Get(firstNodeAddr + "/upload")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	SBC.UpdateEntireBlockChain(string(body))
}

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	blockChainJSON, err := SBC.BlockChainToJSON()
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, blockChainJSON)
}

//UploadBlock Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	height, err := strconv.Atoi(vars["height"])
	if err != nil {
		panic(err)
	}
	hash := vars["hash"]
	block, isExisted := SBC.GetBlock(int32(height), hash)
	if isExisted {
		blockJSON := block.BlockEncodeToJSON()
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, blockJSON)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

//HeartBeatReceive Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	bodyStr := string(body)
	heartbeat := data.HeartBeatData{}
	err = json.Unmarshal([]byte(bodyStr), &heartbeat)
	if err != nil {
		panic(err)
	}
	if heartbeat.Addr != selfAddr {
		Peers.Add(heartbeat.Addr, heartbeat.ID)
		Peers.InjectPeerMapJSON(heartbeat.PeerMapJSON, selfAddr)
		if heartbeat.IfNewBlock {
			blockJSON := heartbeat.BlockJSON
			block := p2.BlockDecodeFromJSON(blockJSON)
			nonce := block.HeaderValue.Nonce
			parentHash := block.HeaderValue.ParentHash
			rootHash := block.Value.Root
			if validNonce(parentHash, nonce, rootHash) {
				if !SBC.CheckParentHash(block) {
					AskForBlock(block.HeaderValue.Height-1, block.HeaderValue.ParentHash)
				}
				SBC.Insert(block)
			} else {
				return
			}
		}
		heartbeat.Hops--
		if heartbeat.Hops > 0 {
			ForwardHeartBeat(heartbeat)
		}
	}
}

//AskForBlock Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	Peers.Rebalance()
	heightStr := strconv.Itoa(int(height))
	peerMap := Peers.Copy()
	for addr := range peerMap {
		addr = addr + "/block/{" + heightStr + "}/{" + hash + "}"
		resp, err := http.Get(addr)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			bodyStr := string(body)
			block := p2.BlockDecodeFromJSON(bodyStr)

			if !SBC.CheckParentHash(block) {
				AskForBlock(block.HeaderValue.Height-1, block.HeaderValue.ParentHash)
			}
			SBC.Insert(block)
			break
		}
	}
}

//ForwardHeartBeat with the number of hop, forward the heartbeat, but don't forward back to the sender
func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	Peers.Rebalance()
	peerMap := Peers.Copy()
	heartbeatJSON, err := json.Marshal(heartBeatData)
	if err != nil {
		panic(err)
	}
	heartbeatBody := []byte(heartbeatJSON)
	for addr := range peerMap {
		heartbeatAddr := addr + "/heartbeat/receive"
		if _, err := http.Post(heartbeatAddr, "application/json", bytes.NewBuffer(heartbeatBody)); err != nil {
			Peers.Delete(addr)
		}
	}
}

//StartHeartBeat send heartbeat
func StartHeartBeat() {
	for {
		postHeartbeat(false, "")
		time.Sleep(10 * time.Second)
	}
}

//StartTryingNonces the POW part
func StartTryingNonces() {
Trying:
	for {
		lastBlockList := SBC.GetLatestBlocks()
		lastBlock := lastBlockList[0]
		lastBlockHeight := lastBlock.HeaderValue.Height
		height := lastBlockHeight + 1
		parentHash := lastBlock.HeaderValue.Hash
		mpt := p1.MerklePatriciaTrie{}
		mpt.Initial()
		mpt.Insert("new", "new block")
		mptRootHash := mpt.Root
		for {
			nonce, err := randomNonce()
			if err != nil {
				panic(err)
			}
			_, isExisted := SBC.Get(height)
			if isExisted {
				continue Trying
			}
			if validNonce(parentHash, nonce, mptRootHash) {
				block := SBC.GenBlock(mpt, nonce)
				SBC.Insert(block)
				blockJSON := block.BlockEncodeToJSON()
				fmt.Println("-----------------found", blockJSON)
				postHeartbeat(true, blockJSON)
			}
		}
	}
}

func postHeartbeat(isNew bool, blockJSON string) {
	Peers.Rebalance()
	peerMapJSON, err := Peers.PeerMapToJSON()
	if err != nil {
		panic(err)
	}
	peerMap := Peers.Copy()
	heartbeat := data.NewHeartBeatData(isNew, SelfID, blockJSON, peerMapJSON, selfAddr)
	heartbeatJSON, err := json.Marshal(heartbeat)
	if err != nil {
		panic(err)
	}
	heartbeatBody := []byte(heartbeatJSON)
	for addr := range peerMap {
		heartbeatAddr := addr + "/heartbeat/receive"
		if _, err := http.Post(heartbeatAddr, "application/json", bytes.NewBuffer(heartbeatBody)); err != nil {
			Peers.Delete(addr)
		}
	}
}

func validNonce(parentHash string, nonce string, rootHash string) bool {
	hashString := parentHash + nonce + rootHash
	result := sha3.Sum256([]byte(hashString))
	rs := hex.EncodeToString(result[:])
	return strings.HasPrefix(rs, "000000")
}

func randomNonce() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

//Canonical show canonical chain
func Canonical(w http.ResponseWriter, r *http.Request) {
	result := ""
	latestBlocks := SBC.GetLatestBlocks()
	for i, block := range latestBlocks {
		parentBlock, isExisted := SBC.GetParentBlock(block)
		result += "Chain: " + strconv.Itoa(i) + ":\n"
		result += block.BlockString()
		for isExisted {
			result += parentBlock.BlockString()
			parentBlock, isExisted = SBC.GetParentBlock(parentBlock)
		}
	}
	fmt.Fprintf(w, result)
}
