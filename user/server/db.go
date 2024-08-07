package main

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"

	"github.com/MrAlias/otel-auto-demo/user/internal"
)

const dsn = "file:user.db?cache=shared&mode=rwc&_busy_timeout=9999999"

const (
	createTable = `CREATE TABLE users (
		id    INTEGER PRIMARY KEY,
		name  TEXT NOT NULL,
		quota INTEGER
	);`

	queryAll  = `SELECT id,name,quota FROM users`
	queryUser = queryAll + ` WHERE name = ?`

	updateQuota = `UPDATE users SET quota = ? WHERE id = ?`
	insertUser  = `INSERT INTO users (name, quota) VALUES (?, ?)`
	decrement   = `UPDATE users SET quota = quota - 1 WHERE name = ? AND quota > 0 RETURNING id, quota`
)

func openDB() (*sql.DB, error) {
	return sql.Open("sqlite3", dsn)
}

func initDB(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, createTable)
	return err
}

func addQuota(ctx context.Context, db *sql.DB, incr, ceil int) (err error) {
	opts := &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false}
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, queryAll)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var u internal.User
		err = rows.Scan(&u.ID, &u.Name, &u.Quota)
		if err != nil {
			return err
		}

		if u.Quota < ceil {
			newQuota := max(ceil, u.Quota+incr)
			_, err = tx.ExecContext(ctx, updateQuota, newQuota, u.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var (
	errInsufficient = errors.New("insufficient quota")
	errUser         = errors.New("unknown user")
)

func useQuota(ctx context.Context, db *sql.DB, name string) (internal.User, error) {
	u := internal.User{Name: name}
	err := db.QueryRowContext(ctx, decrement, name).Scan(&u.ID, &u.Quota)
	if errors.Is(err, sql.ErrNoRows) {
		err = errUser
	}
	if err == nil && u.Quota == 0 {
		err = errInsufficient
	}
	return u, err
}

func addUser(ctx context.Context, db *sql.DB, name string, quota int) error {
	_, err := db.ExecContext(ctx, insertUser, name, quota)
	return err
}

func getUser(ctx context.Context, db *sql.DB, name string) (internal.User, error) {
	u := internal.User{Name: name}
	err := db.QueryRowContext(ctx, queryUser, name).Scan(&u.ID, &u.Quota)
	if errors.Is(err, sql.ErrNoRows) {
		err = errUser
	}
	return u, err
}
