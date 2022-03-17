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
	json.Unmarshal(body, &c)

	fmt.Println(c)

	// TODO update user items in badger DB
	err := n.BadgerWrite([]ClientCart{c})
	resp := APIResp{}
	if err != nil {
		w.WriteHeader(500)
		resp.Status = FAIL
		resp.Error = err

		fmt.Println(err)
	} else {
		w.WriteHeader(201)
		resp.Data = c
		resp.Status = SUCCESS
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(resp)
	w.Write(jsonResp)
	fmt.Println(err)
	fmt.Println("Write Request received: ", c)
}

func (n *Node) FulfilReadRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	userId := query.Get("id") //! type string

	// TODO retrieve user items from badger DB
	c, err := n.BadgerRead(userId)

	resp := APIResp{}
	if err != nil {
		w.WriteHeader(500)
		resp.Status = FAIL
		resp.Error = err
		fmt.Println(err)
	} else {
		w.WriteHeader(200)
		resp.Data = c
		resp.Status = SUCCESS
	}

	// am going to hardcode the response for now since no integration to badger yet
	w.Header().Set("Content-Type", "application/json")

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}
