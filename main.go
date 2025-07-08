package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewStore()
	handler := NewHandler(store)

	http.HandleFunc("/get", handler.GetHandler)
	http.HandleFunc("/set", handler.SetHandler)

	log.Println("Distributed KV Server Started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
