package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

var istLocation *time.Location

func init() {
	var err error
	istLocation, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fallback to fixed zone +5:30
		istLocation = time.FixedZone("IST", int(5.5*3600))
	}
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash verifies a password against bcrypt, werkzeug pbkdf2/scrypt, or plaintext
func CheckPasswordHash(hashedPassword, password string) bool {
	if hashedPassword == "" {
		return false
	}

	// 1. Standard bcrypt ($2a$, $2b$, $2y$)
	if strings.HasPrefix(hashedPassword, "$2a$") ||
		strings.HasPrefix(hashedPassword, "$2b$") ||
		strings.HasPrefix(hashedPassword, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
	}

	// 2. Werkzeug pbkdf2 format: "pbkdf2:sha256:iterations$salt$hash"
	if strings.HasPrefix(hashedPassword, "pbkdf2:") {
		parts := strings.Split(hashedPassword, "$")
		if len(parts) == 3 {
			methodParts := strings.Split(parts[0], ":")
			iterations := 260000
			if len(methodParts) >= 3 {
				if it, err := strconv.Atoi(methodParts[2]); err == nil {
					iterations = it
				}
			}
			salt := parts[1]
			expectedHashHex := parts[2]
			key := pbkdf2.Key([]byte(password), []byte(salt), iterations, len(expectedHashHex)/2, sha256.New)
			computedHex := hex.EncodeToString(key)
			return subtle.ConstantTimeCompare([]byte(computedHex), []byte(expectedHashHex)) == 1
		}
	}

	// 3. Werkzeug scrypt format: "scrypt:32768:8:1$salt$hash"
	if strings.HasPrefix(hashedPassword, "scrypt:") {
		parts := strings.Split(hashedPassword, "$")
		if len(parts) == 3 {
			salt := parts[1]
			expectedHashHex := parts[2]
			key, err := scrypt.Key([]byte(password), []byte(salt), 32768, 8, 1, len(expectedHashHex)/2)
			if err == nil {
				computedHex := hex.EncodeToString(key)
				return subtle.ConstantTimeCompare([]byte(computedHex), []byte(expectedHashHex)) == 1
			}
		}
	}

	// 4. Plain text fallback (for local development or seeded test data)
	return subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(password)) == 1
}

// GenerateRandomToken generates a URL-safe random token
func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// IsStudentEmail checks if an email belongs to an educational institution
func IsStudentEmail(email string) bool {
	testEmails := os.Getenv("TEST_EMAILS")
	if testEmails != "" {
		for _, te := range strings.Split(testEmails, ",") {
			if strings.TrimSpace(te) == email {
				return true
			}
		}
	}
	pattern := `.+@.+\.(edu|ac\.in|ac\.uk|edu\.au|ac\.ca|ac\.nz|ac\.jp|ac\.kr|ac\.us|college\.edu|uni\.edu)$`
	match, _ := regexp.MatchString(pattern, strings.ToLower(email))
	return match
}

// GenerateDefaultAvatar returns a UI-avatars URL for a given name
func GenerateDefaultAvatar(name string) string {
	if strings.TrimSpace(name) == "" {
		return "https://i.ibb.co/QDy827D/default-avatar.png"
	}

	cleanName := strings.TrimSpace(name)
	words := strings.Fields(cleanName)
	initials := ""
	for i, word := range words {
		if i >= 2 {
			break
		}
		initials += string([]rune(word)[0])
	}
	if initials == "" {
		initials = "U"
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

	return fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=%s&color=fff&size=80&font-size=0.6&bold=true",
		url.QueryEscape(initials), backgroundColor)
}

// ToIST converts a time.Time to Indian Standard Time (IST)
func ToIST(t time.Time) time.Time {
	return t.In(istLocation)
}

// FormatISTTime formats a time.Time in Indian Standard Time
func FormatISTTime(t time.Time, format ...string) string {
	if t.IsZero() {
		return "now"
	}
	ist := ToIST(t)
	f := "03:04 PM"
	if len(format) > 0 {
		switch format[0] {
		case "%Y-%m-%d %I:%M %p IST":
			f = "2006-01-02 03:04 PM IST"
		case "%b %d, %I:%M %p":
			f = "Jan 02, 03:04 PM"
		case "%b %d, %Y %H:%M":
			f = "Jan 02, 2006 15:04"
		case "%B %d, %Y":
			f = "January 02, 2006"
		default:
			f = format[0]
		}
	}
	return ist.Format(f)
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

// LoadCities loads cities data from the JSON file
func LoadCities() []string {
	dataPaths := []string{
		"../static/data/cities-name-list.json",
		"static/data/cities-name-list.json",
		"app/static/data/cities-name-list.json",
	}

	for _, p := range dataPaths {
		file, err := os.Open(p)
		if err != nil {
			continue
		}
		defer file.Close()

		var cities []string
		if err := json.NewDecoder(file).Decode(&cities); err == nil {
			return cities
		}
	}
	return []string{"New Delhi", "Mumbai", "Bengaluru", "Hyderabad", "Chennai", "Kolkata", "Pune"}
}

// SaveUploadedFile saves an uploaded file to destination directory and returns relative static path
func SaveUploadedFile(file multipart.File, header *multipart.FileHeader, subDir string) (string, error) {
	origName := filepath.Base(header.Filename)
	ext := strings.ToLower(filepath.Ext(origName))
	if ext == "" {
		ext = ".jpg"
	}

	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), GenerateRandomToken(8), ext)
	uploadDir := filepath.Join("../static/uploads", subDir)
	_ = os.MkdirAll(uploadDir, 0755)

	destPath := filepath.Join(uploadDir, uniqueName)
	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("/static/uploads/%s/%s", subDir, uniqueName), nil
}

// UploadToImgBB uploads an image to ImgBB or falls back to local save
func UploadToImgBB(file multipart.File, header *multipart.FileHeader, apiKey string) string {
	if apiKey != "" {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, file)
		if err == nil {
			b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
			form := url.Values{}
			form.Set("key", apiKey)
			form.Set("image", b64)

			resp, err := http.PostForm("https://api.imgbb.com/1/upload", form)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				var res struct {
					Data struct {
						URL string `json:"url"`
					} `json:"data"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&res); err == nil && res.Data.URL != "" {
					return res.Data.URL
				}
			}
		}
		// Reset file seek
		if seeker, ok := file.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}
	}

	// Local fallback
	localPath, err := SaveUploadedFile(file, header, "avatars")
	if err != nil {
		log.Printf("Failed to save avatar locally: %v", err)
		return ""
	}
	return localPath
}
