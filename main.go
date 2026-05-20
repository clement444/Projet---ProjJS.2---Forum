package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/database"
	"forum/handlers"
)

func main() {
	db := database.Init("forum.db")
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.RegisterPost(db)(w, r)
		} else {
			handlers.RegisterGet(w, r)
		}
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.LoginPost(db)(w, r)
		} else {
			handlers.LoginGet(w, r)
		}
	})

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
