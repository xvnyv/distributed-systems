package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

/* Write would need: ItemID, ItemName, ItemQuantity and UserID.

Read will obtain information from UserID.
*/

// ========== START COORDINATOR WRITE ==========

/* Send individual internal write request to each node */
func (n *Node) sendWriteRequest(c ClientCart, node NodeData, respChannel chan<- ChannelResp) {
	jsonData, _ := json.Marshal(c)
	resp, err := http.Post(fmt.Sprintf("%s/write", node.Ip), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Send Write Request Error: ", err)
		// NOTE: Will have to set a timer to detect timeout for failures in DetermineSuccess if we
		// return like this without sending any failure response to respChannel
		return
	}

	var apiResp APIResp
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode == 500 {
		log.Printf("Internal API Write Request Error: %v\n", apiResp.Error)
	}
	respChannel <- ChannelResp{node.Id, apiResp}

	defer resp.Body.Close()
	log.Println("Write request response body:", body)

}

/* Send requests to all responsible nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) sendWriteRequests(c ClientCart, nodes [REPLICATION_FACTOR]NodeData, coordMutex *sync.Mutex) (bool, map[int]APIResp) {
	var respChannel = make(chan ChannelResp, 10)

	for _, node := range nodes {
		go n.sendWriteRequest(c, node, respChannel)
	}

	// TODO: DETECT NODE FAILURE IN DETERMINE SUCCESS
	return DetermineSuccess(WRITE, respChannel, coordMutex)
}

/* Message handler for write requests for external API to client application */
func (n *Node) handleWriteRequest(w http.ResponseWriter, r *http.Request) {
	var c ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &c)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(APIResp{FAIL, ClientCart{}, "JSON decoding error"})
		log.Printf("Handle Write Request Error: %v\n", err)
		return
	}
	hashKey := HashMD5(c.UserID)
	responsibleNodes := n.GetResponsibleNodes(hashKey)

	// TODO: Vector clock updates should be done by the individual writing nodes
	// Coordinator info should be sent to the writing nodes so that they can update the
	// vector clock correctly
	// If vector clock is not provided, individual nodes should read the vector clock value
	// that is currently being stored and incrememnt the correct index by 1 when updating the object
	// If vector clock is provided, check the nodes to ensure that the provided version exists,
	// then update all replicas to the same cart and vector clock versions to ensure eventual consistency
	if c.VectorClock == nil {
		c.VectorClock = []int{}
		for i := 0; i < len(n.NodeMap); i++ {
			if i == n.Id {
				c.VectorClock = append(c.VectorClock, 1)
			} else {
				c.VectorClock = append(c.VectorClock, 0)
			}
		}
	} else {
		c.VectorClock[n.Id]++
	}

	var coordMutex sync.Mutex
	success, resps := n.sendWriteRequests(c, responsibleNodes, &coordMutex)

	if success {
		w.WriteHeader(201)
	} else {
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")
	coordMutex.Lock()
	for _, v := range resps {
		json.NewEncoder(w).Encode(v)
		break
	}
	coordMutex.Unlock()
}

// ========== END COORDINATOR WRITE ==========

// ========== START COORDINATOR READ ==========

/* Send individual internal read request to each node */
func (n *Node) sendReadRequest(key string, node NodeData, respChannel chan<- ChannelResp) {
	// Add key to query params
	base, _ := url.Parse(fmt.Sprintf("%s/read?id=%s", node.Ip, key)) // key = userID

	// Send read request to node
	resp, err := http.Get(base.String())
	if err != nil {
		log.Println("Send Read Request Error: ", err)
	}

	var apiResp APIResp
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode == 500 {
		log.Println("Internal API Read Request Error:", apiResp.Error)
	}

	respChannel <- ChannelResp{node.Id, apiResp}

	defer resp.Body.Close()
}

/* Send requests to all responsible nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) sendReadRequests(key string, nodes [REPLICATION_FACTOR]NodeData, coordMutex *sync.Mutex) (bool, map[int]APIResp) {
	var respChannel = make(chan ChannelResp, 10)

	for _, node := range nodes {
		go n.sendReadRequest(key, node, respChannel)
	}

	// TODO: DETECT NODE FAILURE IN DETERMINE SUCCESS
	return DetermineSuccess(READ, respChannel, coordMutex)
}

/* Message handler for read requests for external API to client application */
func (n *Node) handleReadRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userId := query.Get("id")

	pos := HashMD5(userId)
	responsibleNodes := n.GetResponsibleNodes(pos)
	var coordMutex sync.Mutex
	success, resps := n.sendReadRequests(userId, responsibleNodes, &coordMutex)

	// TODO: this section has to be edited to catch conflicts in case of success
	if success {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")

	coordMutex.Lock()
	// TODO: change this part when handling conflict -- right now we are just returning one cart version
	// might have to change the APIResp object to return an array of carts instead so that
	// we can return the conflicting versions
	for _, v := range resps {
		json.NewEncoder(w).Encode(v)
		break
	}
	coordMutex.Unlock()
}

// ========== END COORDINATOR READ ==========

func (n *Node) HandleRequests() {
	// Internal API
	http.HandleFunc("/read", n.FulfilReadRequest)
	http.HandleFunc("/write", n.FulfilWriteRequest)
	// http.HandleFunc("/write-success", handleMessage2)
	// http.HandleFunc("/read-success", handleMessage2)
	// http.HandleFunc("/join-request", handleMessage2)
	// http.HandleFunc("/join-broadcast", handleMessage2)
	// http.HandleFunc("/join-offer", handleMessage2)
	// http.HandleFunc("/data-migration", handleMessage2)
	// http.HandleFunc("/handover-request", handleMessage2)
	// http.HandleFunc("/handover-success", handleMessage2)

	// External API
	http.HandleFunc("/write-request", n.handleWriteRequest)
	http.HandleFunc("/read-request", n.handleReadRequest)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", n.Port), nil))
}
