package lib

import (
	"crypto/md5"
	"log"
	"math/big"
	"sort"
)

/*
Contains helper functions for APIs
*/

func HashMD5(s string) int {
	hasher := md5.New()
	hasher.Write([]byte(s))
	hashedBigInt := big.NewInt(0)
	hashedBigInt.SetBytes(hasher.Sum(nil))

	maxRingPosBigInt := big.NewInt(100)
	ringPosBigInt := big.NewInt(0)
	return int(ringPosBigInt.Mod(hashedBigInt, maxRingPosBigInt).Uint64())
}

func (n *Node) GetResponsibleNodes(keyPos int) [REPLICATION_FACTOR]NodeData {
	posArr := []int{}

	for pos := range n.NodeMap {
		posArr = append(posArr, pos)
	}

	sort.Ints(posArr)
	log.Printf("Key position: %d\n", keyPos)
	firstNodePosIndex := -1
	for i, pos := range posArr {
		// if the keyPos is 8, node 0 is at pos 0 and node 1 is at pos 12, the first node index should be 1
		if keyPos <= pos {
			log.Printf("First node position: %d\n", pos)
			firstNodePosIndex = i
			break
		}
	}
	if firstNodePosIndex == -1 {
		firstNodePosIndex = 0
	}

	responsibleNodes := [REPLICATION_FACTOR]NodeData{}
	for i := 0; i < REPLICATION_FACTOR; i++ {
		responsibleNodes[i] = n.NodeMap[posArr[(firstNodePosIndex+i)%len(posArr)]]
	}
	return responsibleNodes
}

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

func (o *ClientCart) IsEqual(b ClientCart) bool {
	if o.UserID != b.UserID {
		return false
	}
	// if o.Items != b.Items {
	// 	return false
	// }
	if !OrderedIntArrayEqual(o.VectorClock, b.VectorClock) {
		return false
	}
	return true
}
