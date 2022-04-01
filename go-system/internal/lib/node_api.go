package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func (n *Node) FulfilWriteRequest(w http.ResponseWriter, r *http.Request) {
	var c ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &c)

	log.Println("Write request received: ", c)

	n.BadgerLock.Lock()
	badgerObject, err := n.BadgerWrite(c)
	n.BadgerLock.Unlock()

	resp := APIResp{}

	if n.hasFailed() {
		log.Printf("Request failed for node %v, fail count: %v\n", n.Id, n.FailCount)
		w.WriteHeader(500)
		resp.Status = SIMULATE_FAIL
		resp.Error = "Node temporary failed."
		return
	}

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
	query := r.URL.Query()

	userId := query.Get("id") //! type string

	log.Println("Read Request received with key: ", userId)

	badgerObject, err := n.BadgerRead(userId)

	resp := APIResp{}

	if n.hasFailed() {
		log.Printf("Request failed for node %v, fail count: %v\n", n.Id, n.FailCount)
		w.WriteHeader(500)
		resp.Status = SIMULATE_FAIL
		resp.Error = "Node temporary failed."
		return
	}

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

func (n *Node) SimulateFailRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	count, err := strconv.Atoi(query.Get("count")) //! type string
	if err != nil {
		log.Println("Error with simluate fail request", err)
	}

	n.FailCount = count
}

/* Calculate new node position and send position to new node */
func (n *Node) handleJoinRequest(w http.ResponseWriter, r *http.Request) {
	// calculate new node position
	newPos := n.GetNewPosition()

	w.Header().Set("Content-Type", "application/json")

	if newPos == -1 {
		// ring is full, send error to new node
		log.Printf("Error: cannot find position for new node, ring is full")
		w.WriteHeader(409)
		return
	}
	w.WriteHeader(200)
	// send back newPos
}

func (n *Node) handleJoinOffer() {

}

func (n *Node) handleJoinBroadcast() {

}
