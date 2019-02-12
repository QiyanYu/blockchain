package main

import (
	"fmt"

	"./p1"
)

func main() {
	mpt := new(p1.MerklePatriciaTrie)
	mpt.Insert("a", "apple")
	mpt.Insert("ab", "apple")
	fmt.Println("hello world!")
}
