package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MrAlias/otel-auto-demo/quota/internal"
	"github.com/MrAlias/otel-auto-demo/user"
	_ "github.com/mattn/go-sqlite3"
)

const dsn = "file:quota.sqlight?cache=shared&mode=rwc&_busy_timeout=9999999"

const (
	createTable = `CREATE TABLE quota (id INTEGER PRIMARY KEY, n INTEGER);`
	queryAll    = `SELECT id,n FROM quota ORDER BY id`
	updateN     = `UPDATE quota SET n = ? WHERE id = ?`
	insertN     = `INSERT INTO quota (id, n) VALUES (?, ?)`
	queryN      = `SELECT n FROM quota WHERE id = ?`
	decrement   = `UPDATE quota SET n = n - 1 WHERE id = ? AND n > 0 RETURNING n`
)

func openDB() (*sql.DB, error) {
	return sql.Open("sqlite3", dsn)
}

func initDB(ctx context.Context, db *sql.DB, uClient *user.Client, quota int) error {
	ids, err := uClient.AllIDs(ctx)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, createTable)
	if err != nil {
		return err
	}

	for _, id := range ids {
		err = errors.Join(err, setQuota(ctx, db, id, quota))
	}
	return err
}

func resetQuotas(ctx context.Context, db *sql.DB, incr, ceil int) (err error) {
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
		var id, n int
		err = rows.Scan(&id, &n)
		if err != nil {
			return err
		}

		if n < ceil {
			n = max(ceil, n+incr)
			_, err = tx.ExecContext(ctx, updateN, n, id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func burnOne(ctx context.Context, db *sql.DB, id int) (*internal.Quota, error) {
	var n int
	err := db.QueryRowContext(ctx, decrement, id).Scan(&n)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &internal.Quota{Remaining: n}, err
}

func setQuota(ctx context.Context, db *sql.DB, id, quota int) error {
	_, err := db.ExecContext(ctx, insertN, id, quota)
	return err
}

var errUser = errors.New("unkown user")

func getQuota(ctx context.Context, db *sql.DB, id int) (*internal.Quota, error) {
	var n int
	err := db.QueryRowContext(ctx, queryN, id).Scan(&n)
	if errors.Is(err, sql.ErrNoRows) {
		err = errUser
	}
	return &internal.Quota{Remaining: n}, err
}
