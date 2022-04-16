package lib

import "fmt"

const (
	NUM_RING_POSITIONS int    = 100
	REPLICATION_FACTOR int    = 3 // increase replication factor after base features are completed
	MIN_READ_SUCCESS   int    = 2
	MIN_WRITE_SUCCESS  int    = 2
	BASE_URL           string = "http://127.0.0.1"
	LOAD_BALANCER_PORT int    = 8080
)

var TEST_NODE_MAP NodeMap = NodeMap{
	0: NodeData{
		Id:       0,
		Ip:       fmt.Sprintf("%s:%d", BASE_URL, 8000),
		Position: 0,
	},
	50: NodeData{
		Id:       1,
		Ip:       fmt.Sprintf("%s:%d", BASE_URL, 8001),
		Position: 50,
	},
	25: NodeData{
		Id:       2,
		Ip:       fmt.Sprintf("%s:%d", BASE_URL, 8002),
		Position: 25,
	},
	75: NodeData{
		Id:       3,
		Ip:       fmt.Sprintf("%s:%d", BASE_URL, 8003),
		Position: 75,
	},
	12: NodeData{
		Id:       4,
		Ip:       fmt.Sprintf("%s:%d", BASE_URL, 8004),
		Position: 12,
	},
}

type STATUS_TYPE int

const (
	FAIL STATUS_TYPE = iota
	SUCCESS
)
