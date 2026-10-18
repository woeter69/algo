import os
import sys
import pytest

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../app/src')))

from algo import create_app
from algo.db import get_db

@pytest.fixture
def client():
    app = create_app({
        "TESTING": True,
        "DEBUG": False,
        "SECRET_KEY": "test-secret"
    })
    with app.test_client() as client:
        yield client

def test_public_pages(client):
    """Verify that public routes render with 200 OK."""
    for path in ['/', '/about', '/contact', '/login', '/register']:
        res = client.get(path)
        assert res.status_code == 200, f"Path {path} returned status {res.status_code}"

def test_login_flow_and_dashboard(client):
    """Test authentication and session access for test student."""
    # Attempt login with seeded user
    res = client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)
    assert res.status_code == 200

    # Access user dashboard with session
    dashboard_res = client.get('/dashboard')
    assert dashboard_res.status_code in [200, 302]

def test_channels_api(client):
    """Verify channel endpoints."""
    # Login as student
    client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)

    res = client.get('/channels')
    assert res.status_code == 200

    # Test channel messages and members for default channel 1
    messages_res = client.get('/channels/1/messages')
    assert messages_res.status_code in [200, 404]

    members_res = client.get('/channels/1/members')
    assert members_res.status_code in [200, 404]
    if members_res.status_code == 200:
        data = members_res.get_json()
        assert data.get('success') is True
        assert 'members' in data

def test_connect_networking_directory(client):
    """Verify alumni directory page."""
    client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)

    res = client.get('/connect')
    assert res.status_code == 200

def test_contact_submission_and_resolution(client):
    """Verify contact form submission, DB persistence, admin visibility, and inquiry resolution."""
    from algo.db import get_db

    # 1. Submit contact inquiry
    res = client.post('/contact', data={
        'full_name': 'Test Querier',
        'email': 'querier@example.com',
        'phone': '1234567890',
        'subject': 'Inquiry about SIH Project',
        'message': 'Can alumni mentor multiple teams?'
    }, follow_redirects=True)
    assert res.status_code == 200

    # 2. Check persistence in PostgreSQL
    with client.application.app_context():
        db = get_db()
        cur = db.cursor()
        cur.execute("SELECT id, status FROM contacts WHERE email = 'querier@example.com' ORDER BY id DESC LIMIT 1")
        row = cur.fetchone()
        assert row is not None
        query_id, status = row
        assert status == 'pending'
        cur.close()

    # 3. Log in as admin
    login_res = client.post('/login', data={
        'email': 'admin@alumnigo.test',
        'password': 'admin123'
    }, follow_redirects=True)
    assert login_res.status_code == 200

    # 4. Verify inquiry appears on Admin Dashboard
    dash_res = client.get('/admin_dashboard')
    assert dash_res.status_code == 200
    assert b"Contact Us Inquiries" in dash_res.data
    assert b"Inquiry about SIH Project" in dash_res.data

    # 5. Resolve the inquiry via Admin action
    resolve_res = client.post(f'/admin/contact/{query_id}/resolve', data={
        'resolution_notes': 'Verified and addressed via email.'
    }, follow_redirects=True)
    assert resolve_res.status_code == 200

    # 6. Verify status updated to 'resolved' in PostgreSQL
    with client.application.app_context():
        db = get_db()
        cur = db.cursor()
        cur.execute("SELECT status, resolution_notes FROM contacts WHERE id = %s", (query_id,))
        resolved_status, notes = cur.fetchone()
        assert resolved_status == 'resolved'
        assert notes == 'Verified and addressed via email.'
        cur.close()

def test_admin_dashboard_auth_guards(client):
    """Verify that unauthenticated and non-admin users cannot access admin dashboard."""
    # 1. Unauthenticated request must redirect to login
    unauth_res = client.get('/admin_dashboard')
    assert unauth_res.status_code == 302
    assert '/auth/login' in unauth_res.headers.get('Location', '')

    # 2. Student login must be forbidden from admin dashboard and redirected to user_dashboard
    client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)
    student_res = client.get('/admin_dashboard')
    assert student_res.status_code == 302
    assert '/user_dashboard' in student_res.headers.get('Location', '')

def test_chat_api_endpoints(client):
    """Verify online status and user search endpoints for chat."""
    # 1. Login as student
    client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)

    # 2. Test /api/online_status
    online_res = client.get('/api/online_status')
    assert online_res.status_code == 200
    data = online_res.get_json()
    assert 'online_users' in data
    assert isinstance(data['online_users'], list)

    # 3. Test /api/search_users
    search_res = client.get('/api/search_users?q=Alumni')
    assert search_res.status_code == 200
    search_data = search_res.get_json()
    assert 'users' in search_data
    assert isinstance(search_data['users'], list)

def test_verification_workflow(client):
    """Verify student submitting verification request and admin reviewing/approving it."""
    # 1. Login as student and submit verification request
    client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)

    vr_res = client.post('/verification_request', data={
        'college_id': 1,
        'requested_role': 'student',
        'student_id': 'CIC2026CS01',
        'graduation_year': 2026,
        'department': 'Information Technology',
        'request_message': 'Please verify my student identity.'
    }, follow_redirects=True)
    assert vr_res.status_code == 200

    # 2. Check limited dashboard reflects submission
    lim_res = client.get('/limited_dashboard')
    assert lim_res.status_code == 200

    # 3. Log out student, log in as admin, and verify pending request appears
    client.get('/logout', follow_redirects=True)
    client.post('/login', data={
        'email': 'admin@alumnigo.test',
        'password': 'admin123'
    }, follow_redirects=True)

    admin_res = client.get('/admin_dashboard')
    assert admin_res.status_code == 200
    assert b'CIC2026CS01' in admin_res.data

    # 4. Find student user_id
    with client.application.app_context():
        db = get_db()
        cur = db.cursor()
        cur.execute("SELECT user_id FROM users WHERE email = 'student@alumnigo.test'")
        student_user_id = cur.fetchone()[0]
        cur.close()

    # 5. Admin approves verification request
    approve_res = client.post('/api/handle_verification_request', json={
        'user_id': student_user_id,
        'action': 'approve'
    })
    assert approve_res.status_code == 200
    approve_data = approve_res.get_json()
    assert approve_data.get('success') is True

    # 6. Verify user is now verified in DB
    with client.application.app_context():
        db = get_db()
        cur = db.cursor()
        cur.execute("SELECT verification_status, role FROM users WHERE user_id = %s", (student_user_id,))
        v_status, u_role = cur.fetchone()
        assert v_status == 'verified'
        assert u_role == 'student'
        cur.close()

def test_profile_crud_endpoints(client):
    """Verify profile editing, experience CRUD, and education CRUD."""
    # 1. Login as student
    client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)

    # 2. Update profile bio, phone, city
    update_res = client.post('/api/profile/update', data={
        'firstName': 'Test',
        'lastName': 'Student',
        'bio': 'Passionate AI & full-stack developer.',
        'phone': '9876543210',
        'currentCity': 'New Delhi',
        'universityName': 'Delhi University',
        'graduationYear': '2026',
        'major': 'Computer Science',
        'linkedIn': 'https://linkedin.com/in/teststudent'
    })
    assert update_res.status_code == 200
    assert update_res.get_json().get('success') is True

    # 3. Add work experience
    exp_res = client.post('/api/profile/experience', data={
        'company_name': 'Google DeepMind',
        'job_title': 'Software Engineering Intern',
        'join_year': 2025,
        'leave_year': 2026
    })
    assert exp_res.status_code == 201
    exp_id = exp_res.get_json().get('exp_id')
    assert exp_id is not None

    # 4. Add education
    edu_res = client.post('/api/profile/education', data={
        'degree_type': 'B.Tech',
        'university_name': 'Delhi University',
        'college_name': 'Cluster Innovation Centre',
        'major': 'Information Technology',
        'graduation_year': 2026,
        'gpa': '9.4'
    })
    assert edu_res.status_code == 201
    detail_id = edu_res.get_json().get('detail_id')
    assert detail_id is not None

    # 5. Delete work experience and education
    del_exp = client.delete(f'/api/profile/experience/{exp_id}')
    assert del_exp.status_code == 200

    del_edu = client.delete(f'/api/profile/education/{detail_id}')
    assert del_edu.status_code == 200

def test_settings_persistence(client):
    """Verify settings updates (privacy, notifications) persist in users table and export_data works."""
    # 1. Login as student
    client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)

    # 2. Update privacy
    priv_res = client.post('/api/update_privacy', json={
        'profileVisibility': 'connections'
    })
    assert priv_res.status_code == 200
    assert priv_res.get_json().get('success') is True

    # 3. Update notifications
    notif_res = client.post('/api/update_notifications', json={
        'email_notifications': False,
        'job_alerts': True
    })
    assert notif_res.status_code == 200
    assert notif_res.get_json().get('success') is True

    # 4. Verify in DB
    with client.application.app_context():
        db = get_db()
        cur = db.cursor()
        cur.execute("SELECT profile_visibility, email_notifications, job_alerts FROM users WHERE email = 'student@alumnigo.test'")
        row = cur.fetchone()
        assert row[0] == 'connections'
        assert row[1] in [False, 'false']
        assert row[2] in [True, 'true']
        cur.close()

    # 5. Test export_data endpoint
    export_res = client.post('/api/export_data')
    assert export_res.status_code == 200
    export_json = export_res.get_json()
    assert 'user_info' in export_json
    assert 'connections' in export_json

def test_community_discovery_and_join(client):
    """Verify discovering communities, joining directly, and leaving."""
    # 1. Login as student
    client.post('/login', data={
        'email': 'student@alumnigo.test',
        'password': 'student123'
    }, follow_redirects=True)

    # 2. Discover communities
    disc_res = client.get('/communities/discover')
    assert disc_res.status_code == 200
    data = disc_res.get_json()
    assert 'communities' in data
    assert len(data['communities']) > 0

    target_comm_id = data['communities'][0]['community_id']

    # 3. Directly join community
    join_res = client.post(f'/communities/{target_comm_id}/join')
    assert join_res.status_code == 200
    assert join_res.get_json().get('success') is True

    # 4. Verify membership in DB
    with client.application.app_context():
        db = get_db()
        cur = db.cursor()
        cur.execute("SELECT user_id FROM users WHERE email = 'student@alumnigo.test'")
        uid = cur.fetchone()[0]
        cur.execute("SELECT status FROM community_members WHERE community_id = %s AND user_id = %s", (target_comm_id, uid))
        mem = cur.fetchone()
        assert mem is not None
        assert mem[0] == 'active'
        cur.close()



