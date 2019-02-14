package main

import (
	"fmt"

	"./p1"
)

func main() {
	mpt := new(p1.MerklePatriciaTrie)
	mpt.Insert("a", "apple")
	mpt.Insert("b", "banana")
	mpt.Insert("a", "new")
	fmt.Println(mpt.Order_nodes())
	// mpt.Delete("b")
	// fmt.Println(mpt.Order_nodes())
}
