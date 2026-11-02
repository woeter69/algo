package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gorilla/mux"
)

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if userID, ok := getSessionUser(r); ok && userID > 0 {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		firstname := strings.TrimSpace(r.FormValue("firstname"))
		lastname := strings.TrimSpace(r.FormValue("lastname"))
		email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")

		if firstname == "" || lastname == "" || email == "" || username == "" || password == "" {
			renderTemplate(w, r, "register.html", pongo2.Context{"error": "All fields are required."})
			return
		}

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)", email, username).Scan(&exists)
		if err != nil {
			log.Printf("Database error checking user existence: %v", err)
			renderTemplate(w, r, "register.html", pongo2.Context{"error": "Database error"})
			return
		}

		if exists {
			renderTemplate(w, r, "register.html", pongo2.Context{"error": "Email or username is already registered."})
			return
		}

		hashedPassword, err := HashPassword(password)
		if err != nil {
			log.Printf("Password hashing error: %v", err)
			renderTemplate(w, r, "register.html", pongo2.Context{"error": "Internal error processing password."})
			return
		}

		token := GenerateRandomToken(32)
		expiry := time.Now().Add(24 * time.Hour)

		// Create user and token in database
		var newUserID int
		err = db.QueryRow(`
			INSERT INTO users (firstname, lastname, email, username, password, verified, verification_status, role, registration_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING user_id
		`, firstname, lastname, email, username, hashedPassword, true, "pending", "unverified", time.Now()).Scan(&newUserID)

		if err != nil {
			log.Printf("Database error creating user: %v", err)
			renderTemplate(w, r, "register.html", pongo2.Context{"error": "Failed to create user account."})
			return
		}

		// Insert verification token
		_, _ = db.Exec(`
			INSERT INTO verification_tokens (email, token, expiry)
			VALUES ($1, $2, $3)
		`, email, token, expiry)

		// Set session and redirect to check_email
		session, _ := store.Get(r, "algo-session")
		session.Values["pending_reg_email"] = email
		_ = session.Save(r, w)

		http.Redirect(w, r, "/check_email", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "register.html", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if userID, ok := getSessionUser(r); ok && userID > 0 {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		emailOrUsername := strings.TrimSpace(r.FormValue("email"))
		if emailOrUsername == "" {
			emailOrUsername = strings.TrimSpace(r.FormValue("username"))
		}
		password := r.FormValue("password")

		var userID int
		var hashedPassword, username, pfpPath, role string
		var loginCount int

		err := db.QueryRow(`
			SELECT user_id, password, login_count, username, COALESCE(pfp_path, ''), COALESCE(role, 'student')
			FROM users 
			WHERE LOWER(email) = LOWER($1) OR username = $2
		`, emailOrUsername, emailOrUsername).Scan(&userID, &hashedPassword, &loginCount, &username, &pfpPath, &role)

		if err != nil {
			if err == sql.ErrNoRows {
				renderTemplate(w, r, "login.html", pongo2.Context{"error": "Incorrect email or username."})
			} else {
				log.Printf("Login query error: %v", err)
				renderTemplate(w, r, "login.html", pongo2.Context{"error": "Database error."})
			}
			return
		}

		if !CheckPasswordHash(hashedPassword, password) {
			renderTemplate(w, r, "login.html", pongo2.Context{"error": "Incorrect password."})
			return
		}

		// Success - set session
		setSessionUser(w, r, userID, username, pfpPath, role)

		// Update last_login and increment login_count
		_, _ = db.Exec("UPDATE users SET last_login=$1, login_count=login_count+1 WHERE user_id=$2", time.Now(), userID)

		nextURL := r.URL.Query().Get("next")
		if loginCount == 0 {
			if nextURL != "" && IsSafeURL(r, nextURL) {
				session, _ := store.Get(r, "algo-session")
				session.Values["post_profile_redirect"] = nextURL
				_ = session.Save(r, w)
			}
			http.Redirect(w, r, "/profile/complete", http.StatusSeeOther)
			return
		}

		if nextURL != "" && IsSafeURL(r, nextURL) {
			http.Redirect(w, r, nextURL, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "login.html", nil)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	clearSession(w, r)
	flashMessage(w, r, "success", "You have been logged out.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleCheckEmail(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "check_email.html", nil)
}

func handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	var email string
	var expiry time.Time
	err := db.QueryRow("SELECT email, expiry FROM verification_tokens WHERE token = $1", token).Scan(&email, &expiry)
	if err != nil {
		renderTemplate(w, r, "token_invalid.html", nil)
		return
	}

	if time.Now().After(expiry) {
		renderTemplate(w, r, "token_expired.html", nil)
		return
	}

	// Update user to verified
	_, _ = db.Exec("UPDATE users SET verified = true, verification_status = 'approved' WHERE LOWER(email) = LOWER($1)", email)
	_, _ = db.Exec("DELETE FROM verification_tokens WHERE token = $1", token)

	http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
}

func handleConfirmation(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "confirmation.html", nil)
}

func handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		emailOrUsername := strings.TrimSpace(r.FormValue("email"))
		if emailOrUsername == "" {
			emailOrUsername = strings.TrimSpace(r.FormValue("username"))
		}

		var email, username string
		err := db.QueryRow("SELECT email, username FROM users WHERE LOWER(email) = LOWER($1) OR username = $2",
			emailOrUsername, emailOrUsername).Scan(&email, &username)

		if err == nil {
			resetToken := GenerateRandomToken(32)
			SendPasswordResetEmail(email, resetToken, username)
		}

		flashMessage(w, r, "info", "If that account exists, password reset instructions have been sent.")
	}

	renderTemplate(w, r, "forgot_password.html", nil)
}
