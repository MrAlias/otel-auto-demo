package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
)

var listenAddr = flag.String("addr", ":8081", "server listen address")

func main() {
	flag.Parse()

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	srv, err := newServer(ctx, *listenAddr)
	if err != nil {
		log.Fatal("Failed to start Authorization server:", err)
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	log.Printf("Authorization server listening at %s ...", *listenAddr)

	select {
	case err = <-errCh:
	case <-ctx.Done():
		err = srv.Shutdown(context.Background())
	}
	if err != nil {
		log.Print("Authorization server error:", err)
	}
	log.Print("Authorization server stopped.")
}
