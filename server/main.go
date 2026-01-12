package main

import (
	"log"
	"net/http"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatalln(err)
	}
	err = http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Close()
	if err != nil {
		log.Fatalln(err)
	}
}