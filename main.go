package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/database"
)

func main() {
	db := database.Init("forum.db")
	defer db.Close()

	_ = db

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
