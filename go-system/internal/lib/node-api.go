package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (n *Node) FulfilWriteRequest(w http.ResponseWriter, r *http.Request) {
	var dao DataObject
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

	w.WriteHeader(200)
	w.Write([]byte(userId))
}
