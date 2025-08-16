package db

import (
	"database/sql"
	"fmt"
	"time"
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
	db, err := sql.Open("mysql", fmt.Sprintf("%s:@tcp(localhost:3306)/usage_based_billing?charset=utf8mb4&parseTime=true", "user"))
	if err != nil {
		return nil, err
	}
	return db, nil
}
