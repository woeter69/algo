package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gorilla/mux"
)

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	var username, role, pfpPath, verificationStatus string
	err := db.QueryRow(`
		SELECT username, COALESCE(role, 'student'), COALESCE(pfp_path, ''), COALESCE(verification_status, 'pending')
		FROM users WHERE user_id = $1
	`, userID).Scan(&username, &role, &pfpPath, &verificationStatus)

	if err != nil {
		clearSession(w, r)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	if role == "unverified" || (role != "admin" && verificationStatus == "pending") {
		http.Redirect(w, r, "/limited_dashboard", http.StatusSeeOther)
		return
	}

	if role == "admin" && r.URL.Path == "/admin_dashboard" {
		handleAdminDashboard(w, r)
		return
	}

	// Dynamic stats
	var totalAlumni, totalConnections, totalChannels, pendingRequests int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'alumni' AND verification_status = 'verified'").Scan(&totalAlumni)
	_ = db.QueryRow(`
		SELECT COUNT(*) FROM connections 
		WHERE (user_id = $1 OR con_user_id = $1) AND status = 'accepted'
	`, userID).Scan(&totalConnections)
	_ = db.QueryRow("SELECT COUNT(*) FROM channels WHERE is_active = true").Scan(&totalChannels)
	_ = db.QueryRow("SELECT COUNT(*) FROM connections WHERE con_user_id = $1 AND status = 'pending'", userID).Scan(&pendingRequests)

	stats := map[string]interface{}{
		"total_alumni":      totalAlumni,
		"total_connections": totalConnections,
		"total_channels":    totalChannels,
		"pending_requests":  pendingRequests,
	}

	// Recent incoming connection requests (limit 4)
	rows, err := db.Query(`
		SELECT u.username, COALESCE(u.role, 'student')
		FROM connections c
		JOIN users u ON c.user_id = u.user_id
		WHERE c.con_user_id = $1 AND c.status = 'pending'
		ORDER BY c.connection_id DESC LIMIT 4
	`, userID)

	var recentRequests []map[string]string
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var uName, uRole string
			if err := rows.Scan(&uName, &uRole); err == nil {
				recentRequests = append(recentRequests, map[string]string{
					"username": uName,
					"role":     uRole,
				})
			}
		}
	}

	context := pongo2.Context{
		"username":        username,
		"role":            role,
		"pfp_path":        pfpPath,
		"stats":           stats,
		"recent_requests": recentRequests,
	}

	renderTemplate(w, r, "user_dashboard.html", context)
}

func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	var username, role, pfpPath string
	err := db.QueryRow("SELECT username, COALESCE(role, 'admin'), COALESCE(pfp_path, '') FROM users WHERE user_id = $1", userID).Scan(&username, &role, &pfpPath)
	if err != nil || role != "admin" {
		http.Redirect(w, r, "/user_dashboard", http.StatusSeeOther)
		return
	}

	var pendingCount, totalUsers, totalVerified, recentRegistrations, pendingInquiriesCount int

	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE verification_status = 'pending'").Scan(&pendingCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE verification_status = 'verified'").Scan(&totalVerified)
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE registration_date >= NOW() - INTERVAL '7 days'").Scan(&recentRegistrations)
	_ = db.QueryRow("SELECT COUNT(*) FROM contacts WHERE status = 'pending'").Scan(&pendingInquiriesCount)

	// Pending verification requests
	pRows, err := db.Query(`
		SELECT 
			u.user_id, 
			u.username, 
			u.email, 
			COALESCE(vr.requested_role, u.role, 'student') as requested_role, 
			COALESCE(u.pfp_path, ''), 
			COALESCE(vr.created_at, u.registration_date) as created_at,
			COALESCE(c.name, ''),
			COALESCE(vr.department, u.department, ''),
			COALESCE(vr.graduation_year, u.graduation_year, 0),
			COALESCE(vr.student_id, ''),
			COALESCE(vr.request_message, '')
		FROM users u
		LEFT JOIN LATERAL (
			SELECT * FROM verification_requests 
			WHERE user_id = u.user_id 
			ORDER BY created_at DESC LIMIT 1
		) vr ON true
		LEFT JOIN communities c ON vr.community_id = c.community_id
		WHERE u.verification_status = 'pending'
		ORDER BY created_at DESC
	`)

	var pendingRequests []map[string]interface{}
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var uid, gradYear int
			var uname, email, reqRole, pfp, cName, dept, vID, msg string
			var createdAt interface{}
			if err := pRows.Scan(&uid, &uname, &email, &reqRole, &pfp, &createdAt, &cName, &dept, &gradYear, &vID, &msg); err == nil {
				pendingRequests = append(pendingRequests, map[string]interface{}{
					"user_id":         uid,
					"username":        uname,
					"email":           email,
					"requested_role":  reqRole,
					"pfp_path":        pfp,
					"created_at":      createdAt,
					"community_name":  cName,
					"department":      dept,
					"graduation_year": gradYear,
					"verification_id": vID,
					"message":         msg,
				})
			}
		}
	}

	// Contact inquiries
	cRows, err := db.Query(`
		SELECT c.id, c.full_name, c.email, COALESCE(c.phone, ''), c.subject, c.message,
		       c.submitted_at, COALESCE(c.status, 'pending'), c.resolved_at, COALESCE(c.resolution_notes, ''),
		       COALESCE(u.username, 'Admin') as resolver_username
		FROM contacts c
		LEFT JOIN users u ON c.resolved_by = u.user_id
		ORDER BY CASE WHEN c.status = 'pending' THEN 0 ELSE 1 END, c.submitted_at DESC
	`)

	var contactInquiries []map[string]interface{}
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var id int
			var fName, email, phone, subject, msg, status, notes, resolver string
			var subAt time.Time
			var resAt sql.NullTime
			if err := cRows.Scan(&id, &fName, &email, &phone, &subject, &msg, &subAt, &status, &resAt, &notes, &resolver); err == nil {
				contactInquiries = append(contactInquiries, map[string]interface{}{
					"id":                id,
					"full_name":         fName,
					"email":             email,
					"phone":             phone,
					"subject":           subject,
					"message":           msg,
					"submitted_at":      subAt,
					"status":            status,
					"resolved_at":       resAt.Time,
					"resolution_notes":  notes,
					"resolver_username": resolver,
				})
			}
		}
	}

	context := pongo2.Context{
		"username":               username,
		"role":                   role,
		"pfp_path":               pfpPath,
		"pending_count":          pendingCount,
		"pending_requests":       pendingRequests,
		"total_users":            totalUsers,
		"total_verified":         totalVerified,
		"recent_registrations":   recentRegistrations,
		"contact_inquiries":      contactInquiries,
		"pending_inquiries_count": pendingInquiriesCount,
	}

	renderTemplate(w, r, "admin_dashboard.html", context)
}

func handleResolveContact(w http.ResponseWriter, r *http.Request) {
	adminID, ok := getSessionUser(r)
	if !ok || getSessionRole(r) != "admin" {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	queryID, _ := strconv.Atoi(vars["query_id"])

	resolutionNotes := strings.TrimSpace(r.FormValue("resolution_notes"))
	if resolutionNotes == "" {
		resolutionNotes = "Resolved via Admin Dashboard"
	}

	var fullName, email, subject, origMsg string
	err := db.QueryRow(`
		UPDATE contacts
		SET status = 'resolved',
		    resolved_by = $1,
		    resolved_at = NOW(),
		    resolution_notes = $2
		WHERE id = $3
		RETURNING full_name, email, subject, message
	`, adminID, resolutionNotes, queryID).Scan(&fullName, &email, &subject, &origMsg)

	if err == nil {
		resolverName := getSessionUsername(r)
		if resolverName == "" {
			resolverName = "Admin"
		}
		SendInquiryResolvedEmail(email, fullName, subject, resolutionNotes, origMsg, resolverName)
		flashMessage(w, r, "success", "Contact inquiry marked as resolved and email sent.")
	} else {
		log.Printf("❌ Failed to resolve contact inquiry: %v", err)
		flashMessage(w, r, "error", "Failed to resolve contact inquiry.")
	}

	http.Redirect(w, r, "/admin_dashboard", http.StatusSeeOther)
}

func handleLimitedDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	var fname, lname, uname, role, vStatus, pfp string
	err := db.QueryRow(`
		SELECT firstname, lastname, username, COALESCE(role, 'student'), COALESCE(verification_status, 'pending'), COALESCE(pfp_path, '')
		FROM users WHERE user_id = $1
	`, userID).Scan(&fname, &lname, &uname, &role, &vStatus, &pfp)

	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userMap := map[string]interface{}{
		"firstname":           fname,
		"lastname":            lname,
		"username":            uname,
		"role":                role,
		"verification_status": vStatus,
		"pfp_path":            pfp,
	}

	var reqID int
	var createdAt time.Time
	var status, reqRole, dept, reviewNotes sql.NullString
	_ = db.QueryRow(`
		SELECT request_id, created_at, status, requested_role, department, review_notes
		FROM verification_requests
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&reqID, &createdAt, &status, &reqRole, &dept, &reviewNotes)

	var vrMap map[string]interface{}
	if reqID > 0 {
		vrMap = map[string]interface{}{
			"request_id":     reqID,
			"created_at":     createdAt,
			"status":         status.String,
			"requested_role": reqRole.String,
			"department":     dept.String,
			"review_notes":   reviewNotes.String,
		}
	}

	renderTemplate(w, r, "limited_dashboard.html", pongo2.Context{
		"user":                 userMap,
		"verification_request": vrMap,
	})
}

func handleVerificationRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		collegeID, _ := strconv.Atoi(r.FormValue("college_id"))
		requestedRole := r.FormValue("requested_role")
		if requestedRole == "" {
			requestedRole = "student"
		}
		studentID := strings.TrimSpace(r.FormValue("student_id"))
		gradYear, _ := strconv.Atoi(r.FormValue("graduation_year"))
		department := strings.TrimSpace(r.FormValue("department"))
		requestMessage := strings.TrimSpace(r.FormValue("request_message"))

		_, err := db.Exec(`
			INSERT INTO verification_requests (
				user_id, community_id, requested_role, student_id,
				graduation_year, department, request_message, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', NOW(), NOW())
		`, userID, collegeID, requestedRole, studentID, gradYear, department, requestMessage)

		if err != nil {
			log.Printf("❌ Failed to create verification request: %v", err)
			flashMessage(w, r, "error", "Failed to submit verification request.")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		_, _ = db.Exec(`
			UPDATE users
			SET verification_status = 'pending',
			    role = $1,
			    department = $2,
			    graduation_year = $3
			WHERE user_id = $4
		`, requestedRole, department, gradYear, userID)

		flashMessage(w, r, "success", "Your verification request has been submitted successfully!")
		http.Redirect(w, r, "/limited_dashboard", http.StatusSeeOther)
		return
	}

	// GET: prepare form data and communities
	var fname, lname, email, userRole, dept string
	var gradYear sql.NullInt32
	_ = db.QueryRow("SELECT firstname, lastname, email, COALESCE(role, 'student'), COALESCE(department, ''), graduation_year FROM users WHERE user_id = $1", userID).
		Scan(&fname, &lname, &email, &userRole, &dept, &gradYear)

	formData := map[string]interface{}{
		"firstname":       fname,
		"lastname":        lname,
		"email":           email,
		"requested_role":  userRole,
		"department":      dept,
		"graduation_year": gradYear.Int32,
	}

	rows, err := db.Query("SELECT community_id, name, COALESCE(location, '') FROM communities WHERE is_active = true ORDER BY name ASC")
	var communities [][]interface{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, loc string
			if err := rows.Scan(&cid, &name, &loc); err == nil {
				communities = append(communities, []interface{}{cid, name, loc})
			}
		}
	}

	renderTemplate(w, r, "verification_request.html", pongo2.Context{
		"form_data":   formData,
		"communities": communities,
	})
}

func handleApiVerificationRequest(w http.ResponseWriter, r *http.Request) {
	adminID, ok := getSessionUser(r)
	if !ok || getSessionRole(r) != "admin" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Unauthorized"})
		return
	}

	var req struct {
		UserID int    `json:"user_id"`
		Action string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Action != "approve" && req.Action != "reject") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid request data"})
		return
	}

	var username, email, currentRole string
	err := db.QueryRow("SELECT username, email, COALESCE(role, 'student') FROM users WHERE user_id = $1", req.UserID).Scan(&username, &email, &currentRole)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "User not found"})
		return
	}

	if req.Action == "approve" {
		var requestedRole sql.NullString
		var communityID sql.NullInt32
		_ = db.QueryRow(`
			SELECT requested_role, community_id FROM verification_requests
			WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1
		`, req.UserID).Scan(&requestedRole, &communityID)

		approvedRole := currentRole
		if requestedRole.Valid && requestedRole.String != "" {
			approvedRole = requestedRole.String
		}

		_, _ = db.Exec(`
			UPDATE users
			SET verification_status = 'verified',
			    role = $1,
			    verified_by = $2,
			    verified_at = NOW()
			WHERE user_id = $3
		`, approvedRole, adminID, req.UserID)

		_, _ = db.Exec(`
			UPDATE verification_requests
			SET status = 'approved',
			    reviewed_by = $1,
			    reviewed_at = NOW(),
			    review_notes = 'Approved by administrator'
			WHERE user_id = $2 AND status = 'pending'
		`, adminID, req.UserID)

		if communityID.Valid && communityID.Int32 > 0 {
			_, _ = db.Exec(`
				INSERT INTO community_members (community_id, user_id, role, status, joined_at)
				VALUES ($1, $2, 'member', 'active', NOW())
				ON CONFLICT (community_id, user_id) DO UPDATE SET status = 'active'
			`, communityID.Int32, req.UserID)
		}
	} else {
		_, _ = db.Exec(`
			UPDATE users
			SET verification_status = 'rejected',
			    role = 'unverified',
			    verified_by = $1,
			    verified_at = NOW()
			WHERE user_id = $2
		`, adminID, req.UserID)

		_, _ = db.Exec(`
			UPDATE verification_requests
			SET status = 'rejected',
			    reviewed_by = $1,
			    reviewed_at = NOW(),
			    review_notes = 'Request rejected by administrator'
			WHERE user_id = $2 AND status = 'pending'
		`, adminID, req.UserID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Verification request %sd for %s", req.Action, username),
		"action":  req.Action,
		"user_id": req.UserID,
	})
}
