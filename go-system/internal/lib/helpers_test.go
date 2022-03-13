package lib

import "testing"

func TestHashMD5(t *testing.T) {
	s := "testing"
	got := HashMD5(s)
	want := 1

	if got != want {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestDataObjectIsEqual(t *testing.T) {
	var testData DataObject = DataObject{
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
