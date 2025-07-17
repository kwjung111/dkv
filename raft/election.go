package raft

type RequestVoteArgs struct {
	Term        int    `json:"term"`
	CandidateID string `json:"candidate_id"`
}

type RequestVoteReply struct {
	Term        int  `json:"term"`
	VoteGranted bool `json:"vote_granted"`
}
