package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func newServer(ctx context.Context, addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/{name}/alloc", handleAlloc)
	mux.HandleFunc("/healthcheck", healthcheck)

	return &http.Server{
		Addr:        addr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     mux,
	}
}

func handleAlloc(w http.ResponseWriter, req *http.Request) {
	name := req.PathValue("name")

	db, err := openDB()
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = db.Close() }()

	ctx := req.Context()
	tracer := otel.Tracer("user")
	ctx, span := tracer.Start(ctx, "userQuota")
	u, err := useQuota(ctx, db, name)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.End()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	span.End()

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(u); err != nil {
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
