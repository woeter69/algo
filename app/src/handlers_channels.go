package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func handleGetChannels(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	communityID, _ := strconv.Atoi(vars["community_id"])

	rows, err := db.Query(`
		SELECT DISTINCT c.channel_id, c.name, c.description, c.channel_type, c.is_private, c.position_order, c.created_at
		FROM channels c
		LEFT JOIN channel_members cm ON c.channel_id = cm.channel_id
		WHERE c.community_id = $1 AND c.is_active = true
		AND (
			c.is_private = false OR
			(c.is_private = true AND cm.user_id = $2)
		)
		ORDER BY c.position_order, c.created_at
	`, communityID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var channels []map[string]interface{}
	for rows.Next() {
		var id int
		var name, desc, cType string
		var isPrivate bool
		var pos int
		var createdAt interface{}
		rows.Scan(&id, &name, &desc, &cType, &isPrivate, &pos, &createdAt)
		channels = append(channels, map[string]interface{}{
			"channel_id":   id,
			"name":         name,
			"description":  desc,
			"channel_type": cType,
			"is_private":   isPrivate,
			"position":     pos,
			"created_at":   createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"channels": channels})
}
