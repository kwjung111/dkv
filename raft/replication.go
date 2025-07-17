package raft

type AppendEntriesArgs struct {
	Term     int    `json:"term"`
	LeaderID string `json:"leader_id"`
	// 실제 로그 복제 구현 시에 아래 필드도 필요:
	// PrevLogIndex, PrevLogTerm, Entries, LeaderCommit
}

type AppendEntriesReply struct {
	Term    int  `json:"term"`
	Success bool `json:"success"`
}
