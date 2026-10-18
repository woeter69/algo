package main

import (
	"context"
	"net/http"
)

type contextKey string

const (
	userContextKey contextKey = "user"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "algo-session")
		userID, ok := session.Values["user_id"]
		if !ok {
			if r.Header.Get("Accept") == "application/json" {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			}
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), userContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getSessionUser(r *http.Request) (int, bool) {
	session, _ := store.Get(r, "algo-session")
	userID, ok := session.Values["user_id"].(int)
	return userID, ok
}
