package main

import (
	"database/sql"
	"net/http"
	"sort"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gorilla/mux"
)

type Conversation struct {
	OtherUserID     int
	Username        string
	PfpPath         string
	LastMessage     sql.NullString
	LastMessageTime sql.NullTime
}

func handleChatList(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	rows, err := db.Query(`
		SELECT DISTINCT other_user_id FROM (
			SELECT CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END as other_user_id
			FROM messages WHERE sender_id = $1 OR receiver_id = $1
			UNION
			SELECT CASE WHEN user_id = $1 THEN con_user_id ELSE user_id END as other_user_id
			FROM connections WHERE (user_id = $1 OR con_user_id = $1) AND status = 'accepted'
		) AS combined_users
	`, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var otherID int
		rows.Scan(&otherID)

		var username, pfpPath, fname, lname string
		db.QueryRow("SELECT username, pfp_path, firstname, lastname FROM users WHERE user_id = $1", otherID).Scan(&username, &pfpPath, &fname, &lname)

		if pfpPath == "" {
			pfpPath = GenerateDefaultAvatar(fname + " " + lname)
		}

		var lastMsg sql.NullString
		var lastTime sql.NullTime
		db.QueryRow(`
			SELECT content, created_at FROM messages 
			WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
			ORDER BY created_at DESC LIMIT 1
		`, userID, otherID).Scan(&lastMsg, &lastTime)

		conversations = append(conversations, Conversation{
			OtherUserID:     otherID,
			Username:        username,
			PfpPath:         pfpPath,
			LastMessage:     lastMsg,
			LastMessageTime: lastTime,
		})
	}

	sort.Slice(conversations, func(i, j int) bool {
		t1 := conversations[i].LastMessageTime.Time
		t2 := conversations[j].LastMessageTime.Time
		return t1.After(t2)
	})

	renderTemplate(w, r, "chat_list.html", pongo2.Context{"conversations": conversations})
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	var otherID int
	var otherUsername, otherPfp, fname, lname string
	err := db.QueryRow("SELECT user_id, username, pfp_path, firstname, lastname FROM users WHERE username = $1", username).Scan(&otherID, &otherUsername, &otherPfp, &fname, &lname)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if otherPfp == "" {
		otherPfp = GenerateDefaultAvatar(fname + " " + lname)
	}

	rows, err := db.Query(`
		SELECT m.sender_id, m.receiver_id, m.content, m.created_at, u.username
		FROM messages m
		JOIN users u ON m.sender_id = u.user_id
		WHERE (m.sender_id=$1 AND m.receiver_id=$2) OR (m.sender_id=$2 AND m.receiver_id=$1)
		ORDER BY m.created_at
	`, userID, otherID)
	
	var conversation []map[string]interface{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sID, rID int
			var content, uname string
			var createdAt time.Time
			rows.Scan(&sID, &rID, &content, &createdAt, &uname)
			conversation = append(conversation, map[string]interface{}{
				"sender_id":   sID,
				"receiver_id": rID,
				"content":     content,
				"created_at":  createdAt,
				"username":    uname,
			})
		}
	}

	renderTemplate(w, r, "chat.html", pongo2.Context{
		"conversation":    conversation,
		"other_user_id":   otherID,
		"other_user_name": otherUsername,
		"other_user_pfp":  otherPfp,
		"current_user_id": userID,
	})
}
