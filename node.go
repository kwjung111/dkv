package main

import "time"

//서버
type Node struct {
	state    NodeState
	store    *Store
	hashring Hashring
}

type NodeState struct {
	ID       string
	Address  string
	Healthy  bool
	LastPing time.Time
}

func InitNode(store *Store, state NodeState, hashring Hashring) *Node {
	return &Node{state: state, store: store, hashring: hashring}
}
