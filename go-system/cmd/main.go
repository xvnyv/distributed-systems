package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	lib "github.com/distributed-systems/go-system/internal/lib"
)

func main() {
	/* go run main.go -id=<id> -port=<port number> -first=<is-first-node> */
	/* -id is set to 0-4, port number ranges from 8000-8004, and -first is true only for the first node indicated */

	idFlagPtr := flag.Int("id", -1, "Node ID")
	portFlagPtr := flag.Int("port", 8000, "Port number")
	firstNodeFlagPtr := flag.Bool("first", false, "Is first node in system?")

	flag.Parse()

	if *idFlagPtr == -1 {
		log.Fatal("Node ID must be specified")
	}

	var badgerLock sync.Mutex
	node := lib.Node{Id: *idFlagPtr, Ip: fmt.Sprintf("%s:%d", lib.BASE_URL, *portFlagPtr), Port: *portFlagPtr, BadgerLock: &badgerLock}
	log.Printf("Node %d joining system\n", node.Id)
	node.JoinSystem(*firstNodeFlagPtr)

	log.Printf("Node %d started\n", node.Id)
	go node.HandleRequests()

	fmt.Println("Press Enter to end")
	fmt.Scanln()
}
