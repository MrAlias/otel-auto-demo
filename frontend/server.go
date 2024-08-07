package main

import (
	"context"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
)

func newServer(ctx context.Context, addr string) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/rolldice/", http.HandlerFunc(rolldice))
	mux.Handle("/rolldice/{player}", http.HandlerFunc(rolldice))

	return &http.Server{
		Addr:        addr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     mux,
	}
}

func rolldice(w http.ResponseWriter, _ *http.Request) {
	roll := 1 + rand.Intn(6)

	resp := strconv.Itoa(roll) + "\n"
	_, _ = io.WriteString(w, resp)
}
