package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gorilla/mux"
)

func handleChannelsView(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Fetch communities user is member of
	rows, err := db.Query(`
		SELECT c.community_id, c.name, c.description, c.college_code, c.location
		FROM communities c
		JOIN community_members cm ON c.community_id = cm.community_id
		WHERE cm.user_id = $1 AND cm.status = 'active'
		ORDER BY c.name
	`, userID)
	if err != nil {
		log.Printf("Error fetching communities: %v", err)
		renderTemplate(w, r, "channels.html", pongo2.Context{"error": "Failed to load communities"})
		return
	}
	defer rows.Close()

	var communities [][]interface{}
	for rows.Next() {
		var id int
		var name, desc, code, loc string
		rows.Scan(&id, &name, &desc, &code, &loc)
		communities = append(communities, []interface{}{id, name, desc, code, loc})
	}

	selectedCommunityIDStr := r.URL.Query().Get("community_id")
	var selectedCommunityID int
	if selectedCommunityIDStr != "" {
		selectedCommunityID, _ = strconv.Atoi(selectedCommunityIDStr)
	} else if len(communities) > 0 {
		selectedCommunityID = communities[0][0].(int)
	}

	var channels [][]interface{}
	if selectedCommunityID > 0 {
		cRows, err := db.Query(`
			SELECT DISTINCT c.channel_id, c.name, c.description, c.channel_type, c.is_private, c.position_order, c.created_at
			FROM channels c
			LEFT JOIN channel_members cm ON c.channel_id = cm.channel_id
			WHERE c.community_id = $1 AND c.is_active = true
			AND (
				c.is_private = false OR
				(c.is_private = true AND cm.user_id = $2)
			)
			ORDER BY c.position_order, c.created_at
		`, selectedCommunityID, userID)
		if err == nil {
			defer cRows.Close()
			for cRows.Next() {
				var id int
				var name, desc, cType string
				var isPrivate bool
				var pos int
				var createdAt interface{}
				cRows.Scan(&id, &name, &desc, &cType, &isPrivate, &pos, &createdAt)
				channels = append(channels, []interface{}{id, name, desc, cType, isPrivate, pos, createdAt})
			}
		}
	}

	renderTemplate(w, r, "channels.html", pongo2.Context{
		"communities":           communities,
		"channels":              channels,
		"selected_community_id": selectedCommunityID,
		"user_id":               userID,
	})
}

func handleCreateCommunity(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("community_name")
		code := r.FormValue("college_code")
		location := r.FormValue("location")
		description := r.FormValue("description")

		var communityID int
		err := db.QueryRow(`
			INSERT INTO communities (name, college_code, location, description, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
			RETURNING community_id
		`, name, code, location, description, userID).Scan(&communityID)

		if err != nil {
			log.Printf("Error creating community: %v", err)
			renderTemplate(w, r, "create_community.html", pongo2.Context{"error": "Failed to create community"})
			return
		}

		// Also add creator as admin member
		db.Exec(`
			INSERT INTO community_members (community_id, user_id, role, status)
			VALUES ($1, $2, 'admin', 'active')
			ON CONFLICT (community_id, user_id) DO NOTHING
		`, communityID, userID)

		// Create default general channel
		db.Exec(`
			INSERT INTO channels (community_id, name, description, channel_type, is_private, created_by)
			VALUES ($1, 'general', 'General discussions', 'text', false, $2)
		`, communityID, userID)

		http.Redirect(w, r, fmt.Sprintf("/channels?community_id=%d", communityID), http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "create_community.html", nil)
}

func handleGetCommunities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	rows, err := db.Query(`
		SELECT c.community_id, c.name, COALESCE(c.description, ''), COALESCE(c.college_code, ''), 
		       COALESCE(c.location, ''), cm.role, cm.joined_at
		FROM communities c
		JOIN community_members cm ON c.community_id = cm.community_id
		WHERE cm.user_id = $1 AND cm.status = 'active'
		ORDER BY c.name
	`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to fetch communities"})
		return
	}
	defer rows.Close()

	var communities []map[string]interface{}
	for rows.Next() {
		var id int
		var name, desc, code, loc, role string
		var joinedAt time.Time
		rows.Scan(&id, &name, &desc, &code, &loc, &role, &joinedAt)
		communities = append(communities, map[string]interface{}{
			"community_id": id,
			"name":         name,
			"description":  desc,
			"college_code": code,
			"location":     loc,
			"role":         role,
			"joined_at":    joinedAt,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"communities": communities})
}

func handleDiscoverCommunities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	rows, err := db.Query(`
		SELECT c.community_id, c.name, COALESCE(c.college_code, ''), COALESCE(c.location, ''), COALESCE(c.description, ''),
		       cm.status as membership_status, cm.role as user_role,
		       (SELECT COUNT(*) FROM community_members WHERE community_id = c.community_id AND status = 'active') as member_count
		FROM communities c
		LEFT JOIN community_members cm ON c.community_id = cm.community_id AND cm.user_id = $1
		ORDER BY c.name
	`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to discover communities"})
		return
	}
	defer rows.Close()

	var communities []map[string]interface{}
	for rows.Next() {
		var id, memberCount int
		var name, code, loc, desc string
		var memStatus, userRole sql.NullString
		rows.Scan(&id, &name, &code, &loc, &desc, &memStatus, &userRole, &memberCount)

		var ms interface{} = nil
		if memStatus.Valid {
			ms = memStatus.String
		}
		var ur interface{} = nil
		if userRole.Valid {
			ur = userRole.String
		}

		communities = append(communities, map[string]interface{}{
			"community_id":      id,
			"name":              name,
			"college_code":      code,
			"location":          loc,
			"description":       desc,
			"membership_status": ms,
			"user_role":         ur,
			"member_count":      memberCount,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"communities": communities})
}

func handleJoinCommunity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	communityID, _ := strconv.Atoi(vars["community_id"])

	var name string
	err := db.QueryRow("SELECT name FROM communities WHERE community_id = $1", communityID).Scan(&name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Community not found"})
		return
	}

	_, err = db.Exec(`
		INSERT INTO community_members (community_id, user_id, role, status)
		VALUES ($1, $2, 'member', 'active')
		ON CONFLICT (community_id, user_id) DO UPDATE SET status = 'active'
	`, communityID, userID)
	if err != nil {
		log.Printf("Error joining community: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to join community"})
		return
	}

	db.Exec("UPDATE users SET community_id = COALESCE(community_id, $1) WHERE user_id = $2", communityID, userID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully joined %s!", name),
	})
}

func handleLeaveCommunity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	communityID, _ := strconv.Atoi(vars["community_id"])

	_, err := db.Exec("DELETE FROM community_members WHERE community_id = $1 AND user_id = $2", communityID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to leave community"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Left community successfully",
	})
}

func handleJoinCommunityRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	communityID, _ := strconv.Atoi(vars["community_id"])

	var reqBody struct {
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&reqBody)

	var name string
	err := db.QueryRow("SELECT name FROM communities WHERE community_id = $1", communityID).Scan(&name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Community not found"})
		return
	}

	var status string
	err = db.QueryRow("SELECT status FROM community_members WHERE community_id = $1 AND user_id = $2", communityID, userID).Scan(&status)
	if err == nil {
		if status == "active" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Already a member of this community"})
			return
		} else if status == "banned" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "You are banned from this community"})
			return
		}
	}

	var reqStatus string
	err = db.QueryRow("SELECT status FROM community_join_requests WHERE community_id = $1 AND user_id = $2", communityID, userID).Scan(&reqStatus)
	if err == nil && reqStatus == "pending" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Join request already pending"})
		return
	}

	var requestID int
	err = db.QueryRow(`
		INSERT INTO community_join_requests (community_id, user_id, message)
		VALUES ($1, $2, $3)
		ON CONFLICT (community_id, user_id) 
		DO UPDATE SET message = $3, status = 'pending', requested_at = CURRENT_TIMESTAMP
		RETURNING request_id
	`, communityID, userID, reqBody.Message).Scan(&requestID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to submit join request"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Join request submitted successfully",
		"request_id": requestID,
	})
}

func handleInviteUserToCommunity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	communityID, _ := strconv.Atoi(vars["community_id"])

	var userRole string
	err := db.QueryRow("SELECT role FROM community_members WHERE community_id = $1 AND user_id = $2 AND status = 'active'", communityID, userID).Scan(&userRole)
	if err != nil || (userRole != "admin" && userRole != "moderator") {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Insufficient permissions"})
		return
	}

	var reqBody struct {
		UserID  int    `json:"user_id"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.UserID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "User ID required"})
		return
	}

	var uname string
	err = db.QueryRow("SELECT username FROM users WHERE user_id = $1", reqBody.UserID).Scan(&uname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "User not found"})
		return
	}

	var memStatus string
	err = db.QueryRow("SELECT status FROM community_members WHERE community_id = $1 AND user_id = $2", communityID, reqBody.UserID).Scan(&memStatus)
	if err == nil && memStatus == "active" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "User is already a member"})
		return
	}

	var invID int
	err = db.QueryRow(`
		INSERT INTO community_invitations (community_id, invited_user_id, invited_by, message)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (community_id, invited_user_id) 
		DO UPDATE SET invited_by = $3, message = $4, status = 'pending', 
		             created_at = CURRENT_TIMESTAMP, expires_at = CURRENT_TIMESTAMP + INTERVAL '7 days'
		RETURNING invitation_id
	`, communityID, reqBody.UserID, userID, reqBody.Message).Scan(&invID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to send invitation"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Invitation sent successfully",
		"invitation_id": invID,
	})
}

func handleGetCommunityMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	communityID, _ := strconv.Atoi(vars["community_id"])

	// Check membership
	var checkRole string
	err := db.QueryRow("SELECT role FROM community_members WHERE community_id = $1 AND user_id = $2 AND status = 'active'", communityID, userID).Scan(&checkRole)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Not a member of this community"})
		return
	}

	rows, err := db.Query(`
		SELECT 
			u.user_id,
			COALESCE(u.firstname, ''),
			COALESCE(u.lastname, ''),
			u.username,
			COALESCE(u.pfp_path, ''),
			COALESCE(u.department, ''),
			COALESCE(u.role, 'member') as user_role,
			cm.role as community_role,
			cm.joined_at,
			CASE WHEN u.last_login > NOW() - INTERVAL '5 minutes' THEN true ELSE false END as is_online
		FROM community_members cm
		JOIN users u ON cm.user_id = u.user_id
		WHERE cm.community_id = $1 AND cm.status = 'active'
		ORDER BY 
			CASE cm.role 
				WHEN 'admin' THEN 1 
				WHEN 'moderator' THEN 2 
				ELSE 3 
			END,
			u.firstname, u.lastname
	`, communityID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to fetch members"})
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var uid int
		var fname, lname, uname, pfp, dept, uRole, cRole string
		var joinedAt time.Time
		var isOnline bool
		rows.Scan(&uid, &fname, &lname, &uname, &pfp, &dept, &uRole, &cRole, &joinedAt, &isOnline)
		members = append(members, map[string]interface{}{
			"user_id":    uid,
			"firstname":  fname,
			"lastname":   lname,
			"username":   uname,
			"pfp_path":   pfp,
			"department": dept,
			"user_role":  uRole,
			"role":       cRole,
			"joined_at":  joinedAt.Format(time.RFC3339),
			"is_online":  isOnline,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"members": members})
}
