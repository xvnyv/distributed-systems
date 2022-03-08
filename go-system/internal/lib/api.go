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
	"sort"
	"sync"
)

func handleWriteRequest(w http.ResponseWriter, r *http.Request) {
	var object DataObject
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &object)
	fmt.Println(err)
	fmt.Println(object)
	query := r.URL.Query()
	fmt.Printf("%v\n", query)
	fmt.Fprintf(w, "Welcome to the HomePage! Object - %v", object)
	fmt.Println("Endpoint Hit: homePage")
}

func handleMessage2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage2!")
	fmt.Println("Endpoint Hit: homePage2")
}

// ==========START COORDINATOR FUNCTIONS==========

func (n *Node) getResponsibleNodes(keyPos int) [REPLICATION_FACTOR]NodeData {
	posArr := []int{}

	for pos, _ := range n.NodeMap {
		posArr = append(posArr, pos)
	}

	sort.Ints(posArr)
	fmt.Printf("Key position: %d\n", keyPos)
	firstNodePosIndex := -1
	for i, pos := range posArr {
		if keyPos <= pos {
			fmt.Printf("First node position: %d\n", pos)
			firstNodePosIndex = i
			break
		}
	}
	if firstNodePosIndex == -1 {
		firstNodePosIndex = 0
	}

	responsibleNodes := [REPLICATION_FACTOR]NodeData{}
	for i := 0; i < REPLICATION_FACTOR; i++ {
		responsibleNodes[i] = n.NodeMap[posArr[(firstNodePosIndex+i)%len(posArr)]]
	}
	return responsibleNodes
}

func (n *Node) sendWriteRequest(object DataObject, node NodeData, successCount *int, mutex sync.Mutex) {
	jsonData, _ := json.Marshal(object)
	resp, err := http.Post(fmt.Sprintf("%s/write-request", node.Ip), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error: ", err)
		return
	}
	// TODO: HANDLE FAILURE TIMEOUT SCENARIO WHEN DEALING WITH HINTED HANDOFF

	// Retrieve response body
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Write request response body:", body)

	mutex.Lock()
	*successCount++
	mutex.Unlock()
}

func (n *Node) sendWriteRequests(object DataObject, nodes [REPLICATION_FACTOR]NodeData) {
	var coordWriteReqMutex sync.Mutex
	successfulWriteCount := 0

	for _, node := range nodes {
		go n.sendWriteRequest(object, node, &successfulWriteCount, coordWriteReqMutex)
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

func (n *Node) sendReadRequest(key string, node NodeData, successCount *int, mutex sync.Mutex) {
	// Add key to query params
	base, _ := url.Parse(fmt.Sprintf("%s/read-request", node.Ip))
	params := url.Values{}
	params.Add("key", key)
	fmt.Println(params)
	base.RawQuery = params.Encode()

	// Send read request to node
	resp, err := http.Get(base.String())
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	// TODO: DECIDE ON SCHEMA FOR RESPONSE DATA
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Read request response body:", body)

	// TODO: HANDLE FAILURE TIMEOUT SCENARIO

	// TODO: SHOULD WE INCLUDE A FAILURE RESPONSE IN THE INTERNAL API SO THAT THE COORDINATOR CAN DIFFERENTIATE BETWEEN
	// ERRORS WHILE PROCESSING VS FAILED NODES? will it actually make sense for the coordinator to retry when there are
	// errors or should it immediately return an error to the client? note: this applies to both read and write requests

	mutex.Lock()
	*successCount++
	mutex.Unlock()
}

func (n *Node) sendReadRequests(key string, nodes [REPLICATION_FACTOR]NodeData) {
	var coordReadReqMutex sync.Mutex
	successfulReadCount := 0

	for _, node := range nodes {
		go n.sendReadRequest(key, node, &successfulReadCount, coordReadReqMutex)
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

func (n *Node) handleUpdate(w http.ResponseWriter, r *http.Request) {
	// Get object from body
	var object DataObject
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	err := json.Unmarshal(body, &object)
	if err != nil {
		log.Println("Error:", err)
		// Send error response to client
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": "error: request body could not be parsed"})
	}
	fmt.Println(object)

	// Get position of object on ring
	pos := HashMD5(object.Key)
	fmt.Printf("Position: %d\n", pos)

	// Get nodes that should store this object
	responsibleNodes := n.getResponsibleNodes(pos)
	fmt.Printf("Responsible nodes: %v\n", responsibleNodes)

	// Send write requests
	n.sendWriteRequests(object, responsibleNodes)

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
}

func (n *Node) handleGet(w http.ResponseWriter, r *http.Request) {
	// Get key from query parameters
	query := r.URL.Query()
	key, ok := query["key"] // key is of type []string but we should only be expecting 1 value
	if !ok {
		log.Printf("Error: Key is not specified")
		// Send error response to client
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": "error: query key was not specified"})
	}
	fmt.Printf("%v\n", query)

	// Get position of key on ring
	pos := HashMD5(key[0])
	fmt.Printf("Position: %d\n", pos)

	// Get nodes that should store this object
	responsibleNodes := n.getResponsibleNodes(pos)
	fmt.Printf("Responsible nodes: %v\n", responsibleNodes)

	// Send read requests
	n.sendReadRequests(key[0], responsibleNodes)
	fmt.Println("yes done sending requests")

	// Send response to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
}

// ==========END COORDINATOR FUNCTIONS==========

func (n *Node) HandleRequests() {
	// Internal API
	http.HandleFunc("/write-request", handleWriteRequest)
	http.HandleFunc("/read-request", handleMessage2)
	http.HandleFunc("/write-success", handleWriteRequest)
	http.HandleFunc("/read-success", handleWriteRequest)
	http.HandleFunc("/join-request", handleWriteRequest)
	http.HandleFunc("/join-broadcast", handleWriteRequest)
	http.HandleFunc("/join-offer", handleWriteRequest)
	http.HandleFunc("/data-migration", handleWriteRequest)
	http.HandleFunc("/handover-request", handleWriteRequest)
	http.HandleFunc("/handover-success", handleWriteRequest)

	// External API
	http.HandleFunc("/update", n.handleUpdate)
	http.HandleFunc("/get", n.handleGet)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", n.Port), nil))
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
