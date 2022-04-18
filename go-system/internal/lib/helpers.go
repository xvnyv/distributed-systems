package lib

import (
	"crypto/md5"
	"log"
	"math/big"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/fatih/color"
)

/*
Contains helper functions for APIs
*/

func ColorLog(s string, chosenColor color.Attribute) {
	color.Set(chosenColor)
	log.Println(s)
	color.Unset()
}

func HashMD5(s string) int {
	hasher := md5.New()
	hasher.Write([]byte(s))
	hashedBigInt := big.NewInt(0)
	hashedBigInt.SetBytes(hasher.Sum(nil))

	maxRingPosBigInt := big.NewInt(100)
	ringPosBigInt := big.NewInt(0)
	return int(ringPosBigInt.Mod(hashedBigInt, maxRingPosBigInt).Uint64())
}

func sortPositions(nodeMap NodeMap) []int {
	posArr := []int{}
	for pos := range nodeMap {
		posArr = append(posArr, pos)
	}
	sort.Ints(posArr)
	return posArr
}

func nodeInSlice(node NodeData, slice []NodeData) bool {
	for _, n := range slice {
		if node.Id == n.Id {
			return true
		}
	}
	return false
}

func GetPredecessor(n NodeData, nodeMap NodeMap) NodeData {
	posArr := sortPositions(nodeMap)

	var predecessor NodeData
	for i, pos := range posArr {
		if pos == n.Position {
			predecessor = nodeMap[posArr[(i+len(posArr)-1)%len(posArr)]]
		}
	}
	return predecessor
}

func GetSuccessor(n NodeData, nodeMap NodeMap) NodeData {
	posArr := sortPositions(nodeMap)

	var successor NodeData
	for i, pos := range posArr {
		if pos == n.Position {
			successor = nodeMap[posArr[(i+len(posArr)+1)%len(posArr)]]
		}
	}
	return successor
}

func (n *Node) GetResponsibleNodes(keyPos int) [REPLICATION_FACTOR]NodeData {
	posArr := sortPositions(n.NodeMap)

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

// Returns position of node, or -1 if node id is not in node map
func (n *Node) GetPositionFromNodeMap(id int) int {
	for _, nodeData := range n.NodeMap {
		if nodeData.Id == id {
			return nodeData.Position
		}
	}
	return -1
}

/*
	Largest gap between all node positions will be found and the middle position of this largest gap will be returned.
	Returns position for new node, -1 if ring is full and new node cannot join

	Middle position will be calculated as:
	- If odd number of positions, middle position is taken (eg. {0,1,2} will mean 1 is chosen)
	- If even number of positions, lower middle position is taken (eg. {0,1,2,3} will mean 1 is chosen)

	When calculating the new position, the indexes that we use to find the gap are excluded from the available spots.
	Eg. to find which position should be selected from {3,4,5,6}, we take (7-2)/2+2=4 (2 and 7 are excluded from available spots)
*/
func (n *Node) GetNewPosition() int {
	posArr := sortPositions(n.NodeMap)
	if len(posArr) == 1 {
		// only one node in the system -- used total ring positions to calculate new position
		return (NUM_RING_POSITIONS-posArr[0])/2 + posArr[0]
	}

	largestGap := 0
	largestGapLowerIndex := -1
	// find largest gap in the rest of the ring
	for i := 0; i < len(posArr)-1; i++ {
		gap := posArr[i+1] - posArr[i]
		if gap > largestGap {
			largestGap = gap
			largestGapLowerIndex = i
		}
	}
	// handle finding gap in loop back from largest to smallest index
	lastGap := (posArr[0] + NUM_RING_POSITIONS) - posArr[len(posArr)-1]
	if lastGap > largestGap {
		largestGap = lastGap
		largestGapLowerIndex = len(posArr) - 1
	}
	log.Printf("Largest gap is between position %d and %d\n", posArr[largestGapLowerIndex], posArr[(largestGapLowerIndex+1)%len(posArr)])
	if largestGap == 1 {
		// ring is full
		return -1
	}
	return (largestGap/2 + posArr[largestGapLowerIndex]) % NUM_RING_POSITIONS
}

func (n *Node) CalculateKeyset(action KeysetAction) (int, int) {
	posArr := sortPositions(n.NodeMap)

	var nodeIndex int
	for i, pos := range posArr {
		if pos == n.Position {
			nodeIndex = i
			break
		}
	}
	startIndex := (nodeIndex + len(posArr) - REPLICATION_FACTOR - 1) % len(posArr)

	var endIndex int
	switch action {
	case MIGRATE:
		endIndex = (nodeIndex + len(posArr) - 1) % len(posArr)
	case DELETE:
		endIndex = (startIndex + 1) % len(posArr)
	}

	// exclusive start, inclusive end
	return posArr[startIndex], posArr[endIndex]
}

/* Returns true if this node is the successor of the new node */
func (n *Node) ShouldMigrateData(newPos int) bool {
	posArr := sortPositions(n.NodeMap)

	for i, pos := range posArr {
		if pos == newPos {
			return posArr[(i+1)%len(posArr)] == n.Position
		}
	}
	// should never reach this return bcos newPos should be in posArr
	return false
}

/* Returns true if this node is one of the 3 subsequent successors of the new node */
func (n *Node) ShouldDeleteData(newPos int) bool {
	posArr := sortPositions(n.NodeMap)

	var newPosIndex int
	for i, pos := range posArr {
		if pos == newPos {
			newPosIndex = i
			break
		}
	}
	for i := 1; i <= REPLICATION_FACTOR; i++ {
		if posArr[(newPosIndex+i)%len(posArr)] == n.Position {
			return true
		}
	}
	return false
}

/* Returns true if position of key falls within the given range (start, end] */
func KeyInRange(key string, start int, end int) bool {
	pos := HashMD5(key)
	// exclude start, include end
	loopbackDelete := start > end && (pos > start || pos <= end)
	regularDelete := start < end && (pos > start && pos <= end)
	return loopbackDelete || regularDelete
}

func (n *Node) DetermineSuccess(successResps map[int]APIResp, failResps map[int]APIResp, requestType RequestType, nodes *[]NodeData, respChannel <-chan ChannelResp, coordMutex *sync.Mutex, wo WriteObject, key string) (bool, map[int]APIResp, map[int]APIResp) {
	/*
		As long as (REPLICATION_FACTOR - MIN_WRITE_SUCCESS + 1) nodes fail, we return an error to the client
		It does not matter if 1 node has already successfully written to disk even if the entire operation fails
		We let the client execute a read before retrying the write to handle that case
		Inspired by DynamoDB: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Programming.Errors.html
	*/
	failureTimer := time.NewTimer(RESPONSE_TIMEOUT)
	var wg sync.WaitGroup
	var minSuccessCount int

	if requestType == READ {
		minSuccessCount = MIN_READ_SUCCESS
	} else {
		minSuccessCount = MIN_WRITE_SUCCESS
	}

	wg.Add(minSuccessCount - len(successResps) - len(failResps))

	go func(successes map[int]APIResp, fails map[int]APIResp) {
		for {
			select {
			case resp := <-respChannel:
				coordMutex.Lock()
				if resp.APIResp.Status == SUCCESS {
					log.Printf("Successful operation for request type %v\n", requestType)
					successes[resp.From.Id] = resp.APIResp
					if len(successes) <= minSuccessCount {
						wg.Done()
					}
				} else if resp.APIResp.Error != TIMEOUT_ERROR {
					fails[resp.From.Id] = resp.APIResp
					if len(fails) >= REPLICATION_FACTOR-minSuccessCount+1 {
						// too many nodes have failed -- return error to client
						log.Printf("Failed operation for request type %v\n", requestType)
						for i := 0; i < (minSuccessCount - len(successes)); i++ {
							wg.Done()
						}
						failureTimer.Stop()
						defer coordMutex.Unlock()
						return
					}
				}
				if (len(successes) + len(fails)) == REPLICATION_FACTOR {
					// defer mutex unlock so that when we break out of this loop,
					// mutex will still be unlocked once the function returns
					failureTimer.Stop()
					defer coordMutex.Unlock()
					return
				}
				coordMutex.Unlock()
			case <-failureTimer.C:
				// all successful responses should have been received by now -- stop listening for responses
				// assume that no more responses will be received after this so we can close respChannel
				log.Println("Failure timer reached -- some nodes failed without announcing")
				remainingWgCount := minSuccessCount - len(successResps)
				switch requestType {
				case WRITE:
					// get nodes with unannounced failures
					failedNodes := []NodeData{}
					for _, curNode := range n.GetResponsibleNodes(HashMD5(wo.Data.UserID)) {
						_, inSuccess := successResps[curNode.Id]
						_, inFail := failResps[curNode.Id]
						if !inSuccess && !inFail {
							failedNodes = append(failedNodes, curNode)
						}
					}

					handoffCh := make(chan ChannelResp, REPLICATION_FACTOR)

					go n.sendHintedReplicas(wo, nodes, failedNodes, handoffCh)
					n.DetermineSuccess(successResps, failResps, requestType, nodes, handoffCh, coordMutex, wo, key)
					// unblock
					for i := 0; i < remainingWgCount; i++ {
						wg.Done()
					}
					return

				case READ:
					// get nodes with unannounced failures
					failedNodes := []NodeData{}
					for _, curNode := range n.GetResponsibleNodes(HashMD5(wo.Data.UserID)) {
						_, inSuccess := successResps[curNode.Id]
						_, inFail := failResps[curNode.Id]
						if !inSuccess && !inFail {
							failedNodes = append(failedNodes, curNode)
						}
					}

					handoffCh := make(chan ChannelResp, REPLICATION_FACTOR)

					go n.getHintedReplicas(key, nodes, failedNodes, handoffCh)
					n.DetermineSuccess(successResps, failResps, requestType, nodes, handoffCh, coordMutex, wo, key)
					// unblock
					for i := 0; i < remainingWgCount; i++ {
						wg.Done()
					}
					return
				}
			}
		}
	}(successResps, failResps)

	wg.Wait()
	coordMutex.Lock()
	defer coordMutex.Unlock()
	if len(successResps) >= minSuccessCount {
		return true, successResps, failResps
	}
	return false, successResps, failResps
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
	if !reflect.DeepEqual(o.VectorClock, b.VectorClock) {
		return false
	}
	return true
}

func ClientCartEqual(c1, c2 ClientCart) bool {
	if c1.UserID != c2.UserID {
		return false
	}
	if !ItemMapEqual(c1.Item, c2.Item) {
		return false
	}
	if !reflect.DeepEqual(c1.VectorClock, c2.VectorClock) {
		return false
	}
	return true
}

func ItemMapEqual(a, b map[int]ItemObject) bool {

	return reflect.DeepEqual(a, b)
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

// checks if map1 is smaller than map2
func VectorClockSmaller(map1, map2 map[int]int) bool {
	for k := range map1 {
		if _, ok := map2[k]; ok {
			if map1[k] > map2[k] {
				return false
			}
		} else {
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
