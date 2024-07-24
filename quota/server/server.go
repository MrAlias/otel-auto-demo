package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
)

func newServer(ctx context.Context, addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/{id}", handleQuota)
	mux.HandleFunc("/user/{id}/burn", handleBurn)
	mux.HandleFunc("/healthcheck", healthcheck)

	return &http.Server{
		Addr:        addr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     mux,
	}
}

func handleQuota(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	if idStr == "" {
		http.Error(w, "Unknown user ID", http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db, err := openDB()
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = db.Close() }()

	quota, err := getQuota(req.Context(), db, id)
	if err != nil {
		if errors.Is(err, errUser) {
			http.Error(w, "Unknown user", http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(quota); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}

func handleBurn(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	if idStr == "" {
		http.Error(w, "Unknown user ID", http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db, err := openDB()
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = db.Close() }()

	quota, err := burnOne(req.Context(), db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(quota); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}

func healthcheck(w http.ResponseWriter, _ *http.Request) {
	db, err := openDB()
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = db.Close() }()

	fmt.Fprint(w, "healthy")
}
