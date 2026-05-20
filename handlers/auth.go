package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RegisterGet(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/register.html"))
	tmpl.Execute(w, nil)
}

func RegisterPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		tmpl := template.Must(template.ParseFiles("templates/register.html"))

		if username == "" || email == "" || password == "" {
			tmpl.Execute(w, map[string]string{"Error": "Tous les champs sont obligatoires."})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(
			"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
			username, email, string(hash),
		)
		if err != nil {
			tmpl.Execute(w, map[string]string{"Error": "Nom d'utilisateur ou email déjà utilisé."})
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func LoginGet(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	tmpl.Execute(w, nil)
}

func LoginPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")

		tmpl := template.Must(template.ParseFiles("templates/login.html"))

		var userID int
		var hash string
		err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&userID, &hash)
		if err != nil {
			tmpl.Execute(w, map[string]string{"Error": "Email ou mot de passe incorrect."})
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
			tmpl.Execute(w, map[string]string{"Error": "Email ou mot de passe incorrect."})
			return
		}

		token := make([]byte, 32)
		rand.Read(token)
		sessionToken := hex.EncodeToString(token)
		expires := time.Now().Add(24 * time.Hour)

		db.Exec(
			"DELETE FROM sessions WHERE user_id = ?",
			userID,
		)
		db.Exec(
			"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
			userID, sessionToken, expires,
		)

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expires,
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
