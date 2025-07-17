package raft

import (
	"encoding/json"
	"net/http"
)

func (n *Node) HandleRequestVote(w http.ResponseWriter, r *http.Request) {
	var args RequestVoteArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	reply := RequestVoteReply{
		Term:        n.term,
		VoteGranted: false,
	}

	// 요청 Term이 내 Term보다 낮으면 거절
	if args.Term < n.term {
		json.NewEncoder(w).Encode(reply)
		return
	}

	// Term 갱신 필요 시
	if args.Term > n.term {
		n.term = args.Term
		n.votedFor = ""
		n.state = Follower
		n.resetElectionTimer()
	}

	// 아직 투표 안 했거나, 이전에 요청한 후보에게 투표한 경우
	if n.votedFor == "" || n.votedFor == args.CandidateID {
		n.votedFor = args.CandidateID
		reply.VoteGranted = true
		reply.Term = n.term
		n.voteCh <- true // voteCh에 알림 → 타이머 리셋
	}

	json.NewEncoder(w).Encode(reply)
}

func (n *Node) HandleAppendEntries(w http.ResponseWriter, r *http.Request) {
	var args AppendEntriesArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	reply := AppendEntriesReply{
		Term:    n.term,
		Success: false,
	}

	// 요청 term이 내 term보다 작으면 거절
	if args.Term < n.term {
		json.NewEncoder(w).Encode(reply)
		return
	}

	// 더 높은 term이면 Follower 전환
	if args.Term > n.term {
		n.term = args.Term
		n.votedFor = ""
		n.state = Follower
	}

	n.resetElectionTimer()
	n.appendCh <- true

	reply.Term = n.term
	reply.Success = true
	json.NewEncoder(w).Encode(reply)
}
