package main

import (
	"fmt"

	"./p2"
)

func main() {
	// mpt := new(p1.MerklePatriciaTrie)
	// mpt.Initial()
	// mpt.Insert("hello", "world")
	// mpt.Insert("charles", "ge")
	// //fmt.Println(mpt.Order_nodes())
	// b1 := new(p2.Block)
	// b1.Initial(1, "genesis", mpt)
	// fmt.Println(b1.EncodeToJSON())
	// mpt.Insert("ab", "new")
	// fmt.Println(mpt.Order_nodes())
	block := p2.DecodeFromJSON(`{"hash":"11342f36e3ee819df408a6935b4255353d414398a78319fdb441be9797727fb8","height":1,"mpt":{"hello":"world"},"parentHash":"genesis","size":509,"timeStamp":1551117605}`)
	fmt.Println(block.EncodeToJSON())
}
