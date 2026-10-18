package main

import (
	"database/sql"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/gorilla/mux"
)

func handleProfileView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var id int
	var fname, lname, uname, email, pfp, role, bio, college, dept string
	var gradYear sql.NullInt32

	err := db.QueryRow(`
		SELECT user_id, firstname, lastname, username, email, pfp_path, role, bio, college, graduation_year, department
		FROM users WHERE username = $1
	`, username).Scan(
		&id, &fname, &lname, &uname, &email, &pfp, &role, &bio, &college, &gradYear, &dept,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	user := map[string]interface{}{
		"user_id":         id,
		"firstname":       fname,
		"lastname":        lname,
		"username":        uname,
		"email":           email,
		"pfp_path":        pfp,
		"role":            role,
		"bio":             bio,
		"college":         college,
		"graduation_year": gradYear.Int32,
		"department":      dept,
	}

	renderTemplate(w, r, "profile.html", pongo2.Context{"profile_user": user})
}

func handleSettingsView(w http.ResponseWriter, r *http.Request) {
	_, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Logic to update settings
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "settings.html", nil)
}
