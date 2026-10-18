package main

import (
	"encoding/json"
	"net/http"
)

func handleConnectView(w http.ResponseWriter, r *http.Request) {
	_, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "connect.html", nil)
}

func handleRequestsView(w http.ResponseWriter, r *http.Request) {
	_, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "requests.html", nil)
}

func handleSearchAlumni(w http.ResponseWriter, r *http.Request) {
	_, ok := getSessionUser(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	
	// Simplified search logic
	rows, err := db.Query(`
		SELECT user_id, firstname, lastname, username, pfp_path, role, college
		FROM users
		WHERE (firstname ILIKE $1 OR lastname ILIKE $1 OR username ILIKE $1)
		AND profile_visibility = 'public'
		LIMIT 20
	`, "%"+query+"%")
	
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var fname, lname, uname, pfp, role, college string
		rows.Scan(&id, &fname, &lname, &uname, &pfp, &role, &college)
		users = append(users, map[string]interface{}{
			"user_id":   id,
			"firstname": fname,
			"lastname":  lname,
			"username":  uname,
			"pfp_path":  pfp,
			"role":      role,
			"college":   college,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"users": users})
}
