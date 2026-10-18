package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// GenerateRandomToken generates a random URL-safe token
func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// IsStudentEmail checks if an email belongs to an educational institution
func IsStudentEmail(email string) bool {
	pattern := `.+@.+\.(edu|ac\.in|ac\.uk|edu\.au|ac\.ca|ac\.nz|ac\.jp|ac\.kr|ac\.us|college\.edu|uni\.edu)$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// GenerateDefaultAvatar returns a UI-avatars URL for a given name
func GenerateDefaultAvatar(name string) string {
	if name == "" {
		return "https://i.ibb.co/QDy827D/default-avatar.png"
	}

	cleanName := strings.TrimSpace(name)
	words := strings.Fields(cleanName)
	initials := ""
	for i, word := range words {
		if i >= 2 {
			break
		}
		initials += string(word[0])
	}
	initials = strings.ToUpper(initials)

	colors := []string{
		"6D28D9", "3B82F6", "10B981", "F59E0B", "EF4444",
		"8B5CF6", "06B6D4", "F97316", "EC4899", "84CC16",
	}
	
	hash := 0
	for _, char := range cleanName {
		hash += int(char)
	}
	backgroundColor := colors[hash%len(colors)]

	return fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=%s&color=fff&size=80&font-size=0.6&bold=true", url.QueryEscape(initials), backgroundColor)
}

// IsSafeURL checks if a target URL is safe for redirection
func IsSafeURL(r *http.Request, target string) bool {
	if target == "" {
		return false
	}
	u, err := url.Parse(target)
	if err != nil {
		return false
	}
	if u.Host == "" && u.Scheme == "" {
		return true
	}
	return u.Host == r.Host && (u.Scheme == "http" || u.Scheme == "https")
}
