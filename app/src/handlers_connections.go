package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
)

func handleRequestsView(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	var currentUsername, currentFname, currentLname, currentRole string
	var currentPfp sql.NullString
	err := db.QueryRow(`
		SELECT username, firstname, lastname, role, pfp_path 
		FROM users WHERE user_id = $1
	`, userID).Scan(&currentUsername, &currentFname, &currentLname, &currentRole, &currentPfp)
	if err != nil {
		log.Printf("Error fetching user for requests: %v", err)
		flashMessage(w, r, "Error loading requests page", "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	currentUser := map[string]interface{}{
		"username":  currentUsername,
		"firstname": currentFname,
		"lastname":  currentLname,
		"role":      currentRole,
		"pfp_path":  currentPfp.String,
	}

	// Incoming pending requests
	rows, err := db.Query(`
		SELECT c.connection_id, c.user_id, COALESCE(c.request, ''),
		       u.firstname, u.lastname, u.username, COALESCE(u.pfp_path, '')
		FROM connections c
		JOIN users u ON c.user_id = u.user_id
		WHERE c.con_user_id = $1 AND c.status = 'pending'
		ORDER BY c.connection_id DESC
	`, userID)

	var pendingRequests []map[string]interface{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var connID, reqUserID int
			var message, fname, lname, uname, pfp string
			rows.Scan(&connID, &reqUserID, &message, &fname, &lname, &uname, &pfp)
			fullName := strings.TrimSpace(fname + " " + lname)
			if pfp == "" {
				pfp = GenerateDefaultAvatar(fullName)
			}
			pendingRequests = append(pendingRequests, map[string]interface{}{
				"connection_id": connID,
				"user_id":       reqUserID,
				"message":       message,
				"created_at":    "Recently",
				"name":          fullName,
				"username":      uname,
				"avatar":        pfp,
			})
		}
	}

	// Sent pending requests
	sRows, err := db.Query(`
		SELECT c.connection_id, c.con_user_id, COALESCE(c.request, ''),
		       u.firstname, u.lastname, u.username, COALESCE(u.pfp_path, '')
		FROM connections c
		JOIN users u ON c.con_user_id = u.user_id
		WHERE c.user_id = $1 AND c.status = 'pending'
		ORDER BY c.connection_id DESC
	`, userID)

	var sentRequests []map[string]interface{}
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var connID, reqUserID int
			var message, fname, lname, uname, pfp string
			sRows.Scan(&connID, &reqUserID, &message, &fname, &lname, &uname, &pfp)
			fullName := strings.TrimSpace(fname + " " + lname)
			if pfp == "" {
				pfp = GenerateDefaultAvatar(fullName)
			}
			sentRequests = append(sentRequests, map[string]interface{}{
				"connection_id": connID,
				"user_id":       reqUserID,
				"message":       message,
				"created_at":    "Recently",
				"name":          fullName,
				"username":      uname,
				"avatar":        pfp,
			})
		}
	}

	// Accepted connections
	cRows, err := db.Query(`
		SELECT DISTINCT
			CASE WHEN c.user_id = $1 THEN c.con_user_id ELSE c.user_id END as other_user_id,
			CASE WHEN c.user_id = $1 THEN u2.firstname ELSE u1.firstname END as firstname,
			CASE WHEN c.user_id = $1 THEN u2.lastname ELSE u1.lastname END as lastname,
			CASE WHEN c.user_id = $1 THEN u2.username ELSE u1.username END as username,
			CASE WHEN c.user_id = $1 THEN COALESCE(u2.pfp_path, '') ELSE COALESCE(u1.pfp_path, '') END as avatar,
			c.connection_id
		FROM connections c
		JOIN users u1 ON c.user_id = u1.user_id
		JOIN users u2 ON c.con_user_id = u2.user_id
		WHERE (c.user_id = $1 OR c.con_user_id = $1) AND c.status = 'accepted'
		ORDER BY c.connection_id DESC
	`, userID)

	var connections []map[string]interface{}
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var otherID, connID int
			var fname, lname, uname, avatar string
			cRows.Scan(&otherID, &fname, &lname, &uname, &avatar, &connID)
			fullName := strings.TrimSpace(fname + " " + lname)
			if avatar == "" {
				avatar = GenerateDefaultAvatar(fullName)
			}
			connections = append(connections, map[string]interface{}{
				"user_id":       otherID,
				"name":          fullName,
				"username":      uname,
				"avatar":        avatar,
				"connected_at":  "Recently",
				"connection_id": connID,
			})
		}
	}

	renderTemplate(w, r, "requests.html", pongo2.Context{
		"pending_requests":  pendingRequests,
		"sent_requests":     sentRequests,
		"connections":       connections,
		"connections_count": len(connections),
		"current_user":      currentUser,
	})
}

func handleConnectView(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	var currentUsername, currentFname, currentLname, currentRole string
	var currentPfp, currentUni, currentGrad, currentCity sql.NullString
	err := db.QueryRow(`
		SELECT username, firstname, lastname, role, pfp_path, university_name, graduation_year, current_city 
		FROM users WHERE user_id = $1
	`, userID).Scan(&currentUsername, &currentFname, &currentLname, &currentRole, &currentPfp, &currentUni, &currentGrad, &currentCity)
	if err != nil {
		log.Printf("Error fetching current user for connect: %v", err)
		flashMessage(w, r, "Error loading connections page", "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	currentUser := map[string]interface{}{
		"username":        currentUsername,
		"firstname":       currentFname,
		"lastname":        currentLname,
		"role":            currentRole,
		"pfp_path":        currentPfp.String,
		"university_name": currentUni.String,
		"graduation_year": currentGrad.String,
		"current_city":    currentCity.String,
	}

	// Fetch all verified users except current user
	rows, err := db.Query(`
		SELECT u.user_id, u.firstname, u.lastname, u.username, 
		       COALESCE(u.university_name, ''), COALESCE(u.graduation_year::text, ''), 
		       COALESCE(u.current_city, ''), COALESCE(u.pfp_path, ''), COALESCE(u.role, 'unverified'),
		       COALESCE(we.company_name, ''), COALESCE(we.job_title, ''),
		       COALESCE(STRING_AGG(i.name, ', '), '') as interests
		FROM users u
		LEFT JOIN work_experience we ON u.user_id = we.user_id AND we.leave_year IS NULL
		LEFT JOIN user_interests ui ON u.user_id = ui.user_id
		LEFT JOIN interests i ON ui.interest_id = i.interest_id
		WHERE u.user_id != $1 AND u.verification_status = 'verified'
		GROUP BY u.user_id, u.firstname, u.lastname, u.username, u.university_name, 
		         u.graduation_year, u.current_city, u.pfp_path, u.role,
		         we.company_name, we.job_title
		ORDER BY u.lastname, u.firstname
	`, userID)
	if err != nil {
		log.Printf("Error fetching verified users for connect: %v", err)
		flashMessage(w, r, "Error loading connections page", "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	defer rows.Close()

	// Fetch outgoing connections
	outRows, err := db.Query(`
		SELECT con_user_id, status, connection_id FROM connections WHERE user_id = $1
	`, userID)
	type connStatus struct {
		status string
		connID int
	}
	connectionsMap := make(map[int]connStatus)
	if err == nil {
		defer outRows.Close()
		for outRows.Next() {
			var targetID, connID int
			var status string
			outRows.Scan(&targetID, &status, &connID)
			connectionsMap[targetID] = connStatus{status: status, connID: connID}
		}
	}

	// Fetch incoming connections
	inRows, err := db.Query(`
		SELECT user_id, status, connection_id FROM connections WHERE con_user_id = $1
	`, userID)
	reverseConnectionsMap := make(map[int]connStatus)
	if err == nil {
		defer inRows.Close()
		for inRows.Next() {
			var senderID, connID int
			var status string
			inRows.Scan(&senderID, &status, &connID)
			reverseConnectionsMap[senderID] = connStatus{status: status, connID: connID}
		}
	}

	var peopleData []map[string]interface{}
	uniMap := make(map[string]bool)
	gradMap := make(map[string]bool)
	locMap := make(map[string]bool)

	for rows.Next() {
		var id int
		var fname, lname, uname, uni, grad, loc, pfp, role, comp, title, rawInterests string
		rows.Scan(&id, &fname, &lname, &uname, &uni, &grad, &loc, &pfp, &role, &comp, &title, &rawInterests)

		name := strings.TrimSpace(fname + " " + lname)
		avatar := pfp
		if avatar == "" {
			avatar = GenerateDefaultAvatar(name)
		}

		uniDisplay := uni
		if uniDisplay == "" {
			uniDisplay = "Not specified"
		} else {
			uniMap[uniDisplay] = true
		}

		gradDisplay := grad
		if gradDisplay == "" {
			gradDisplay = "Not specified"
		} else {
			gradMap[gradDisplay] = true
		}

		locDisplay := loc
		if locDisplay == "" {
			locDisplay = "Not specified"
		} else {
			locMap[locDisplay] = true
		}

		var interests []string
		if rawInterests != "" {
			for _, item := range strings.Split(rawInterests, ",") {
				tr := strings.TrimSpace(item)
				if tr != "" {
					interests = append(interests, tr)
				}
			}
		}

		connStatusStr := "none"
		connID := 0
		if c, found := connectionsMap[id]; found {
			connStatusStr = c.status
			connID = c.connID
		} else if rc, found := reverseConnectionsMap[id]; found {
			if rc.status == "accepted" {
				connStatusStr = "connected"
			} else if rc.status == "pending" {
				connStatusStr = "received_request"
				connID = rc.connID
			}
		}

		userData := map[string]interface{}{
			"id":                id,
			"name":              name,
			"username":          uname,
			"university":        uniDisplay,
			"graduation_year":   gradDisplay,
			"location":          locDisplay,
			"avatar":            avatar,
			"role":              role,
			"company":           comp,
			"title":             title,
			"interests":         interests,
			"connection_status": connStatusStr,
			"connection_id":     connID,
		}
		peopleData = append(peopleData, userData)
	}

	var universities, graduationYears, locations []string
	for k := range uniMap {
		universities = append(universities, k)
	}
	sort.Strings(universities)

	for k := range gradMap {
		graduationYears = append(graduationYears, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(graduationYears)))

	for k := range locMap {
		locations = append(locations, k)
	}
	sort.Strings(locations)

	renderTemplate(w, r, "connect.html", pongo2.Context{
		"people":           peopleData,
		"universities":     universities,
		"graduation_years": graduationYears,
		"locations":        locations,
		"current_user":     currentUser,
	})
}

func handleSendConnectionRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Authentication required"})
		return
	}

	var req struct {
		UserID  interface{} `json:"user_id"`
		Message string      `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid request payload"})
		return
	}

	var targetUserID int
	switch v := req.UserID.(type) {
	case float64:
		targetUserID = int(v)
	case int:
		targetUserID = v
	case string:
		targetUserID, _ = strconv.Atoi(v)
	}

	if targetUserID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "User ID is required"})
		return
	}

	if userID == targetUserID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Cannot connect to yourself"})
		return
	}

	var connID int
	var status string
	err := db.QueryRow(`
		SELECT connection_id, status FROM connections 
		WHERE (user_id = $1 AND con_user_id = $2) OR (user_id = $2 AND con_user_id = $1)
	`, userID, targetUserID).Scan(&connID, &status)

	if err == nil {
		if status == "accepted" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Already connected"})
			return
		} else if status == "pending" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Connection request already sent"})
			return
		}
	}

	_, err = db.Exec(`
		INSERT INTO connections (user_id, con_user_id, request, status)
		VALUES ($1, $2, $3, 'pending')
	`, userID, targetUserID, req.Message)
	if err != nil {
		log.Printf("Error inserting connection request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Error sending connection request"})
		return
	}

	var targetFname, targetLname string
	db.QueryRow("SELECT firstname, lastname FROM users WHERE user_id = $1", targetUserID).Scan(&targetFname, &targetLname)
	targetName := strings.TrimSpace(targetFname + " " + targetLname)
	if targetName == "" {
		targetName = "User"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Connection request sent to %s", targetName),
		"status":  "pending",
	})
}

func handleRespondConnectionRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Authentication required"})
		return
	}

	var req struct {
		ConnectionID interface{} `json:"connection_id"`
		Action       string      `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid request payload"})
		return
	}

	var connID int
	switch v := req.ConnectionID.(type) {
	case float64:
		connID = int(v)
	case int:
		connID = v
	case string:
		connID, _ = strconv.Atoi(v)
	}

	if connID == 0 || (req.Action != "accept" && req.Action != "reject") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid request"})
		return
	}

	var senderID, targetID int
	var currentStatus string
	err := db.QueryRow(`
		SELECT user_id, con_user_id, status FROM connections 
		WHERE connection_id = $1 AND con_user_id = $2 AND status = 'pending'
	`, connID, userID).Scan(&senderID, &targetID, &currentStatus)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Connection request not found"})
		return
	}

	newStatus := "accepted"
	if req.Action == "reject" {
		newStatus = "denied"
	}

	_, err = db.Exec(`UPDATE connections SET status = $1 WHERE connection_id = $2`, newStatus, connID)
	if err != nil {
		log.Printf("Error updating connection request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Error processing request"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Connection request %sed", req.Action),
		"status":  newStatus,
	})
}

func handleGetConnectionRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Authentication required"})
		return
	}

	rows, err := db.Query(`
		SELECT c.connection_id, c.user_id, COALESCE(c.request, ''),
		       u.firstname, u.lastname, u.username, COALESCE(u.pfp_path, '')
		FROM connections c
		JOIN users u ON c.user_id = u.user_id
		WHERE c.con_user_id = $1 AND c.status = 'pending'
		ORDER BY c.connection_id DESC
	`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Error loading requests"})
		return
	}
	defer rows.Close()

	var requests []map[string]interface{}
	for rows.Next() {
		var connID, reqUserID int
		var message, fname, lname, uname, pfp string
		rows.Scan(&connID, &reqUserID, &message, &fname, &lname, &uname, &pfp)
		fullName := strings.TrimSpace(fname + " " + lname)
		if pfp == "" {
			pfp = GenerateDefaultAvatar(fullName)
		}
		requests = append(requests, map[string]interface{}{
			"connection_id": connID,
			"user_id":       reqUserID,
			"message":       message,
			"created_at":    "Recently",
			"name":          fullName,
			"username":      uname,
			"avatar":        pfp,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"requests": requests,
	})
}

func handleCancelConnectionRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Authentication required"})
		return
	}

	var req struct {
		ConnectionID interface{} `json:"connection_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid request payload"})
		return
	}

	var connID int
	switch v := req.ConnectionID.(type) {
	case float64:
		connID = int(v)
	case int:
		connID = v
	case string:
		connID, _ = strconv.Atoi(v)
	}

	if connID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Connection ID is required"})
		return
	}

	var senderID, targetID int
	var currentStatus string
	err := db.QueryRow(`
		SELECT user_id, con_user_id, status FROM connections 
		WHERE connection_id = $1 AND user_id = $2 AND status = 'pending'
	`, connID, userID).Scan(&senderID, &targetID, &currentStatus)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Connection request not found"})
		return
	}

	_, err = db.Exec(`DELETE FROM connections WHERE connection_id = $1`, connID)
	if err != nil {
		log.Printf("Error deleting connection request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Error cancelling request"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Connection request cancelled successfully",
	})
}

func handleSearchAlumni(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")

	rows, err := db.Query(`
		SELECT user_id, firstname, lastname, username, COALESCE(pfp_path, ''), role, COALESCE(college, '')
		FROM users
		WHERE user_id != $1 AND (firstname ILIKE $2 OR lastname ILIKE $2 OR username ILIKE $2)
		AND profile_visibility = 'public'
		LIMIT 20
	`, userID, "%"+query+"%")

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
		if pfp == "" {
			pfp = GenerateDefaultAvatar(strings.TrimSpace(fname + " " + lname))
		}
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
