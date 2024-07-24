package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
)

var listenAddr = flag.String("addr", ":8083", "server listen address")

func main() {
	flag.Parse()

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log.Printf("Starting User server at %s ...", *listenAddr)
	srv := newServer(ctx, *listenAddr)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	log.Println("User server started.")

	var err error
	select {
	case err = <-errCh:
	case <-ctx.Done():
		err = srv.Shutdown(context.Background())
	}
	if err != nil {
		log.Print("User server error:", err)
	}
	log.Print("User server stopped.")
}
