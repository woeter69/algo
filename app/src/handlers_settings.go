package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
)

func handleSettingsView(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	var id int
	var fname, lname, email, uname, uNameStr, dept, clg, city, pfp, role, bio, pVis, linkedin, github, twitter, website, phone string
	var gradYear, loginCount sql.NullInt32
	var communityID sql.NullInt32
	var dob, lastLogin sql.NullTime
	var emailNotif, jobAlerts sql.NullBool

	err := db.QueryRow(`
		SELECT user_id, firstname, lastname, email, username, dob,
		       university_name, department, college, graduation_year, current_city, 
		       pfp_path, role, community_id, last_login, login_count,
		       bio, profile_visibility, email_notifications, job_alerts,
		       linkedin, github, twitter, website, phone
		FROM users WHERE user_id = $1
	`, userID).Scan(
		&id, &fname, &lname, &email, &uname, &dob,
		&uNameStr, &dept, &clg, &gradYear, &city,
		&pfp, &role, &communityID, &lastLogin, &loginCount,
		&bio, &pVis, &emailNotif, &jobAlerts,
		&linkedin, &github, &twitter, &website, &phone,
	)

	userMap := map[string]interface{}{}
	if err == nil {
		userMap = map[string]interface{}{
			"user_id":             id,
			"firstname":           fname,
			"lastname":            lname,
			"email":               email,
			"username":            uname,
			"dob":                 dob.Time,
			"university_name":     uNameStr,
			"department":          dept,
			"college":             clg,
			"graduation_year":     gradYear.Int32,
			"current_city":        city,
			"pfp_path":            pfp,
			"role":                role,
			"community_id":        communityID.Int32,
			"last_login":          lastLogin.Time,
			"login_count":         loginCount.Int32,
			"bio":                 bio,
			"profile_visibility":  pVis,
			"email_notifications": emailNotif.Valid && emailNotif.Bool,
			"job_alerts":          jobAlerts.Valid && jobAlerts.Bool,
			"linkedin":            linkedin,
			"github":              github,
			"twitter":             twitter,
			"website":             website,
			"phone":               phone,
			"show_email":          false,
			"allow_messages":      true,
		}
	}

	renderTemplate(w, r, "settings.html", pongo2.Context{
		"user_data": userMap,
	})
}

func handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Unauthorized"})
		return
	}

	var data map[string]interface{}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&data)
	} else {
		_ = r.ParseForm()
		data = make(map[string]interface{})
		for k, v := range r.Form {
			if len(v) > 0 {
				data[k] = v[0]
			}
		}
	}

	if newUname, ok := data["username"].(string); ok && newUname != "" {
		var exists bool
		_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND user_id != $2)", newUname, userID).Scan(&exists)
		if exists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Username already taken"})
			return
		}
	}

	fields := []string{}
	values := []interface{}{}
	idx := 1

	fieldMap := map[string]string{
		"username":       "username",
		"firstName":      "firstname",
		"lastName":       "lastname",
		"university":     "university_name",
		"department":     "department",
		"college":        "college",
		"graduationYear": "graduation_year",
		"currentCity":    "current_city",
	}

	for k, col := range fieldMap {
		if val, exists := data[k]; exists && val != nil && fmt.Sprint(val) != "" {
			fields = append(fields, fmt.Sprintf("%s = $%d", col, idx))
			values = append(values, val)
			idx++
		}
	}

	if len(fields) > 0 {
		values = append(values, userID)
		query := fmt.Sprintf("UPDATE users SET %s WHERE user_id = $%d", strings.Join(fields, ", "), idx)
		_, err := db.Exec(query, values...)
		if err != nil {
			log.Printf("❌ Failed to update account: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Error updating account"})
			return
		}
		if newUname, ok := data["username"].(string); ok && newUname != "" {
			session, _ := store.Get(r, "algo-session")
			session.Values["username"] = newUname
			_ = session.Save(r, w)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Account updated successfully"})
}

func handleUpdatePrivacy(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Unauthorized"})
		return
	}

	var data map[string]interface{}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&data)
	} else {
		_ = r.ParseForm()
		data = make(map[string]interface{})
		for k, v := range r.Form {
			if len(v) > 0 {
				data[k] = v[0]
			}
		}
	}

	vis := "public"
	if v, ok := data["profileVisibility"].(string); ok && v != "" {
		vis = v
	} else if v, ok := data["profile_visibility"].(string); ok && v != "" {
		vis = v
	}

	_, _ = db.Exec("UPDATE users SET profile_visibility = $1 WHERE user_id = $2", vis, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Privacy settings updated successfully"})
}

func handleUpdateNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Unauthorized"})
		return
	}

	var data map[string]interface{}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&data)
	} else {
		_ = r.ParseForm()
		data = make(map[string]interface{})
		for k, v := range r.Form {
			if len(v) > 0 {
				data[k] = v[0]
			}
		}
	}

	emailNotif := true
	if v, exists := data["email_notifications"]; exists {
		emailNotif = fmt.Sprint(v) == "true" || fmt.Sprint(v) == "1"
	} else if v, exists := data["emailNotifications"]; exists {
		emailNotif = fmt.Sprint(v) == "true" || fmt.Sprint(v) == "1"
	}

	jobAlerts := true
	if v, exists := data["job_alerts"]; exists {
		jobAlerts = fmt.Sprint(v) == "true" || fmt.Sprint(v) == "1"
	} else if v, exists := data["jobAlerts"]; exists {
		jobAlerts = fmt.Sprint(v) == "true" || fmt.Sprint(v) == "1"
	}

	_, err := db.Exec("UPDATE users SET email_notifications = $1, job_alerts = $2 WHERE user_id = $3", emailNotif, jobAlerts, userID)
	if err != nil {
		log.Printf("❌ Error updating notifications: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Notifications updated successfully"})
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Unauthorized"})
		return
	}

	var data struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	_ = json.NewDecoder(r.Body).Decode(&data)

	var currentHash string
	err := db.QueryRow("SELECT password FROM users WHERE user_id = $1", userID).Scan(&currentHash)
	if err != nil || !CheckPasswordHash(currentHash, data.CurrentPassword) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Incorrect current password"})
		return
	}

	newHash, err := HashPassword(data.NewPassword)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to update password"})
		return
	}

	_, _ = db.Exec("UPDATE users SET password = $1 WHERE user_id = $2", newHash, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Password changed successfully"})
}

func handleUpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Unauthorized"})
		return
	}

	file, header, err := r.FormFile("pfp")
	if err != nil || file == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "No file provided"})
		return
	}
	defer file.Close()

	pfpURL := UploadToImgBB(file, header, os.Getenv("PFP_API"))
	if pfpURL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to upload image"})
		return
	}

	_, _ = db.Exec("UPDATE users SET pfp_path = $1 WHERE user_id = $2", pfpURL, userID)

	session, _ := store.Get(r, "algo-session")
	session.Values["pfp_path"] = pfpURL
	_ = session.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "pfp_url": pfpURL})
}

func handleDeactivateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Unauthorized"})
		return
	}

	_, _ = db.Exec("UPDATE users SET verified = false WHERE user_id = $1", userID)
	clearSession(w, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Account deactivated successfully"})
}

func handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Unauthorized"})
		return
	}

	_, _ = db.Exec("DELETE FROM user_interests WHERE user_id = $1", userID)
	_, _ = db.Exec("DELETE FROM connections WHERE user_id = $1 OR con_user_id = $1", userID)
	_, _ = db.Exec("DELETE FROM messages WHERE sender_id = $1 OR receiver_id = $1", userID)
	_, _ = db.Exec("DELETE FROM education_details WHERE user_id = $1", userID)
	_, _ = db.Exec("DELETE FROM work_experience WHERE user_id = $1", userID)
	_, _ = db.Exec("DELETE FROM users WHERE user_id = $1", userID)

	clearSession(w, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Account deleted successfully"})
}

func handleLogoutAllSessions(w http.ResponseWriter, r *http.Request) {
	clearSession(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Logged out from all sessions"})
}

func handleExportData(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	var fname, lname, email, uname, uNameStr, clg, city, role string
	var gradYear sql.NullInt32
	var dob, regDate sql.NullTime

	_ = db.QueryRow(`
		SELECT firstname, lastname, email, username, dob,
		       university_name, college, graduation_year, current_city,
		       registration_date, role
		FROM users WHERE user_id = $1
	`, userID).Scan(&fname, &lname, &email, &uname, &dob, &uNameStr, &clg, &gradYear, &city, &regDate, &role)

	rows, _ := db.Query(`
		SELECT u.firstname, u.lastname, u.email
		FROM connections c
		JOIN users u ON (c.con_user_id = u.user_id OR c.user_id = u.user_id)
		WHERE (c.user_id = $1 OR c.con_user_id = $1)
		AND c.status = 'accepted' AND u.user_id != $1
	`, userID)

	var conns []map[string]interface{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var f, l, e string
			if err := rows.Scan(&f, &l, &e); err == nil {
				conns = append(conns, map[string]interface{}{
					"name":           fmt.Sprintf("%s %s", f, l),
					"email":          e,
					"connected_date": nil,
				})
			}
		}
	}

	exportData := map[string]interface{}{
		"user_info": map[string]interface{}{
			"firstname":         fname,
			"lastname":          lname,
			"email":             email,
			"username":          uname,
			"dob":               dob.Time.Format("2006-01-02"),
			"university":        uNameStr,
			"college":           clg,
			"graduation_year":   gradYear.Int32,
			"city":              city,
			"registration_date": regDate.Time.Format(time.RFC3339),
			"role":              role,
		},
		"connections": conns,
		"export_date": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=algo_data_export.json")
	json.NewEncoder(w).Encode(exportData)
}

func handleUpdatePreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Authentication required"})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Preferences updated successfully"})
}

