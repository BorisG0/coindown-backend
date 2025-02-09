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

type CreateSessionRequest struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

type CreateSessionResponse struct {
	Token string `json:"token"`
}

type SessionInfo struct {
	DateTime   string `json:"datetime"`
	CoinResult string `json:"coin_result"`
}

// generateToken creates a random 16-byte token and returns it as a 32-character hex string
func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func createSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate token
	token, err := generateToken()
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store in database
	datetime := req.Date + " " + req.Time
	if _, err := db.Exec(`INSERT INTO sessions (token, datetime) VALUES (?, ?)`, token, datetime); err != nil {
		log.Printf("Failed to insert session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	json.NewEncoder(w).Encode(CreateSessionResponse{Token: token})
	fmt.Printf("Session created: %v, %v\n", token, datetime)
}

func viewSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Get token from URL path
	token := r.URL.Path[len("/session/"):]
	if token == "" {
		http.Error(w, "Token not provided", http.StatusBadRequest)
		return
	}

	// Query the database
	var session SessionInfo
	err := db.QueryRow("SELECT datetime, COALESCE(coin_result, '') FROM sessions WHERE token = ?", token).
		Scan(&session.DateTime, &session.CoinResult)

	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return session info as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
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
	http.HandleFunc("/create", createSession)
	http.HandleFunc("/session/", viewSession)
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("main over")
}
