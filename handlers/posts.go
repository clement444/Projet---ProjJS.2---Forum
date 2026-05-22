package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type PostFull struct {
	ID         int
	Title      string
	Content    string
	Username   string
	UserID     int
	ImagePath  string
	CreatedAt  string
	Categories []string
}

type Comment struct {
	ID        int
	Content   string
	Username  string
	UserID    int
	CreatedAt string
}

type PostDetailData struct {
	Post     PostFull
	Comments []Comment
	Username string
	UserID   int
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

func saveImage(r *http.Request) (string, error) {
	file, header, err := r.FormFile("image")
	if err != nil {
		return "", nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowed[ext] {
		return "", nil
	}

	buf := make([]byte, 16)
	rand.Read(buf)
	filename := hex.EncodeToString(buf) + ext
	dst, err := os.Create("static/uploads/" + filename)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	io.Copy(dst, file)
	return "static/uploads/" + filename, nil
}

func CreatePostPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, username := GetSessionUser(db, r)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		r.ParseMultipartForm(10 << 20)

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

		imagePath, err := saveImage(r)
		if err != nil {
			http.Error(w, "Erreur lors de l'upload de l'image", http.StatusInternalServerError)
			return
		}

		result, err := db.Exec(
			"INSERT INTO posts (user_id, title, content, image_path) VALUES (?, ?, ?, ?)",
			userID, title, content, imagePath,
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

func PostDetail(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/post/")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}

		currentUserID, currentUsername := GetSessionUser(db, r)

		var p PostFull
		err = db.QueryRow(`
			SELECT p.id, p.title, p.content, u.username, u.id, COALESCE(p.image_path, ''), p.created_at
			FROM posts p JOIN users u ON p.user_id = u.id
			WHERE p.id = ?`, id,
		).Scan(&p.ID, &p.Title, &p.Content, &p.Username, &p.UserID, &p.ImagePath, &p.CreatedAt)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		catRows, _ := db.Query(`
			SELECT c.name FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?`, id,
		)
		defer catRows.Close()
		for catRows.Next() {
			var name string
			catRows.Scan(&name)
			p.Categories = append(p.Categories, name)
		}

		commentRows, _ := db.Query(`
			SELECT c.id, c.content, u.username, u.id, c.created_at
			FROM comments c JOIN users u ON c.user_id = u.id
			WHERE c.post_id = ?
			ORDER BY c.created_at ASC`, id,
		)
		defer commentRows.Close()
		var comments []Comment
		for commentRows.Next() {
			var c Comment
			commentRows.Scan(&c.ID, &c.Content, &c.Username, &c.UserID, &c.CreatedAt)
			comments = append(comments, c)
		}

		tmpl := template.Must(template.ParseFiles("templates/post.html"))
		tmpl.Execute(w, PostDetailData{
			Post:     p,
			Comments: comments,
			Username: currentUsername,
			UserID:   currentUserID,
		})
	}
}
