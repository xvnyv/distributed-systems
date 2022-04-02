package lib

import (
	"fmt"
	"testing"
)

func TestHashMD5(t *testing.T) {
	s := "testing"
	got := HashMD5(s)
	want := 1

	if got != want {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestGetNewPositionOdd(t *testing.T) {
	testNodeMap := NodeMap{
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
	n := Node{Id: 4, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8004), Port: 8004, NodeMap: testNodeMap}
	got := n.GetNewPosition()
	want := 87

	if got != want {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestGetNewPositionEven(t *testing.T) {
	testNodeMap := NodeMap{
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
	}
	n := Node{Id: 2, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8002), Port: 8002, NodeMap: testNodeMap}
	got := n.GetNewPosition()
	want := 75

	if got != want {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestGetNewPositionOneItem(t *testing.T) {
	testNodeMap := NodeMap{
		0: NodeData{
			Id:       0,
			Ip:       fmt.Sprintf("%s:%d", BASE_URL, 8000),
			Position: 0,
		},
	}
	n := Node{Id: 0, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8000), Port: 8000, NodeMap: testNodeMap}
	got := n.GetNewPosition()
	want := 50

	if got != want {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestGetNewPositionFull(t *testing.T) {
	testNodeMap := NodeMap{}
	for i := 0; i < NUM_RING_POSITIONS; i++ {
		testNodeMap[i] = NodeData{
			Id:       i,
			Ip:       fmt.Sprintf("%s:%d", BASE_URL, 8000+i),
			Position: i,
		}
	}
	n := Node{Id: 2, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8002), Port: 8002, NodeMap: testNodeMap}
	got := n.GetNewPosition()
	want := -1

	if got != want {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestShouldMigrateData(t *testing.T) {
	// testing true
	n := Node{Id: 2, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8002), Port: 8002, Position: 25, NodeMap: TEST_NODE_MAP}
	got := n.ShouldMigrateData(12)
	want := true

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}

	// testing false
	n = Node{Id: 2, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8002), Port: 8002, Position: 25, NodeMap: TEST_NODE_MAP}
	got = n.ShouldMigrateData(50)
	want = false

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}

	// testing loopback
	n = Node{Id: 0, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8000), Port: 8000, Position: 0, NodeMap: TEST_NODE_MAP}
	got = n.ShouldMigrateData(75)
	want = true

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}
}

func TestShouldDeleteData(t *testing.T) {
	// testing first delete node
	n := Node{Id: 2, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8002), Port: 8002, Position: 25, NodeMap: TEST_NODE_MAP}
	got := n.ShouldDeleteData(12)
	want := true

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}

	// testing last delete node + loopback
	n = Node{Id: 0, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8000), Port: 8000, Position: 0, NodeMap: TEST_NODE_MAP}
	got = n.ShouldDeleteData(25)
	want = true

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}

	// testing false
	n = Node{Id: 4, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8004), Port: 8004, Position: 12, NodeMap: TEST_NODE_MAP}
	got = n.ShouldDeleteData(25)
	want = false

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}
}

func TestCalculateDeleteKeyset(t *testing.T) {
	// with loopback
	n := Node{Id: 1, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8001), Port: 8001, Position: 50, NodeMap: TEST_NODE_MAP}
	gotStart, gotEnd := n.CalculateKeyset(DELETE)
	wantStart := 75
	wantEnd := 0

	if gotStart != wantStart {
		t.Errorf("Expected %d, got %d", wantStart, gotStart)
	}

	if gotEnd != wantEnd {
		t.Errorf("Expected %d, got %d", wantEnd, gotEnd)
	}

	// without loopback
	n = Node{Id: 3, Ip: fmt.Sprintf("%s:%d", BASE_URL, 8003), Port: 8003, Position: 75, NodeMap: TEST_NODE_MAP}
	gotStart, gotEnd = n.CalculateKeyset(DELETE)
	wantStart = 0
	wantEnd = 12

	if gotStart != wantStart {
		t.Errorf("Expected %d, got %d", wantStart, gotStart)
	}

	if gotEnd != wantEnd {
		t.Errorf("Expected %d, got %d", wantEnd, gotEnd)
	}
}

func TestKeyInRange(t *testing.T) {
	// testing in range normal
	got := KeyInRange("123", 0, 12) // keyPos = 8
	want := true

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}

	// testing in range loopback
	got = KeyInRange("132", 75, 0) // keyPos = 88
	want = true

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}

	// testing not in range loopback
	got = KeyInRange("132", 0, 12)
	want = false

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}

	// testing not in range normal
	got = KeyInRange("123", 75, 0)
	want = false

	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}
}

func TestDataObjectIsEqual(t *testing.T) {
	var testData ClientCart = ClientCart{
		UserID: "hello",
		// Value:       "world",
		VectorClock: []int{1, 0, 234, 347, 2, 34, 6, 6, 235, 7},
	}

	if !testData.IsEqual(testData) {
		t.Errorf("Expected %v, got %v", true, false)
	}
}

func TestOrderedIntArrayEqual(t *testing.T) {
	testIntArray := []int{1, 2, 3, 4, 5}
	if !OrderedIntArrayEqual(testIntArray, testIntArray) {
		t.Errorf("Expected %v, got %v", true, false)
	}
}

func TestUnorderedIntArrayEqual(t *testing.T) {
	testIntArray1 := []int{1, 2, 3, 4, 5}
	testIntArray2 := []int{5, 4, 3, 2, 1}
	if !UnorderedIntArrayEqual(testIntArray2, testIntArray1) {
		t.Errorf("Expected %v, got %v", true, false)
	}
}

func TestUorderedStringArrayEqual(t *testing.T) {
	testStringArray1 := []string{"Hello", "World"}
	testStringArray2 := []string{"World", "Hello"}
	if !UnorderedStringArrayEqual(testStringArray1, testStringArray2) {
		t.Errorf("Expected %v, got %v", true, false)
	}
}

func TestVectorClockIsSmaller(t *testing.T) {
	type testItem struct {
		arr1     []int
		arr2     []int
		expected bool
	}
	testItems := make([]testItem, 3)
	testItems[0] = testItem{[]int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 4}, false}
	testItems[1] = testItem{[]int{1, 2, 3, 4, 4}, []int{1, 2, 3, 4, 4}, true}
	testItems[2] = testItem{[]int{1, 2, 3, 4, 5}, []int{1, 2, 4, 4, 5}, true}
	for i := 0; i < len(testItems); i++ {
		if VectorClockSmaller(testItems[i].arr1, testItems[i].arr2) != testItems[i].expected {
			t.Errorf("array1:  %v, arary2: %v", testItems[i].arr1, testItems[i].arr2)
			t.Errorf("Expected %v, got %v", testItems[i].expected, VectorClockSmaller(testItems[i].arr1, testItems[i].arr2))
		}
	}
}

func TestMax(t *testing.T) {
	type testItem struct {
		a        int
		b        int
		expected int
	}
	testItems := make([]testItem, 5)
	testItems[0] = testItem{1, 3, 3}
	testItems[1] = testItem{1, 1, 1}
	testItems[2] = testItem{4, 1, 4}
	testItems[3] = testItem{-5, -1, -1}
	testItems[4] = testItem{-5, 6, 6}
	for i := 0; i < len(testItems); i++ {
		maxVal := Max(testItems[i].a, testItems[i].b)
		if maxVal != testItems[i].expected {
			t.Errorf("Expected %v, got %v", testItems[i].expected, maxVal)
		}
	}
}

func TestItemObjectEqual(t *testing.T) {
	type testItem struct {
		item1    ItemObject
		item2    ItemObject
		expected bool
	}
	testItems := make([]testItem, 5)
	testItems[0] = testItem{ItemObject{Id: 1, Name: "Popcorn", Quantity: 10}, ItemObject{Id: 1, Name: "Popcorn", Quantity: 10}, true}
	testItems[1] = testItem{ItemObject{Id: 1, Name: "Popcorn", Quantity: 10}, ItemObject{Id: 2, Name: "Popcorn", Quantity: 10}, false}
	testItems[2] = testItem{ItemObject{Id: 1, Name: "Popcorn", Quantity: 10}, ItemObject{Id: 1, Name: "Fish", Quantity: 10}, false}
	testItems[3] = testItem{ItemObject{Id: 1, Name: "Popcorn", Quantity: 10}, ItemObject{Id: 1, Name: "Popcorn", Quantity: 11}, false}
	testItems[4] = testItem{ItemObject{Id: 1, Name: "Popcorn", Quantity: 10}, ItemObject{Id: 2, Name: "Fish", Quantity: 11}, false}
	for i := 0; i < len(testItems); i++ {
		isItemObjectEqual := ItemObjectEqual(testItems[i].item1, testItems[i].item2)
		if isItemObjectEqual != testItems[i].expected {
			t.Errorf("Expected %v, got %v", testItems[i].expected, isItemObjectEqual)
		}
	}
}

func TestItemMapEqual(t *testing.T) {
	type testItem struct {
		item1    map[int]ItemObject
		item2    map[int]ItemObject
		expected bool
	}
	testItems := make([]testItem, 6)
	testItems[0] = testItem{
		item1: map[int]ItemObject{
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 1124},
			1: {Id: 1, Name: "Popcorn", Quantity: 10}},
		item2: map[int]ItemObject{
			1: {Id: 1, Name: "Popcorn", Quantity: 10},
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 1124}},
		expected: true}
	testItems[1] = testItem{ //change quantity
		item1: map[int]ItemObject{
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 1124},
			1: {Id: 1, Name: "Popcorn", Quantity: 10}},
		item2: map[int]ItemObject{
			1: {Id: 1, Name: "Popcorn", Quantity: 10},
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 114}},
		expected: false}
	testItems[2] = testItem{ //change number of items
		item1: map[int]ItemObject{
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 1124},
			1: {Id: 1, Name: "Popcorn", Quantity: 10}},
		item2: map[int]ItemObject{
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 1124}},
		expected: false}
	testItems[3] = testItem{ //change name
		item1: map[int]ItemObject{
			0: {Id: 0, Name: "Popcorn sweeter", Quantity: 1124},
			1: {Id: 1, Name: "Popcorn", Quantity: 10}},
		item2: map[int]ItemObject{
			1: {Id: 1, Name: "Popcorn", Quantity: 10},
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 1124}},
		expected: false}
	testItems[4] = testItem{ //change id inside
		item1: map[int]ItemObject{
			0: {Id: 4, Name: "Popcorn sweet", Quantity: 1124},
			1: {Id: 1, Name: "Popcorn", Quantity: 10}},
		item2: map[int]ItemObject{
			1: {Id: 1, Name: "Popcorn", Quantity: 10},
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 1124}},
		expected: false}
	testItems[5] = testItem{ //change id outside
		item1: map[int]ItemObject{
			4: {Id: 0, Name: "Popcorn sweet", Quantity: 1124},
			1: {Id: 1, Name: "Popcorn", Quantity: 10}},
		item2: map[int]ItemObject{
			1: {Id: 1, Name: "Popcorn", Quantity: 10},
			0: {Id: 0, Name: "Popcorn sweet", Quantity: 1124}},
		expected: false}
	for i := 0; i < len(testItems); i++ {
		isItemMapEqual := ItemMapEqual(testItems[i].item1, testItems[i].item2)
		if isItemMapEqual != testItems[i].expected {
			t.Errorf("Expected %v, got %v", testItems[i].expected, isItemMapEqual)
		}
	}
}
func TestClientCartEqual(t *testing.T) {
	type testItem struct {
		c1       ClientCart
		c2       ClientCart
		expected bool
	}
	testItems := make([]testItem, 5)
	testItems[0] = testItem{
		c1: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				}},
			VectorClock: []int{1, 2, 3, 4, 5}},
		c2: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				}},
			VectorClock: []int{1, 2, 3, 4, 5}},
		expected: true}
	testItems[1] = testItem{
		c1: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				}},
			VectorClock: []int{1, 2, 3, 4, 5}},
		c2: ClientCart{
			UserID: "3",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				}},
			VectorClock: []int{1, 2, 3, 4, 5}},
		expected: false}
	testItems[2] = testItem{
		c1: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				}},
			VectorClock: []int{1, 2, 3, 4, 5}},
		c2: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				13: {
					Id:       13,
					Name:     "Popcorn",
					Quantity: 451,
				}},
			VectorClock: []int{1, 2, 3, 4, 5}},
		expected: false}
	testItems[3] = testItem{
		c1: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				}},
			VectorClock: []int{1, 2, 3, 6, 5}},
		c2: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				}},
			VectorClock: []int{1, 2, 3, 4, 5}},
		expected: false}
	testItems[4] = testItem{
		c1: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				}},
			VectorClock: []int{1, 2, 3, 6, 5}},
		c2: ClientCart{
			UserID: "3",
			Item: map[int]ItemObject{
				13: {
					Id:       13,
					Name:     "Popcorn",
					Quantity: 451,
				}},
			VectorClock: []int{1, 2, 3, 6, 5}},
		expected: false}
	for i := 0; i < len(testItems); i++ {
		isClientCartEqual := ClientCartEqual(testItems[i].c1, testItems[i].c2)
		if isClientCartEqual != testItems[i].expected {
			t.Errorf("Expected %v, got %v", testItems[i].expected, isClientCartEqual)
		}
	}
}

func TestMergeClientCarts(t *testing.T) {
	type testItem struct {
		c1       ClientCart
		c2       ClientCart
		expected ClientCart
	}
	testItems := make([]testItem, 5)
	testItems[0] = testItem{
		//test whether two different merge => A vs B => [A,B]
		c1: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				12: {
					Id:       12,
					Name:     "Pencil",
					Quantity: 123,
				},
			},
			VectorClock: []int{1, 2, 3, 4, 5},
		}, c2: ClientCart{
			UserID: "4",
			Item: map[int]ItemObject{
				15: {
					Id:       15,
					Name:     "Orange",
					Quantity: 123,
				},
			},
			VectorClock: []int{1, 2, 3, 7, 4},
		}, expected: ClientCart{
			UserID: "4",
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
			UserID: "4",
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
			UserID: "4",
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
			UserID: "4",
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
			UserID: "4",
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
			UserID: "4",
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
			UserID: "4",
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
		clientCartsEq := ClientCartEqual(MergeClientCarts(testItems[i].c1, testItems[i].c2), testItems[i].expected)
		if !clientCartsEq {
			t.Errorf("test Number %v", i)
			t.Errorf("Expected %v, got %v", testItems[i].expected, MergeClientCarts(testItems[i].c1, testItems[i].c2))
		}
	}
}
