package raft

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func (n *Node) sendHeartbeats() {
	n.mu.Lock()
	currentTerm := n.term
	n.mu.Unlock()

	for _, peer := range n.peers {
		go func(peer string) {
			args := AppendEntriesArgs{
				Term:     currentTerm,
				LeaderID: n.id,
			}

			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(args); err != nil {
				log.Printf("[%s] Failed to encode heartbeat: %v", n.id, err)
				return
			}

			resp, err := http.Post("http://"+peer+"/raft/append-entries", "application/json", &buf)
			if err != nil {
				log.Printf("[%s] Failed to send heartbeat to %s: %v", n.id, peer, err)
				return
			}
			defer resp.Body.Close()

			var reply AppendEntriesReply
			if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
				log.Printf("[%s] Invalid heartbeat reply from %s", n.id, peer)
				return
			}

			// 상대 term이 더 높으면 → Follower 전환
			if reply.Term > currentTerm {
				n.mu.Lock()
				n.term = reply.Term
				n.state = Follower
				n.votedFor = ""
				n.resetElectionTimer()
				n.mu.Unlock()
			}
		}(peer)
	}
}
