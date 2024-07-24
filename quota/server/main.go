package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/MrAlias/otel-auto-demo/user"
)

const defaultQuota = 5

var (
	listenAddr = flag.String("addr", ":8082", "server listen address")
	usersAddr  = flag.String("users", "http://localhost:8083", "users server address")
)

func main() {
	flag.Parse()

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	uClient := user.NewClient(http.DefaultClient, *usersAddr)

	db, err := openDB()
	if err != nil {
		log.Print("Quota database error:", err)
	}
	err = initDB(ctx, db, uClient, defaultQuota)
	err = errors.Join(err, db.Close())
	if err != nil {
		log.Print("Quota database initialization error:", err)
	}

	sErr := syncUsers(ctx, uClient, defaultQuota)
	rErr := refreshQuotas(ctx, 3*time.Second, 3, defaultQuota*2)

	log.Printf("Starting Quota server at %s ...", *listenAddr)
	srv := newServer(ctx, *listenAddr)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	log.Println("Quota server started.")

	select {
	case err = <-errCh:
		stop()
	case err = <-rErr:
		stop()
	case err = <-sErr:
		stop()
	case <-ctx.Done():
		err = srv.Shutdown(context.Background())

	}
	if err != nil {
		log.Print("Quota server error:", err)
	}
	log.Print("Quota server stopped.")
}

func syncUsers(ctx context.Context, uClient *user.Client, quota int) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)

		db, err := openDB()
		if err != nil {
			errCh <- err
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			ids, err := uClient.AllIDs(ctx)
			if err != nil {
				continue
			}

			for _, id := range ids {
				_, err := getQuota(ctx, db, id)
				if errors.Is(err, errUser) {
					_ = setQuota(ctx, db, id, quota)
				}
			}
		}
	}()
	return errCh
}

func refreshQuotas(ctx context.Context, d time.Duration, incr, ceil int) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)

		ticker := time.NewTicker(d)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				db, err := openDB()
				if err != nil {
					errCh <- err
					return
				}
				err = resetQuotas(ctx, db, incr, ceil)
				if err != nil {
					log.Print("Failed to reset quotas: ", err)
				}
			}
		}
	}()
	return errCh
}
