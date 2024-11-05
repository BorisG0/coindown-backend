package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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
	err := decoder.Decode(&createInfo)
	if err != nil {
		fmt.Println("Error: ", err)
		fmt.Fprintf(w, "input error")
		return
	}

	fmt.Printf("here's your link date: %v, time: %v\n", createInfo.Date, createInfo.Time)
	fmt.Fprintf(w, "here's your link date: %v, time: %v", createInfo.Date, createInfo.Time)
}

func main() {
	fmt.Println("Server started")
	http.HandleFunc("/create", createLink)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
