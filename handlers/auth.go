package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

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
