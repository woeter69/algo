package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func handleGetChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	communityID, _ := strconv.Atoi(vars["community_id"])

	// Check if user is member of community
	var checkRole string
	err := db.QueryRow("SELECT role FROM community_members WHERE community_id = $1 AND user_id = $2 AND status = 'active'", communityID, userID).Scan(&checkRole)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Not a member of this community"})
		return
	}

	rows, err := db.Query(`
		SELECT DISTINCT c.channel_id, c.name, COALESCE(c.description, ''), c.channel_type, 
		       c.is_private, c.position_order, c.created_at, u.username as created_by_username
		FROM channels c
		LEFT JOIN channel_members cm ON c.channel_id = cm.channel_id
		LEFT JOIN users u ON c.created_by = u.user_id
		WHERE c.community_id = $1 AND c.is_active = true
		AND (
			c.is_private = false OR
			(c.is_private = true AND cm.user_id = $2)
		)
		ORDER BY c.position_order, c.created_at
	`, communityID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to fetch channels"})
		return
	}
	defer rows.Close()

	var channels []map[string]interface{}
	for rows.Next() {
		var id, pos int
		var name, desc, cType string
		var isPrivate bool
		var createdAt time.Time
		var createdByUsername sql.NullString
		rows.Scan(&id, &name, &desc, &cType, &isPrivate, &pos, &createdAt, &createdByUsername)
		channels = append(channels, map[string]interface{}{
			"channel_id":          id,
			"name":                name,
			"description":         desc,
			"channel_type":        cType,
			"is_private":          isPrivate,
			"position_order":      pos,
			"created_at":          createdAt.Format(time.RFC3339),
			"created_by_username": createdByUsername.String,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"channels": channels})
}

func handleCreateChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	communityID, _ := strconv.Atoi(vars["community_id"])

	// Check permissions in community
	var role string
	err := db.QueryRow("SELECT role FROM community_members WHERE community_id = $1 AND user_id = $2 AND status = 'active'", communityID, userID).Scan(&role)
	sessionRole := getSessionRole(r)
	if err != nil && sessionRole != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Insufficient permissions"})
		return
	}

	var reqBody struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ChannelType string `json:"channel_type"`
		IsPrivate   bool   `json:"is_private"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || strings.TrimSpace(reqBody.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Channel name required"})
		return
	}

	if reqBody.ChannelType == "" {
		reqBody.ChannelType = "text"
	}

	// Check duplicate name
	var existingID int
	err = db.QueryRow(`
		SELECT channel_id FROM channels 
		WHERE community_id = $1 AND LOWER(name) = LOWER($2) AND is_active = true
	`, communityID, strings.TrimSpace(reqBody.Name)).Scan(&existingID)
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Channel name already exists"})
		return
	}

	var newChannel struct {
		ID          int
		Name        string
		Description string
		ChannelType string
		IsPrivate   bool
	}
	err = db.QueryRow(`
		INSERT INTO channels (community_id, name, description, channel_type, is_private, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING channel_id, name, COALESCE(description, ''), channel_type, is_private
	`, communityID, strings.TrimSpace(reqBody.Name), reqBody.Description, reqBody.ChannelType, reqBody.IsPrivate, userID).Scan(
		&newChannel.ID, &newChannel.Name, &newChannel.Description, &newChannel.ChannelType, &newChannel.IsPrivate,
	)
	if err != nil {
		log.Printf("Error creating channel: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to create channel"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Channel created successfully",
		"channel": map[string]interface{}{
			"channel_id":   newChannel.ID,
			"name":         newChannel.Name,
			"description":  newChannel.Description,
			"channel_type": newChannel.ChannelType,
			"is_private":   newChannel.IsPrivate,
			"community_id": communityID,
		},
	})
}

func handleGetChannelMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	channelID, _ := strconv.Atoi(vars["channel_id"])

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	var communityID int
	var isPrivate bool
	err := db.QueryRow(`
		SELECT community_id, is_private FROM channels 
		WHERE channel_id = $1 AND is_active = true
	`, channelID).Scan(&communityID, &isPrivate)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Channel not found"})
		return
	}

	// Check community membership
	var cRole string
	err = db.QueryRow(`
		SELECT role FROM community_members 
		WHERE community_id = $1 AND user_id = $2 AND status = 'active'
	`, communityID, userID).Scan(&cRole)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Access denied"})
		return
	}

	// If private channel, check channel membership
	if isPrivate {
		var dummy int
		err = db.QueryRow("SELECT 1 FROM channel_members WHERE channel_id = $1 AND user_id = $2", channelID, userID).Scan(&dummy)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Access denied to private channel"})
			return
		}
	}

	rows, err := db.Query(`
		SELECT 
			cm.message_id, cm.channel_id, cm.user_id, cm.content, cm.message_type,
			cm.reply_to_message_id, cm.is_edited, cm.is_deleted, cm.created_at,
			u.username, COALESCE(u.pfp_path, ''),
			COALESCE(reply_msg.content, '') as reply_content,
			COALESCE(reply_user.username, '') as reply_username
		FROM channel_messages cm
		JOIN users u ON cm.user_id = u.user_id
		LEFT JOIN channel_messages reply_msg ON cm.reply_to_message_id = reply_msg.message_id
		LEFT JOIN users reply_user ON reply_msg.user_id = reply_user.user_id
		WHERE cm.channel_id = $1 AND cm.is_deleted = false
		ORDER BY cm.created_at DESC
		LIMIT $2 OFFSET $3
	`, channelID, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to fetch messages"})
		return
	}
	defer rows.Close()

	var messages []map[string]interface{}
	var messageIDs []int

	for rows.Next() {
		var msgID, chID, uID int
		var content, mType, uname, pfp, replyContent, replyUname string
		var replyID sql.NullInt32
		var isEdited, isDeleted bool
		var createdAt time.Time

		rows.Scan(&msgID, &chID, &uID, &content, &mType, &replyID, &isEdited, &isDeleted, &createdAt,
			&uname, &pfp, &replyContent, &replyUname)

		var repIDVal interface{} = nil
		if replyID.Valid {
			repIDVal = int(replyID.Int32)
		}

		messageIDs = append(messageIDs, msgID)
		messages = append(messages, map[string]interface{}{
			"message_id":          msgID,
			"channel_id":          chID,
			"user_id":             uID,
			"content":             content,
			"message_type":        mType,
			"reply_to_message_id": repIDVal,
			"is_edited":           isEdited,
			"is_deleted":          isDeleted,
			"created_at":          createdAt.Format(time.RFC3339),
			"username":            uname,
			"pfp_path":            pfp,
			"reply_content":       replyContent,
			"reply_username":      replyUname,
			"reactions":           []map[string]interface{}{},
		})
	}

	// Fetch reactions for these messages
	if len(messageIDs) > 0 {
		var idPlaceholders []string
		var args []interface{}
		for i, id := range messageIDs {
			idPlaceholders = append(idPlaceholders, "$"+strconv.Itoa(i+1))
			args = append(args, id)
		}
		reactionQuery := "SELECT message_id, emoji, COUNT(*) FROM message_reactions WHERE message_id IN (" +
			strings.Join(idPlaceholders, ",") + ") GROUP BY message_id, emoji"
		rRows, err := db.Query(reactionQuery, args...)
		if err == nil {
			defer rRows.Close()
			reactionMap := make(map[int][]map[string]interface{})
			for rRows.Next() {
				var mID, count int
				var emoji string
				rRows.Scan(&mID, &emoji, &count)
				reactionMap[mID] = append(reactionMap[mID], map[string]interface{}{
					"emoji": emoji,
					"count": count,
				})
			}
			for _, msg := range messages {
				mID := msg["message_id"].(int)
				if rx, ok := reactionMap[mID]; ok {
					msg["reactions"] = rx
				}
			}
		}
	}

	// Reverse messages to display in chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"page":     page,
		"limit":    limit,
	})
}

func handleSendChannelMessageHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	channelID, _ := strconv.Atoi(vars["channel_id"])

	var reqBody struct {
		Content          string `json:"content"`
		MessageType      string `json:"message_type"`
		ReplyToMessageID *int   `json:"reply_to_message_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || strings.TrimSpace(reqBody.Content) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Message content required"})
		return
	}

	if reqBody.MessageType == "" {
		reqBody.MessageType = "text"
	}

	var communityID int
	var isPrivate bool
	err := db.QueryRow("SELECT community_id, is_private FROM channels WHERE channel_id = $1 AND is_active = true", channelID).Scan(&communityID, &isPrivate)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Channel not found"})
		return
	}

	// Check community membership
	var cRole string
	err = db.QueryRow("SELECT role FROM community_members WHERE community_id = $1 AND user_id = $2 AND status = 'active'", communityID, userID).Scan(&cRole)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Access denied"})
		return
	}

	// If private channel, check channel membership and permissions
	if isPrivate {
		var canSend bool
		err = db.QueryRow("SELECT can_send_messages FROM channel_members WHERE channel_id = $1 AND user_id = $2", channelID, userID).Scan(&canSend)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Access denied to private channel"})
			return
		}
		if !canSend {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "You cannot send messages in this channel"})
			return
		}
	}

	var msgID int
	var createdAt time.Time
	err = db.QueryRow(`
		INSERT INTO channel_messages (channel_id, user_id, content, message_type, reply_to_message_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING message_id, created_at
	`, channelID, userID, strings.TrimSpace(reqBody.Content), reqBody.MessageType, reqBody.ReplyToMessageID).Scan(&msgID, &createdAt)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to send message"})
		return
	}

	var username string
	var pfp sql.NullString
	db.QueryRow("SELECT username, pfp_path FROM users WHERE user_id = $1", userID).Scan(&username, &pfp)

	responseData := map[string]interface{}{
		"message_id":          msgID,
		"channel_id":          channelID,
		"user_id":             userID,
		"content":             reqBody.Content,
		"message_type":        reqBody.MessageType,
		"reply_to_message_id": reqBody.ReplyToMessageID,
		"is_edited":           false,
		"is_deleted":          false,
		"created_at":          createdAt.Format(time.RFC3339),
		"username":            username,
		"pfp_path":            pfp.String,
		"reactions":           []interface{}{},
	}

	// Broadcast via WebSocket hub instantly
	if hub != nil {
		wsMsg := WSMessage{
			Type:      NewMessage,
			ChannelID: channelID,
			UserID:    userID,
			Username:  username,
			Content:   reqBody.Content,
			MessageID: strconv.Itoa(msgID),
			CreatedAt: createdAt.Format(time.RFC3339),
			PfpPath:   pfp.String,
			Data:      responseData,
			Timestamp: time.Now(),
		}
		go hub.broadcastToChannel(wsMsg)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Message sent successfully",
		"data":    responseData,
	})
}

func handleGetChannelMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	channelID, _ := strconv.Atoi(vars["channel_id"])

	var communityID int
	var isPrivate bool
	err := db.QueryRow("SELECT community_id, is_private FROM channels WHERE channel_id = $1", channelID).Scan(&communityID, &isPrivate)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Channel not found"})
		return
	}

	var rows *sql.Rows
	if isPrivate {
		rows, err = db.Query(`
			SELECT u.user_id, u.username, COALESCE(u.firstname, ''), COALESCE(u.lastname, ''), 
			       COALESCE(u.pfp_path, ''), cm.role, cm.joined_at,
			       CASE WHEN u.last_login > NOW() - INTERVAL '15 minutes' THEN true ELSE false END as is_online
			FROM channel_members cm
			JOIN users u ON cm.user_id = u.user_id
			WHERE cm.channel_id = $1
			ORDER BY is_online DESC, u.username ASC
		`, channelID)
	} else {
		rows, err = db.Query(`
			SELECT u.user_id, u.username, COALESCE(u.firstname, ''), COALESCE(u.lastname, ''), 
			       COALESCE(u.pfp_path, ''), cm.role, cm.joined_at,
			       CASE WHEN u.last_login > NOW() - INTERVAL '15 minutes' THEN true ELSE false END as is_online
			FROM community_members cm
			JOIN users u ON cm.user_id = u.user_id
			WHERE cm.community_id = $1 AND cm.status = 'active'
			ORDER BY is_online DESC, u.username ASC
		`, communityID)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to get channel members"})
		return
	}
	defer rows.Close()

	var membersList []map[string]interface{}
	for rows.Next() {
		var uid int
		var uname, fname, lname, pfp, role string
		var joinedAt time.Time
		var isOnline bool
		rows.Scan(&uid, &uname, &fname, &lname, &pfp, &role, &joinedAt, &isOnline)
		membersList = append(membersList, map[string]interface{}{
			"user_id":   uid,
			"username":  uname,
			"firstname": fname,
			"lastname":  lname,
			"pfp_path":  pfp,
			"role":      role,
			"joined_at": joinedAt.Format(time.RFC3339),
			"is_online": isOnline,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"members": membersList,
		"count":   len(membersList),
	})
}

func handleAddReaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := getSessionUser(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Authentication required"})
		return
	}

	vars := mux.Vars(r)
	messageID, _ := strconv.Atoi(vars["message_id"])

	var reqBody struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || strings.TrimSpace(reqBody.Emoji) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Emoji required"})
		return
	}
	emoji := strings.TrimSpace(reqBody.Emoji)

	var channelID, communityID int
	var isPrivate bool
	err := db.QueryRow(`
		SELECT cm.channel_id, c.community_id, c.is_private
		FROM channel_messages cm
		JOIN channels c ON cm.channel_id = c.channel_id
		WHERE cm.message_id = $1 AND cm.is_deleted = false
	`, messageID).Scan(&channelID, &communityID, &isPrivate)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Message not found"})
		return
	}

	// Check community membership
	var cRole string
	err = db.QueryRow(`
		SELECT role FROM community_members 
		WHERE community_id = $1 AND user_id = $2 AND status = 'active'
	`, communityID, userID).Scan(&cRole)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Access denied"})
		return
	}

	// Check if reaction already exists
	var rxID int
	err = db.QueryRow(`
		SELECT reaction_id FROM message_reactions 
		WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`, messageID, userID, emoji).Scan(&rxID)

	if err == nil {
		// Toggle: delete existing reaction
		db.Exec("DELETE FROM message_reactions WHERE reaction_id = $1", rxID)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "Reaction removed"})
		return
	}

	// Insert reaction
	_, err = db.Exec(`
		INSERT INTO message_reactions (message_id, user_id, emoji)
		VALUES ($1, $2, $3)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
	`, messageID, userID, emoji)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to manage reaction"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Reaction added"})
}
