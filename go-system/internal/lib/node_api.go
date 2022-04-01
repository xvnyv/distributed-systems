package lib

import (
	"encoding/json"
	"fmt"
	"io"
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

	// create response
	apiResp := JoinResp{}
	w.Header().Set("Content-Type", "application/json")

	if newPos == -1 {
		// ring is full, send error to new node
		log.Printf("Error: cannot find position for new node, ring is full")
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
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		// return immediately since APIResp could not be marshalled
		w.WriteHeader(500)
		return
	}
	w.Write(encodedResp)
	log.Println("Join offer sent with assigned position ", newPos)
}

func (n *Node) handleJoinBroadcast(w http.ResponseWriter, r *http.Request) {

}

func (n *Node) HandleRequests() {
	// Internal API
	http.HandleFunc("/read", n.FulfilReadRequest)
	http.HandleFunc("/write", n.FulfilWriteRequest)
	http.HandleFunc("/simulate-fail", n.SimulateFailRequest)
	// http.HandleFunc("/write-success", handleMessage2)
	// http.HandleFunc("/read-success", handleMessage2)
	http.HandleFunc("/join-request", n.handleJoinRequest)
	http.HandleFunc("/join-broadcast", n.handleJoinBroadcast)
	// http.HandleFunc("/data-migration", handleMessage2)
	// http.HandleFunc("/handover-request", handleMessage2)
	// http.HandleFunc("/handover-success", handleMessage2)

	// External API
	http.HandleFunc("/write-request", n.handleWriteRequest)
	http.HandleFunc("/read-request", n.handleReadRequest)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", n.Port), nil))
}

func (n *Node) JoinSystem(init bool) {
	if init {
		// first node in the system, initialise itself
		n.Position = 0
		n.NodeMap = map[int]NodeData{n.Position: {Id: n.Id, Ip: n.Ip, Port: n.Port, Position: n.Position}}
		log.Println(n.NodeMap)
		return
	}
	// send request to join
	resp, err := http.Get(fmt.Sprintf("%s:%d/join-request", BASE_URL, LOAD_BALANCER_PORT))
	if err != nil {
		// end program if cannot join
		log.Fatalf("Join Request Error: %s\n", err)
	}
	defer resp.Body.Close()
	// parse API response
	apiResp := JoinResp{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode == 409 {
		// end program if cannot join
		log.Fatalf("Join Request Error: %s\n", apiResp.Error)
	}

	// set new node's variables
	n.Position = apiResp.Data.Position
	n.NodeMap = apiResp.Data.NodeMap
	n.NodeMap[n.Id] = NodeData{Id: n.Id, Ip: n.Ip, Port: n.Port, Position: n.Position}
	log.Printf("Position received: %d\n", n.Position)
	log.Printf("Node map received: %v\n", n.NodeMap)

	// jsonData, _ := json.Marshal(n.NodeMap[n.Id])
	// // announce position to all other nodes
	// for _, nodeData := range n.NodeMap {
	// 	resp, err := http.Post(fmt.Sprintf("%s/join-broadcast", nodeData.Ip), "application/json", bytes.NewBuffer(jsonData))
	// 	if err != nil {
	// 		// end program if cannot announce join
	// 		log.Fatalf("Join Broadcast Error: %s\n", err)
	// 	}
	// 	defer resp.Body.Close()
	// 	// parse API response
	// 	apiResp := APIResp{}
	// 	body, _ := io.ReadAll(resp.Body)
	// 	json.Unmarshal(body, &apiResp)

	// }
}
