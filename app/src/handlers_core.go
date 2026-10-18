package main

import (
	"net/http"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "home.html", nil)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "about.html", nil)
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Logic to handle contact form submission
		// For now just redirect back
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
		return
	}
	renderTemplate(w, r, "contact.html", nil)
}

func handleThanks(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "thanks.html", nil)
}
