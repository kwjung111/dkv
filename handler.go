package main

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	store   *Store
	cluster *Cluster
}

func NewHandler(s *Store, c *Cluster) *Handler {
	return &Handler{store: s, cluster: c}
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}
	val, ok := h.store.Get(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte(val))
}

func (h *Handler) SetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	h.store.Set(req.Key, req.Value)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ClusterStateHandler(w http.ResponseWriter, r *http.Request) {
	data, err := h.cluster.GetState()
	if err != nil {
		http.Error(w, "failed to get cluster state", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
