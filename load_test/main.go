package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

type Object struct {
	Id       int
	Name     string
	Quantity int
}

type WriteData struct {
	UserID      string
	Item        map[int]Object
	VectorClock map[int]int
}

type AttackData struct {
	metrics   vegeta.Metrics
	histogram vegeta.Histogram
	encoder   vegeta.Encoder
}

const BASE_URL = "http://localhost:8080"

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

const FULL_BAR = 25
const BAR_SYMBOL = "#"
const NUM_RECORDS = 1000

var ch chan AttackData

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
		id := rand.Intn(NUM_RECORDS)
		tgt.URL = fmt.Sprintf("%s/read-request?id=%d", BASE_URL, id)
		return nil
	}
}

func NewWriteTargeter(numNodes int, clockVal int) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		tgt.Method = "POST"
		tgt.URL = fmt.Sprintf("%s/write-request", BASE_URL)

		id := rand.Intn(NUM_RECORDS)
		randomCart := getRandomCart()
		vectorClock := map[int]int{}
		for i := 0; i < numNodes; i++ {
			vectorClock[i] = clockVal
		}
		data := WriteData{UserID: strconv.Itoa(id), Item: randomCart, VectorClock: vectorClock}
		bytesData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling write data")
		}
		tgt.Body = bytesData
		return nil
	}
}

func attack(fname string, typeStr string, attacker *vegeta.Attacker, targeter vegeta.Targeter, rate vegeta.Rate, duration time.Duration) {
	var histogram = vegeta.Histogram{
		Buckets: []time.Duration{
			0,
			5 * time.Millisecond,
			10 * time.Millisecond,
			20 * time.Millisecond,
			30 * time.Millisecond,
			50 * time.Millisecond,
			100 * time.Millisecond,
			500 * time.Millisecond,
			time.Second,
		},
	}
	out, err := os.Create(fmt.Sprintf("results_bin2/%s-%s.bin", fname, typeStr))
	if err != nil {
		fmt.Printf("Error creating %s-%s.bin: %s", fname, typeStr, err)
	}
	enc := vegeta.NewEncoder(out)
	attackData := AttackData{encoder: enc, histogram: histogram}
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		attackData.metrics.Add(res)
		attackData.histogram.Add(res)
		err := attackData.encoder.Encode(res)
		if err != nil {
			fmt.Println("Error encoding results: ", err)
		}
	}
	attackData.metrics.Close()
	ch <- attackData
}

func main() {
	fnameFlag := flag.String("f", "results", "Enter filename to store attack results")
	typeFlag := flag.String("t", "read", "Enter request type: read,write")
	requestRateFlag := flag.Int("r", 1, "Enter request rate (/s)")
	numNodesFlag := flag.Int("n", 5, "Enter number of nodes")
	clockFlag := flag.Int("c", 0, "Enter vector clock value to be set")
	flag.Parse()

	var targeter vegeta.Targeter
	var secondTargeter vegeta.Targeter
	var rate vegeta.Rate
	switch *typeFlag {
	case "read":
		targeter = NewReadTargeter()
		rate = vegeta.Rate{Freq: *requestRateFlag, Per: time.Second}
	case "write":
		targeter = NewWriteTargeter(*numNodesFlag, *clockFlag)
		rate = vegeta.Rate{Freq: *requestRateFlag, Per: time.Second}
	case "read-write":
		targeter = NewReadTargeter()
		secondTargeter = NewWriteTargeter(*numNodesFlag, *clockFlag)
		rate = vegeta.Rate{Freq: *requestRateFlag / 2, Per: time.Second}
	default:
		fmt.Println("Error: invalid request type: ", *typeFlag)
		os.Exit(1)
	}

	duration := time.Minute
	// duration := 10 * time.Second

	// keepAlive := vegeta.KeepAlive(false)
	// attacker := vegeta.NewAttacker(keepAlive)
	timeout := vegeta.Timeout(time.Minute)
	attacker := vegeta.NewAttacker(timeout)

	ch = make(chan AttackData, 2)
	numAttackers := 1
	go attack(*fnameFlag, *typeFlag, attacker, targeter, rate, duration)
	if *typeFlag == "read-write" {
		go attack(*fnameFlag, *typeFlag, attacker, secondTargeter, rate, duration)
		numAttackers++
	}

	metricsArr := []vegeta.Metrics{}
	histogramArr := []vegeta.Histogram{}
	for i := 0; i < numAttackers; i++ {
		attackData := <-ch
		metricsArr = append(metricsArr, attackData.metrics)
		histogramArr = append(histogramArr, attackData.histogram)
	}

	textOut, err := os.Create(fmt.Sprintf("results_text2/%s.txt", *fnameFlag))
	if err != nil {
		fmt.Printf("Error creating %s.txt: %s", *fnameFlag, err)
	}

	jsonOut, err := os.Create(fmt.Sprintf("results_json2/%s.json", *fnameFlag))
	if err != nil {
		fmt.Printf("Error creating %s.json: %s", *fnameFlag, err)
	}

	earliest := metricsArr[0].Earliest.String()
	avgRate := metricsArr[0].Rate
	totalRequests := metricsArr[0].Requests
	successRequests := float64(metricsArr[0].Requests) * metricsArr[0].Success
	successRate := metricsArr[0].Success
	avgLatency := metricsArr[0].Latencies.Mean
	statusCodes := metricsArr[0].StatusCodes

	for i, metrics := range metricsArr {
		if i > 0 {
			avgRate += metrics.Rate
			totalRequests += metrics.Requests
			successRequests += (float64(metrics.Requests) * metrics.Success)
			successRate = (successRate + metrics.Success) / 2
			avgLatency = (avgLatency + metrics.Latencies.Mean) / 2
			for code, cnt := range metrics.StatusCodes {
				if curCnt, ok := statusCodes[code]; ok {
					curCnt += cnt
					statusCodes[code] = curCnt
				} else {
					statusCodes[code] = cnt
				}
			}
		}
	}

	metricsJsonBytes, _ := json.Marshal(metricsArr)
	jsonOut.Write(metricsJsonBytes)
	jsonOut.Close()

	metricsString := ""
	metricsString += fmt.Sprintf("First request sent at: %s\n", earliest)
	metricsString += fmt.Sprintf("Number of nodes: %d\n", *numNodesFlag)
	metricsString += fmt.Sprintf("Request type: %s\n", *typeFlag)
	metricsString += fmt.Sprintf("Average request rate: %.2f/s\n\n", avgRate)
	metricsString += fmt.Sprintf("Total requests: %v\n", totalRequests)
	metricsString += fmt.Sprintf("Successful requests: %v\n", successRequests)
	metricsString += fmt.Sprintf("Success rate: %v\n", successRate)
	metricsString += fmt.Sprintf("Average latency: %v\n", avgLatency)
	metricsString += "Status Codes: ["
	for code, cnt := range statusCodes {
		metricsString += fmt.Sprintf("%s: %d, ", code, cnt)
	}
	metricsString = strings.TrimSuffix(metricsString, ", ")
	metricsString += "]\n\n"
	metricsString += fmt.Sprintf("%-25s\t%-8s%-10.2s\n", "Buckets", "#", "%")

	for i, cnt := range histogramArr[0].Counts {
		lowerBucket, upperBucket := histogramArr[0].Buckets.Nth(i)
		if len(histogramArr) == 2 {
			cnt += histogramArr[1].Counts[i]
		}
		percentage := float64(cnt) / float64(totalRequests)
		barCnt := int(percentage * float64(FULL_BAR))
		if cnt > 0 && barCnt == 0 {
			barCnt = 1
		}
		bar := ""
		for i := 0; i < barCnt; i++ {
			bar += BAR_SYMBOL
		}
		bucketText := fmt.Sprintf("[%-8s%8s]", lowerBucket+",", upperBucket)
		metricsString += fmt.Sprintf("%-25s\t%-8d%-10.2f%s\n", bucketText, cnt, percentage*100, bar)
	}
	textOut.WriteString(metricsString)
	textOut.Close()
}
