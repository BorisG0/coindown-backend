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
	res, err := db.Exec(stmt, datetime)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	// stmt = `select last_insert_rowid()`

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		log.Fatalf("Failed to get last insert id: %v", err)
	}
	fmt.Printf("result: %v\n", lastInsertId)

	fmt.Printf("link creation(date: %v, time: %v, id: %v)\n", createInfo.Date, createInfo.Time, lastInsertId)
	fmt.Fprintf(w, "https://coindown.com/%v", lastInsertId)
}

func main() {
	db, err = sql.Open("sqlite3", "coindown.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable := `
	create table if not exists sessions (
		id integer primary key autoincrement,
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
