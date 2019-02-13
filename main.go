package main

import (
	"fmt"
)

func main() {
	/*mpt := new(p1.MerklePatriciaTrie)
	mpt.Insert("a", "apple")
	mpt.Insert("ab", "apple")*/
	var hexArray []uint8
	hexArray = []uint8{1, 2}
	hexArray = hexArray[2:]
	fmt.Println(hexArray)
}
