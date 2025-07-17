package raft

type LogEntry struct {
	Term    int // Term in which the command was received
	Command string
	Key     string
	Value   string
}
