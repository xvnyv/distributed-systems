package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
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

func getResponsibleNodes(keyPos int, nodeMap NodeMap) [REPLICATION_FACTOR]NodeData {
	posArr := []int{}

	for pos, _ := range nodeMap {
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
		responsibleNodes[i] = nodeMap[posArr[(firstNodePosIndex+i)%len(posArr)]]
	}
	return responsibleNodes
}

func (n *Node) sendWriteRequest(node NodeData) {
}

func (n *Node) sendWriteRequests(nodes [REPLICATION_FACTOR]NodeData) {

}

func (n *Node) handleUpdate(w http.ResponseWriter, r *http.Request) {
	// Get object from body
	var object DataObject
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &object)
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Println(object)

	// Get position of object on ring
	pos := HashMD5(object.Key)
	fmt.Printf("Position: %d\n", pos)

	// Get nodes that should store this object
	responsibleNodes := getResponsibleNodes(pos, n.NodeMap)
	fmt.Printf("Responsible nodes: %v\n", responsibleNodes)

	// Send write requests
	n.sendWriteRequests(responsibleNodes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
}

func (n *Node) handleGet(w http.ResponseWriter, r *http.Request) {

}

func handleTest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func (n *Node) HandleRequests() {
	// Internal API
	http.HandleFunc("/", handleTest)
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

	// log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", n.Port), nil))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 8000), nil))
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
