package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type NodeState struct {
	ID       string
	Address  string
	Healthy  bool
	LastPing time.Time
}

type ClusterState struct {
	mu     sync.Mutex
	selfID string
	Nodes  map[string]*NodeState
}

var cluster = ClusterState{
	Nodes: make(map[string]*NodeState),
}

func heartbeat(node *NodeState) {
	resp, err := http.Get(fmt.Sprintf("http://%s/ping", node.Address))
	cluster.mu.Lock()
	defer cluster.mu.Unlock()

	if err != nil || resp.StatusCode != http.StatusOK {
		node.Healthy = false
	} else {
		node.Healthy = true
		node.LastPing = time.Now()
	}
}

func SetSelf(id string) {
	cluster.selfID = id
}

func RegisterNode(id string, address string) {
	cluster.mu.Lock()
	defer cluster.mu.Unlock()
	newState := NodeState{
		ID:       id,
		Address:  address,
		Healthy:  false,
		LastPing: time.Now(),
	}
	cluster.Nodes[id] = &newState
}

func StartHeartBeat(interval time.Duration) {
	for {
		cluster.mu.Lock()
		for id, node := range cluster.Nodes {
			if id == cluster.selfID {
				continue
			}
			go heartbeat(node)
		}
		cluster.mu.Unlock()
		time.Sleep(interval)
	}
}

func MonitorState() {
	for {
		time.Sleep(5 * time.Second)
		cluster.mu.Lock()
		fmt.Println("===cluster status===")
		for _, node := range cluster.Nodes {
			if node.ID == cluster.selfID {
				continue
			} else {
				fmt.Printf("[%s] Status : %v , %v, LastPing : %v,\n", node.ID, node.Healthy, node.Address, node.LastPing)
			}
		}
		cluster.mu.Unlock()
	}
}
