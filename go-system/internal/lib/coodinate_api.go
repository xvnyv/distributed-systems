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

	"github.com/fatih/color"
)

/* Send individual internal write request to each node without using a goroutine (ie. synchronously) */
func (n *Node) sendWriteRequestSync(wo WriteObject, node NodeData) APIResp {
	jsonData, _ := json.Marshal(wo)
	var apiResp APIResp
	resp, err := http.Post(fmt.Sprintf("%s/write", node.Ip), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Send Write Request Error: ", err)
		if strings.Contains(err.Error(), "connection refused") {
			apiResp.Error = TIMEOUT_ERROR
			apiResp.Status = FAIL
			return apiResp
		}
	}

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	json.Unmarshal(body, &apiResp)
	return apiResp
}

/* Send individual internal write request to each node using a goroutine (ie. asynchronously) */
func (n *Node) sendWriteRequestAsync(wo WriteObject, node NodeData, respChannel chan<- ChannelResp) {
	apiResp := n.sendWriteRequestSync(wo, node)
	respChannel <- ChannelResp{node, apiResp}
}

/* Send requests to all responsible nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) sendWriteRequests(wo WriteObject, nodes [REPLICATION_FACTOR]NodeData, coordMutex *sync.Mutex) (bool, map[int]APIResp, map[int]APIResp) {
	var respChannel = make(chan ChannelResp, 10)

	for _, node := range nodes {
		go n.sendWriteRequestAsync(wo, node, respChannel)
	}

	successResps := map[int]APIResp{}
	failResps := map[int]APIResp{}
	nodesCopy := make([]NodeData, len(nodes))
	copy(nodesCopy, nodes[:])
	success, _, _ := n.DetermineSuccess(successResps, failResps, WRITE, &nodesCopy, respChannel, coordMutex, wo, "")
	return success, successResps, failResps
}

/* Send requests to unresponsive nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) sendHintedReplica(wo WriteObject, node NodeData, nodes *[]NodeData) ChannelResp {
	var hintedHandoffNode NodeData

	// set hint to original node
	wo.Hint = node.Id

	for {
		// get unused predecessor to send node to hinted handoff
		hintedHandoffNode = GetSuccessor(hintedHandoffNode, n.NodeMap)
		if hintedHandoffNode.Id == node.Id {
			// all nodes have already been tried for storing the replicas
			return ChannelResp{From: node, APIResp: APIResp{Status: FAIL, Error: "No nodes left to hand off replica"}}
		}
		if !nodeInSlice(hintedHandoffNode, *nodes) {
			log.Printf("Handing off replica to Node %d\n", hintedHandoffNode.Id)
			apiResp := n.sendWriteRequestSync(wo, hintedHandoffNode)
			if apiResp.Status == SUCCESS || apiResp.Error == TIMEOUT_ERROR {
				if apiResp.Status == SUCCESS {
					log.Printf("Successfully handed off replica to Node %d\n", hintedHandoffNode.Id)
				}
				*nodes = append(*nodes, hintedHandoffNode)
				return ChannelResp{From: hintedHandoffNode, APIResp: apiResp}
			}
		}
	}
}

/* Send all hinted replicas */
func (n *Node) sendHintedReplicas(wo WriteObject, nodes *[]NodeData, failedNodes []NodeData, handoffCh chan<- ChannelResp) {
	log.Printf("Sending hinted replicas to %d nodes\n", len(failedNodes))
	for _, nodeData := range failedNodes {
		// no point sending the hinted write requests synchronously because we'll have to lock the section where we find the unused successor until we get a successful response anyway
		// otherwise there may be a case of multiple hinted replicas being stored at the same node
		channelResp := n.sendHintedReplica(wo, nodeData, nodes)
		handoffCh <- channelResp
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

	// create WriteObject and send requests
	wo := WriteObject{Hint: -1, Data: c}
	success, successResps, failResps := n.sendWriteRequests(wo, responsibleNodes, &coordMutex)

	if success {
		w.WriteHeader(201)
	} else {
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")
	coordMutex.Lock()
	returnResp := successResps
	if !success {
		returnResp = failResps
	}
	for _, v := range returnResp {
		if success {
			v.Data = BadgerObject{UserID: c.UserID, Versions: []ClientCart{c}}
		}
		json.NewEncoder(w).Encode(v)
		break
	}
	coordMutex.Unlock()
}

// ========================================================================================================================
// ========================================================================================================================
// ========================================================================================================================

/* Send requests to unresponsive nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) getHintedReplica(key string, node NodeData, nodes *[]NodeData) ChannelResp {
	var hintedHandoffNode NodeData

	for {
		// get unused predecessor to send node to hinted handoff
		hintedHandoffNode = GetSuccessor(hintedHandoffNode, n.NodeMap)
		if hintedHandoffNode.Id == node.Id {
			// all nodes have already been tried for storing the replicas
			return ChannelResp{From: node, APIResp: APIResp{Status: FAIL, Error: "No nodes left to hand off replica"}}
		}
		if !nodeInSlice(hintedHandoffNode, *nodes) {
			log.Printf("Reading replica from Node %d\n", hintedHandoffNode.Id)
			apiResp := n.sendReadRequestSync(key, hintedHandoffNode)
			if apiResp.Status == SUCCESS || apiResp.Error == TIMEOUT_ERROR {
				if apiResp.Status == SUCCESS {
					log.Printf("Successfully read replica from Node %d\n", hintedHandoffNode.Id)
				}
				*nodes = append(*nodes, hintedHandoffNode)
				return ChannelResp{From: hintedHandoffNode, APIResp: apiResp}
			}
		}
	}
}

/* Send all hinted replicas */
func (n *Node) getHintedReplicas(key string, nodes *[]NodeData, failedNodes []NodeData, handoffCh chan<- ChannelResp) {
	log.Printf("Sending hinted replicas to %d nodes\n", len(failedNodes))
	for _, nodeData := range failedNodes {
		// no point sending the hinted write requests synchronously because we'll have to lock the section where we find the unused successor until we get a successful response anyway
		// otherwise there may be a case of multiple hinted replicas being stored at the same node
		channelResp := n.getHintedReplica(key, nodeData, nodes)
		handoffCh <- channelResp
	}
}

/* Send individual internal read request to each node */
func (n *Node) sendReadRequestSync(key string, node NodeData) APIResp {
	// Add key to query params
	base, _ := url.Parse(fmt.Sprintf("%s/read?id=%s", node.Ip, key)) // key = userID

	var apiResp APIResp
	// Send read request to node
	resp, err := http.Get(base.String())
	if err != nil {
		log.Println("Send Read Request Error: ", err)
		if strings.Contains(err.Error(), "connection refused") {
			apiResp.Error = TIMEOUT_ERROR
			apiResp.Status = FAIL
			return apiResp
		}
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode == 500 {
		log.Println("Internal API Read Request Error:", apiResp.Error)
	}

	defer resp.Body.Close()
	json.Unmarshal(body, &apiResp)
	return apiResp
}

/* Send individual internal read request to each node */
func (n *Node) sendReadRequestAsync(key string, node NodeData, respChannel chan<- ChannelResp) {
	apiResp := n.sendReadRequestSync(key, node)
	respChannel <- ChannelResp{node, apiResp}
}

/* Send requests to all responsible nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) sendReadRequests(key string, nodes [REPLICATION_FACTOR]NodeData, coordMutex *sync.Mutex) (bool, map[int]APIResp, map[int]APIResp) {
	var respChannel = make(chan ChannelResp, 10)

	for _, node := range nodes {
		go n.sendReadRequestAsync(key, node, respChannel)
	}
	successResps := map[int]APIResp{}
	failResps := map[int]APIResp{}
	nodesCopy := make([]NodeData, len(nodes))
	copy(nodesCopy, nodes[:])
	success, _, _ := n.DetermineSuccess(successResps, failResps, READ, &nodesCopy, respChannel, coordMutex, WriteObject{}, key)
	return success, successResps, failResps
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
	success, successResps, failResps := n.sendReadRequests(userId, responsibleNodes, &coordMutex)

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

	returnResp := successResps
	if !success {
		returnResp = failResps
	}
	for _, v := range returnResp {
		log.Println(v)
		json.NewEncoder(w).Encode(v)
		break
	}
	coordMutex.Unlock()

}
