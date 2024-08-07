package main

import (
	"context"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"

	"github.com/MrAlias/otel-auto-demo/user"
)

type serviceKeyType int

const userKey serviceKeyType = 0

func newServer(ctx context.Context, listenAddr, userAddr string) *http.Server {
	client := user.NewClient(http.DefaultClient, userAddr)
	if err := client.HealthCheck(ctx); err != nil {
		log.Print("Cannot reach User service: ", err)
	} else {
		log.Print("Connected to User service at ", userAddr)
	}
	ctx = context.WithValue(ctx, userKey, client)

	mux := http.NewServeMux()

	mux.Handle("/rolldice/", http.HandlerFunc(rolldice))
	mux.Handle("/rolldice/{player}", http.HandlerFunc(rolldice))

	return &http.Server{
		Addr:        listenAddr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     mux,
	}
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	player := r.PathValue("player")

	client, ok := ctx.Value(userKey).(*user.Client)
	if !ok {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	if err := client.UseQuota(ctx, player); err != nil {
		if errors.Is(err, user.ErrInsufficient) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		} else {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	roll := 1 + rand.Intn(6)

	resp := strconv.Itoa(roll) + "\n"
	_, _ = io.WriteString(w, resp)
}
