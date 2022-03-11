package lib

type MessageType int

type NodeData struct {
	Id       int
	Ip       string
	Port int
	Position int
}

type NodeMap map[int]NodeData //int refers to position in the ring

type Node struct {
	Id           int
	Ip           string
	Port         int
	Position     int
	NodeMap      NodeMap
	Successors   []int
	Predecessors []int
}

type Message struct {
	Id       int
	Sender   int
	Receiver int
	Type     MessageType
	Content  string
	MetaData string // may contain intended receiver
	itemObject map[int]ItemObject
}


type DataObject struct {
	UserID string
	Items map[int]ItemObject
	VectorClock []int
 }
 
 type ItemObject struct{
 	Id int
	Name string
	Quantity int
 }

const (
	WriteRequest MessageType = iota //Object with optional vector clock
	ReadRequest                     //Key, Maybe need vector clock
	WriteSuccess                    //NodeIds, MessageId 201 or 202 SUCCESS
	ReadSuccess                     //Object, MessageId, 200 SUCCESS
	JoinRequest                     //NodeId
	//new node contact random node to request to join

	JoinOffer //Position, NodeMap
	//response to the JoinRequest message (tells node where the node should be)

	JoinBroadcast //Position
	//to tell all nodes that the new node is in the ring

	DataMigration //[]ObjectData, NodeId
	//after new node joins, this message contains data for new node to store

	HandoverRequest //[]ObjectData, MessageId
	//node containing hinted data trying to hondover data to the dead node

	HandoverSuccess //MessageId
	//this will be sent after dead node revives and stores the hinted data
)
