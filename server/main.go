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
	"strings"

	"github.com/gorilla/websocket"
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
var getUserStatement *sql.Stmt
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}
var conns map[int]*websocket.Conn

func main() {
	var err error
	var schema *sql.Stmt

	db, err = sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatalln("sql.Open() error:", err)
	}
	defer db.Close()

	schema, err = db.Prepare(`
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	key TEXT UNIQUE
);
	`)
	if err != nil {
		log.Fatalln("db.Prepare() error:", err)
	}
	_, err = schema.Exec()
	if err != nil {
		log.Fatalln("schema.Exec() error:", err)
	}
	schema.Close()

	insertUserStatement, err = db.Prepare(`
INSERT INTO users(username, password) VALUES(?, ?)
	`)
	if err != nil {
		log.Fatalln("db.Prepare() error:", err)
	}
	defer insertUserStatement.Close()

	getPasswordStatement, err = db.Prepare(`
SELECT password FROM users WHERE username = ?
	`)
	if err != nil {
		log.Fatalln("db.Prepare() error:", err)
	}
	defer getPasswordStatement.Close()

	updateKeyStatement, err = db.Prepare(`
UPDATE users SET key = ? WHERE username = ?
	`)
	if err != nil {
		log.Fatalln("db.Prepare() error:", err)
	}
	defer updateKeyStatement.Close()

	getUserStatement, err = db.Prepare(`
SELECT id, username FROM users WHERE key = ?
	`)
	if err != nil {
		log.Fatalln("db.Prepare() error:", err)
	}
	defer getUserStatement.Close()

	conns = make(map[int]*websocket.Conn)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/ws", wsHandler)

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
		log.Println("io.ReadAll() error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &login)
	if err != nil {
		log.Println("json.Unmarshal() error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if login.Username == "" || login.Password == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err = getPasswordStatement.QueryRow(login.Username).Scan(&password)
	if err != nil {
		log.Println("getPasswordStatement.QueryRow().Scan() error:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if login.Password != password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	b := make([]byte, 32)
	rand.Read(b)
	key := hex.EncodeToString(b)
	_, err = updateKeyStatement.Exec(key, login.Username)
	if err != nil {
		log.Println("updateKeyStatement.Exec() error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"key\":\"%s\"}", key)))
	log.Printf("User %s logged in\n", login.Username)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var id int
	var username string
	var key string
	var conn *websocket.Conn
	var err error
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	key = auth[len(prefix):]
	err = getUserStatement.QueryRow(key).Scan(&id, &username)
	if err != nil {
		log.Println("getIDStatement.QueryRow().Scan() error:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	_, err = updateKeyStatement.Exec(username, "")
	if err != nil {
		log.Println("updateKeyStatement.Exec() error:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrader.Upgrade() error:", err)
		return
	}
	defer conn.Close()
	conns[id] = conn
	defer delete(conns, id)
	log.Printf("User %s connected\n", username)
	// listen
	log.Printf("User %s disconnected\n", username)
}