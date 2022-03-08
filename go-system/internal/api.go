package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	// "github.com/distributed-systems/go-system/lib"
)

// type DataObject struct {
// 	//the thing we store
// 	Key         string `json: key`
// 	Value       string `json: value` //base64
// 	VectorClock []int  `json: context`
// }

// json body = {Key: "heloosadf" , Value: , VectorClock}

func handleWriteRequest(w http.ResponseWriter, r *http.Request) {
	var object DataObject
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &object)
	fmt.Println(err)
	fmt.Println(object)
	query := r.URL.Query()
	fmt.Printf("%v\n", query)
	fmt.Fprintf(w, "Welcome to the HomePage! Query - %v", query)
	fmt.Println("Endpoint Hit: homePage")
}

func handleMessage2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage2!")
	fmt.Println("Endpoint Hit: homePage2")
}

func HandleRequests() {
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
	log.Fatal(http.ListenAndServe(":8000", nil))
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
