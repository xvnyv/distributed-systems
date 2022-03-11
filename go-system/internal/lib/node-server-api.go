package lib

import (
	"fmt"
	"log"
	"net/http"
)

/* Write would need: ItemID, ItemName, ItemQuantity and UserID.

Read will obtain information from UserID.
*/

func (n *Node) HandleRequests() {
	// Internal API
	// http.HandleFunc("/write-request", n.handleWriteRequest)
	// http.HandleFunc("/read-request", handleMessage2)
	// http.HandleFunc("/write-success", n.handleWriteRequest)
	// http.HandleFunc("/read-success", n.handleWriteRequest)
	// http.HandleFunc("/join-request", n.handleWriteRequest)
	// http.HandleFunc("/join-broadcast", n.handleWriteRequest)
	// http.HandleFunc("/join-offer", n.handleWriteRequest)

	// http.HandleFunc("/handover-request", n.handleWriteRequest)
	// http.HandleFunc("/handover-success", n.handleWriteRequest)

	http.HandleFunc("/update", n.handleUpdate)
	http.HandleFunc("/get", n.handleGet)
	http.HandleFunc("/write", n.handleGet)
	http.HandleFunc("/data-migration", n.handleGet)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", n.Port), nil))
}
