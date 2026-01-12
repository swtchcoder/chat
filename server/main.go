package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const addr = ":3333"
var db *sql.DB
var insertUserStatement *sql.Stmt
var getPasswordStatement *sql.Stmt
var updateKeyStatement *sql.Stmt

func main() {
	var err error
	var schema *sql.Stmt

	db, err = sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	schema, err = db.Prepare(`
CREATE TABLE IF NOT EXISTS users (
	username TEXT PRIMARY KEY,
	password TEXT NOT NULL,
	key TEXT UNIQUE
);
	`)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = schema.Exec()
	if err != nil {
		log.Fatalln(err)
	}
	schema.Close()

	insertUserStatement, err = db.Prepare(`
INSERT INTO users(username, password) VALUES(?, ?)
	`)
	if err != nil {
		log.Fatalln(err)
	}
	defer insertUserStatement.Close()

	getPasswordStatement, err = db.Prepare(`
SELECT password FROM users WHERE username = ?
	`)
	if err != nil {
		log.Fatalln(err)
	}
	defer getPasswordStatement.Close()

	updateKeyStatement, err = db.Prepare(`
UPDATE users SET key = ? WHERE username = ?
	`)
	if err != nil {
		log.Fatalln(err)
	}
	defer updateKeyStatement.Close()

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)

	log.Printf("Listening on %s\n", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error
	var login Login
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	body, err = io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &login)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if login.Username == "" || login.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = insertUserStatement.Exec(login.Username, login.Password) 
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	log.Printf("User %s just registered\n", login.Username)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error
	var login Login
	var password string
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	body, err = io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &login)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if login.Username == "" || login.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = getPasswordStatement.QueryRow(login.Username).Scan(&password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if login.Password != password {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b := make([]byte, 32)
	rand.Read(b)
	key := hex.EncodeToString(b)
	_, err = updateKeyStatement.Exec(key, login.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"key\":\"%s\"}", key)))
	log.Printf("User %s logged in, Key: %s\n", login.Username, key)
}