package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func getChatHistory(currentUserID int) [][]interface{} {
	rows, err := db.Query(`
		SELECT DISTINCT other_user_id FROM (
			SELECT CASE WHEN m.sender_id = $1 THEN m.receiver_id ELSE m.sender_id END as other_user_id
			FROM messages m
			WHERE m.sender_id = $1 OR m.receiver_id = $1
			UNION
			SELECT CASE WHEN c.user_id = $1 THEN c.con_user_id ELSE c.user_id END as other_user_id
			FROM connections c
			WHERE (c.user_id = $1 OR c.con_user_id = $1) AND c.status = 'accepted'
		) AS combined_users
	`, currentUserID)
	if err != nil {
		log.Printf("Error querying chat history user IDs: %v", err)
		return nil
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var uid int
		rows.Scan(&uid)
		userIDs = append(userIDs, uid)
	}

	type chatItem struct {
		otherID     int
		username    string
		pfpPath     string
		lastMsg     string
		lastTime    time.Time
		hasLastTime bool
	}

	var items []chatItem

	for _, otherID := range userIDs {
		var username, fname, lname string
		var pfp sql.NullString
		err := db.QueryRow(`
			SELECT username, pfp_path, firstname, lastname 
			FROM users WHERE user_id = $1
		`, otherID).Scan(&username, &pfp, &fname, &lname)
		if err != nil {
			continue
		}

		pfpPath := pfp.String
		if pfpPath == "" {
			pfpPath = GenerateDefaultAvatar(strings.TrimSpace(fname + " " + lname))
		}

		var lastContent sql.NullString
		var lastCreatedAt sql.NullTime
		err = db.QueryRow(`
			SELECT content, created_at 
			FROM messages 
			WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
			ORDER BY created_at DESC LIMIT 1
		`, currentUserID, otherID).Scan(&lastContent, &lastCreatedAt)

		lastMessage := ""
		var msgTime time.Time
		hasTime := false

		if err == nil && lastContent.Valid {
			lastMessage = lastContent.String
			msgTime = lastCreatedAt.Time
			hasTime = true
		} else {
			var isConnected int
			db.QueryRow(`
				SELECT 1 FROM connections 
				WHERE ((user_id = $1 AND con_user_id = $2) OR (user_id = $2 AND con_user_id = $1)) 
				AND status = 'accepted'
			`, currentUserID, otherID).Scan(&isConnected)
			if isConnected == 1 {
				lastMessage = "You're now connected! Start a conversation."
			}
		}

		items = append(items, chatItem{
			otherID:     otherID,
			username:    username,
			pfpPath:     pfpPath,
			lastMsg:     lastMessage,
			lastTime:    msgTime,
			hasLastTime: hasTime,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].hasLastTime && items[j].hasLastTime {
			return items[i].lastTime.After(items[j].lastTime)
		}
		if items[i].hasLastTime {
			return true
		}
		return false
	})

	var result [][]interface{}
	for _, it := range items {
		var timeVal interface{}
		if it.hasLastTime {
			timeVal = it.lastTime
		}
		result = append(result, []interface{}{
			it.otherID,
			it.username,
			it.pfpPath,
			it.lastMsg,
			timeVal,
		})
	}

	return result
}

func handleChatList(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	conversations := getChatHistory(userID)
	renderTemplate(w, r, "chat_list.html", pongo2.Context{
		"conversations": conversations,
	})
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
	var otherUsername, fname, lname string
	var pfp sql.NullString
	err := db.QueryRow(`
		SELECT user_id, username, pfp_path, firstname, lastname 
		FROM users WHERE username = $1
	`, username).Scan(&otherID, &otherUsername, &pfp, &fname, &lname)
	if err != nil {
		flashMessage(w, r, "User not found", "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	otherPfp := pfp.String
	if otherPfp == "" {
		otherPfp = GenerateDefaultAvatar(strings.TrimSpace(fname + " " + lname))
	}

	rows, err := db.Query(`
		SELECT m.sender_id, m.receiver_id, m.content, m.created_at, u.username
		FROM messages m
		JOIN users u ON m.sender_id = u.user_id
		WHERE (m.sender_id = $1 AND m.receiver_id = $2) OR (m.sender_id = $2 AND m.receiver_id = $1)
		ORDER BY m.created_at ASC
	`, userID, otherID)

	var conversation [][]interface{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sID, rID int
			var content, uname string
			var createdAt time.Time
			rows.Scan(&sID, &rID, &content, &createdAt, &uname)
			conversation = append(conversation, []interface{}{
				sID,
				rID,
				content,
				createdAt,
				uname,
			})
		}
	}

	chatHistory := getChatHistory(userID)

	renderTemplate(w, r, "chat.html", pongo2.Context{
		"conversation":    conversation,
		"other_user_id":   otherID,
		"other_user_name": otherUsername,
		"other_user_pfp":  otherPfp,
		"chat_history":    chatHistory,
		"current_user_id": userID,
	})
}

func handleOnlineStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "online_users": []int{}})
		return
	}

	rows, err := db.Query(`
		SELECT user_id FROM users WHERE last_login > NOW() - INTERVAL '15 minutes'
	`)
	var onlineUsers []int
	foundSelf := false

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var uid int
			rows.Scan(&uid)
			onlineUsers = append(onlineUsers, uid)
			if uid == userID {
				foundSelf = true
			}
		}
	}

	if !foundSelf {
		onlineUsers = append(onlineUsers, userID)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"online_users": onlineUsers,
	})
}

func handleUploadChatFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	err := r.ParseMultipartForm(20 << 20) // 20 MB max
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to parse upload"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		file, header, err = r.FormFile("image")
	}
	if err != nil || header == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "No file provided"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}
	uniqueFilename := fmt.Sprintf("%s%s", strings.ReplaceAll(uuid.New().String(), "-", ""), ext)

	uploadDir := filepath.Join("..", "static", "uploads", "chat")
	os.MkdirAll(uploadDir, 0755)

	destPath := filepath.Join(uploadDir, uniqueFilename)
	out, err := os.Create(destPath)
	if err != nil {
		log.Printf("Error saving chat file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Upload failed"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to write file"})
		return
	}

	fileURL := fmt.Sprintf("/static/uploads/chat/%s", uniqueFilename)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"file_url":  fileURL,
		"image_url": fileURL,
		"filename":  header.Filename,
	})
}

func handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "users": []interface{}{}})
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	var rows *sql.Rows
	var err error

	if q != "" {
		param := "%" + q + "%"
		rows, err = db.Query(`
			SELECT user_id, username, firstname, lastname, COALESCE(pfp_path, ''), 
			       COALESCE(role, 'Member'), COALESCE(department, ''), COALESCE(university_name, '')
			FROM users
			WHERE user_id != $1 AND (
				username ILIKE $2 OR firstname ILIKE $2 OR lastname ILIKE $2 OR email ILIKE $2
			)
			ORDER BY firstname ASC
			LIMIT 20
		`, userID, param)
	} else {
		rows, err = db.Query(`
			SELECT DISTINCT u.user_id, u.username, u.firstname, u.lastname, COALESCE(u.pfp_path, ''),
			       COALESCE(u.role, 'Member'), COALESCE(u.department, ''), COALESCE(u.university_name, '')
			FROM users u
			LEFT JOIN connections c ON (
				(c.user_id = $1 AND c.con_user_id = u.user_id) OR
				(c.con_user_id = $1 AND c.user_id = u.user_id)
			) AND c.status = 'accepted'
			WHERE u.user_id != $1
			ORDER BY u.firstname ASC
			LIMIT 20
		`, userID)
	}

	if err != nil {
		log.Printf("Error searching users: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "users": []interface{}{}})
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var uid int
		var uname, fname, lname, pfp, role, dept, uni string
		rows.Scan(&uid, &uname, &fname, &lname, &pfp, &role, &dept, &uni)

		fullName := strings.TrimSpace(fname + " " + lname)
		if fullName == "" {
			fullName = uname
		}
		if pfp == "" {
			pfp = "/static/images/default_pfp.png"
		}

		headline := ""
		if dept != "" && uni != "" {
			headline = fmt.Sprintf("%s at %s", dept, uni)
		} else if dept != "" {
			headline = dept
		} else if uni != "" {
			headline = uni
		} else {
			headline = fmt.Sprintf("%s Member", strings.Title(role))
		}

		users = append(users, map[string]interface{}{
			"user_id":   uid,
			"username":  uname,
			"name":      fullName,
			"pfp_path":  pfp,
			"role":      strings.Title(role),
			"headline":  headline,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
	})
}
