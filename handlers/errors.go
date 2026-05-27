package handlers

import (
	"html/template"
	"net/http"
)

type ErrorData struct {
	Code    int
	Message string
}

func renderError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	tmpl := template.Must(template.ParseFiles("templates/error.html"))
	tmpl.Execute(w, ErrorData{Code: code, Message: message})
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	renderError(w, http.StatusNotFound, "La page que vous cherchez n'existe pas.")
}

func Forbidden(w http.ResponseWriter, r *http.Request) {
	renderError(w, http.StatusForbidden, "Vous n'avez pas la permission d'effectuer cette action.")
}

func InternalError(w http.ResponseWriter, r *http.Request) {
	renderError(w, http.StatusInternalServerError, "Une erreur interne s'est produite.")
}
