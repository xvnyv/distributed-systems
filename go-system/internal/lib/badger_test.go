package lib

import (
	"reflect"
	"strconv"
	"testing"
)

var testNode Node = Node{Id: 1, Ip: "hello"}

/**
type DataObject struct {
	UserID      string
	Items       map[int]ItemObject
	VectorClock []int
}

type ItemObject struct {
	Id       int
	Name     string
	Quantity int
}
*/

var testData DataObject = DataObject{
	UserID:      "hello",
	Items:       map[int]ItemObject{testItem.Id: testItem},
	VectorClock: []int{1, 0, 234, 347, 2, 34, 6, 6, 235, 7},
}

var testItem ItemObject = ItemObject{
	Id:       3,
	Name:     "hello",
	Quantity: 5,
}

var testDataArray = []DataObject{testData}

var keylst []string = make([]string, 0)

func TestBadgerReadWriteDelete(t *testing.T) {

	err := testNode.BadgerWrite(testDataArray)
	if err != nil {
		t.Fatal("Write Failed")
	}

	newDataobject, err := testNode.BadgerRead(testData.UserID)

	if err != nil {
		t.Errorf("Read error: %v ", err)
	}

	//try deep equal
	if !reflect.DeepEqual(newDataobject, testData) {
		t.Errorf("Expected %v, got %v", testData, newDataobject)
	}

	testNode.BadgerDelete([]string{testData.UserID})
	newDataobject, err = testNode.BadgerRead(testData.UserID)
	if err.Error() != "Key not found" {
		t.Errorf("ggwp %v ", err)
	}
}

func TestBadgerGetKeys(t *testing.T) {
	numberOfTestObjects := 100

	dataObjectlst := make([]DataObject, 0)
	for i := 0; i < numberOfTestObjects; i++ {
		tempObject := DataObject{
			UserID:      "adsfh" + strconv.Itoa(i),
			Items:       map[int]ItemObject{i: ItemObject{Id: i, Name: "object" + strconv.Itoa(i), Quantity: i}},
			VectorClock: []int{i, i, i, i, i, i, i, i},
		}
		keylst = append(keylst, tempObject.UserID)
		dataObjectlst = append(dataObjectlst, tempObject)
	}
	err := testNode.BadgerWrite(dataObjectlst)
	if err != nil {
		t.Errorf("ggwp %v ", err)
	}
	result, err := testNode.BadgerGetKeys()
	if err != nil {
		t.Errorf("ggwp %v ", err)
	}
	if !UnorderedStringArrayEqual(result, keylst) {
		t.Errorf("Expected %v, got %v", keylst, result)
	}

	err = testNode.BadgerDelete(keylst)
	if err != nil {
		t.Errorf("ggwp %v ", err)
	}

}
