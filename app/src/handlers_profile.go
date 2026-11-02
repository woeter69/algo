package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gorilla/mux"
)

func normalizeDegreeType(degree string) string {
	if degree == "" {
		return "Bachelors"
	}
	cleaned := strings.ToLower(strings.ReplaceAll(degree, ".", ""))
	cleaned = strings.TrimSpace(cleaned)

	validMap := map[string]string{
		"btech":     "B Tech",
		"b tech":    "B Tech",
		"mtech":     "M Tech",
		"m tech":    "M Tech",
		"be":        "B.E.",
		"me":        "M.E.",
		"bsc":       "B.Sc.",
		"msc":       "M.Sc.",
		"bca":       "BCA",
		"mca":       "MCA",
		"mba":       "MBA",
		"bba":       "BBA",
		"phd":       "PHD",
		"doctorate": "Doctorate",
		"diploma":   "Diploma",
		"masters":   "Masters",
		"bachelors": "Bachelors",
	}

	for k, v := range validMap {
		if strings.Contains(cleaned, k) {
			return v
		}
	}
	return "Bachelors"
}

func handleProfileView(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := getSessionUser(r)

	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" || username == "profile" {
		if uName := getSessionUsername(r); uName != "" {
			username = uName
		} else {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}
	}

	var query string
	var queryArg interface{}
	if targetID, err := strconv.Atoi(username); err == nil && targetID > 0 {
		query = "SELECT user_id, firstname, lastname, email, username, dob, graduation_year, university_name, department, college, current_city, pfp_path, registration_date, role, enrollment_number, community_id, last_login, login_count, bio, linkedin, github, twitter, website, phone FROM users WHERE user_id = $1"
		queryArg = targetID
	} else {
		query = "SELECT user_id, firstname, lastname, email, username, dob, graduation_year, university_name, department, college, current_city, pfp_path, registration_date, role, enrollment_number, community_id, last_login, login_count, bio, linkedin, github, twitter, website, phone FROM users WHERE username = $1"
		queryArg = username
	}

	var id int
	var fname, lname, email, uname, uNameStr, dept, clg, city, pfp, role, enrollNum, bio, linkedin, github, twitter, website, phone string
	var gradYear, loginCount sql.NullInt32
	var communityID sql.NullInt32
	var dob, regDate, lastLogin sql.NullTime

	err := db.QueryRow(query, queryArg).Scan(
		&id, &fname, &lname, &email, &uname, &dob, &gradYear, &uNameStr, &dept, &clg,
		&city, &pfp, &regDate, &role, &enrollNum, &communityID, &lastLogin, &loginCount,
		&bio, &linkedin, &github, &twitter, &website, &phone,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Sliced user_data for template index compatibility (user_data.0, user_data.1, ...)
	userData := []interface{}{
		id, fname, lname, email, uname, dob.Time, gradYear.Int32, uNameStr, dept, clg,
		city, pfp, regDate.Time, role, enrollNum, communityID.Int32, lastLogin.Time, loginCount.Int32,
		bio, linkedin, github, twitter, website, phone,
	}

	// User interests
	iRows, err := db.Query(`
		SELECT i.name FROM interests i
		JOIN user_interests ui ON i.interest_id = ui.interest_id
		WHERE ui.user_id = $1
	`, id)
	var userInterests []string
	if err == nil {
		defer iRows.Close()
		for iRows.Next() {
			var iname string
			if err := iRows.Scan(&iname); err == nil {
				userInterests = append(userInterests, iname)
			}
		}
	}

	// Education details
	eRows, err := db.Query(`
		SELECT degree_type, university_name, college_name, major, graduation_year, gpa, detail_id
		FROM education_details WHERE user_id = $1
		ORDER BY graduation_year DESC NULLS LAST
	`, id)
	var educationRows [][]interface{}
	if err == nil {
		defer eRows.Close()
		for eRows.Next() {
			var deg, uName, cName, major string
			var gYear sql.NullInt32
			var gpa sql.NullFloat64
			var detID int
			if err := eRows.Scan(&deg, &uName, &cName, &major, &gYear, &gpa, &detID); err == nil {
				educationRows = append(educationRows, []interface{}{deg, uName, cName, major, gYear.Int32, gpa.Float64, detID})
			}
		}
	}
	var educationData []interface{}
	if len(educationRows) > 0 {
		educationData = educationRows[0]
	}

	// Work experience
	wRows, err := db.Query(`
		SELECT company_name, job_title, join_year, leave_year, exp_id
		FROM work_experience WHERE user_id = $1
		ORDER BY join_year DESC NULLS LAST
	`, id)
	var workExperience [][]interface{}
	if err == nil {
		defer wRows.Close()
		for wRows.Next() {
			var comp, title string
			var jYear, lYear sql.NullInt32
			var expID int
			if err := wRows.Scan(&comp, &title, &jYear, &lYear, &expID); err == nil {
				workExperience = append(workExperience, []interface{}{comp, title, jYear.Int32, lYear.Int32, expID})
			}
		}
	}

	var connectionsCount int
	_ = db.QueryRow(`
		SELECT COUNT(*) FROM connections 
		WHERE (user_id = $1 OR con_user_id = $1) AND status = 'accepted'
	`, id).Scan(&connectionsCount)

	var communityName string
	if communityID.Valid && communityID.Int32 > 0 {
		_ = db.QueryRow("SELECT name FROM communities WHERE community_id = $1", communityID.Int32).Scan(&communityName)
	}

	var currentUserInfo []interface{}
	if currentUserID > 0 && currentUserID != id {
		var curID int
		var curFname, curLname, curEmail, curUname, curPfp, curRole string
		if err := db.QueryRow("SELECT user_id, firstname, lastname, email, username, pfp_path, role FROM users WHERE user_id = $1", currentUserID).
			Scan(&curID, &curFname, &curLname, &curEmail, &curUname, &curPfp, &curRole); err == nil {
			currentUserInfo = []interface{}{curID, curFname, curLname, curEmail, curUname, curPfp, curRole}
		}
	}

	socialLinks := map[string]interface{}{
		"linkedin": linkedin,
		"github":   github,
		"twitter":  twitter,
		"website":  website,
	}

	context := pongo2.Context{
		"user_data":          userData,
		"user_interests":     userInterests,
		"education_data":     educationData,
		"education_list":     educationRows,
		"work_experience":    workExperience,
		"connections_count":  connectionsCount,
		"community_name":     communityName,
		"current_user_info":  currentUserInfo,
		"user_bio":           bio,
		"user_skills":        userInterests,
		"user_social_links":  socialLinks,
		"user_phone":         phone,
	}

	renderTemplate(w, r, "profile.html", context)
}

func handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	_ = r.ParseMultipartForm(10 << 20) // 10 MB

	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	phone := r.FormValue("phone")
	currentCity := r.FormValue("currentCity")
	bio := r.FormValue("bio")
	universityName := r.FormValue("universityName")
	gradYearStr := r.FormValue("graduationYear")
	major := r.FormValue("major")
	gpaStr := r.FormValue("gpa")
	degree := r.FormValue("degree")
	linkedin := r.FormValue("linkedIn")
	github := r.FormValue("github")
	twitter := r.FormValue("twitter")
	website := r.FormValue("website")

	var pfpURL string
	file, header, err := r.FormFile("profilePicture")
	if err == nil && file != nil {
		defer file.Close()
		pfpURL = UploadToImgBB(file, header, os.Getenv("PFP_API"))
		if pfpURL != "" {
			session, _ := store.Get(r, "algo-session")
			session.Values["pfp_path"] = pfpURL
			_ = session.Save(r, w)
		}
	}

	updateFields := []string{}
	updateValues := []interface{}{}
	argIdx := 1

	if firstName != "" {
		updateFields = append(updateFields, fmt.Sprintf("firstname = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(firstName))
		argIdx++
	}
	if lastName != "" {
		updateFields = append(updateFields, fmt.Sprintf("lastname = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(lastName))
		argIdx++
	}
	if phone != "" {
		updateFields = append(updateFields, fmt.Sprintf("phone = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(phone))
		argIdx++
	}
	if currentCity != "" {
		updateFields = append(updateFields, fmt.Sprintf("current_city = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(currentCity))
		argIdx++
	}
	if bio != "" {
		updateFields = append(updateFields, fmt.Sprintf("bio = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(bio))
		argIdx++
	}
	if universityName != "" {
		updateFields = append(updateFields, fmt.Sprintf("university_name = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(universityName))
		argIdx++
	}
	if gradYear, err := strconv.Atoi(gradYearStr); err == nil && gradYear > 0 {
		updateFields = append(updateFields, fmt.Sprintf("graduation_year = $%d", argIdx))
		updateValues = append(updateValues, gradYear)
		argIdx++
	}
	if major != "" {
		updateFields = append(updateFields, fmt.Sprintf("department = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(major))
		argIdx++
	}
	if linkedin != "" {
		updateFields = append(updateFields, fmt.Sprintf("linkedin = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(linkedin))
		argIdx++
	}
	if github != "" {
		updateFields = append(updateFields, fmt.Sprintf("github = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(github))
		argIdx++
	}
	if twitter != "" {
		updateFields = append(updateFields, fmt.Sprintf("twitter = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(twitter))
		argIdx++
	}
	if website != "" {
		updateFields = append(updateFields, fmt.Sprintf("website = $%d", argIdx))
		updateValues = append(updateValues, strings.TrimSpace(website))
		argIdx++
	}
	if pfpURL != "" {
		updateFields = append(updateFields, fmt.Sprintf("pfp_path = $%d", argIdx))
		updateValues = append(updateValues, pfpURL)
		argIdx++
	}

	if len(updateFields) > 0 {
		updateValues = append(updateValues, userID)
		query := fmt.Sprintf("UPDATE users SET %s WHERE user_id = $%d", strings.Join(updateFields, ", "), argIdx)
		_, err = db.Exec(query, updateValues...)
		if err != nil {
			log.Printf("❌ Error updating profile: %v", err)
		}
	}

	// Update education details
	if universityName != "" || degree != "" || major != "" || gradYearStr != "" {
		degNorm := normalizeDegreeType(degree)
		gradYear, _ := strconv.Atoi(gradYearStr)
		gpa, _ := strconv.ParseFloat(gpaStr, 64)

		var detailID int
		err := db.QueryRow("SELECT detail_id FROM education_details WHERE user_id = $1", userID).Scan(&detailID)
		if err == nil {
			_, _ = db.Exec(`
				UPDATE education_details 
				SET university_name = COALESCE(NULLIF($1, ''), university_name),
				    degree_type = COALESCE(NULLIF($2, ''), degree_type),
				    major = COALESCE(NULLIF($3, ''), major),
				    graduation_year = CASE WHEN $4 > 0 THEN $4 ELSE graduation_year END,
				    gpa = CASE WHEN $5 > 0 THEN $5 ELSE gpa END
				WHERE detail_id = $6
			`, universityName, degNorm, major, gradYear, gpa, detailID)
		} else {
			_, _ = db.Exec(`
				INSERT INTO education_details (user_id, university_name, degree_type, major, graduation_year, gpa)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, userID, universityName, degNorm, major, gradYear, gpa)
		}
	}

	// Update interests
	interestsStr := r.FormValue("interests")
	if interestsStr != "" {
		parts := strings.Split(interestsStr, ",")
		for _, part := range parts {
			iname := strings.TrimSpace(part)
			if iname == "" {
				continue
			}
			var iID int
			err := db.QueryRow("INSERT INTO interests (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING interest_id", iname).Scan(&iID)
			if err == nil && iID > 0 {
				_, _ = db.Exec("INSERT INTO user_interests (user_id, interest_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, iID)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Profile updated successfully!",
		"pfp_path": pfpURL,
	})
}

func handleAddExperience(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	var data struct {
		CompanyName string      `json:"company_name"`
		JobTitle    string      `json:"job_title"`
		JoinYear    interface{} `json:"join_year"`
		LeaveYear   interface{} `json:"leave_year"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&data)
	} else {
		_ = r.ParseForm()
		data.CompanyName = r.FormValue("company_name")
		data.JobTitle = r.FormValue("job_title")
		data.JoinYear = r.FormValue("join_year")
		data.LeaveYear = r.FormValue("leave_year")
	}

	if data.CompanyName == "" || data.JobTitle == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Company and title are required"})
		return
	}

	var jYear, lYear sql.NullInt32
	if jy, err := strconv.Atoi(fmt.Sprint(data.JoinYear)); err == nil && jy > 0 {
		jYear = sql.NullInt32{Int32: int32(jy), Valid: true}
	}
	if ly, err := strconv.Atoi(fmt.Sprint(data.LeaveYear)); err == nil && ly > 0 {
		lYear = sql.NullInt32{Int32: int32(ly), Valid: true}
	}

	var expID int
	err := db.QueryRow(`
		INSERT INTO work_experience (user_id, company_name, job_title, join_year, leave_year)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING exp_id
	`, userID, data.CompanyName, data.JobTitle, jYear, lYear).Scan(&expID)

	if err != nil {
		log.Printf("❌ Error inserting experience: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to add experience"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"exp_id":  expID,
		"message": "Experience added!",
	})
}

func handleDeleteExperience(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	vars := mux.Vars(r)
	expID, _ := strconv.Atoi(vars["exp_id"])

	_, err := db.Exec("DELETE FROM work_experience WHERE exp_id = $1 AND user_id = $2", expID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to delete experience"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Experience deleted!",
	})
}

func handleAddEducation(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	var data struct {
		DegreeType     string      `json:"degree_type"`
		UniversityName string      `json:"university_name"`
		CollegeName    string      `json:"college_name"`
		Major          string      `json:"major"`
		GraduationYear interface{} `json:"graduation_year"`
		GPA            interface{} `json:"gpa"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&data)
	} else {
		_ = r.ParseForm()
		data.DegreeType = r.FormValue("degree_type")
		data.UniversityName = r.FormValue("university_name")
		data.CollegeName = r.FormValue("college_name")
		data.Major = r.FormValue("major")
		data.GraduationYear = r.FormValue("graduation_year")
		data.GPA = r.FormValue("gpa")
	}

	if data.UniversityName == "" && data.CollegeName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Institution name is required"})
		return
	}

	degreeNorm := normalizeDegreeType(data.DegreeType)
	var gradYear sql.NullInt32
	if gy, err := strconv.Atoi(fmt.Sprint(data.GraduationYear)); err == nil && gy > 0 {
		gradYear = sql.NullInt32{Int32: int32(gy), Valid: true}
	}
	var gpa sql.NullFloat64
	if gp, err := strconv.ParseFloat(fmt.Sprint(data.GPA), 64); err == nil && gp > 0 {
		gpa = sql.NullFloat64{Float64: gp, Valid: true}
	}

	var detailID int
	err := db.QueryRow(`
		INSERT INTO education_details (user_id, degree_type, university_name, college_name, major, graduation_year, gpa)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING detail_id
	`, userID, degreeNorm, data.UniversityName, data.CollegeName, data.Major, gradYear, gpa).Scan(&detailID)

	if err != nil {
		log.Printf("❌ Error inserting education: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to add education"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"detail_id": detailID,
		"message":   "Education added!",
	})
}

func handleDeleteEducation(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	vars := mux.Vars(r)
	detailID, _ := strconv.Atoi(vars["detail_id"])

	_, err := db.Exec("DELETE FROM education_details WHERE detail_id = $1 AND user_id = $2", detailID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to delete education"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Education deleted!",
	})
}

func handleCompleteProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	cities := LoadCities()
	cutoffDate := time.Now().AddDate(-16, 0, 0).Format("2006-01-02")

	if r.Method == http.MethodPost {
		_ = r.ParseMultipartForm(10 << 20)

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		dob := r.FormValue("dob")
		bio := r.FormValue("bio")
		role := r.FormValue("role")
		uniName := r.FormValue("uni_name")
		clgName := r.FormValue("clg_name")
		gradYearStr := r.FormValue("grad_year")
		city := r.FormValue("city")
		degree := r.FormValue("degree")
		major := r.FormValue("major")
		gpaStr := r.FormValue("gpa")
		company := r.FormValue("company")
		jobTitle := r.FormValue("job_title")
		joinYearStr := r.FormValue("join_year")
		leaveYearStr := r.FormValue("leave_year")

		var pfpURL string
		file, header, err := r.FormFile("pfp")
		if err == nil && file != nil {
			defer file.Close()
			pfpURL = UploadToImgBB(file, header, os.Getenv("PFP_API"))
		}

		gradYear, _ := strconv.Atoi(gradYearStr)
		verificationStatus := "pending"
		if role == "admin" {
			verificationStatus = "verified"
		}

		_, err = db.Exec(`
			UPDATE users SET 
				firstname = $1, lastname = $2, dob = $3,
				university_name = $4, college = $5, graduation_year = $6, current_city = $7, 
				pfp_path = COALESCE(NULLIF($8, ''), pfp_path), role = $9, verification_status = $10,
				bio = $11
			WHERE user_id = $12
		`, firstName, lastName, dob, uniName, clgName, gradYear, city, pfpURL, role, verificationStatus, bio, userID)

		if err != nil {
			log.Printf("❌ Failed to complete profile: %v", err)
			renderTemplate(w, r, "complete_profile.html", pongo2.Context{
				"cities":      cities,
				"cutoff_date": cutoffDate,
				"error":       "Failed to save profile. Please check your details.",
			})
			return
		}

		// Education details
		if uniName != "" || degree != "" || major != "" {
			gpa, _ := strconv.ParseFloat(gpaStr, 64)
			degNorm := normalizeDegreeType(degree)
			_, _ = db.Exec(`
				INSERT INTO education_details (user_id, degree_type, major, university_name, college_name, graduation_year, gpa)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT DO NOTHING
			`, userID, degNorm, major, uniName, clgName, gradYear, gpa)
		}

		// Work experience
		if company != "" && jobTitle != "" {
			jYear, _ := strconv.Atoi(joinYearStr)
			lYear, _ := strconv.Atoi(leaveYearStr)
			_, _ = db.Exec(`
				INSERT INTO work_experience (user_id, company_name, job_title, join_year, leave_year)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT DO NOTHING
			`, userID, company, jobTitle, jYear, lYear)
		}

		// Update session
		session, _ := store.Get(r, "algo-session")
		session.Values["role"] = role
		if pfpURL != "" {
			session.Values["pfp_path"] = pfpURL
		}
		if firstName != "" && lastName != "" {
			session.Values["username"] = fmt.Sprintf("%s %s", firstName, lastName)
		}
		_ = session.Save(r, w)

		http.Redirect(w, r, "/interests", http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, "complete_profile.html", pongo2.Context{
		"cities":      cities,
		"cutoff_date": cutoffDate,
	})
}

func handleInterests(w http.ResponseWriter, r *http.Request) {
	userID, ok := getSessionUser(r)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		selected := r.Form["interests"]

		_, _ = db.Exec("DELETE FROM user_interests WHERE user_id = $1", userID)
		for _, sID := range selected {
			if iID, err := strconv.Atoi(sID); err == nil && iID > 0 {
				_, _ = db.Exec("INSERT INTO user_interests (user_id, interest_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, iID)
			}
		}

		session, _ := store.Get(r, "algo-session")
		redirectURL, _ := session.Values["post_profile_redirect"].(string)
		delete(session.Values, "post_profile_redirect")
		_ = session.Save(r, w)

		if redirectURL != "" && IsSafeURL(r, redirectURL) {
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	rows, err := db.Query("SELECT interest_id, name FROM interests ORDER BY name")
	var dbInterests [][]interface{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				dbInterests = append(dbInterests, []interface{}{id, name})
			}
		}
	}

	renderTemplate(w, r, "interests.html", pongo2.Context{
		"db_interests": dbInterests,
	})
}
