package lib

import (
	"fmt"
	"strconv"
	"testing"
)

var testNode Node = Node{Id: 1, Ip: "hello"}

var testData DataObject = DataObject{
	Key:         "hello",
	Value:       "world",
	VectorClock: []int{1, 0, 234, 347, 2, 34, 6, 6, 235, 7},
}

var testDataArray = []DataObject{testData}

var keylst []string = make([]string, 0)

func TestBadgerReadWriteDelete(t *testing.T) {

	err := testNode.BadgerWrite(testDataArray)
	if err != nil {
		t.Fatal("Write Failed")
	}

	newDataobject, err := testNode.BadgerRead(testData.Key)

	if err != nil {
		t.Errorf("ggwp %v ", err)
	}

	if !newDataobject.IsEqual(testData) {
		t.Errorf("Expected %v, got %v", testData, newDataobject)
	}

	testNode.BadgerDelete([]string{testData.Key})
	newDataobject, err = testNode.BadgerRead(testData.Key)
	if err.Error() != "Key not found" {
		t.Errorf("ggwp %v ", err)
	}
}

func TestBadgerGetKeys(t *testing.T) {
	numberOfTestObjects := 100

	dataObjectlst := make([]DataObject, 0)
	for i := 0; i < numberOfTestObjects; i++ {
		tempObject := DataObject{
			Key:         "adsfh" + strconv.Itoa(i),
			Value:       fmt.Sprintf("adsfh%v", i),
			VectorClock: []int{i, i, i, i, i, i, i, i},
		}
		keylst = append(keylst, tempObject.Key)
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
