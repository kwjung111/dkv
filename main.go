package main

import (
	"log"
	"net/http"
	"time"
)

func main() {

	// for clustering
	selfID := "node-A"

	//register nodes
	SetSelf(selfID)
	RegisterNode("node-A", "localhost:8080")
	RegisterNode("node-B", "localhost:8082")
	RegisterNode("node-C", "localhost:8084")

	store := NewStore()
	handler := NewHandler(store)

	http.HandleFunc("/get", handler.GetHandler)
	http.HandleFunc("/set", handler.SetHandler)
	http.HandleFunc("/ping", handler.PingHandler)

	// for clustering
	go MonitorState()
	go StartHeartBeat(2 * time.Second)

	log.Println("Distributed KV Server Started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
