package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleOAuth handles Google OAuth 2.0 authentication
type GoogleOAuth struct {
	Config   *oauth2.Config
	DB       *sql.DB
	Sessions map[string]*UserSession // Simple in-memory session store
}

// UserSession represents a user session
type UserSession struct {
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	PfpPath   string    `json:"pfp_path"`
	GoogleID  string    `json:"google_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GoogleUserInfo represents user info from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// NewGoogleOAuth creates a new Google OAuth handler
func NewGoogleOAuth(db *sql.DB) *GoogleOAuth {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Println("⚠️  Google OAuth credentials not found in environment variables")
		log.Println("📝 Please set the following environment variables:")
		log.Println("   GOOGLE_CLIENT_ID=your_google_client_id")
		log.Println("   GOOGLE_CLIENT_SECRET=your_google_client_secret")
		log.Println("🔄 Continuing without OAuth (WebSocket will still work)")
		return nil
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "", // Will be set dynamically based on request
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuth{
		Config:   config,
		DB:       db,
		Sessions: make(map[string]*UserSession),
	}
}

// setupOAuthRoutes configures OAuth endpoints
func (oauth *GoogleOAuth) setupOAuthRoutes(r *mux.Router) {
	r.HandleFunc("/auth/google", oauth.handleGoogleLogin)
	r.HandleFunc("/auth/google/callback", oauth.handleGoogleCallback)
	r.HandleFunc("/auth/logout", oauth.handleLogout)
	r.HandleFunc("/auth/user", oauth.handleGetUser)

	log.Println("🔐 OAuth endpoints configured:")
	log.Println("   🚀 /auth/google - Start Google OAuth")
	log.Println("   🔄 /auth/google/callback - OAuth callback")
	log.Println("   👋 /auth/logout - Logout user")
	log.Println("   👤 /auth/user - Get current user")
}

// handleGoogleLogin initiates Google OAuth flow
func (oauth *GoogleOAuth) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set redirect URL based on current request
	oauth.Config.RedirectURL = oauth.getRedirectURL(r)

	// Generate random state for security
	state := oauth.generateState()

	// Store state in a simple way (in production, use secure session storage)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	// Get authorization URL
	authURL := oauth.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	log.Printf("🚀 OAuth login initiated for %s", r.RemoteAddr)

	// Return JSON response with auth URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"auth_url": authURL,
		"state":    state,
	})
}

// handleGoogleCallback handles OAuth callback
func (oauth *GoogleOAuth) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	// Get state from cookie
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, "Missing state cookie", http.StatusBadRequest)
		return
	}

	// Verify state parameter
	state := r.URL.Query().Get("state")
	if state != stateCookie.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Set redirect URL
	oauth.Config.RedirectURL = oauth.getRedirectURL(r)

	// Exchange code for token
	ctx := context.Background()
	token, err := oauth.Config.Exchange(ctx, code)
	if err != nil {
		log.Printf("❌ Token exchange failed: %v", err)
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	userInfo, err := oauth.getUserInfo(ctx, token)
	if err != nil {
		log.Printf("❌ Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Create or update user in database
	user, err := oauth.createOrUpdateUser(userInfo)
	if err != nil {
		log.Printf("❌ Failed to create/update user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionToken := oauth.createSession(user)

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 days
	})

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	log.Printf("✅ User %s (ID: %d) logged in via Google OAuth", user.Username, user.UserID)

	// Redirect to main app or return JSON
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"user":    user,
			"token":   sessionToken,
		})
	} else {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// handleLogout logs out the user
func (oauth *GoogleOAuth) handleLogout(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get session token
	sessionCookie, err := r.Cookie("session_token")
	if err == nil {
		// Remove session from memory
		delete(oauth.Sessions, sessionCookie.Value)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	log.Printf("👋 User logged out from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleGetUser returns current user info
func (oauth *GoogleOAuth) handleGetUser(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	user := oauth.getCurrentUser(r)
	if user == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Helper methods

func (oauth *GoogleOAuth) getRedirectURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/auth/google/callback", scheme, r.Host)
}

func (oauth *GoogleOAuth) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (oauth *GoogleOAuth) getUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := oauth.Config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (oauth *GoogleOAuth) createOrUpdateUser(userInfo *GoogleUserInfo) (*UserSession, error) {
	if oauth.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	// Check if user exists by Google ID
	var userID int
	var username, email string

	err := oauth.DB.QueryRow(
		"SELECT user_id, username, email FROM users WHERE google_id = $1",
		userInfo.ID,
	).Scan(&userID, &username, &email)

	if err == sql.ErrNoRows {
		// Check if email already exists (link accounts)
		err = oauth.DB.QueryRow(
			"SELECT user_id, username FROM users WHERE email = $1",
			userInfo.Email,
		).Scan(&userID, &username)

		if err == sql.ErrNoRows {
			// Create new user
			username = oauth.generateUsername(userInfo.Email)
			err = oauth.DB.QueryRow(`
				INSERT INTO users (username, email, google_id, pfp_path, created_at)
				VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
				RETURNING user_id
			`, username, userInfo.Email, userInfo.ID, userInfo.Picture).Scan(&userID)

			if err != nil {
				return nil, fmt.Errorf("failed to create user: %v", err)
			}

			log.Printf("👤 Created new user: %s (ID: %d)", username, userID)
		} else if err != nil {
			return nil, fmt.Errorf("database error: %v", err)
		} else {
			// Link Google account to existing email account
			_, err = oauth.DB.Exec(`
				UPDATE users 
				SET google_id = $1, pfp_path = $2, updated_at = CURRENT_TIMESTAMP
				WHERE user_id = $3
			`, userInfo.ID, userInfo.Picture, userID)

			if err != nil {
				return nil, fmt.Errorf("failed to link account: %v", err)
			}

			log.Printf("🔗 Linked Google account to existing user: %s (ID: %d)", username, userID)
		}
	} else if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	} else {
		// Update existing user info
		_, err = oauth.DB.Exec(`
			UPDATE users 
			SET email = $1, pfp_path = $2, updated_at = CURRENT_TIMESTAMP
			WHERE google_id = $3
		`, userInfo.Email, userInfo.Picture, userInfo.ID)

		if err != nil {
			return nil, fmt.Errorf("failed to update user: %v", err)
		}

		log.Printf("🔄 Updated existing user: %s (ID: %d)", username, userID)
	}

	return &UserSession{
		UserID:    userID,
		Username:  username,
		Email:     userInfo.Email,
		PfpPath:   userInfo.Picture,
		GoogleID:  userInfo.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}, nil
}

func (oauth *GoogleOAuth) generateUsername(email string) string {
	// Extract base username from email
	parts := strings.Split(email, "@")
	baseUsername := strings.ToLower(parts[0])

	// Remove special characters and limit length
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	baseUsername = reg.ReplaceAllString(baseUsername, "")
	if len(baseUsername) > 20 {
		baseUsername = baseUsername[:20]
	}

	// Check if username exists and add number if needed
	username := baseUsername
	counter := 1

	for {
		var exists int
		err := oauth.DB.QueryRow("SELECT 1 FROM users WHERE username = $1", username).Scan(&exists)
		if err == sql.ErrNoRows {
			break // Username is available
		}
		username = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++
	}

	return username
}

func (oauth *GoogleOAuth) createSession(user *UserSession) string {
	sessionToken := oauth.generateState() // Reuse state generator for session token
	oauth.Sessions[sessionToken] = user
	return sessionToken
}

func (oauth *GoogleOAuth) getCurrentUser(r *http.Request) *UserSession {
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	session, exists := oauth.Sessions[sessionCookie.Value]
	if !exists || session.ExpiresAt.Before(time.Now()) {
		// Session expired or doesn't exist
		if exists {
			delete(oauth.Sessions, sessionCookie.Value)
		}
		return nil
	}

	return session
}

// Middleware to require authentication
func (oauth *GoogleOAuth) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := oauth.getCurrentUser(r)
		if user == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Add user to request context (optional)
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Helper function to set CORS headers (reuse existing one)
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
