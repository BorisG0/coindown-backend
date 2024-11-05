package main

import (
	"fmt"
	"log"
	"net/http"
)

func createLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	fmt.Fprint(w, "here's your link")
}

func main() {
	fmt.Println("Server started")
	http.HandleFunc("/create", createLink)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
