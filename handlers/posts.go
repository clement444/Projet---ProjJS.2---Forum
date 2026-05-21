package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
)

type Post struct {
	ID        int
	Title     string
	Username  string
	CreatedAt string
}

type HomeData struct {
	Posts    []Post
	Username string
}

func Home(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		_, username := GetSessionUser(db, r)

		rows, err := db.Query(`
			SELECT p.id, p.title, u.username, p.created_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			ORDER BY p.created_at DESC
		`)
		if err != nil {
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var p Post
			rows.Scan(&p.ID, &p.Title, &p.Username, &p.CreatedAt)
			posts = append(posts, p)
		}

		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, HomeData{Posts: posts, Username: username})
	}
}
