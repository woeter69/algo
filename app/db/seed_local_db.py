"""
Seed script for local development and testing.
Creates initial community, test users (admin, student, alumni), channels, and relationships.
"""
import os
import psycopg2
from werkzeug.security import generate_password_hash

DB_URL = os.getenv("DATABASE_URL", "postgresql://postgres@localhost:5432/alumni_platform?sslmode=disable")

def seed():
    conn = psycopg2.connect(DB_URL)
    conn.autocommit = True
    cur = conn.cursor()

    print("🌱 Seeding community...")
    cur.execute("""
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
    """)
    comm_id = cur.fetchone()[0]
    print(f"✅ Community created/found: ID {comm_id}")

    print("🌱 Seeding users...")
    users_data = [
        {
            "firstname": "Admin",
            "lastname": "Master",
            "email": "admin@alumnigo.test",
            "username": "admin",
            "password": generate_password_hash("admin123"),
            "role": "admin",
            "verified": True,
            "verification_status": "approved",
            "university_name": "University of Delhi",
            "college": "Cluster Innovation Centre",
            "department": "Information Technology & Mathematical Innovations",
            "current_city": "New Delhi",
            "bio": "Lead administrator for AlumniGo platform."
        },
        {
            "firstname": "Aarav",
            "lastname": "Sharma",
            "email": "student@alumnigo.test",
            "username": "teststudent",
            "password": generate_password_hash("student123"),
            "role": "student",
            "verified": True,
            "verification_status": "approved",
            "university_name": "University of Delhi",
            "college": "Cluster Innovation Centre",
            "department": "Information Technology",
            "graduation_year": 2026,
            "current_city": "New Delhi",
            "bio": "3rd year B.Tech student interested in Distributed Systems and Go."
        },
        {
            "firstname": "Priya",
            "lastname": "Verma",
            "email": "alumni@alumnigo.test",
            "username": "testalumni",
            "password": generate_password_hash("alumni123"),
            "role": "alumni",
            "verified": True,
            "verification_status": "approved",
            "university_name": "University of Delhi",
            "college": "Cluster Innovation Centre",
            "department": "Information Technology",
            "graduation_year": 2022,
            "current_city": "Bengaluru",
            "bio": "Senior Software Engineer @ Google. CIC Batch of 2022. Happy to mentor!"
        }
    ]

    user_ids = {}
    for u in users_data:
        cur.execute("""
            INSERT INTO users (
                firstname, lastname, email, username, password, role,
                verified, verification_status, university_name, college,
                department, graduation_year, current_city, bio, community_id
            ) VALUES (
                %(firstname)s, %(lastname)s, %(email)s, %(username)s, %(password)s, %(role)s,
                %(verified)s, %(verification_status)s, %(university_name)s, %(college)s,
                %(department)s, %(graduation_year)s, %(current_city)s, %(bio)s, %(community_id)s
            ) ON CONFLICT (username) DO UPDATE SET
                password = EXCLUDED.password,
                verified = EXCLUDED.verified,
                verification_status = EXCLUDED.verification_status
            RETURNING user_id;
        """, {**u, "graduation_year": u.get("graduation_year"), "community_id": comm_id})
        user_ids[u["username"]] = cur.fetchone()[0]

    admin_id = user_ids["admin"]
    student_id = user_ids["teststudent"]
    alumni_id = user_ids["testalumni"]
    print(f"✅ Users ready: Admin ({admin_id}), Student ({student_id}), Alumni ({alumni_id})")

    print("🌱 Seeding community membership & permissions...")
    for uid, role in [(admin_id, 'admin'), (student_id, 'member'), (alumni_id, 'member')]:
        cur.execute("""
            INSERT INTO community_members (community_id, user_id, role, status)
            VALUES (%s, %s, %s, 'active')
            ON CONFLICT (community_id, user_id) DO NOTHING;
        """, (comm_id, uid, role))

    cur.execute("""
        INSERT INTO admin_permissions (admin_user_id, community_id, can_verify_students, can_verify_alumni, can_manage_admins)
        VALUES (%s, %s, TRUE, TRUE, TRUE)
        ON CONFLICT (admin_user_id, community_id) DO NOTHING;
    """, (admin_id, comm_id))

    print("🌱 Seeding education and work experience...")
    cur.execute("""
        INSERT INTO education_details (user_id, degree_type, university_name, college_name, major, graduation_year, gpa)
        VALUES
            (%s, 'B Tech', 'University of Delhi', 'Cluster Innovation Centre', 'IT & Mathematical Innovations', 2026, 8.8),
            (%s, 'B Tech', 'University of Delhi', 'Cluster Innovation Centre', 'IT & Mathematical Innovations', 2022, 9.2)
        ON CONFLICT DO NOTHING;
    """, (student_id, alumni_id))

    cur.execute("""
        INSERT INTO work_experience (user_id, company_name, job_title, join_year, leave_year)
        VALUES (%s, 'Google', 'Senior Software Engineer', 2022, NULL)
        ON CONFLICT DO NOTHING;
    """, (alumni_id,))

    print("🌱 Seeding channels...")
    channel_names = [
        ('general', 'General discussion for all CIC members', 'text', 1),
        ('announcements', 'Official announcements from CIC administration', 'announcement', 2),
        ('introductions', 'Introduce yourself to the CIC community', 'text', 3),
        ('projects', 'Share and discuss your innovation projects', 'text', 4),
        ('startups', 'Startup discussions and entrepreneurship', 'text', 5),
        ('job-opportunities', 'Job postings and career opportunities', 'text', 6),
        ('tech-talk', 'Technology discussions and trends', 'text', 7),
    ]

    for name, desc, ctype, order in channel_names:
        cur.execute("""
            INSERT INTO channels (community_id, name, description, channel_type, created_by, position_order)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON CONFLICT DO NOTHING;
        """, (comm_id, name, desc, ctype, admin_id, order))

    cur.execute("SELECT channel_id, name FROM channels WHERE community_id = %s;", (comm_id,))
    channels = cur.fetchall()
    print(f"✅ {len(channels)} channels confirmed.")

    for ch_id, ch_name in channels:
        cur.execute("""
            INSERT INTO channel_messages (channel_id, user_id, content, message_type)
            VALUES (%s, %s, %s, 'text')
            ON CONFLICT DO NOTHING;
        """, (ch_id, admin_id, f"Welcome to #{ch_name}! Start connecting and sharing with your peers."))

    print("🌱 Seeding direct messages & connection...")
    cur.execute("""
        INSERT INTO connections (user_id, con_user_id, request, status)
        VALUES (%s, %s, 'Hi Priya! I would love mentorship on backend engineering.', 'accepted')
        ON CONFLICT DO NOTHING;
    """, (student_id, alumni_id))

    cur.execute("""
        INSERT INTO messages (sender_id, receiver_id, content)
        VALUES
            (%s, %s, 'Hi Priya, thanks for connecting! Looking forward to your guidance.'),
            (%s, %s, 'Hey Aarav! Glad to connect. Feel free to ask anything about career or tech.')
        ON CONFLICT DO NOTHING;
    """, (student_id, alumni_id, alumni_id, student_id))

    cur.close()
    conn.close()
    print("🎉 Database seeded successfully!")

if __name__ == "__main__":
    seed()
