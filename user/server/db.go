package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MrAlias/otel-auto-demo/user/internal"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const dsn = "file:users?mode=memory&cache=shared"

const (
	createTable = `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`
	addUser     = `INSERT INTO 'users' ('name') VALUES (?)`
	queryID     = `SELECT id, name FROM 'users' WHERE name = ?`
	queryAll    = `SELECT id, name FROM 'users'`
)

var users = []string{
	"Alice", "Bob", "Carol", "Dan", "Erin", "Faythe", "Grace", "Heidi", "Ivan",
	"Judy", "Mallory", "Niaj", "Olivia", "Peggy", "Rupert", "Sybil", "Trent",
	"Victor", "Walter",
}

func init() {
	var err error
	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	for _, user := range users {
		_, err = db.Exec(addUser, user)
		if err != nil {
			panic(err)
		}
	}
}

func getAll(ctx context.Context) ([]*internal.User, error) {
	rows, err := db.QueryContext(ctx, queryAll)
	if err != nil {
		return nil, err
	}

	var out []*internal.User
	for rows.Next() {
		u := new(internal.User)
		err := rows.Scan(&u.ID, &u.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func getUserID(ctx context.Context, name string) (*internal.User, error) {
	var u internal.User
	err := db.QueryRowContext(ctx, queryID, name).Scan(&u.ID, &u.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}
