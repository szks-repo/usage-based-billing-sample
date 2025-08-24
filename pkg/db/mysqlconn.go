package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/szks-repo/usage-based-billing-sample/pkg/types/ctxkey"
)

var conn *sql.DB

func MustInit() {
	db, err := open()
	if err != nil {
		panic(err)
	}
	db.SetConnMaxIdleTime(time.Minute)
	db.SetConnMaxLifetime(time.Minute)
	db.SetMaxOpenConns(100)

	conn = db
}

func Get() *sql.DB {
	return conn
}

func Close() {
	if conn != nil {
		conn.Close()
	}
}

func open() (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(localhost:3306)/usage_based_billing?charset=utf8mb4&parseTime=true", "user", "password"))
	if err != nil {
		return nil, err
	}
	return db, nil
}

type DBConnection interface {
	PrepareContext(ctx context.Context, query string, args ...any) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func GetTxn(ctx context.Context) (DBConnection, error) {
	if txn, ok := ctx.Value(ctxkey.Txn{}).(DBConnection); ok {
		return txn, nil
	}
	return nil, errors.New("txn not set")
}
