package main

import (
	"context"
	"net/http"
)

type contextKey string

const (
	userContextKey contextKey = "user"
	roleContextKey contextKey = "role"
)

// authMiddleware ensures user is authenticated
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "algo-session")
		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			if r.Header.Get("Accept") == "application/json" || r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
				http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			}
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, userID)
		if role, ok := session.Values["role"].(string); ok {
			ctx = context.WithValue(ctx, roleContextKey, role)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// adminMiddleware ensures user is an administrator
func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "algo-session")
		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		role, _ := session.Values["role"].(string)
		if role != "admin" {
			http.Redirect(w, r, "/user_dashboard", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, userID)
		ctx = context.WithValue(ctx, roleContextKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// verifiedUserMiddleware ensures user is verified or admin
func verifiedUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "algo-session")
		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		role, _ := session.Values["role"].(string)
		if role == "admin" {
			next.ServeHTTP(w, r)
			return
		}

		// Check verification status from database
		var verificationStatus, dbRole string
		err := db.QueryRow("SELECT verification_status, role FROM users WHERE user_id = $1", userID).Scan(&verificationStatus, &dbRole)
		if err != nil || verificationStatus != "verified" || dbRole == "unverified" {
			http.Redirect(w, r, "/limited_dashboard", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getSessionUser(r *http.Request) (int, bool) {
	session, _ := store.Get(r, "algo-session")
	if session == nil || session.Values == nil {
		return 0, false
	}
	userID, ok := session.Values["user_id"].(int)
	return userID, ok
}

func getSessionRole(r *http.Request) string {
	session, _ := store.Get(r, "algo-session")
	if session == nil || session.Values == nil {
		return ""
	}
	role, _ := session.Values["role"].(string)
	return role
}

func getSessionUsername(r *http.Request) string {
	session, _ := store.Get(r, "algo-session")
	if session == nil || session.Values == nil {
		return ""
	}
	username, _ := session.Values["username"].(string)
	return username
}

func setSessionUser(w http.ResponseWriter, r *http.Request, userID int, username, pfpPath, role string) {
	session, _ := store.Get(r, "algo-session")
	session.Values["user_id"] = userID
	session.Values["username"] = username
	session.Values["pfp_path"] = pfpPath
	session.Values["role"] = role
	_ = session.Save(r, w)
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "algo-session")
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
	_ = session.Save(r, w)
}

func flashMessage(w http.ResponseWriter, r *http.Request, category, message string) {
	session, _ := store.Get(r, "algo-session")
	var flashes [][]string
	if raw, ok := session.Values["_flashes"]; ok {
		if fList, ok := raw.([][]string); ok {
			flashes = fList
		}
	}
	flashes = append(flashes, []string{category, message})
	session.Values["_flashes"] = flashes
	_ = session.Save(r, w)
}

func popFlashes(w http.ResponseWriter, r *http.Request) [][]string {
	session, _ := store.Get(r, "algo-session")
	var flashes [][]string
	if raw, ok := session.Values["_flashes"]; ok {
		if fList, ok := raw.([][]string); ok {
			flashes = fList
		}
	}
	delete(session.Values, "_flashes")
	_ = session.Save(r, w)
	return flashes
}
