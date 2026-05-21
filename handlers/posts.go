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

type Category struct {
	ID   int
	Name string
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

func getCategories(db *sql.DB) []Category {
	rows, err := db.Query("SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var cats []Category
	for rows.Next() {
		var c Category
		rows.Scan(&c.ID, &c.Name)
		cats = append(cats, c)
	}
	return cats
}

func CreatePostGet(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, username := GetSessionUser(db, r)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/create_post.html"))
		tmpl.Execute(w, map[string]interface{}{
			"Username":   username,
			"Categories": getCategories(db),
		})
	}
}

func CreatePostPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, username := GetSessionUser(db, r)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		title := r.FormValue("title")
		content := r.FormValue("content")
		categoryIDs := r.Form["categories"]

		tmpl := template.Must(template.ParseFiles("templates/create_post.html"))

		if title == "" || content == "" || len(categoryIDs) == 0 {
			tmpl.Execute(w, map[string]interface{}{
				"Username":   username,
				"Categories": getCategories(db),
				"Error":      "Titre, contenu et au moins une catégorie sont obligatoires.",
			})
			return
		}

		result, err := db.Exec(
			"INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)",
			userID, title, content,
		)
		if err != nil {
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			return
		}

		postID, _ := result.LastInsertId()
		for _, catID := range categoryIDs {
			db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
