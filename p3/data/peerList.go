package data

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

//PeerList peerlist struct store the information of peers
type PeerList struct {
	selfID    int32
	peerMap   map[string]int32
	maxLength int32
	mux       sync.Mutex
}

//NewPeerList return peerList struct
func NewPeerList(id int32, maxLength int32) PeerList {
	initialPeerMap := make(map[string]int32)
	peerList := PeerList{selfID: id, peerMap: initialPeerMap, maxLength: maxLength}
	return peerList
}

//Add add peer into peerlist
func (peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	peers.peerMap[addr] = id
	peers.mux.Unlock()
}

//Delete delete peer in peerlist
func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.peerMap, addr)
	peers.mux.Unlock()
}

//Rebalance rebalance the peerlist
func (peers *PeerList) Rebalance() {
	peers.mux.Lock()
	maxLen := int(peers.maxLength)
	nowLen := len(peers.peerMap)
	if nowLen > maxLen {
		tempMap := make(map[int32]string)
		r := ring.New(nowLen + 1)
		selfIDNumber := peers.selfID
		var tempSlice []int32
		tempSlice = append(tempSlice, selfIDNumber)
		for addr, id := range peers.peerMap {
			tempSlice = append(tempSlice, id)
			tempMap[id] = addr
		}
		sort.Slice(tempSlice, func(i, j int) bool { return tempSlice[i] < tempSlice[j] })
		var index int
		for i, v := range tempSlice {
			if selfIDNumber == v {
				index = i
			}
			r.Value = v
			r = r.Next()
		}
		r = r.Move(index + maxLen/2 + 1)
		for i := 0; i < maxLen+1; i++ {
			x, _ := r.Prev().Value.(int32)
			r = r.Prev()
			delete(tempMap, x)
		}
		for _, v := range tempMap {
			delete(peers.peerMap, v)
		}
	}
	peers.mux.Unlock()
}

//Show help to demonstrate the peerlist
func (peers *PeerList) Show() string {
	rs := ""
	rs += fmt.Sprintf("Self Id: %v, Max length: %v", peers.selfID, peers.maxLength)
	rs += "\n"
	for k, v := range peers.peerMap {
		rs += fmt.Sprintf("%v: %v", k, v)
		rs += "\n"
	}
	return rs

}

//Register register
func (peers *PeerList) Register(id int32) {
	peers.selfID = id
	fmt.Printf("SelfId=%v\n", id)
}

//Copy copy the peers map
func (peers *PeerList) Copy() map[string]int32 {
	copyPeersMap := make(map[string]int32)
	for k, v := range peers.peerMap {
		copyPeersMap[k] = v
	}
	return copyPeersMap
}

//GetSelfID return the self ID
func (peers *PeerList) GetSelfID() int32 {
	return peers.selfID
}

//PeerMapToJSON marshal peermap to JSON
func (peers *PeerList) PeerMapToJSON() (string, error) {
	str, err := json.Marshal(peers.peerMap)
	if err != nil {
		panic(err)
	}
	return string(str), err
}

//InjectPeerMapJSON when receive peerMap JSON, add peers into own peersMap
func (peers *PeerList) InjectPeerMapJSON(peerMapJSONStr string, selfAddr string) {
	var injectPeerMap map[string]int32
	if err := json.Unmarshal([]byte(peerMapJSONStr), &injectPeerMap); err != nil {
		panic(err)
	}
	for k, v := range injectPeerMap {
		if k != selfAddr {
			peers.Add(k, v)
		}
	}
}

//TestPeerListRebalance test
func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
