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
	Likes      int
	Dislikes   int
}

type EditPostData struct {
	Post       PostFull
	Categories []Category
	Selected   map[int]bool
	Username   string
	Error      string
}

type Comment struct {
	ID        int
	Content   string
	Username  string
	UserID    int
	CreatedAt string
	Likes     int
	Dislikes  int
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

		db.QueryRow(`SELECT COALESCE(SUM(CASE WHEN value=1 THEN 1 ELSE 0 END),0), COALESCE(SUM(CASE WHEN value=-1 THEN 1 ELSE 0 END),0) FROM likes WHERE post_id = ?`, id).Scan(&p.Likes, &p.Dislikes)

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
			db.QueryRow(`SELECT COALESCE(SUM(CASE WHEN value=1 THEN 1 ELSE 0 END),0), COALESCE(SUM(CASE WHEN value=-1 THEN 1 ELSE 0 END),0) FROM likes WHERE comment_id = ?`, c.ID).Scan(&c.Likes, &c.Dislikes)
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

func DeletePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		userID, username := GetSessionUser(db, r)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		postID := r.FormValue("post_id")

		var ownerID int
		err := db.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&ownerID)
		if err != nil || ownerID != userID {
			http.Error(w, "Interdit", http.StatusForbidden)
			return
		}

		db.Exec("DELETE FROM post_categories WHERE post_id = ?", postID)
		db.Exec("DELETE FROM likes WHERE post_id = ?", postID)
		db.Exec("DELETE FROM comments WHERE post_id = ?", postID)
		db.Exec("DELETE FROM posts WHERE id = ?", postID)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func EditPostGet(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, username := GetSessionUser(db, r)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}

		var p PostFull
		err = db.QueryRow(`
			SELECT p.id, p.title, p.content, u.username, u.id, COALESCE(p.image_path, ''), p.created_at
			FROM posts p JOIN users u ON p.user_id = u.id
			WHERE p.id = ?`, id,
		).Scan(&p.ID, &p.Title, &p.Content, &p.Username, &p.UserID, &p.ImagePath, &p.CreatedAt)
		if err != nil || p.UserID != userID {
			http.Error(w, "Interdit", http.StatusForbidden)
			return
		}

		selected := map[int]bool{}
		rows, _ := db.Query("SELECT category_id FROM post_categories WHERE post_id = ?", id)
		defer rows.Close()
		for rows.Next() {
			var cid int
			rows.Scan(&cid)
			selected[cid] = true
		}

		tmpl := template.Must(template.ParseFiles("templates/edit_post.html"))
		tmpl.Execute(w, EditPostData{
			Post:       p,
			Categories: getCategories(db),
			Selected:   selected,
			Username:   username,
		})
	}
}

func EditPostPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, username := GetSessionUser(db, r)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		postID := r.FormValue("post_id")
		title := r.FormValue("title")
		content := r.FormValue("content")
		categoryIDs := r.Form["categories"]

		var ownerID int
		err := db.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&ownerID)
		if err != nil || ownerID != userID {
			http.Error(w, "Interdit", http.StatusForbidden)
			return
		}

		if title == "" || content == "" || len(categoryIDs) == 0 {
			http.Redirect(w, r, "/post/edit?id="+postID, http.StatusSeeOther)
			return
		}

		db.Exec(
			"UPDATE posts SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			title, content, postID,
		)
		db.Exec("DELETE FROM post_categories WHERE post_id = ?", postID)
		for _, catID := range categoryIDs {
			db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
		}

		http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
	}
}
