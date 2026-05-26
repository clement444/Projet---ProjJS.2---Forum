package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
)

func Like(db *sql.DB) http.HandlerFunc {
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

		val, err := strconv.Atoi(r.FormValue("value"))
		if err != nil || (val != 1 && val != -1) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		postID := r.FormValue("post_id")
		commentID := r.FormValue("comment_id")
		redirectURL := r.FormValue("redirect")

		var existingID, existingValue int
		if postID != "" {
			err = db.QueryRow("SELECT id, value FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingID, &existingValue)
		} else {
			err = db.QueryRow("SELECT id, value FROM likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingID, &existingValue)
		}

		if err == nil {
			if existingValue == val {
				db.Exec("DELETE FROM likes WHERE id = ?", existingID)
			} else {
				db.Exec("UPDATE likes SET value = ? WHERE id = ?", val, existingID)
			}
		} else {
			if postID != "" {
				db.Exec("INSERT INTO likes (user_id, post_id, value) VALUES (?, ?, ?)", userID, postID, val)
			} else {
				db.Exec("INSERT INTO likes (user_id, comment_id, value) VALUES (?, ?, ?)", userID, commentID, val)
			}
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}
