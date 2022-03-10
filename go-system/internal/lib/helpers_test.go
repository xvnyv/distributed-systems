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

func TestObjectIsEqual(t *testing.T) {
	var testData DataObject = DataObject{
		Key:         "hello",
		Value:       "world",
		VectorClock: []int{1, 0, 234, 347, 2, 34, 6, 6, 235, 7},
	}

	if !testData.IsEqual(testData) {
		t.Errorf("Expected %v, got %v", true, false)
	}
}
