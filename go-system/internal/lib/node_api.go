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

// func (n *Node) SimulateFailRequest(w http.ResponseWriter, r *http.Request) {
// 	query := r.URL.Query()

// 	count, err := strconv.Atoi(query.Get("count")) //! type string
// 	if err != nil {
// 		log.Println("Error with simluate fail request", err)
// 	}

// 	n.FailCount = count
// }
