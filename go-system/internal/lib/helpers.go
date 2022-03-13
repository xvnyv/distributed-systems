package lib

import (
	"crypto/md5"
	"fmt"
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
func (ring *Ring) AllocateKey(key string) (NodeData, string) {
	// nodeMap := ringServer.Ring.RingNodeDataMap
	hashKey := HashMD5(key)
	fmt.Printf("this is the hash below: \n")
	fmt.Println(hashKey)

	nodeId := hashKey % 10 //based on the hash generated, we will modulo it to find out which node will take responsibility.

	// to do: replication is not accounted for, need to send to other nodes also in case node down.
	return ring.NodeDataMap[nodeId], strconv.Itoa(hashKey)

func OrderedIntArrayEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func UnorderedIntArrayEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	exists := make(map[int]bool)
	for _, value := range a {
		exists[value] = true
	}
	for _, value := range b {
		if !exists[value] {
			return false
		}
	}
	return true

}

func UnorderedStringArrayEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	exists := make(map[string]bool)
	for _, value := range a {
		exists[value] = true
	}
	for _, value := range b {
		if !exists[value] {
			return false
		}
	}
	return true

}

func (o *DataObject) IsEqual(b DataObject) bool {
	if o.Key != b.Key {
		return false
	}
	if o.Value != b.Value {
		return false
	}
	if !OrderedIntArrayEqual(o.VectorClock, b.VectorClock) {
		return false
	}
	return true
}
