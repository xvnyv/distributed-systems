package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func (n *Node) FulfilWriteRequest(w http.ResponseWriter, r *http.Request) {
	var c ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &c)

	log.Println("Write request received: ", c)

	err := n.BadgerWrite([]ClientCart{c})

	resp := APIResp{}
	if err != nil {
		w.WriteHeader(500)
		resp.Status = FAIL
		resp.Error = err.Error()
	} else {
		w.WriteHeader(201)
		resp.Data = c
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
	log.Println("Write request completed for", c)
}

func (n *Node) FulfilReadRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	userId := query.Get("id") //! type string

	log.Println("Read Request received with key: ", userId)

	c, err := n.BadgerRead(userId)

	resp := APIResp{}
	if err != nil {
		w.WriteHeader(500)
		resp.Status = FAIL
		resp.Error = err.Error()
		log.Printf("Error: %v", err)
	} else {
		w.WriteHeader(200)
		resp.Data = c
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
	log.Println("Read request completed for", c)
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
