-- no CREATE DATABASE IF NOT EXISTS. In Postgres you run this once as superuser:
-- CREATE DATABASE alumni_platform;

-- Then connect to alumni_platform and run the rest:

-- Communities table (referenced by users and admin_permissions)
CREATE TABLE communities (
    community_id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    college_code VARCHAR(10),
    location VARCHAR(255),
    established_year INT,
    website VARCHAR(255),
    created_by INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Users table
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    firstname TEXT NOT NULL,
    lastname TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    registration_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verified BOOLEAN DEFAULT FALSE,
    dob DATE,
    graduation_year INT,
    university_name TEXT,
    department TEXT,
    college TEXT,
    current_city TEXT,
    pfp_path TEXT,
    role TEXT NOT NULL DEFAULT 'student' CHECK (role IN ('student', 'alumni', 'staff', 'admin')),
    enrollment_number TEXT UNIQUE,
    community_id INT,
    last_login TIMESTAMP,
    login_count INT DEFAULT 0,
    verified_by INT,
    verified_at TIMESTAMP,
    verification_status TEXT DEFAULT 'pending',
    phone VARCHAR(20),
    bio TEXT,
    linkedin VARCHAR(300),
    github VARCHAR(300),
    twitter VARCHAR(300),
    website VARCHAR(300),
    profile_visibility VARCHAR(20) DEFAULT 'public',
    email_notifications VARCHAR(20) DEFAULT 'enabled',
    job_alerts VARCHAR(20) DEFAULT 'enabled',
    student_id TEXT,
    alumni_id TEXT,
    employee_id TEXT,
    department_role TEXT,
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE SET NULL,
    FOREIGN KEY (verified_by) REFERENCES users(user_id) ON DELETE SET NULL
);

ALTER TABLE communities ADD CONSTRAINT fk_communities_created_by FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE SET NULL;

-- Create indexes for users table
CREATE INDEX idx_users_community ON users(community_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_verification_status ON users(verification_status);

CREATE TABLE verification_tokens (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expiry TIMESTAMP NOT NULL
);

CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    email TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expiry TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE interests (
    interest_id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE user_interests (
    user_id INT NOT NULL,
    interest_id INT NOT NULL,
    PRIMARY KEY(user_id, interest_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (interest_id) REFERENCES interests(interest_id) ON DELETE CASCADE
);

INSERT INTO interests (name) VALUES
('Career Guidance'),
('Internships / Job Opportunities'),
('Research Projects'),
('Startups & Entrepreneurship'),
('Technology / Software Development'),
('Data Science / AI / Machine Learning'),
('Engineering / Design'),
('Business / Management'),
('Networking'),
('Clubs & Societies'),
('Sports & Fitness'),
('Cultural Activities'),
('Events & Workshops'),
('Competitions / Hackathons'),
('Mentorship'),
('Volunteering / Social Work'),
('Travel'),
('Music / Arts / Creative Work');

CREATE TABLE connections (
    connection_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    con_user_id INT NOT NULL,
    request TEXT,
    status TEXT CHECK (status IN ('pending','accepted','denied')) DEFAULT 'pending',
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (con_user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE education_details (
    detail_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    degree_type TEXT CHECK (degree_type IN ('Bachelors','Masters','PHD','Doctorate','B Tech','M Tech','B.E.','M.E.','B.Sc.','M.Sc.','BCA','MCA','MBA','BBA','Diploma')) NOT NULL,
    university_name TEXT NOT NULL,
    college_name TEXT,
    major TEXT NOT NULL,
    graduation_year INT,
    gpa NUMERIC(4,2) CHECK (gpa >= 0.00 AND gpa <= 10.00),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    subject TEXT,
    message TEXT NOT NULL,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'resolved')),
    resolved_by INT,
    resolved_at TIMESTAMP,
    resolution_notes TEXT,
    FOREIGN KEY (resolved_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE INDEX idx_contacts_status ON contacts(status);

CREATE TABLE work_experience (
    exp_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    company_name TEXT NOT NULL,
    job_title TEXT NOT NULL,
    join_year INT NOT NULL,
    leave_year INT,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE messages (
    message_id SERIAL PRIMARY KEY,
    sender_id INT NOT NULL,
    receiver_id INT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (receiver_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Admin permissions table
CREATE TABLE admin_permissions (
    permission_id SERIAL PRIMARY KEY,
    admin_user_id INT NOT NULL,
    community_id INT,
    can_verify_students BOOLEAN DEFAULT TRUE,
    can_verify_alumni BOOLEAN DEFAULT TRUE,
    can_manage_admins BOOLEAN DEFAULT FALSE,
    granted_by INT,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(admin_user_id, community_id),
    FOREIGN KEY (admin_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE SET NULL,
    FOREIGN KEY (granted_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create indexes for admin_permissions table
CREATE INDEX idx_admin_permissions_admin ON admin_permissions(admin_user_id);
CREATE INDEX idx_admin_permissions_community ON admin_permissions(community_id);

-- Verification requests table (referenced by communities)
CREATE TABLE verification_requests (
    request_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    community_id INT,
    requested_role TEXT NOT NULL DEFAULT 'student',
    student_id TEXT,
    graduation_year INT,
    department TEXT,
    university_name TEXT,
    college TEXT,
    request_message TEXT,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by INT,
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE SET NULL,
    FOREIGN KEY (reviewed_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create indexes for verification_requests table
CREATE INDEX idx_verification_requests_user_id ON verification_requests(user_id);
CREATE INDEX idx_verification_requests_community ON verification_requests(community_id);
CREATE INDEX idx_verification_requests_community_id ON verification_requests(community_id);
CREATE INDEX idx_verification_requests_status ON verification_requests(status);

-- Community members table (users belonging to communities)
CREATE TABLE community_members (
    member_id SERIAL PRIMARY KEY,
    community_id INT NOT NULL,
    user_id INT NOT NULL,
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('admin', 'moderator', 'member')),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    invited_by INT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'banned', 'left')),
    UNIQUE(community_id, user_id),
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (invited_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Community join requests table
CREATE TABLE community_join_requests (
    request_id SERIAL PRIMARY KEY,
    community_id INT NOT NULL,
    user_id INT NOT NULL,
    message TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP,
    reviewed_by INT,
    review_notes TEXT,
    UNIQUE(community_id, user_id),
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (reviewed_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Community invitations table
CREATE TABLE community_invitations (
    invitation_id SERIAL PRIMARY KEY,
    community_id INT NOT NULL,
    invited_user_id INT NOT NULL,
    invited_by INT NOT NULL,
    message TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '7 days'),
    responded_at TIMESTAMP,
    UNIQUE(community_id, invited_user_id),
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE,
    FOREIGN KEY (invited_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (invited_by) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Channels table for Discord-like chat
CREATE TABLE channels (
    channel_id SERIAL PRIMARY KEY,
    community_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    channel_type VARCHAR(20) DEFAULT 'text' CHECK (channel_type IN ('text', 'voice', 'announcement')),
    is_private BOOLEAN DEFAULT FALSE,
    created_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    position_order INT DEFAULT 0,
    FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Channel messages table
CREATE TABLE channel_messages (
    message_id SERIAL PRIMARY KEY,
    channel_id INT NOT NULL,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file', 'system')),
    reply_to_message_id INT,
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL,
    FOREIGN KEY (reply_to_message_id) REFERENCES channel_messages(message_id) ON DELETE SET NULL
);

-- Channel members table (for private channels)
CREATE TABLE channel_members (
    channel_id INT NOT NULL,
    user_id INT NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('admin', 'moderator', 'member')),
    can_send_messages BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (channel_id, user_id),
    FOREIGN KEY (channel_id) REFERENCES channels(channel_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Message reactions table
CREATE TABLE message_reactions (
    reaction_id SERIAL PRIMARY KEY,
    message_id INT NOT NULL,
    user_id INT NOT NULL,
    emoji VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, user_id, emoji),
    FOREIGN KEY (message_id) REFERENCES channel_messages(message_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create indexes for community system
CREATE INDEX idx_community_members_community ON community_members(community_id);
CREATE INDEX idx_community_members_user ON community_members(user_id);
CREATE INDEX idx_community_members_role ON community_members(role);
CREATE INDEX idx_community_join_requests_community ON community_join_requests(community_id);
CREATE INDEX idx_community_join_requests_user ON community_join_requests(user_id);
CREATE INDEX idx_community_join_requests_status ON community_join_requests(status);
CREATE INDEX idx_community_invitations_community ON community_invitations(community_id);
CREATE INDEX idx_community_invitations_user ON community_invitations(invited_user_id);
CREATE INDEX idx_community_invitations_status ON community_invitations(status);

-- Create indexes for channels system
CREATE INDEX idx_channels_community ON channels(community_id);
CREATE INDEX idx_channels_active ON channels(is_active);
CREATE INDEX idx_channel_messages_channel ON channel_messages(channel_id);
CREATE INDEX idx_channel_messages_user ON channel_messages(user_id);
CREATE INDEX idx_channel_messages_created_at ON channel_messages(created_at);
CREATE INDEX idx_channel_members_user ON channel_members(user_id);
CREATE INDEX idx_message_reactions_message ON message_reactions(message_id);
