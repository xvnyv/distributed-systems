package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func (n *Node) FulfilWriteRequest(w http.ResponseWriter, r *http.Request) {
	var dao ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &dao)
	fmt.Println(dao)

	// TODO update user items in badger DB

	fmt.Println(err)
	fmt.Println("Write Request received: ", dao)
}

func (n *Node) FulfilReadRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	userId := query.Get("id") //! type string

	// TODO retrieve user items from badger DB
	// am going to hardcode the response for now since no integration to badger yet
	if userId == "123" {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		var itemObj ItemObject = ItemObject{
			Id:       1,
			Name:     "Pen",
			Quantity: 44,
		}

		var testData ClientCartDTO = ClientCartDTO{
			UserID:      "hello",
			Item:        itemObj,
			VectorClock: []int{1, 0, 234, 347, 2, 34, 6, 6, 235, 7},
		}
		jsonResp, err := json.Marshal(testData)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}
	w.WriteHeader(404)
}
