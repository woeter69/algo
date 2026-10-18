package main

import (
	"log"
	"net/http"
	"time"

	"github.com/flosch/pongo2/v6"
	"golang.org/x/crypto/bcrypt"
)

func handleRegister(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "algo-session")
	if _, ok := session.Values["user_id"]; ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		firstname := r.FormValue("firstname")
		lastname := r.FormValue("lastname")
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)", email, username).Scan(&exists)
		if err != nil {
			log.Printf("Database error: %v", err)
			renderTemplate(w, r, "register.html", pongo2.Context{"error": "Database error"})
			return
		}

		if exists {
			renderTemplate(w, r, "register.html", pongo2.Context{"error": "Email or username already registered"})
			return
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		// Verification logic omitted for simplicity in this first pass, or we can implement it
		// For now, let's just create the user as verified to make it work
		_, err = db.Exec(`
			INSERT INTO users (firstname, lastname, email, username, password, verified)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, firstname, lastname, email, username, string(hashedPassword), true)

		if err != nil {
			log.Printf("Database error: %v", err)
			renderTemplate(w, r, "register.html", pongo2.Context{"error": "Failed to create user"})
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "register.html", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "algo-session")
	if _, ok := session.Values["user_id"]; ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		emailOrUsername := r.FormValue("email")
		password := r.FormValue("password")

		var userID int
		var hashedPassword, username, pfpPath, role string
		var loginCount int

		err := db.QueryRow(`
			SELECT user_id, password, login_count, username, pfp_path, role 
			FROM users 
			WHERE email=$1 OR username=$2
		`, emailOrUsername, emailOrUsername).Scan(&userID, &hashedPassword, &loginCount, &username, &pfpPath, &role)

		if err != nil {
			renderTemplate(w, r, "login.html", pongo2.Context{"error": "Incorrect email or username"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			renderTemplate(w, r, "login.html", pongo2.Context{"error": "Incorrect password"})
			return
		}

		// Success
		session.Values["user_id"] = userID
		session.Values["username"] = username
		session.Values["pfp_path"] = pfpPath
		session.Values["role"] = role
		session.Save(r, w)

		db.Exec("UPDATE users SET last_login=$1, login_count=login_count+1 WHERE user_id=$2", time.Now(), userID)

		if loginCount == 0 {
			http.Redirect(w, r, "/profile/complete", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
		return
	}

	renderTemplate(w, r, "login.html", nil)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "algo-session")
	delete(session.Values, "user_id")
	delete(session.Values, "username")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
