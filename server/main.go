package main

import (
	"log"
	"net/http"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const addr = ":3333"
var db *sql.DB

func main() {
	var err error
	var schema *sql.Stmt

	db, err = sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatalln(err)
	}
	schema, err = db.Prepare(`
CREATE TABLE IF NOT EXISTS users (
	username TEXT PRIMARY KEY,
	password TEXT NOT NULL,
	key TEXT
);
	`)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = schema.Exec()
	if err != nil {
		log.Fatalln(err)
	}

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Close()
	if err != nil {
		log.Fatalln(err)
	}
}