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
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

/* Send individual internal write request to each node */
func (n *Node) sendWriteRequest(c ClientCart, node NodeData, respChannel chan<- ChannelResp) {

	jsonData, _ := json.Marshal(c)
	var apiResp APIResp
	resp, err := http.Post(fmt.Sprintf("%s/write", node.Ip), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Send Write Request Error: ", err)
		if strings.Contains(err.Error(), "connection refused") {
			apiResp.Error = "Timeout"
			apiResp.Status = FAIL
			respChannel <- ChannelResp{node.Id, apiResp}
			return
		}
	}

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	json.Unmarshal(body, &apiResp)
	respChannel <- ChannelResp{node.Id, apiResp}
}

/* Send requests to all responsible nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) sendWriteRequests(c ClientCart, nodes [REPLICATION_FACTOR]NodeData, coordMutex *sync.Mutex) (bool, map[int]APIResp, map[int]APIResp) {
	var respChannel = make(chan ChannelResp, 10)

	for _, node := range nodes {
		go n.sendWriteRequest(c, node, respChannel)
	}

	return DetermineSuccess(WRITE, respChannel, coordMutex)
}

/* Send requests to unresponsive nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) hintedWriteRequest(c ClientCart, node NodeData) {
	// resps contains the failed nodes' responses
	var respChannel = make(chan ChannelResp, 10)
	timer := time.NewTimer(time.Second * 15)
	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ticker.C:
			go n.sendWriteRequest(c, node, respChannel)
			resp := <-respChannel
			if resp.APIResp.Status == SUCCESS {
				// great
				ticker.Stop()
				timer.Stop()
				log.Printf("Node %v has revived \n", node.Id)
				return
			}
		case <-timer.C:
			// end liao
			log.Printf("Node %v permanently failed\n", node.Id)
			ticker.Stop()
			return
		}
	}
}

/* Message handler for write requests for external API to client application */
func (n *Node) handleWriteRequest(w http.ResponseWriter, r *http.Request) {
	ColorLog(fmt.Sprintf("Coordinator Node:%v WRITE REQUEST FROM CLIENT RECEIVED", n.Id), color.FgCyan)

	// ? FEATURE: if node fails, it can still coordinate so that hinted handoff will be executed.
	// ? If we allow coordinator to fail, the write request gets dropped without any backup
	// ? TBC whether coordinator should fail...

	var c ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &c)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(APIResp{FAIL, BadgerObject{}, "JSON decoding error"})
		log.Printf("Handle Write Request Error: %v\n", err)
		return
	}
	log.Printf("Input: %s\n", c.UserID)
	hashKey := HashMD5(c.UserID)
	responsibleNodes := n.GetResponsibleNodes(hashKey)
	log.Printf("Responsible nodes: %+v", responsibleNodes)

	// TODO: Vector clock updates should be done by the individual writing nodes
	// Coordinator info should be sent to the writing nodes so that they can update the
	// vector clock correctly
	// If vector clock is not provided, individual nodes should read the vector clock value
	// that is currently being stored and incrememnt the correct index by 1 when updating the object
	// If vector clock is provided, check the nodes to ensure that the provided version exists,
	// then update all replicas to the same cart and vector clock versions to ensure eventual consistency
	if c.VectorClock == nil {
		c.VectorClock = make(map[int]int, 0)
		for i := 0; i < len(n.NodeMap); i++ {
			c.VectorClock[i] = 0
		}
	}

	var coordMutex sync.Mutex

	// update vector clock using coordinator's ID
	c.VectorClock[n.Id] += 1

	success, SuccessResps, FailResps := n.sendWriteRequests(c, responsibleNodes, &coordMutex)

	if success {
		w.WriteHeader(201)
	} else {
		w.WriteHeader(500)
	}
	for id := range FailResps {
		for _, nodeData := range n.NodeMap {
			if nodeData.Id == id {
				go n.hintedWriteRequest(c, nodeData)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	coordMutex.Lock()
	for _, v := range SuccessResps {
		json.NewEncoder(w).Encode(v)
		break
	}
	coordMutex.Unlock()
}

/* Send individual internal read request to each node */
func (n *Node) sendReadRequest(key string, node NodeData, respChannel chan<- ChannelResp) {
	// Add key to query params
	base, _ := url.Parse(fmt.Sprintf("%s/read?id=%s", node.Ip, key)) // key = userID

	// Send read request to node
	resp, err := http.Get(base.String())
	if err != nil {
		log.Println("Send Read Request Error: ", err)
		if strings.Contains(err.Error(), "connection refused") {
			return
		}
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
func (n *Node) sendReadRequests(key string, nodes [REPLICATION_FACTOR]NodeData, coordMutex *sync.Mutex) (bool, map[int]APIResp, map[int]APIResp) {
	var respChannel = make(chan ChannelResp, 10)

	for _, node := range nodes {
		go n.sendReadRequest(key, node, respChannel)
	}

	// TODO: DETECT NODE FAILURE IN DETERMINE SUCCESS
	return DetermineSuccess(READ, respChannel, coordMutex)
}

/* Message handler for read requests for external API to client application */
func (n *Node) handleReadRequest(w http.ResponseWriter, r *http.Request) {
	ColorLog(fmt.Sprintf("Coordinator Node:%v READ REQUEST FROM CLIENT RECEIVED", n.Id), color.FgCyan)
	query := r.URL.Query()
	userId := query.Get("id")

	log.Printf("Input: %s\n", userId)
	pos := HashMD5(userId)
	responsibleNodes := n.GetResponsibleNodes(pos)
	log.Printf("Responsible nodes: %+v", responsibleNodes)

	var coordMutex sync.Mutex
	success, SuccessResps, _ := n.sendReadRequests(userId, responsibleNodes, &coordMutex)

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

	// check

	for _, v := range SuccessResps {
		json.NewEncoder(w).Encode(v)
		break

	}
	coordMutex.Unlock()

}
