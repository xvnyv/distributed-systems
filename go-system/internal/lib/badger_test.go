package lib

import "testing"

func TestBadgerWrite(t *testing.T) {
	node := Node{Id: 1, Ip: "hello"}
	node.badger_start()
}
