package main

import (
	"log"
	"net/http"
	"time"
)

func main() {

	// for clustering
	selfID := "node-A"
	cluster := InitCluster()

	//register nodes
	cluster.SetSelf(selfID)
	cluster.RegisterNode("node-A", "localhost:8080")
	cluster.RegisterNode("node-B", "localhost:8082")
	cluster.RegisterNode("node-C", "localhost:8084")

	store := NewStore()
	handler := NewHandler(store)

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
