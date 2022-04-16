package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/fatih/color"
)

func (n *Node) FulfilWriteRequest(w http.ResponseWriter, r *http.Request) {
	ColorLog("INTERNAL ENDPOINT /write HIT", color.FgYellow)
	var c ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &c)

	log.Println("Write request received with key: ", c.UserID)

	n.BadgerLock.Lock()
	badgerObject, err := n.BadgerWrite(c)
	n.BadgerLock.Unlock()

	resp := APIResp{}

	// if n.hasFailed() {
	// 	log.Printf("Request failed for node %v", n.Id)
	// 	w.WriteHeader(500)
	// 	resp.Status = SIMULATE_FAIL
	// 	resp.Error = "Node temporary failed."
	// 	return
	// }

	if err != nil {
		w.WriteHeader(500)
		resp.Status = FAIL
		resp.Error = err.Error()
	} else {
		w.WriteHeader(201)
		resp.Data = badgerObject
		resp.Status = SUCCESS
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		// return immediately since APIResp could not be marshalled
		w.WriteHeader(500)
		return
	}
	w.Write(jsonResp)
	log.Println("Write request completed for", badgerObject)
}

func (n *Node) FulfilReadRequest(w http.ResponseWriter, r *http.Request) {
	ColorLog("INTERNAL ENDPOINT /read HIT", color.FgYellow)
	query := r.URL.Query()

	userId := query.Get("id") //! type string

	log.Println("Read Request received with key: ", userId)

	badgerObject, err := n.BadgerRead(userId)

	resp := APIResp{}

	// if n.hasFailed() {
	// 	log.Printf("Request failed for node %v, fail count: %v\n", n.Id, n.FailCount)
	// 	w.WriteHeader(500)
	// 	resp.Status = SIMULATE_FAIL
	// 	resp.Error = "Node temporary failed."
	// 	return
	// }

	if err != nil {
		w.WriteHeader(500)
		resp.Status = FAIL
		resp.Error = err.Error()
		log.Printf("Error: %v", err)
	} else {
		w.WriteHeader(200)
		resp.Data = badgerObject
		resp.Status = SUCCESS
	}

	w.Header().Set("Content-Type", "application/json")

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		// return immediately since APIResp could not be marshalled
		w.WriteHeader(500)
		return
	}
	w.Write(jsonResp)
	log.Println("Read request completed for", badgerObject)
}

// func (n *Node) SimulateFailRequest(w http.ResponseWriter, r *http.Request) {
// 	ColorLog("INTERNAL ENDPOINT /simulate-fail HIT", color.FgYellow)
// 	query := r.URL.Query()

// 	count, err := strconv.Atoi(query.Get("count")) //! type string
// 	if err != nil {
// 		log.Println("Error with simluate fail request", err)
// 	}

// 	n.FailCount = count
// }

/* Calculate new node position and send position to new node */
func (n *Node) handleJoinRequest(w http.ResponseWriter, r *http.Request) {
	ColorLog("ENDPOINT /join-request HIT", color.FgMagenta)

	query := r.URL.Query()
	newNodeId, err := strconv.Atoi(query.Get("node"))
	if err != nil {
		log.Println("Error parising Node ID: ", err)
		w.WriteHeader(400)
		return
	}
	// check if node is existing node rejoining system
	newPos := n.GetPositionFromNodeMap(newNodeId)
	log.Println("Existing position of node: ", newPos)
	if newPos == -1 {
		// calculate node position if is new node
		newPos = n.GetNewPosition()
	}

	// create response
	apiResp := JoinResp{}
	w.Header().Set("Content-Type", "application/json")

	if newPos == -1 {
		// ring is full, send error to new node
		log.Println("Error: cannot find position for new node, ring is full")
		apiResp.Status = FAIL
		apiResp.Error = "Cannot find position for node, ring is full"
		w.WriteHeader(409)
		return
	} else {
		// send back newPos
		w.WriteHeader(200)
		data := JoinOfferObject{Position: newPos, NodeMap: n.NodeMap}
		apiResp.Status = SUCCESS
		apiResp.Data = data
	}

	encodedResp, err := json.Marshal(apiResp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s\n", err)
		// return immediately since APIResp could not be marshalled
		w.WriteHeader(500)
		return
	}
	w.Write(encodedResp)
	log.Println("Join offer sent with assigned position ", newPos)
}

func (n *Node) handleJoinBroadcast(w http.ResponseWriter, r *http.Request) {
	/*
		As an example, when node 5 joins the ring at position 12, it should acquire all the keys
		from position 50 to position 12 (ie. (50,75], (75,0], (0,12]).
		All these keys will currently be stored in node 3 at position 25, so node 5 only has to contact node 3 to get all the keys.
		Node 3 can then delete the keys at position (50,75] as well since node 5 will be taking care of those keys instead.
		Node 1 can also delete the keys at (75,0] and
		node 4 can delete the keys at (0-12].
	*/
	ColorLog("ENDPOINT /join-broadcast HIT", color.FgMagenta)
	var newNode NodeData
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &newNode)

	// extract position of new node and add to its node map
	n.NodeMap[newNode.Position] = newNode

	log.Printf("New node with ID %d added to NodeMap\n", newNode.Id)
	log.Printf("New node information: %+v\n", newNode)

	resp := MigrateResp{}

	// only migrate data and delete unneeded keys when minimum num of nodes were already in the system
	if len(n.NodeMap) > REPLICATION_FACTOR {
		var allKeys []string
		shouldMigrateData := n.ShouldMigrateData(newNode.Position)
		shouldDeleteData := n.ShouldDeleteData(newNode.Position)

		log.Printf("Migrating data? %v\n", shouldMigrateData)
		log.Printf("Deleting data? %v\n", shouldDeleteData)

		if shouldMigrateData || shouldDeleteData {
			allKeys, _ = n.BadgerGetKeys()
		}

		// check if node is in charge of migrating keys to new node
		if shouldMigrateData {
			start, end := n.CalculateKeyset(MIGRATE)
			migrateData := []BadgerObject{}
			for _, key := range allKeys {
				if KeyInRange(key, start, end) {
					obj, err := n.BadgerRead(key)
					if err != nil {
						log.Fatalf("Error happened while getting migration objects. Err: %s\n", err)
						w.WriteHeader(500)
						return
					}
					migrateData = append(migrateData, obj)
				}
			}
			resp.Data = migrateData
			log.Printf("Migration of %d keys at (%d, %d] completed\n", len(migrateData), start, end)
		}

		// check if node can delete any keys
		if shouldDeleteData {
			start, end := n.CalculateKeyset(DELETE)
			count := 0
			for _, key := range allKeys {
				if KeyInRange(key, start, end) {
					err := n.BadgerDelete([]string{key})
					if err != nil {
						log.Fatalf("Error happened while deleting objects. Err: %s\n", err)
						w.WriteHeader(500)
						return
					}
					count++
				}
			}
			log.Printf("Deletion of %d keys at (%d, %d] completed\n", count, start, end)
		}
	}

	resp.Status = SUCCESS
	encodedResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s\n", err)
		// return immediately since APIResp could not be marshalled
		w.WriteHeader(500)
		return
	}
	w.Write(encodedResp)
	log.Println("Completed handling join broadcast")
}
