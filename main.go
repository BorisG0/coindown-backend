package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var err error

type CreateInfo struct {
	Date string
	Time string
}

// generateToken creates a random 16-byte token and returns it as a 32-character hex string
func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func createLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var createInfo CreateInfo

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&createInfo)
	if err != nil {
		fmt.Println("Error at decode: ", err)
		fmt.Fprintf(w, "input error")
		return
	}

	datetime := createInfo.Date + " " + createInfo.Time

	// Generate a random token
	token, err := generateToken()
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	stmt := `INSERT INTO sessions (token, datetime) VALUES (?, ?)`
	_, err = db.Exec(stmt, token, datetime)
	if err != nil {
		log.Printf("Failed to insert data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Printf("link creation(date: %v, time: %v, token: %v)\n", createInfo.Date, createInfo.Time, token)
	fmt.Fprintf(w, "https://coindown.com/%v", token)
}

func main() {
	db, err = sql.Open("sqlite3", "coindown.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable := `
	create table if not exists sessions (
		token text primary key,
		datetime datetime,
		coin_result text,
		created_at datetime default current_timestamp
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Println("Server started")
	http.HandleFunc("/create", createLink)
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("main over")
}
