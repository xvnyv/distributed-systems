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

// TODO shift this into the handleWriteRequest and handleReadRequest function
// leaving it out here as a global channel will mean that the write request handler could potentially read
// messages meant for the read request handler or other concurrent write/read requests
var respChannel = make(chan APIResp, 10)

/* Write would need: ItemID, ItemName, ItemQuantity and UserID.

Read will obtain information from UserID.
*/

// ========== START COORDINATOR WRITE ==========

/* Send individual internal write request to each node */
func (n *Node) sendWriteRequest(c ClientCart, node NodeData, successCount *int, mutex *sync.Mutex) {
	jsonData, _ := json.Marshal(c)
	resp, err := http.Post(fmt.Sprintf("%s/write", node.Ip), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Send Write Request Error: ", err)
		return
	}

	var apiResp APIResp
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode == 201 {
		// TODO change this section to simply send the response directly back to the handler
		mutex.Lock()
		*successCount++
		if *successCount == MIN_WRITE_SUCCESS {
			respChannel <- apiResp
		}
		mutex.Unlock()
	}

	if resp.StatusCode == 500 {
		// log error for debugging
		// TODO handle case of internal failure -- send response code 500 back to client
		log.Printf("Internal API Write Request Error: %v\n", apiResp.Error)
	}

	// TODO: HANDLE FAILURE TIMEOUT SCENARIO WHEN DEALING WITH HINTED HANDOFF
	defer resp.Body.Close()
	log.Println("Write request response body:", body)

}

/* Send requests to all responsible nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) sendWriteRequests(c ClientCart, nodes [REPLICATION_FACTOR]NodeData) {
	var coordWriteReqMutex sync.Mutex
	successfulWriteCount := 0

	for _, node := range nodes {
		go n.sendWriteRequest(c, node, &successfulWriteCount, &coordWriteReqMutex)
	}

	// TODO: ADD CHANNEL HERE TO DETECT TIMEOUT
	for {
		coordWriteReqMutex.Lock()
		if successfulWriteCount >= MIN_WRITE_SUCCESS {
			break
		}
		coordWriteReqMutex.Unlock()
	}
}

/* Message handler for write requests for external API to client application */
func (n *Node) handleWriteRequest(w http.ResponseWriter, r *http.Request) {
	var c ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &c)
	if err != nil {
		w.WriteHeader(400)
		log.Printf("Handle Write Request Error: %v\n", err)
		return
	}
	hashKey := HashMD5(c.UserID)
	responsibleNodes := n.GetResponsibleNodes(hashKey)
	//
	//this parts needs to go to badger to commit the
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
	n.sendWriteRequests(c, responsibleNodes)

	w.Header().Set("Content-Type", "application/json")
	resp := <-respChannel
	json.NewEncoder(w).Encode(resp)
}

// ========== END COORDINATOR WRITE ==========

// ========== START COORDINATOR READ ==========

/* Send individual internal read request to each node */
func (n *Node) sendReadRequest(key string, node NodeData, successCount *int, mutex *sync.Mutex) {
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

	if resp.StatusCode == 200 {
		mutex.Lock()
		*successCount++
		if *successCount == MIN_READ_SUCCESS {
			//TODO handle conflict
			respChannel <- apiResp
		}
		mutex.Unlock()
	}

	if resp.StatusCode == 500 {
		// log error for debugging
		//TODO handle if 2 pass and 1 fail
		// extra mile: handle programmatic error
		log.Println("Internal API Read Request Error:", apiResp.Error)
	}
	// TODO: DECIDE ON SCHEMA FOR RESPONSE DATA
	defer resp.Body.Close()

	// TODO: HANDLE FAILURE TIMEOUT SCENARIO

	// TODO: SHOULD WE INCLUDE A FAILURE RESPONSE IN THE INTERNAL API SO THAT THE COORDINATOR CAN DIFFERENTIATE BETWEEN
	// ERRORS WHILE PROCESSING VS FAILED NODES? will it actually make sense for the coordinator to retry when there are
	// errors or should it immediately return an error to the client? note: this applies to both read and write requests

	// internal error 500 for error while processing; failed node no response -- timeout
}

/* Send requests to all responsible nodes concurrently and wait for minimum required nodes to succeed */
func (n *Node) sendReadRequests(key string, nodes [REPLICATION_FACTOR]NodeData) {
	var coordReadReqMutex sync.Mutex
	successfulReadCount := 0

	for _, node := range nodes {
		go n.sendReadRequest(key, node, &successfulReadCount, &coordReadReqMutex)
	}

	// TODO: ADD CHANNEL HERE TO DETECT TIMEOUT
	for {
		coordReadReqMutex.Lock()
		if successfulReadCount >= MIN_READ_SUCCESS {
			break
		}
		coordReadReqMutex.Unlock()
	}
}

/* Message handler for read requests for external API to client application */
func (n *Node) handleReadRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userId := query.Get("id")

	pos := HashMD5(userId)
	responsibleNodes := n.GetResponsibleNodes(pos)
	n.sendReadRequests(userId, responsibleNodes)

	resp := <-respChannel
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ========== END COORDINATOR READ ==========

func (n *Node) HandleRequests() {
	// Internal API
	http.HandleFunc("/write-request", n.handleWriteRequest)
	http.HandleFunc("/read-request", n.handleReadRequest)
	// http.HandleFunc("/write-success", handleMessage2)
	// http.HandleFunc("/read-success", handleMessage2)
	// http.HandleFunc("/join-request", handleMessage2)
	// http.HandleFunc("/join-broadcast", handleMessage2)
	// http.HandleFunc("/join-offer", handleMessage2)
	// http.HandleFunc("/data-migration", handleMessage2)
	// http.HandleFunc("/handover-request", handleMessage2)
	// http.HandleFunc("/handover-success", handleMessage2)

	// http.HandleFunc("/update", handleMessage2)
	http.HandleFunc("/read", n.FulfilReadRequest)
	http.HandleFunc("/write", n.FulfilWriteRequest)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", n.Port), nil))
}

// func get() {
// 	resp, err := http.Get("http://localhost:8000/write-request/hello")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	//We Read the response body on the line below.
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	//Convert the body to type string
// 	sb := string(body)
// 	log.Printf(sb)
// 	fmt.Print(sb)
// }
