package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Node struct {
	mu sync.Mutex

	id       string
	state    State
	term     int
	votedFor string

	log       []LogEntry
	commitIdx int

	peers    []string
	selfAddr string

	electionTimeout   time.Duration
	heartbeatInterval time.Duration

	electionTimer *time.Timer

	//channels
	voteCh   chan bool
	appendCh chan bool
	commitCh chan LogEntry
}

func NewNode(id string, selfAddr string, peers []string) *Node {
	node := &Node{
		id:        id,
		selfAddr:  selfAddr,
		peers:     peers,
		state:     Follower,
		term:      0,
		votedFor:  "",
		log:       make([]LogEntry, 0),
		commitIdx: -1,

		electionTimeout:   randomElectionTimeout(),
		heartbeatInterval: 100 * time.Millisecond,

		voteCh:   make(chan bool),
		appendCh: make(chan bool),
		commitCh: make(chan LogEntry),
	}
	node.resetElectionTimer()
	return node
}

func (n *Node) resetElectionTimer() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.electionTimer != nil {
		n.electionTimer.Stop()
	}
	n.electionTimer = time.NewTimer(n.electionTimeout)
}

func randomElectionTimeout() time.Duration {
	min := 150 * time.Millisecond
	max := 300 * time.Millisecond
	return time.Duration(min + time.Duration((max-min)*time.Duration(time.Now().UnixNano()%int64(max-min))))
}

func (n *Node) run() {
	for {
		switch n.state {
		case Follower:
			select {
			case <-n.voteCh:
				// 투표 받음 : 타이머 리셋
				n.resetElectionTimer()
			case <-n.appendCh:
				// 하트비트 받음 : 타이머 리셋
				n.resetElectionTimer()
			case <-n.electionTimer.C:
				n.mu.Lock()
				n.state = Candidate
				n.mu.Unlock()
				log.Printf("[%s] Election timeout, transitioning to Candidate state", n.id)
			}
		case Candidate:
			n.startElection()
		case Leader:
			n.sendHeartbeats()
			time.Sleep(n.heartbeatInterval)
		}
	}
}

func (n *Node) startElection() {
	n.mu.Lock()
	n.term++
	n.votedFor = n.id
	currentTerm := n.term
	n.state = Candidate
	n.resetElectionTimer()
	n.mu.Unlock()

	log.Printf("[%s] Starting election for term %d", n.id, n.term)

	votes := int32(1) //자기 자신에게 투표
	majority := (len(n.peers) / 2) + 1

	for _, peer := range n.peers {
		go func(peer string) {
			args := RequestVoteArgs{
				Term:        currentTerm,
				CandidateID: n.id,
			}

			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(args); err != nil {
				log.Printf("[%s] Encode Error: %v", n.id, err)
				return
			}

			resp, err := http.Post(fmt.Sprintf("http://%s/raft/request_vote", peer), "application/json", &buf)
			if err != nil {
				log.Printf("[%s] Failed to contact %s: %v", n.id, peer, err)
				return
			}
			defer resp.Body.Close()

			var reply RequestVoteReply
			if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
				log.Printf("[%s] Decode Error from %s: %v", n.id, peer, err)
				return
			}

			if reply.VoteGranted {
				v := atomic.AddInt32(&votes, 1)
				if v >= int32(majority) {
					n.mu.Lock()
					if n.state == Candidate && n.term == currentTerm {
						log.Printf("[%s] Won election for term %d!", n.id, currentTerm)
						n.state = Leader
					}
					n.mu.Unlock()

				} else if reply.Term > currentTerm {
					// 더 높은 Term 응답 : Follower 로 전환함.
					n.mu.Lock()
					n.term = reply.Term
					n.state = Follower
					n.votedFor = ""
					n.resetElectionTimer()
					n.mu.Unlock()
				}
			}
		}(peer)
	}
}
