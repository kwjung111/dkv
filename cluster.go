package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Status int

const (
	Unkown Status = iota
	Initializing
	Running
)

func (s Status) String() string {
	switch s {
	case Unkown:
		return "Unkown"
	case Initializing:
		return "Initializing"
	case Running:
		return "Running"
	default:
		return "Unknown Status"
	}
}

type Cluster struct {
	State *ClusterState
}

type ClusterState struct {
	mu       sync.Mutex
	selfID   string
	LeaderID string
	Status   Status
	Nodes    map[string]*NodeState
}

func InitCluster(id string) *Cluster {
	state := ClusterState{Status: Initializing, LeaderID: "None", selfID: id, Nodes: make(map[string]*NodeState)}
	return &Cluster{State: &state}
}

func (c *Cluster) SetLeader(id string) {
	c.State.mu.Lock()
	defer c.State.mu.Unlock()
	c.State.LeaderID = id
}

func (c *Cluster) GetLeader() string {
	return c.State.LeaderID
}

func (c *Cluster) GetState() ([]byte, error) {
	state := struct {
		SelfID   string            `json:"self_id"`
		LeaderID string            `json:"leader_id"`
		Status   Status            `json:"status"`
		Nodes    map[string]string `json:"nodes"`
	}{
		SelfID:   c.State.selfID,
		LeaderID: c.State.LeaderID,
		Status:   c.State.Status,
		Nodes:    make(map[string]string),
	}
	return json.Marshal(state)
}

func (c *Cluster) heartbeat(node *NodeState) {
	timeoutSec := 2
	client := http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/ping", node.Address))
	c.State.mu.Lock()
	defer c.State.mu.Unlock()

	if err != nil || resp.StatusCode != http.StatusOK {
		node.Healthy = false
	} else {
		node.Healthy = true
		node.LastPing = time.Now()
	}
}

func (c *Cluster) RegisterNode(id string, address string) {
	c.State.mu.Lock()
	defer c.State.mu.Unlock()
	newState := NodeState{
		ID:       id,
		Address:  address,
		Healthy:  false,
		LastPing: time.Now(),
	}
	c.State.Nodes[id] = &newState
}

func (c *Cluster) StartHeartBeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		var wg sync.WaitGroup
		c.State.mu.Lock()
		for id, node := range c.State.Nodes {
			if id == c.State.selfID {
				continue
			}
			wg.Add(1)
			go func(node *NodeState) {
				defer wg.Done()
				c.heartbeat(node)
			}(node)
		}
		c.State.mu.Unlock()
		wg.Wait()
	}
}

func (c *Cluster) MonitorState() {
	for {
		time.Sleep(5 * time.Second)
		c.State.mu.Lock()
		fmt.Println("===cluster status===")
		for _, node := range c.State.Nodes {
			if node.ID == c.State.selfID {
				continue
			} else {
				fmt.Printf("[%s] Status : %v , %v, LastPing : %v,\n", node.ID, node.Healthy, node.Address, node.LastPing)
			}
		}
		c.State.mu.Unlock()
	}
}

func (c *Cluster) CheckState() {
	log.Println("클러스터 상태 질의 시작")
	for _, node := range c.State.Nodes {
		if node.Healthy {
			//타임아웃 로직 필요
			resp, err := http.Get(fmt.Sprintf("http://%s/cluster/state", node.Address))
			if err != nil || resp.StatusCode != http.StatusOK {
				continue
			} else {
				//클러스터 메타데이터 파싱 후 동기화
			}
		}
	}
}

// 리더 선출 시작
func StartElection() {

}

// 리더 확인
func Check() {
	// 클러스터에 연결되어있다고 가정 시
	// 클러스터 상태 질의
	// 클러스터에 노드가 없으면 자동으로 리더 승격
	// 리더 없으면 리더 선출 시작
	//
}
