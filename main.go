package main

import (
	"log"
	"net/http"

	"forum/database"
	"forum/handlers"
)

func main() {
	db := database.Init("forum.db")
	defer db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handlers.Home(db))
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
	http.HandleFunc("/logout", handlers.Logout(db))
	http.HandleFunc("/comment/create", handlers.CreateComment(db))
	http.HandleFunc("/post/delete", handlers.DeletePost(db))
	http.HandleFunc("/comment/delete", handlers.DeleteComment(db))
	http.HandleFunc("/post/", handlers.PostDetail(db))
	http.HandleFunc("/post/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreatePostPost(db)(w, r)
		} else {
			handlers.CreatePostGet(db)(w, r)
		}
	})

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
