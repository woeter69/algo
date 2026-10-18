package main

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
)

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "algo-session")
	userID, ok := session.Values["user_id"]
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	role := session.Values["role"].(string)
	username := session.Values["username"].(string)

	context := pongo2.Context{
		"user_id":  userID,
		"username": username,
		"role":     role,
	}

	if role == "admin" {
		renderTemplate(w, r, "admin_dashboard.html", context)
	} else {
		renderTemplate(w, r, "user_dashboard.html", context)
	}
}
