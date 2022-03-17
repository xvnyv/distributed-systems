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

func (n *Node) handleWriteRequest(w http.ResponseWriter, r *http.Request) {
	var dao ClientCartDTO
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &dao)
	fmt.Println(err)

	hashKey := HashMD5(dao.UserID) % 10
	responsibleNodes := n.GetResponsibleNodes(hashKey)
	var coordWriteReqMutex sync.Mutex
	successfulWriteCount := 0
	go n.sendWriteRequest(dao, responsibleNodes[0], &successfulWriteCount, coordWriteReqMutex)

	for {
		coordWriteReqMutex.Lock()
		if successfulWriteCount >= 1 { //! for now, there is no replication, only 1 node
			break
		}
		coordWriteReqMutex.Unlock()
	}
	fmt.Println("Write request success for user id:", dao.UserID)

	// writeRequestMessage := Message{
	// 	Id:         1,
	// 	Sender:     n.Id,
	// 	Receiver:   responsibleNodes[0].Id,
	// 	Type:       WriteRequest,
	// 	MetaData:   strconv.Itoa(hashKey), // placeholder for the hash key
	// 	itemObject: dao.Items,
	// }

	// requestBody, _ := json.Marshal(writeRequestMessage)
	// writeRequestUrl := fmt.Sprintf("http://%s/write-request", n.Ip)

	// res, err := http.Post(writeRequestUrl, "application/json", bytes.NewReader(requestBody))

	// if err != nil {
	// 	fmt.Println(err)
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }
	// defer res.Body.Close()
	// resBody, _ := ioutil.ReadAll(res.Body)
	// // Echo response back to Frontend
	// if res.StatusCode == 200 {
	// 	fmt.Println("Successfully wrote to node. Response:", string(resBody))
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte(string(resBody)))
	// } else {
	// 	fmt.Println("Failed to write to node. Reason:", string(resBody))
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte(string(resBody)))
	// }
}

func (n *Node) handleReadRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userId := query.Get("id")

	hashKey := HashMD5(userId) % 10
	responsibleNodes := n.GetResponsibleNodes(hashKey)
	var coordWriteReqMutex sync.Mutex
	successfulReadCount := 0
	dao, err := n.sendReadRequest(userId, responsibleNodes[0], &successfulReadCount, coordWriteReqMutex)

	if err != nil {
		log.Println("Error: ", err)
		return
	}

	for {
		coordWriteReqMutex.Lock()
		if successfulReadCount >= 1 { //! for now, there is no replication, only 1 node
			break
		}
		coordWriteReqMutex.Unlock()
	}
	fmt.Println("Read request success:", dao)
}

func handleMessage2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage2!")
	fmt.Println("Endpoint Hit: homePage2")
}

func (n *Node) sendWriteRequest(dao ClientCartDTO, node NodeData, successCount *int, mutex sync.Mutex) {
	jsonData, _ := json.Marshal(dao)
	resp, err := http.Post(fmt.Sprintf("%s/write", node.Ip), "application/json", bytes.NewBuffer(jsonData))
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

func (n *Node) sendWriteRequests(object ClientCartDTO, nodes [REPLICATION_FACTOR]NodeData) {
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

func (n *Node) sendReadRequest(key string, node NodeData, successCount *int, mutex sync.Mutex) (ClientCartDTO, error) {
	// Add key to query params
	base, _ := url.Parse(fmt.Sprintf("%s/read", node.Ip))
	params := url.Values{}
	params.Add("key", key)
	fmt.Println(params)
	base.RawQuery = params.Encode()

	var dao ClientCartDTO

	// Send read request to node
	res, err := http.Get(base.String())
	if err != nil {
		log.Println("Error: ", err)
		return dao, err
	}

	// TODO: DECIDE ON SCHEMA FOR RESPONSE DATA
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	fmt.Println("Read request response body:", body)
	decodeError := json.Unmarshal(body, &dao)
	fmt.Println(decodeError)

	// TODO: HANDLE FAILURE TIMEOUT SCENARIO

	// TODO: SHOULD WE INCLUDE A FAILURE RESPONSE IN THE INTERNAL API SO THAT THE COORDINATOR CAN DIFFERENTIATE BETWEEN
	// ERRORS WHILE PROCESSING VS FAILED NODES? will it actually make sense for the coordinator to retry when there are
	// errors or should it immediately return an error to the client? note: this applies to both read and write requests

	mutex.Lock()
	*successCount++
	mutex.Unlock()
	return dao, nil
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
	var object ClientCartDTO
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
	pos := HashMD5(object.UserID)
	fmt.Printf("Position: %d\n", pos)

	// Get nodes that should store this object
	responsibleNodes := n.GetResponsibleNodes(pos)
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
	responsibleNodes := n.GetResponsibleNodes(pos)
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
	http.HandleFunc("/get", n.FulfilReadRequest)
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
