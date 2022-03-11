package lib

import (
	"fmt"
	"crypto/md5"
	"math/big"
	"strconv"
)

func HashMD5(s string) int {
	hasher := md5.New()
	hasher.Write([]byte(s))
	hashedBigInt := big.NewInt(0)
	hashedBigInt.SetBytes(hasher.Sum(nil))

	maxRingPosBigInt := big.NewInt(100)
	ringPosBigInt := big.NewInt(0)
	return int(ringPosBigInt.Mod(hashedBigInt, maxRingPosBigInt).Uint64())
}

// Function to allocate the given UserID to a node and return that nodeData and keyHash
func (n *Node) AllocateKey(key string) (NodeData, string) {
	// nodeMap := ringServer.Ring.RingNodeDataMap
	keyHash := HashMD5(key)
	fmt.Printf("this is the hash below: \n")
	fmt.Println(keyHash) 

	
	nodeId := keyHash % 10 //based on the hash generated, we will modulo it to find out which node will take responsibility.

	// to do: replication is not accounted for, need to send to other nodes also in case node down.
	return n.NodeMap[nodeId], strconv.Itoa(keyHash)
}
