package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "home.html", nil)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "about.html", nil)
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fullName := strings.TrimSpace(r.FormValue("full_name"))
		email := strings.TrimSpace(r.FormValue("email"))
		phone := strings.TrimSpace(r.FormValue("phone"))
		subject := strings.TrimSpace(r.FormValue("subject"))
		message := strings.TrimSpace(r.FormValue("message"))

		if fullName == "" || email == "" || message == "" {
			renderTemplate(w, r, "contact.html", pongo2.Context{
				"error": "Full Name, Email, and Message are required.",
			})
			return
		}

		var inquiryID int
		err := db.QueryRow(`
			INSERT INTO contacts (full_name, email, phone, subject, message, submitted_at, status)
			VALUES ($1, $2, $3, $4, $5, $6, 'pending')
			RETURNING id
		`, fullName, email, phone, subject, message, time.Now()).Scan(&inquiryID)

		if err != nil {
			log.Printf("❌ Failed to save contact message: %v", err)
			renderTemplate(w, r, "contact.html", pongo2.Context{
				"error": "Failed to submit inquiry. Please try again.",
			})
			return
		}

		// Send async auto-confirmation email
		SendInquiryConfirmationEmail(email, fullName, subject, message, inquiryID)

		flashMessage(w, r, "Thank you for reaching out! Your inquiry has been submitted.", "success")
		http.Redirect(w, r, "/thanks", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "contact.html", nil)
}

func handleThanks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, r, "thanks.html", nil)
}

func handleRecommendations(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "recommendations.html", nil)
}
