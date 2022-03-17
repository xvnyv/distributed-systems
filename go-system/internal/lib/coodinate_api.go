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

var respChannel = make(chan APIResp, 10)

func (n *Node) sendWriteRequest(c ClientCart, node NodeData, successCount *int, mutex *sync.Mutex) {
	jsonData, _ := json.Marshal(c)
	resp, err := http.Post(fmt.Sprintf("%s/write", node.Ip), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	var apiResp APIResp
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode == 201 {
		mutex.Lock()
		*successCount++
		if *successCount == MIN_WRITE_SUCCESS {
			respChannel <- apiResp
		}
		mutex.Unlock()
	}

	if resp.StatusCode == 500 {
		// log error for debugging
		//TODO handle if 2 pass and 1 fail
		// extra mile: handle programmatic error

		log.Fatal(apiResp.Error)
	}

	// TODO: HANDLE FAILURE TIMEOUT SCENARIO WHEN DEALING WITH HINTED HANDOFF
	// Retrieve response body
	defer resp.Body.Close()
	fmt.Println("Write request response body:", body)

}

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

/* Write would need: ItemID, ItemName, ItemQuantity and UserID.

Read will obtain information from UserID.
*/

func (n *Node) handleWriteRequest(w http.ResponseWriter, r *http.Request) {
	var c ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &c)
	if err != nil {
		w.WriteHeader(400)
		log.Fatal(err)
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

	fmt.Println("Write request success for user id:", c.UserID)
	w.Header().Set("Content-Type", "application/json")
	resp := <-respChannel
	json.NewEncoder(w).Encode(resp)
}

func (n *Node) handleReadRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userId := query.Get("id")

	pos := HashMD5(userId)
	responsibleNodes := n.GetResponsibleNodes(pos)
	n.sendReadRequests(userId, responsibleNodes)

	w.Header().Set("Content-Type", "application/json")
	resp := <-respChannel
	json.NewEncoder(w).Encode(resp)
}

func handleMessage2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage2!")
	fmt.Println("Endpoint Hit: homePage2")
}

func (n *Node) sendReadRequest(key string, node NodeData, successCount *int, mutex *sync.Mutex) {
	// Add key to query params
	base, _ := url.Parse(fmt.Sprintf("%s/read?id=%s", node.Ip, key)) // key = userID

	// Send read request to node
	resp, err := http.Get(base.String())
	if err != nil {
		log.Println("Error: ", err)
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

		log.Fatal(apiResp.Error)
	}
	// TODO: DECIDE ON SCHEMA FOR RESPONSE DATA
	defer resp.Body.Close()

	// TODO: HANDLE FAILURE TIMEOUT SCENARIO

	// TODO: SHOULD WE INCLUDE A FAILURE RESPONSE IN THE INTERNAL API SO THAT THE COORDINATOR CAN DIFFERENTIATE BETWEEN
	// ERRORS WHILE PROCESSING VS FAILED NODES? will it actually make sense for the coordinator to retry when there are
	// errors or should it immediately return an error to the client? note: this applies to both read and write requests

	// internal error 500 for error while processing; failed node no response -- timeout
}

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

// ==========END COORDINATOR FUNCTIONS==========

func (n *Node) HandleRequests() {
	// Internal API
	http.HandleFunc("/write-request", n.handleWriteRequest)
	http.HandleFunc("/read-request", n.handleReadRequest)
	// http.HandleFunc("/write-success", handleMessage2)
	http.HandleFunc("/read-success", handleMessage2)
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

func get() {
	resp, err := http.Get("http://localhost:8000/write-request/hello")
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	log.Printf(sb)
	fmt.Print(sb)
}
