package main

import (
	"database/sql"
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

	stmt := `INSERT INTO sessions (datetime) VALUES (?)`
	_, err = db.Exec(stmt, datetime)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	fmt.Printf("here's your link date: %v, time: %v\n", createInfo.Date, createInfo.Time)
	fmt.Fprintf(w, "here's your link date: %v, time: %v", createInfo.Date, createInfo.Time)
}

func main() {
	db, err = sql.Open("sqlite3", "../coindown.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Server started")
	http.HandleFunc("/create", createLink)
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("main over")
}
