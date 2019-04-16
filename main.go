package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"./p3"
)

func main() {
	router := p3.NewRouter()
	if len(os.Args) > 1 {
		args1 := os.Args[1]
		id, err := strconv.Atoi(args1)
		if err != nil {
			panic(err)
		}
		firstNodeID, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}
		p3.SelfID = int32(id)
		p3.FirstNodeID = int32(firstNodeID)
		log.Fatal(http.ListenAndServe(":"+args1, router))
	} else {
		log.Fatal(http.ListenAndServe(":6686", router))
	}
}
