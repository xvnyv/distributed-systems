package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func (n *Node) FulfilWriteRequest(w http.ResponseWriter, r *http.Request) {
	var c ClientCart
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &c)
	fmt.Println(c)

	// TODO update user items in badger DB
	n.BadgerWrite([]ClientCart{c})

	fmt.Println(err)
	fmt.Println("Write Request received: ", c)
}

func (n *Node) FulfilReadRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	userId := query.Get("id") //! type string

	// TODO retrieve user items from badger DB
	clientCart, err := n.BadgerRead(userId)
	if err != nil {
		log.Panic(err.Error())
	}
	// am going to hardcode the response for now since no integration to badger yet
	if userId == "123" {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		// var itemObj ItemObject = ItemObject{
		// 	Id:       1,
		// 	Name:     "Pen",
		// 	Quantity: 44,
		// }

		// var testData ClientCartDTO = ClientCartDTO{
		// 	UserID:      "hello",
		// 	Item:        itemObj,
		// 	VectorClock: []int{1, 0, 234, 347, 2, 34, 6, 6, 235, 7},
		// }
		jsonResp, err := json.Marshal(clientCart)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}
	w.WriteHeader(404)
}
