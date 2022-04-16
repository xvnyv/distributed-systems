package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

var joinWg sync.WaitGroup

func (n *Node) HandleRequests() {
	// Internal API
	http.HandleFunc("/read", n.FulfilReadRequest)
	http.HandleFunc("/write", n.FulfilWriteRequest)
	// http.HandleFunc("/simulate-fail", n.SimulateFailRequest)
	http.HandleFunc("/join-request", n.handleJoinRequest)
	http.HandleFunc("/join-broadcast", n.handleJoinBroadcast)
	// http.HandleFunc("/handover-request", handleMessage2)
	// http.HandleFunc("/handover-success", handleMessage2)

	// External API
	http.HandleFunc("/write-request", n.handleWriteRequest)
	http.HandleFunc("/read-request", n.handleReadRequest)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%v", n.Port),
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func (n *Node) UpdateNginx() {
	servers := ""
	for _, nodeData := range n.NodeMap {
		servers += fmt.Sprintf("\t\tserver localhost:%d;\n", nodeData.Port)
	}
	if runtime.GOOS != "windows" {
		confPathCmd := "ps aux | grep nginx | grep \"[c]onf\""

		// get conf file path
		out, err := exec.Command("bash", "-c", confPathCmd).Output()
		if err != nil {
			log.Fatalf("Error getting nginx conf file path: %s\n", err)
		}
		confPath := string(out)
		if confPath == "" {
			// nginx not currently running
			log.Fatal("Error: nginx is not running")
		}
		confPath = strings.TrimSpace(strings.Split(confPath, "nginx -c ")[1])

		writeFileCmd := fmt.Sprintf(`cat << EOF > '%s'
events {}

http {
	upstream powerpuffgirls {
		%s
	}

	server {
		listen 8080;
		server_name localhost;
		location / {
			proxy_pass http://powerpuffgirls;
			proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
         # Simple requests
         if (\$request_method ~* "(GET|POST)") {
            add_header "Access-Control-Allow-Origin"  'http://localhost:3000' always;
         }
         # Preflighted requests
         if (\$request_method = OPTIONS ) {
            add_header "Access-Control-Allow-Origin"  'http://localhost:3000' always;
            add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
            add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
            return 200;
         }
		}
	}
}
`, confPath, servers)

		// write new config to file
		out, err = exec.Command("bash", "-c", writeFileCmd).Output()
		if err != nil {
			log.Fatalf("Error writing new config to file: %s\n", err)
		}

		// stop nginx
		out, err = exec.Command("nginx", "-s", "stop").Output()
		if err != nil {
			log.Fatalf("Error stopping nginx: %s\n", err)
		}

		// start nginx with new config
		out, err = exec.Command("nginx", "-c", confPath).Output()
		if err != nil {
			log.Fatalf("Error restarting nginx: %s\n", err)
		}
	}
}

func (n *Node) sendJoinBroadcast(nodeData NodeData, jsonData []byte) {
	resp, err := http.Post(fmt.Sprintf("%s/join-broadcast", nodeData.Ip), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// end program if cannot announce join
		log.Fatalf("Join Broadcast Error: %s\n", err)
	}

	defer resp.Body.Close()
	// parse API response
	apiResp := MigrateResp{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	log.Printf("Received JOIN BROADCAST response from Node %d with status code %d\n", nodeData.Id, resp.StatusCode)

	// handle API response
	if resp.StatusCode != 200 {
		// end program if cannot join
		log.Fatalf("Join Broadcast Error: %s\n", apiResp.Error)
	}

	// store migrated data
	if apiResp.Data != nil {
		n.BadgerMigrateWrite(apiResp.Data)
		log.Printf("%d records stored\n", len(apiResp.Data))
		// for debugging purposes
		keys := []string{}
		for _, data := range apiResp.Data {
			keys = append(keys, data.UserID)
		}
		log.Printf("Keys stored: %v", keys)
	}

	joinWg.Done()
}

func (n *Node) JoinSystem(init bool) {
	if init {
		// first node in the system, initialise itself
		n.Position = 0
		n.NodeMap = map[int]NodeData{n.Position: {Id: n.Id, Ip: n.Ip, Port: n.Port, Position: n.Position}}
		log.Println(n.NodeMap)
		return
	}
	// send request to join
	resp, err := http.Get(fmt.Sprintf("%s:%d/join-request?node=%d", BASE_URL, LOAD_BALANCER_PORT, n.Id))
	if err != nil {
		// end program if cannot join
		log.Fatalf("Join Request Error: %s\n", err)
	}
	defer resp.Body.Close()
	// parse API response
	apiResp := JoinResp{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode != 200 {
		// end program if cannot join
		log.Fatalf("Join Request Error: %s\n", apiResp.Error)
	}

	// set new node's variables
	n.Position = apiResp.Data.Position
	n.NodeMap = apiResp.Data.NodeMap
	log.Printf("Position received: %d\n", n.Position)
	log.Printf("Node map received: %v\n", n.NodeMap)

	newNodeData := NodeData{Id: n.Id, Ip: n.Ip, Port: n.Port, Position: n.Position}

	// skip join-broadcast if node has just returned from temporary failure
	if _, ok := n.NodeMap[n.Position]; !ok {
		jsonData, _ := json.Marshal(newNodeData)
		// announce position to all other nodes
		for _, nodeData := range n.NodeMap {
			joinWg.Add(1)
			go n.sendJoinBroadcast(nodeData, jsonData)
		}
		joinWg.Wait()
		n.NodeMap[n.Position] = newNodeData
	}

	log.Println("Joining process completed")
}
