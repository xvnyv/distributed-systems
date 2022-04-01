package lib

import (
	"crypto/md5"
	"log"
	"math/big"
	"sort"
	"sync"
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

func DetermineSuccess(requestType RequestType, respChannel <-chan ChannelResp, coordMutex *sync.Mutex) (bool, map[int]APIResp) {
	// As long as (REPLICATION_FACTOR - MIN_WRITE_SUCCESS + 1) nodes fail, we return an error to the client
	// It does not matter if 1 node has already successfully written to disk even if the entire operation fails
	// We let the client execute a read before retrying the write to handle that case
	// Inspired by DynamoDB: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Programming.Errors.html

	successResps := map[int]APIResp{}
	failResps := map[int]APIResp{}
	var wg sync.WaitGroup
	var minSuccessCount int

	if requestType == READ {
		minSuccessCount = MIN_READ_SUCCESS
	} else {
		minSuccessCount = MIN_WRITE_SUCCESS
	}

	wg.Add(minSuccessCount)

	go func(successes map[int]APIResp, fails map[int]APIResp) {
	Loop:
		for {
			select {
			case resp := <-respChannel:
				coordMutex.Lock()
				if resp.APIResp.Status == SUCCESS {
					log.Printf("Successful operation for request type %v\n", requestType)
					successes[resp.From] = resp.APIResp
					if len(successes) <= minSuccessCount {
						wg.Done()
					}
				} else {
					fails[resp.From] = resp.APIResp
					if len(fails) >= REPLICATION_FACTOR-minSuccessCount+1 {
						// too many nodes have failed -- return error to client
						if resp.APIResp.Status == SIMULATE_FAIL {
							log.Printf("Simulate failure operation for request type %v\n", requestType)
						} else {
							log.Printf("Failed operation for request type %v\n", requestType)
						}
						for i := 0; i < (minSuccessCount - len(successes)); i++ {
							wg.Done()
						}
						defer coordMutex.Unlock()
						break Loop
					}
				}
				if (len(successes) + len(fails)) == REPLICATION_FACTOR {
					// defer mutex unlock so that when we break out of this loop,
					// mutex will still be unlocked once the function returns
					defer coordMutex.Unlock()
					break Loop
				}
				coordMutex.Unlock()

				// TODO: ADD CHANNEL HERE TO DETECT TIMEOUT
				// IF DID NOT HIT minSuccessCount, THEN WE SEND AN ERROR TO THE CLIENT
				// IF minSuccessCount IS HIT BUT NODES TIMED OUT, SEND HINTED HANDOFF

				// TODO: actually, we will need to determine how we are going to simulate failed nodes
				// will we get an error while sending the API request like connection rejected?
				// or will the nodes simply not respond?
			}
		}
	}(successResps, failResps)

	wg.Wait()
	coordMutex.Lock()
	defer coordMutex.Unlock()
	if len(successResps) >= minSuccessCount {
		return true, successResps
	}
	return false, failResps
}

/*
checks if node has fail count > 0,
if yes, decrement fail count return true.
else, return false.
*/
func (n *Node) hasFailed() bool {
	if n.FailCount > 0 {
		n.FailCount--
		return true
	}
	return false
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

// type ClientCart struct {
// 	UserID      string
// 	Item        map[int]ItemObject
// 	VectorClock []int
// }

//returns node contains clientCartSelf, receives clientCartReceived
func MergeClientCarts(clientCartSelf, clientCartReceived ClientCart) ClientCart {
	var output ClientCart
	output.UserID = clientCartReceived.UserID
	newmap := make(map[int]ItemObject)
	for key, value := range clientCartSelf.Item {
		var currentKey = key
		var currentObject = value
		if val, ok := clientCartReceived.Item[currentKey]; ok {
			if currentObject.Quantity < val.Quantity {
				currentObject = val
			}
		}
		newmap[currentKey] = currentObject
	}

	for key, value := range clientCartReceived.Item {
		if _, ok := newmap[key]; ok {
		} else {
			newmap[key] = value
		}
	}

	output.Item = newmap

	newVectorClock := make([]int, len(clientCartSelf.VectorClock))
	for key, value := range clientCartSelf.VectorClock {
		newVectorClock[key] = Max(value, clientCartReceived.VectorClock[key])
	}

	output.VectorClock = newVectorClock

	return output
}

func ClientCartEqual(c1, c2 ClientCart) bool {
	if c1.UserID != c2.UserID {
		return false
	}
	if !ItemMapEqual(c1.Item, c2.Item) {
		return false
	}
	if !OrderedIntArrayEqual(c1.VectorClock, c2.VectorClock) {
		return false
	}
	return true
}

func ItemMapEqual(a, b map[int]ItemObject) bool {
	if len(a) != len(b) {
		return false
	}
	for key, vala := range a {
		if valb, ok := b[key]; ok {
			if !ItemObjectEqual(vala, valb) {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func ItemObjectEqual(a, b ItemObject) bool {
	if a.Id != b.Id {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if a.Quantity != b.Quantity {
		return false
	}
	return true
}

// checks if arr1 is smaller than arr2
func VectorClockSmaller(arr1, arr2 []int) bool {
	for i := 0; i < len(arr2); i++ {
		if arr1[i] > arr2[i] {
			return false
		}
	}
	return true
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
