package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Cluster struct {
	State *ClusterState
}

type ClusterState struct {
	mu     sync.Mutex
	selfID string
	Nodes  map[string]*NodeState
}

func InitCluster() *Cluster {
	state := ClusterState{Nodes: make(map[string]*NodeState)}
	return &Cluster{State: &state}
}

func (c *Cluster) SetSelf(id string) {
	c.State.selfID = id
}

func (c *Cluster) heartbeat(node *NodeState) {
	resp, err := http.Get(fmt.Sprintf("http://%s/ping", node.Address))
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
	for {
		c.State.mu.Lock()
		for id, node := range c.State.Nodes {
			if id == c.State.selfID {
				continue
			}
			go c.heartbeat(node)
		}
		c.State.mu.Unlock()
		time.Sleep(interval)
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
