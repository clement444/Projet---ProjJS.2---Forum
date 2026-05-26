package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
)

func CreateComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, username := GetSessionUser(db, r)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		postID := r.FormValue("post_id")
		content := r.FormValue("content")

		if postID == "" || content == "" {
			http.Redirect(w, r, fmt.Sprintf("/post/%s", postID), http.StatusSeeOther)
			return
		}

		db.Exec(
			"INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)",
			postID, userID, content,
		)

		http.Redirect(w, r, fmt.Sprintf("/post/%s", postID), http.StatusSeeOther)
	}
}

func DeleteComment(db *sql.DB) http.HandlerFunc {
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

		commentID := r.FormValue("comment_id")
		postID := r.FormValue("post_id")

		var ownerID int
		err := db.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&ownerID)
		if err != nil || ownerID != userID {
			http.Error(w, "Interdit", http.StatusForbidden)
			return
		}

		db.Exec("DELETE FROM likes WHERE comment_id = ?", commentID)
		db.Exec("DELETE FROM comments WHERE id = ?", commentID)

		http.Redirect(w, r, fmt.Sprintf("/post/%s", postID), http.StatusSeeOther)
	}
}

func EditComment(db *sql.DB) http.HandlerFunc {
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

		commentID := r.FormValue("comment_id")
		postID := r.FormValue("post_id")
		content := r.FormValue("content")

		if content == "" {
			http.Redirect(w, r, fmt.Sprintf("/post/%s", postID), http.StatusSeeOther)
			return
		}

		var ownerID int
		err := db.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&ownerID)
		if err != nil || ownerID != userID {
			http.Error(w, "Interdit", http.StatusForbidden)
			return
		}

		db.Exec(
			"UPDATE comments SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			content, commentID,
		)

		http.Redirect(w, r, fmt.Sprintf("/post/%s", postID), http.StatusSeeOther)
	}
}
