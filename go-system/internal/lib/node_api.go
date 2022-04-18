package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fatih/color"
)

func (n *Node) tryHintedHandoff(wo WriteObject) {
	ticker := time.NewTicker(HINTED_HANDOFF_INTERVAL)
	respCh := make(chan ChannelResp)

	intendedNode := n.NodeMap[n.GetPositionFromNodeMap(wo.Hint)]
	wo.Hint = NIL_HINT

	log.Printf("Starting to send hinted replica with userId %s to Node %d\n", wo.Data.UserID, intendedNode.Id)

	for {
		select {
		case <-ticker.C:
			// send hinted replica
			log.Printf("Sending replica with userId %s to Node %d\n", wo.Data.UserID, intendedNode.Id)
			go n.sendWriteRequestAsync(wo, intendedNode, respCh)
		case chResp := <-respCh:
			if chResp.APIResp.Status == SUCCESS {
				// successfully handed off data
				log.Printf("Successfully sent replica with userId %s to Node %d\n", wo.Data.UserID, intendedNode.Id)
				ticker.Stop()
				delete(n.HintedStorage, chResp.APIResp.Data.UserID)
				log.Printf("Hinted Storage: %+v\n", n.HintedStorage)
				log.Printf("Node %d has revived \n", intendedNode.Id)
				break
			} else if chResp.APIResp.Error != TIMEOUT_ERROR {
				log.Fatalf("Node %d could not successfully store data, probably a bug in the code", intendedNode.Id)
			}
		}
	}
}

func (n *Node) FulfilHintedHandoff(wo WriteObject, w *http.ResponseWriter) {
	log.Println("Received hinted replica")
	// store hinted replica
	bo := BadgerObject{UserID: wo.Data.UserID, Versions: []ClientCart{wo.Data}}
	n.HintedStorage[bo.UserID] = bo
	log.Printf("Hinted Storage: %+v\n", n.HintedStorage)

	go n.tryHintedHandoff(wo)
	// return success
	(*w).WriteHeader(201)
	resp := APIResp{}
	resp.Status = SUCCESS

	(*w).Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		// return immediately since APIResp could not be marshalled
		(*w).WriteHeader(500)
		return
	}
	(*w).Write(jsonResp)
	return
}

func (n *Node) FulfilWriteRequest(w http.ResponseWriter, r *http.Request) {
	ColorLog("INTERNAL ENDPOINT /write HIT", color.FgYellow)
	var wo WriteObject
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(body, &wo)

	// handle hinted handoff since normal write and hinted handoff use the same endpoint
	if wo.Hint != NIL_HINT {
		n.FulfilHintedHandoff(wo, &w)
		return
	}

	// handle normal write
	c := wo.Data
	log.Println("Write request received with key: ", c.UserID)

	n.BadgerLock.Lock()
	badgerObject, err := n.BadgerWrite(c)
	n.BadgerLock.Unlock()

	resp := APIResp{}

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

	var badgerObject BadgerObject
	var err error
	// Check whether key exists in HintedMap
	if _, ok := n.HintedStorage[userId]; ok {
		//return response if found in hintedstorage
		badgerObject = n.HintedStorage[userId]
	} else {
		badgerObject, err = n.BadgerRead(userId)
	}

	resp := APIResp{}

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

func (n *Node) FulfilHintedRead(wo WriteObject, w *http.ResponseWriter) {
	log.Println("Received hinted replica")
	// store hinted replica
	bo := BadgerObject{UserID: wo.Data.UserID, Versions: []ClientCart{wo.Data}}
	n.HintedStorage[bo.UserID] = bo
	log.Printf("Hinted Storage: %+v\n", n.HintedStorage)

	go n.tryHintedHandoff(wo)
	// return success
	(*w).WriteHeader(201)
	resp := APIResp{}
	resp.Status = SUCCESS

	(*w).Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		// return immediately since APIResp could not be marshalled
		(*w).WriteHeader(500)
		return
	}
	(*w).Write(jsonResp)
	return
}

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
