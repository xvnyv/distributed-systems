package main

import (
	"flag"
	"fmt"
	"log"

	lib "github.com/distributed-systems/go-system/internal/lib"
)

func main() {
	/* go run main.go -id=<id> -port=<port number> -first=<is-first-node> */
	/* -id is set to 0-4, port number ranges from 8000-8004, and -first is true only for the first node indicated */

	idFlagPtr := flag.Int("id", -1, "Node ID")
	portFlagPtr := flag.Int("port", 8000, "Port number")
	firstNodeFlagPtr := flag.Bool("first", false, "Is first node in system?")
	// TODO: REMOVE posPtr WHEN IMPLEMENTING JOINING
	posPtr := flag.Int("pos", 0, "Position in ring")
	flag.Parse()

	if *idFlagPtr == -1 {
		log.Fatal("Node ID must be specified")
	}

	pos := -1
	if *firstNodeFlagPtr {
		pos = 0
	} else {
		// TODO: REMOVE THIS SECTION WHEN IMPLEMENTING JOINING
		pos = *posPtr
	}

	node := lib.Node{Id: *idFlagPtr, Ip: fmt.Sprintf("http://127.0.0.1:%d", *portFlagPtr), Port: *portFlagPtr, Position: pos, NodeMap: lib.TEMP_NODE_MAP}
	log.Printf("Node %d started\n", node.Id)
	go node.HandleRequests()

	fmt.Println("Press Enter to end")
	fmt.Scanln()
}
