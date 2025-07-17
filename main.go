package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kwjung111/dkv/raft"
)

func main() {

	node := raft.NewNode("node-A", "localhost:8080", []string{
		"localhost:8081",
		"localhost:8082",
	})

	http.HandleFunc("/raft/request-vote", node.HandleRequestVote)
	http.HandleFunc("/raft/append-entries", node.HandleAppendEntries)
}

func main2() {

	// for clustering
	selfID := "node-A"
	cluster := InitCluster(selfID)

	//register nodes
	cluster.RegisterNode("node-A", "localhost:8080")
	cluster.RegisterNode("node-B", "localhost:8082")
	cluster.RegisterNode("node-C", "localhost:8084")
	//리더 선출

	store := NewStore()
	handler := NewHandler(store, cluster)

	http.HandleFunc("/get", handler.GetHandler)
	http.HandleFunc("/set", handler.SetHandler)
	http.HandleFunc("/ping", handler.PingHandler)
	http.HandleFunc("/cluster/state", handler.ClusterStateHandler)

	// for clustering
	go cluster.MonitorState()
	go cluster.StartHeartBeat(2 * time.Second)

	log.Println("Distributed KV Server Started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
