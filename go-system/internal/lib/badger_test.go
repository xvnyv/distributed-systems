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

var itemObj ItemObject = ItemObject{
	Id:       1,
	Name:     "Pen",
	Quantity: 44,
}

var testData ClientCart = ClientCart{
	UserID: "hello",
	Item: map[int]ItemObject{1: {
		Id:       1,
		Name:     "shift",
		Quantity: 1,
	}},
	VectorClock: []int{1, 0, 234, 347, 2, 34, 6, 6, 235, 7},
}

var keylst []string = make([]string, 0)

func TestBadgerReadWriteDelete(t *testing.T) {

	_, err := testNode.BadgerWrite(testData)
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
	numberOfTestObjects := 10

	for i := 0; i < numberOfTestObjects; i++ {
		tempObject := ClientCart{
			UserID: "adsfh" + strconv.Itoa(i),
			Item: map[int]ItemObject{1: {
				Id:       1,
				Name:     "shift",
				Quantity: 1,
			}},
			VectorClock: []int{i, i, i, i, i, i, i, i},
		}
		keylst = append(keylst, tempObject.UserID)
		_, err := testNode.BadgerWrite(tempObject)
		if err != nil {
			t.Errorf("ggwp %v ", err)
		}
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

func TestWriteConflictClientCarts(t *testing.T) {
	type testItem struct {
		c1       ClientCart
		c2       ClientCart
		expected ClientCart
	}
	testItems := make([]testItem, 3)
	testItems[0] = testItem{
		//test whether two different merge => A vs B => [A,B]
		c1: ClientCart{
			UserID: "5",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				},
			},
			VectorClock: []int{1, 2, 3, 4, 5},
		}, c2: ClientCart{
			UserID: "5",
			Item: map[int]ItemObject{
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 123,
				},
			},
			VectorClock: []int{1, 2, 3, 7, 4},
		}, expected: ClientCart{
			UserID: "5",
			Item: map[int]ItemObject{
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 123,
				},
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				},
			},
			VectorClock: []int{1, 2, 3, 7, 5}, //test whether vector clock updates
		}}
	testItems[1] = testItem{
		//test whether quantity that is larger is taken
		// c1.orange.qty = 12123 vs c2.orange.qty = 123 => c1.orange.qty
		// c1.pencil.qty = 123 vs c2.pencil.qty = 122 => c2.pencil.qty
		c1: ClientCart{
			UserID: "6",
			Item: map[int]ItemObject{
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 12123, //larger
				},
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123, //smaller
				},
			},
			VectorClock: []int{1, 2, 3, 6, 5},
		}, c2: ClientCart{
			UserID: "6",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 126, //larger
				},
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 123, //smaller
				},
			},
			VectorClock: []int{10, 2, 3, 4, 5},
		}, expected: ClientCart{
			UserID: "6",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 126, //larger
				},
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 12123, //larger
				},
			},
			VectorClock: []int{10, 2, 3, 6, 5}, //test whether vector clock updates
		}}
	testItems[2] = testItem{
		//test whether two different merge => [A,B,C] vs [B,C,D] => [A,B,C,D]
		c1: ClientCart{
			UserID: "7",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil", //missing in c2
					Quantity: 123,
				},
				13: {
					Id:       13,
					Name:     "Pen",
					Quantity: 1283, //smaller
				},
				15: {
					Id:       15,
					Name:     "Ruler",
					Quantity: 12905, //larger
				},
			},
			VectorClock: []int{1, 2, 3, 6, 5},
		}, c2: ClientCart{
			UserID: "7",
			Item: map[int]ItemObject{
				13: {
					Id:       13,
					Name:     "Pen",
					Quantity: 12003, //larger
				},
				15: {
					Id:       15,
					Name:     "Ruler",
					Quantity: 1290, //smaller
				},
				14: {
					Id:       14,
					Name:     "scissors", //missing in c1
					Quantity: 1290,
				},
			},
			VectorClock: []int{10, 2, 3, 4, 5},
		}, expected: ClientCart{
			UserID: "7",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil", //missing in c2
					Quantity: 123,
				},
				13: {
					Id:       13,
					Name:     "Pen",
					Quantity: 12003, //larger
				},
				15: {
					Id:       15,
					Name:     "Ruler",
					Quantity: 12905, //larger
				},
				14: {
					Id:       14,
					Name:     "scissors", //missing in c1
					Quantity: 1290,       //
				},
			},
			VectorClock: []int{10, 2, 3, 6, 5}, //test whether vector clock updates
		}}

	for i := 0; i < len(testItems); i++ {
		_, err := testNode.BadgerWrite(testItems[i].c1)
		if err != nil {
			t.Errorf("Writing error: %v", err.Error())
		}
		_, err = testNode.BadgerWrite(testItems[i].c2)
		if err != nil {
			t.Errorf("Writing error: %v", err.Error())
		}
		res, err := testNode.BadgerRead(testItems[i].c1.UserID)
		if err != nil {
			t.Errorf("Reading error: %v", err.Error())
		}

		clientCartsEq := ClientCartEqual(res, testItems[i].expected)

		if !clientCartsEq {
			t.Errorf("test Number %v", i)
			t.Errorf("Expected %v, got %v", testItems[i].expected, res)
		}
	}
}

func TestOverwriteConflictClientCarts(t *testing.T) {
	type testItem struct {
		c1       ClientCart
		c2       ClientCart
		expected ClientCart
	}
	testItems := make([]testItem, 3)
	testItems[0] = testItem{
		//test whether two different merge => A vs B => [A,B]
		c1: ClientCart{
			UserID: "8",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				},
			},
			VectorClock: []int{1, 2, 3, 4, 5}, //smaller vector clock
		}, c2: ClientCart{
			UserID: "8",
			Item: map[int]ItemObject{
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 123,
				},
			},
			VectorClock: []int{1, 2, 3, 7, 5}, //strictly larger vector clock
		}, expected: ClientCart{
			UserID: "8",
			Item: map[int]ItemObject{
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 123,
				},
			},
			VectorClock: []int{1, 2, 3, 7, 5}, //test whether vector clock overwritten by strictly larger
		}}
	testItems[1] = testItem{
		//test whether quantity that is larger is taken
		// c1.orange.qty = 12123 vs c2.orange.qty = 123 => c1.orange.qty
		// c1.pencil.qty = 123 vs c2.pencil.qty = 122 => c2.pencil.qty
		c1: ClientCart{
			UserID: "9",
			Item: map[int]ItemObject{
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 12123, //larger
				},
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123, //smaller
				},
			},
			VectorClock: []int{1, 2, 3, 6, 5},
		}, c2: ClientCart{
			UserID: "9",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 126, //larger
				},
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 123, //smaller
				},
			},
			VectorClock: []int{1, 2, 3, 4, 5},
		}, expected: ClientCart{
			UserID: "9",
			Item: map[int]ItemObject{
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 12123, //larger
				},
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123, //smaller
				},
			},
			VectorClock: []int{1, 2, 3, 6, 5},
		}}
	testItems[2] = testItem{
		c1: ClientCart{ // test delete
			UserID: "75",
			Item: map[int]ItemObject{
				13: {
					Id:       13,
					Name:     "Pen",
					Quantity: 12003, //larger
				},
				15: {
					Id:       15,
					Name:     "Ruler",
					Quantity: 1290, //smaller
				},
				14: {
					Id:       14,
					Name:     "scissors", //missing in c1
					Quantity: 1290,
				},
			},
			VectorClock: []int{10, 2, 3, 4, 5},
		}, c2: ClientCart{
			UserID: "75",
			Item: map[int]ItemObject{
				13: {
					Id:       13,
					Name:     "Pen",
					Quantity: 12003, //larger
				},
				15: {
					Id:       15,
					Name:     "Ruler",
					Quantity: 1290, //smaller
				},
			},
			VectorClock: []int{10, 2, 3, 4, 6},
		}, expected: ClientCart{
			UserID: "75",
			Item: map[int]ItemObject{
				13: {
					Id:       13,
					Name:     "Pen",
					Quantity: 12003, //larger
				},
				15: {
					Id:       15,
					Name:     "Ruler",
					Quantity: 1290, //smaller
				},
			},
			VectorClock: []int{10, 2, 3, 4, 6},
		}}

	for i := 0; i < len(testItems); i++ {
		_, err := testNode.BadgerWrite(testItems[i].c1)
		if err != nil {
			t.Errorf("Writing error: %v", err.Error())
		}
		_, err = testNode.BadgerWrite(testItems[i].c2)
		if err != nil {
			t.Errorf("Writing error: %v", err.Error())
		}
		res, err := testNode.BadgerRead(testItems[i].c1.UserID)
		if err != nil {
			t.Errorf("Reading error: %v", err.Error())
		}

		clientCartsEq := ClientCartEqual(res, testItems[i].expected)

		if !clientCartsEq {
			t.Errorf("test Number %v", i)
			t.Errorf("Expected %v, got %v", testItems[i].expected, res)
		}
	}
}
