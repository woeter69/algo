package main

import (
	"database/sql"
	"fmt"
	"log"
)

// SeedDatabase seeds initial development data
func SeedDatabase(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	log.Println("🌱 Seeding community...")
	var commID int
	err := db.QueryRow(`
		INSERT INTO communities (name, description, college_code, location, established_year, website)
		VALUES (
			'Cluster Innovation Centre',
			'Delhi University premier innovation hub fostering creativity, entrepreneurship, and technological advancement.',
			'CIC',
			'University of Delhi, North Campus',
			2017,
			'https://cic.du.ac.in'
		) ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
		RETURNING community_id;
	`).Scan(&commID)
	if err != nil {
		return fmt.Errorf("error inserting community: %w", err)
	}
	log.Printf("✅ Community ready: ID %d", commID)

	log.Println("🌱 Seeding users...")
	adminHash, _ := HashPassword("admin123")
	studentHash, _ := HashPassword("student123")
	alumniHash, _ := HashPassword("alumni123")

	type seedUser struct {
		fname, lname, email, uname, password, role string
		verified                                    bool
		vStatus, uni, col, dept                    string
		gradYear                                   interface{}
		city, bio                                  string
	}

	users := []seedUser{
		{
			fname: "Admin", lname: "Master", email: "admin@alumnigo.test", uname: "admin",
			password: adminHash, role: "admin", verified: true, vStatus: "verified",
			uni: "University of Delhi", col: "Cluster Innovation Centre",
			dept: "Information Technology & Mathematical Innovations", gradYear: nil,
			city: "New Delhi", bio: "Lead administrator for AlumniGo platform.",
		},
		{
			fname: "Aarav", lname: "Sharma", email: "student@alumnigo.test", uname: "teststudent",
			password: studentHash, role: "student", verified: true, vStatus: "verified",
			uni: "University of Delhi", col: "Cluster Innovation Centre",
			dept: "Information Technology", gradYear: 2026,
			city: "New Delhi", bio: "3rd year B.Tech student interested in Distributed Systems and Go.",
		},
		{
			fname: "Priya", lname: "Verma", email: "alumni@alumnigo.test", uname: "testalumni",
			password: alumniHash, role: "alumni", verified: true, vStatus: "verified",
			uni: "University of Delhi", col: "Cluster Innovation Centre",
			dept: "Information Technology", gradYear: 2022,
			city: "Bengaluru", bio: "Senior Software Engineer @ Google. CIC Batch of 2022. Happy to mentor!",
		},
	}

	userIDs := make(map[string]int)
	for _, u := range users {
		var uID int
		err := db.QueryRow(`
			INSERT INTO users (
				firstname, lastname, email, username, password, role,
				verified, verification_status, university_name, college,
				department, graduation_year, current_city, bio, community_id
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10,
				$11, $12, $13, $14, $15
			) ON CONFLICT (username) DO UPDATE SET
				password = EXCLUDED.password,
				verified = EXCLUDED.verified,
				verification_status = EXCLUDED.verification_status
			RETURNING user_id;
		`, u.fname, u.lname, u.email, u.uname, u.password, u.role,
			u.verified, u.vStatus, u.uni, u.col,
			u.dept, u.gradYear, u.city, u.bio, commID).Scan(&uID)
		if err != nil {
			return fmt.Errorf("error seeding user %s: %w", u.uname, err)
		}
		userIDs[u.uname] = uID
	}

	adminID := userIDs["admin"]
	studentID := userIDs["teststudent"]
	alumniID := userIDs["testalumni"]
	log.Printf("✅ Users ready: Admin (%d), Student (%d), Alumni (%d)", adminID, studentID, alumniID)

	log.Println("🌱 Seeding community membership & permissions...")
	memberships := []struct {
		uid  int
		role string
	}{
		{adminID, "admin"},
		{studentID, "member"},
		{alumniID, "member"},
	}
	for _, m := range memberships {
		_, err = db.Exec(`
			INSERT INTO community_members (community_id, user_id, role, status)
			VALUES ($1, $2, $3, 'active')
			ON CONFLICT (community_id, user_id) DO NOTHING;
		`, commID, m.uid, m.role)
		if err != nil {
			log.Printf("Warning inserting community member: %v", err)
		}
	}

	_, _ = db.Exec(`
		INSERT INTO admin_permissions (admin_user_id, community_id, can_verify_students, can_verify_alumni, can_manage_admins)
		VALUES ($1, $2, TRUE, TRUE, TRUE)
		ON CONFLICT (admin_user_id, community_id) DO NOTHING;
	`, adminID, commID)

	log.Println("🌱 Seeding education and work experience...")
	_, _ = db.Exec(`
		INSERT INTO education_details (user_id, degree_type, university_name, college_name, major, graduation_year, gpa)
		VALUES
			($1, 'B Tech', 'University of Delhi', 'Cluster Innovation Centre', 'IT & Mathematical Innovations', 2026, 8.8),
			($2, 'B Tech', 'University of Delhi', 'Cluster Innovation Centre', 'IT & Mathematical Innovations', 2022, 9.2)
		ON CONFLICT DO NOTHING;
	`, studentID, alumniID)

	_, _ = db.Exec(`
		INSERT INTO work_experience (user_id, company_name, job_title, join_year, leave_year)
		VALUES ($1, 'Google', 'Senior Software Engineer', 2022, NULL)
		ON CONFLICT DO NOTHING;
	`, alumniID)

	log.Println("🌱 Seeding channels...")
	channelNames := []struct {
		name, desc, ctype string
		order             int
	}{
		{"general", "General discussion for all CIC members", "text", 1},
		{"announcements", "Official announcements from CIC administration", "announcement", 2},
		{"introductions", "Introduce yourself to the CIC community", "text", 3},
		{"projects", "Share and discuss your innovation projects", "text", 4},
		{"startups", "Startup discussions and entrepreneurship", "text", 5},
		{"job-opportunities", "Job postings and career opportunities", "text", 6},
		{"tech-talk", "Technology discussions and trends", "text", 7},
	}

	for _, ch := range channelNames {
		_, _ = db.Exec(`
			INSERT INTO channels (community_id, name, description, channel_type, created_by, position_order)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING;
		`, commID, ch.name, ch.desc, ch.ctype, adminID, ch.order)
	}

	cRows, err := db.Query("SELECT channel_id, name FROM channels WHERE community_id = $1;", commID)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var chID int
			var chName string
			cRows.Scan(&chID, &chName)
			_, _ = db.Exec(`
				INSERT INTO channel_messages (channel_id, user_id, content, message_type)
				VALUES ($1, $2, $3, 'text')
				ON CONFLICT DO NOTHING;
			`, chID, adminID, fmt.Sprintf("Welcome to #%s! Start connecting and sharing with your peers.", chName))
		}
	}

	log.Println("🌱 Seeding direct messages & connection...")
	_, _ = db.Exec(`
		INSERT INTO connections (user_id, con_user_id, request, status)
		VALUES ($1, $2, 'Hi Priya! I would love mentorship on backend engineering.', 'accepted')
		ON CONFLICT DO NOTHING;
	`, studentID, alumniID)

	_, _ = db.Exec(`
		INSERT INTO messages (sender_id, receiver_id, content)
		VALUES
			($1, $2, 'Hi Priya, thanks for connecting! Looking forward to your guidance.'),
			($2, $1, 'Hey Aarav! Glad to connect. Feel free to ask anything about career or tech.')
		ON CONFLICT DO NOTHING;
	`, studentID, alumniID)

	log.Println("🎉 Database seeded successfully!")
	return nil
}
