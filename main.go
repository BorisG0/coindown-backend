package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB
var err error

type CreateSessionRequest struct {
	Timestamp int64 `json:"timestamp"`
}

type CreateSessionResponse struct {
	Token string `json:"token"`
}

type SessionInfo struct {
	Timestamp  int64  `json:"timestamp"`
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

	// Store in database - now storing timestamp directly
	if _, err := db.Exec(`INSERT INTO sessions (token, timestamp) VALUES ($1, $2)`, token, req.Timestamp); err != nil {
		log.Printf("Failed to insert session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	json.NewEncoder(w).Encode(CreateSessionResponse{Token: token})
	localDateTime := time.Unix(req.Timestamp, 0).Format("2006-01-02 15:04:05")
	fmt.Printf("Session created: %v, %v\n", token, localDateTime)
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

	// Query the database - now reading timestamp directly
	var session SessionInfo
	err := db.QueryRow("SELECT timestamp, COALESCE(coin_result, '') FROM sessions WHERE token = $1", token).
		Scan(&session.Timestamp, &session.CoinResult)

	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If timestamp is in the past and no coin result exists, generate one
	if session.Timestamp <= (time.Now().Unix()+1) && session.CoinResult == "" {
		// Generate random coin flip result
		randomBytes := make([]byte, 1)
		if _, err := rand.Read(randomBytes); err != nil {
			log.Printf("Failed to generate random number: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Use the random byte to determine heads or tails
		if randomBytes[0]%2 == 0 {
			session.CoinResult = "heads"
		} else {
			session.CoinResult = "tails"
		}

		// Save the result to the database
		_, err = db.Exec("UPDATE sessions SET coin_result = $1 WHERE token = $2", session.CoinResult, token)
		if err != nil {
			log.Printf("Failed to update coin result: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Printf("New result for %v: %v", token, session.CoinResult)
	}

	// Return session info as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Build connection string from environment variables
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		timestamp BIGINT,
		coin_result TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
