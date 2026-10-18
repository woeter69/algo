package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/flosch/pongo2/v6"
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

		_, err := db.Exec(`
			INSERT INTO communities (name, college_code, location, description, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		`, name, code, location, description, userID)

		if err != nil {
			log.Printf("Error creating community: %v", err)
			renderTemplate(w, r, "create_community.html", pongo2.Context{"error": "Failed to create community"})
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "create_community.html", nil)
}

func handleGetCommunities(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.Query(`
		SELECT c.community_id, c.name, c.description, c.college_code, c.location, cm.role, cm.joined_at
		FROM communities c
		JOIN community_members cm ON c.community_id = cm.community_id
		WHERE cm.user_id = $1 AND cm.status = 'active'
		ORDER BY c.name
	`, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"communities": communities})
}

