package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const dsn = "file:auth?mode=memory&cache=shared"

func init() {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}

	createTbl := `CREATE TABLE auth (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`
	_, err = db.Exec(createTbl)
	if err != nil {
		panic(err)
	}

	addUser := `INSERT INTO 'auth' ('name') VALUES ('Alice');`
	_, err = db.Exec(addUser)
	if err != nil {
		panic(err)
	}
}

type server struct {
	*http.Server

	db *sql.DB
}

func newServer(ctx context.Context, addr string) (*server, error) {
	srv := new(server)

	var err error
	srv.db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/{user}", srv.check)

	srv.Server = &http.Server{
		Addr:        addr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     mux,
	}
	return srv, nil
}

const userQuery = "SELECT 1 FROM auth WHERE name = ?"

func (s *server) check(w http.ResponseWriter, req *http.Request) {
	user := req.PathValue("user")
	if user == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var i int
	err := s.db.QueryRowContext(req.Context(), userQuery, user).Scan(&i)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	default:
		fmt.Fprint(w, "Authorized")
	}
}
