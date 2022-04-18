package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

type Object struct {
	Id       int
	Name     string
	Quantity int
}

type WriteData struct {
	UserID string
	Item   map[int]Object
}

var BASE_URL = "http://localhost:8080"
var ITEMS = []Object{
	{Id: 1, Name: "pencil", Quantity: 1},
	{Id: 2, Name: "pen", Quantity: 1},
	{Id: 3, Name: "paper", Quantity: 1},
	{Id: 4, Name: "notebook", Quantity: 1},
	{Id: 5, Name: "backpack", Quantity: 1},
	{Id: 6, Name: "water bottle", Quantity: 1},
	{Id: 7, Name: "eraser", Quantity: 1},
	{Id: 8, Name: "glue", Quantity: 1},
	{Id: 9, Name: "tape", Quantity: 1},
	{Id: 10, Name: "highlighter", Quantity: 1},
}

func getRandomCart() map[int]Object {
	numItems := rand.Intn(5) + 1
	start := rand.Intn(len(ITEMS))
	cart := map[int]Object{}

	var curItem Object
	for i := 0; i < numItems; i++ {
		curItem = ITEMS[(start+i)%len(ITEMS)]
		curItem.Quantity = rand.Intn(10) + 1
		cart[curItem.Id] = curItem
	}
	return cart
}

func file(name string, create bool) (*os.File, error) {
	switch name {
	case "stdin":
		return os.Stdin, nil
	case "stdout":
		return os.Stdout, nil
	default:
		if create {
			return os.Create(name)
		}
		return os.Open(name)
	}
}

func NewReadTargeter() vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		tgt.Method = "GET"

		id := rand.Intn(100)
		// id := 81
		tgt.URL = fmt.Sprintf("%s/read-request?id=%d", BASE_URL, id)
		return nil
	}
}

func NewWriteTargeter() vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		tgt.Method = "POST"
		tgt.URL = fmt.Sprintf("%s/write-request", BASE_URL)

		id := rand.Intn(100)
		randomCart := getRandomCart()
		data := WriteData{UserID: strconv.Itoa(id), Item: randomCart}
		bytesData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling write data")
		}
		tgt.Body = bytesData
		return nil
	}
}

func main() {
	fnameFlag := flag.String("f", "results.bin", "Enter filename to store attack results")
	flag.Parse()
	rate := vegeta.Rate{Freq: 3, Per: time.Second}
	duration := 10 * time.Second

	targeter := NewWriteTargeter()

	keepAlive := vegeta.KeepAlive(false)
	attacker := vegeta.NewAttacker(keepAlive)

	var metrics vegeta.Metrics
	var histogram = vegeta.Histogram{
		Buckets: []time.Duration{
			0,
			100 * time.Millisecond,
			300 * time.Millisecond,
			500 * time.Millisecond,
			700 * time.Millisecond,
			900 * time.Millisecond,
			1 * time.Second,
		},
	}

	out, err := os.Create(*fnameFlag)
	if err != nil {
		fmt.Printf("Error creating %s: %s", *fnameFlag, err)
	}
	enc := vegeta.NewEncoder(out)

	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)
		histogram.Add(res)
		err := enc.Encode(res)
		if err != nil {
			fmt.Println("Error encoding results: ", err)
		}
	}
	metrics.Close()

	fmt.Printf("Average latency: %v\n", metrics.Latencies.Mean)
	fmt.Printf("Success rate: %v\n", metrics.Success)
	fmt.Printf("Total requests: %v\n", metrics.Requests)
	fmt.Printf("Average request rate: %v\n", metrics.Rate)

	fmt.Printf("\n\n%+v  \n", metrics)

	fmt.Printf("Histogram: %+v\n", histogram.Counts)
}
