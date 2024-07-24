package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

func newServer(ctx context.Context, addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", handleUsers)
	mux.HandleFunc("/users/{user}", handleUser)
	mux.HandleFunc("/healthcheck", healthcheck)

	return &http.Server{
		Addr:        addr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     mux,
	}
}

func handleUsers(w http.ResponseWriter, req *http.Request) {
	users, err := getAll(req.Context())
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}

func handleUser(w http.ResponseWriter, req *http.Request) {
	name := req.PathValue("user")
	if name == "" {
		http.Error(w, "Invalid user name", http.StatusNotFound)
		return
	}

	u, err := getUserID(req.Context(), name)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	if u == nil {
		http.Error(w, "Unknown user", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(u); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}

func healthcheck(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "healthy")
}
