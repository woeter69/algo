from flask import Blueprint, render_template, session, redirect, url_for, flash, request
from algo.db import get_db
from algo.auth.decorators import login_required
from algo.auth import user_roles

bp = Blueprint('dashboard', __name__)

@bp.route('/dashboard')
@bp.route('/user_dashboard')
@login_required
def user_dashboard():
    user_id = session["user_id"]

    db = get_db()
    cur = db.cursor()
    cur.execute(
        "SELECT username, role, pfp_path, verification_status FROM users WHERE user_id = %s",
        (user_id,),
    )
    row = cur.fetchone()
    cur.close()

    if not row:
        flash("User not found", "error")
        return redirect(url_for("auth.login"))

    username, role, pfp_path, verification_status = row

    if role == "unverified" or (role != "admin" and verification_status == "pending"):
        return redirect(url_for("dashboard.limited_dashboard"))

    # Fetch dynamic user metrics
    cur = db.cursor()
    try:
        cur.execute("SELECT COUNT(*) FROM users WHERE role = 'alumni' AND verification_status = 'verified'")
        total_alumni = cur.fetchone()[0]

        cur.execute("""
            SELECT COUNT(*) FROM connections 
            WHERE (user_id = %s OR con_user_id = %s) AND status = 'accepted'
        """, (user_id, user_id))
        total_connections = cur.fetchone()[0]

        cur.execute("SELECT COUNT(*) FROM channels WHERE is_active = true")
        total_channels = cur.fetchone()[0]

        cur.execute("SELECT COUNT(*) FROM connections WHERE con_user_id = %s AND status = 'pending'", (user_id,))
        pending_requests = cur.fetchone()[0]

        # Recent activities (incoming connection requests)
        cur.execute("""
            SELECT u.username, u.role
            FROM connections c
            JOIN users u ON c.user_id = u.user_id
            WHERE c.con_user_id = %s AND c.status = 'pending'
            ORDER BY c.connection_id DESC LIMIT 4
        """, (user_id,))
        recent_requests = [{"username": r[0], "role": r[1]} for r in cur.fetchall()]

        stats = {
            "total_alumni": total_alumni,
            "total_connections": total_connections,
            "total_channels": total_channels,
            "pending_requests": pending_requests,
        }
    except Exception as e:
        import logging
        logging.error(f"Error fetching dashboard metrics: {e}")
        stats = {
            "total_alumni": 0,
            "total_connections": 0,
            "total_channels": 0,
            "pending_requests": 0,
        }
        recent_requests = []
    finally:
        cur.close()

    return render_template(
        "user_dashboard.html",
        username=username,
        role=role,
        pfp_path=pfp_path,
        stats=stats,
        recent_requests=recent_requests,
    )

@bp.route("/admin_dashboard", methods=["GET", "POST"])
@login_required
@user_roles.admin_required
def admin_dashboard():
    """Admin dashboard for managing users and verification requests"""
    user_id = session["user_id"]
    try:
        db = get_db()
        cur = db.cursor()
        cur.execute(
            "SELECT username, role, pfp_path FROM users WHERE user_id = %s", (user_id,)
        )
        admin_info = cur.fetchone()
        if not admin_info:
            flash("Admin information not found", "error")
            return redirect(url_for("dashboard.user_dashboard"))
            
        username, role, pfp_path = admin_info
        
        try:
            cur.execute(
                "SELECT COUNT(*) FROM users WHERE verification_status = 'pending'"
            )
            pending_count_result = cur.fetchone()
            pending_count = pending_count_result[0] if pending_count_result else 0
            cur.execute(
                """
                SELECT 
                    u.user_id, 
                    u.username, 
                    u.email, 
                    COALESCE(vr.requested_role, u.role) as requested_role, 
                    u.pfp_path, 
                    COALESCE(vr.created_at, u.registration_date) as created_at,
                    c.name as community_name,
                    COALESCE(vr.department, u.department) as department,
                    COALESCE(vr.graduation_year, u.graduation_year) as graduation_year,
                    vr.student_id as verification_id,
                    vr.request_message as message
                FROM users u
                LEFT JOIN LATERAL (
                    SELECT * FROM verification_requests 
                    WHERE user_id = u.user_id 
                    ORDER BY created_at DESC LIMIT 1
                ) vr ON true
                LEFT JOIN communities c ON vr.community_id = c.community_id
                WHERE u.verification_status = 'pending'
                ORDER BY created_at DESC
                """
            )
            pending_requests_data = cur.fetchall()
            pending_requests = []
            for row in pending_requests_data:
                pending_requests.append(
                    {
                        "user_id": row[0],
                        "username": row[1],
                        "email": row[2],
                        "requested_role": row[3] or "student",
                        "pfp_path": row[4],
                        "created_at": row[5],
                        "community_name": row[6],
                        "department": row[7],
                        "graduation_year": row[8],
                        "verification_id": row[9],
                        "message": row[10],
                    }
                )
        except Exception as e:
            import logging
            logging.error(f"Error loading pending requests: {e}")
            pending_count = 0
            pending_requests = []
            
        try:
            cur.execute("SELECT COUNT(*) FROM users")
            total_users_result = cur.fetchone()
            total_users = total_users_result[0] if total_users_result else 0
        except Exception as e:
            total_users = 0
            
        try:
            cur.execute(
                "SELECT COUNT(*) FROM users WHERE verification_status = 'verified'"
            )
            verified_users_result = cur.fetchone()
            verified_users = verified_users_result[0] if verified_users_result else 0
        except Exception as e:
            verified_users = 0
            
        try:
            cur.execute(
                """
                SELECT COUNT(*) FROM users 
                WHERE registration_date >= NOW() - INTERVAL '7 days'
                """
            )
            recent_registrations_result = cur.fetchone()
            recent_registrations = (
                recent_registrations_result[0] if recent_registrations_result else 0
            )
        except Exception as e:
            recent_registrations = 0
            
        try:
            cur.execute("SELECT COUNT(*) FROM contacts WHERE status = 'pending'")
            pending_inquiries_result = cur.fetchone()
            pending_inquiries_count = pending_inquiries_result[0] if pending_inquiries_result else 0
            cur.execute("""
                SELECT c.id, c.full_name, c.email, c.phone, c.subject, c.message,
                       c.submitted_at, c.status, c.resolved_at, c.resolution_notes,
                       u.username as resolver_username
                FROM contacts c
                LEFT JOIN users u ON c.resolved_by = u.user_id
                ORDER BY CASE WHEN c.status = 'pending' THEN 0 ELSE 1 END, c.submitted_at DESC
            """)
            inquiries_data = cur.fetchall()
            contact_inquiries = []
            for row in inquiries_data:
                contact_inquiries.append({
                    "id": row[0],
                    "full_name": row[1],
                    "email": row[2],
                    "phone": row[3],
                    "subject": row[4],
                    "message": row[5],
                    "submitted_at": row[6],
                    "status": row[7] or 'pending',
                    "resolved_at": row[8],
                    "resolution_notes": row[9],
                    "resolver_username": row[10],
                })
        except Exception as e:
            pending_inquiries_count = 0
            contact_inquiries = []

        cur.close()
        
        return render_template(
            "admin_dashboard.html",
            username=username,
            role=role,
            pfp_path=pfp_path,
            pending_count=pending_count,
            pending_requests=pending_requests,
            total_users=total_users,
            total_verified=verified_users,
            recent_registrations=recent_registrations,
            contact_inquiries=contact_inquiries,
            pending_inquiries_count=pending_inquiries_count,
        )
    except Exception as e:
        flash("Error loading admin dashboard", "error")
        return redirect(url_for("dashboard.user_dashboard"))

@bp.route("/admin/contact/<int:query_id>/resolve", methods=["POST"])
@login_required
@user_roles.admin_required
def resolve_contact_query(query_id):
    """Mark a contact inquiry as resolved by the admin."""
    user_id = session["user_id"]
    resolution_notes = request.form.get("resolution_notes", "").strip()
    db = get_db()
    cur = db.cursor()
    try:
        final_notes = resolution_notes if resolution_notes else "Resolved via Admin Dashboard"
        cur.execute("""
            UPDATE contacts
            SET status = 'resolved',
                resolved_by = %s,
                resolved_at = NOW(),
                resolution_notes = %s
            WHERE id = %s
            RETURNING full_name, email, subject, message
        """, (user_id, final_notes, query_id))
        row = cur.fetchone()
        db.commit()

        if row:
            full_name, querier_email, subject, orig_msg = row
            resolver_username = session.get("username", "Admin")
            from algo.utils import send_inquiry_resolved_email
            send_inquiry_resolved_email(
                to_email=querier_email,
                full_name=full_name,
                subject=subject,
                resolution_notes=final_notes,
                original_message=orig_msg,
                resolver_name=resolver_username,
            )
            flash(f"Contact inquiry marked as resolved and email notification sent to {querier_email}.", "success")
        else:
            flash("Contact inquiry marked as resolved.", "success")
    except Exception as e:
        import logging
        logging.error(f"Error resolving contact query {query_id}: {e}")
        flash("Failed to resolve contact inquiry.", "error")
    finally:
        cur.close()

    return redirect(url_for("dashboard.admin_dashboard"))

@bp.route("/limited_dashboard", methods=["GET"])
@login_required
def limited_dashboard():
    """Dashboard for unverified users with limited access"""
    user_id = session["user_id"]
    db = get_db()
    cur = db.cursor()
    try:
        cur.execute(
            """
            SELECT firstname, lastname, username, role, verification_status, pfp_path 
            FROM users WHERE user_id = %s
            """,
            (user_id,),
        )
        user_row = cur.fetchone()
        if not user_row:
            flash("User not found", "error")
            return redirect(url_for("auth.login"))
            
        user = {
            "firstname": user_row[0],
            "lastname": user_row[1],
            "username": user_row[2],
            "role": user_row[3],
            "verification_status": user_row[4],
            "pfp_path": user_row[5],
        }
        
        cur.execute("""
            SELECT request_id, created_at, status, requested_role, department, review_notes
            FROM verification_requests
            WHERE user_id = %s
            ORDER BY created_at DESC
            LIMIT 1
        """, (user_id,))
        vr_row = cur.fetchone()
        verification_req = None
        if vr_row:
            verification_req = {
                "request_id": vr_row[0],
                "created_at": vr_row[1],
                "status": vr_row[2],
                "requested_role": vr_row[3],
                "department": vr_row[4],
                "review_notes": vr_row[5],
            }
            
        return render_template(
            "limited_dashboard.html", user=user, verification_request=verification_req
        )
    except Exception as e:
        import logging
        logging.error(f"Error loading limited dashboard: {e}")
        flash("An error occurred loading the dashboard.", "error")
        return redirect(url_for("core.home"))
    finally:
        if cur:
            cur.close()

@bp.route("/verification_request", methods=["GET", "POST"])
@login_required
def verification_request():
    """Handle verification requests from users"""
    user_id = session["user_id"]
    db = get_db()
    cur = db.cursor()
    try:
        if request.method == "POST":
            college_id = request.form.get("college_id", type=int)
            requested_role = request.form.get("requested_role", "student")
            student_id = request.form.get("student_id", "").strip()
            graduation_year = request.form.get("graduation_year", type=int)
            department = request.form.get("department", "").strip()
            request_message = request.form.get("request_message", "").strip()

            cur.execute("""
                INSERT INTO verification_requests (
                    user_id, community_id, requested_role, student_id,
                    graduation_year, department, request_message, status, created_at, updated_at
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, 'pending', NOW(), NOW())
                RETURNING request_id
            """, (user_id, college_id, requested_role, student_id, graduation_year, department, request_message))

            cur.execute("""
                UPDATE users
                SET verification_status = 'pending',
                    role = %s,
                    department = %s,
                    graduation_year = %s
                WHERE user_id = %s
            """, (requested_role, department, graduation_year, user_id))

            db.commit()
            flash(
                "Your verification request has been submitted successfully! An admin will review it soon.",
                "success",
            )
            return redirect(url_for("dashboard.limited_dashboard"))

        cur.execute(
            """
            SELECT firstname, lastname, email, role, department, graduation_year
            FROM users WHERE user_id = %s
            """,
            (user_id,),
        )
        user_info = cur.fetchone()
        if not user_info:
            flash("User information not found", "error")
            return redirect(url_for("dashboard.limited_dashboard"))
        firstname, lastname, email, user_role, department, grad_year = user_info
        form_data = {
            "firstname": firstname,
            "lastname": lastname,
            "email": email,
            "requested_role": user_role if user_role != "unverified" else "student",
            "graduation_year": grad_year,
            "department": department or "",
        }

        cur.execute("SELECT community_id, name, location FROM communities WHERE is_active = true ORDER BY name ASC")
        communities = cur.fetchall()

        return render_template("verification_request.html", form_data=form_data, communities=communities)
    except Exception as e:
        db.rollback()
        import logging
        logging.error(f"Error in verification request: {e}")
        flash("An error occurred. Please try again.", "error")
        return redirect(url_for("dashboard.user_dashboard"))
    finally:
        cur.close()

@bp.route("/api/handle_verification_request", methods=["POST"])
@login_required
@user_roles.admin_required
def handle_verification_request():
    """Handle approve/reject verification requests"""
    try:
        data = request.get_json()
        user_id = data.get("user_id")
        action = data.get("action")
        if not user_id or action not in ["approve", "reject"]:
            return ({"success": False, "message": "Invalid request data"}, 400)
        admin_id = session["user_id"]
        db = get_db()
        cur = db.cursor()
        cur.execute(
            """
            SELECT username, email, role FROM users 
            WHERE user_id = %s AND verification_status = 'pending'
            """,
            (user_id,),
        )
        user_info = cur.fetchone()
        if not user_info:
            return (
                {
                    "success": False,
                    "message": "User not found or not pending verification",
                },
                404,
            )
        username, email, current_role = user_info
        if action == "approve":
            cur.execute("""
                SELECT requested_role, community_id 
                FROM verification_requests 
                WHERE user_id = %s 
                ORDER BY created_at DESC LIMIT 1
            """, (user_id,))
            req_info = cur.fetchone()
            approved_role = req_info[0] if req_info and req_info[0] else current_role
            comm_id = req_info[1] if req_info else None

            cur.execute(
                """
                UPDATE users 
                SET verification_status = 'verified', 
                    role = %s,
                    verified_by = %s, 
                    verified_at = NOW()
                WHERE user_id = %s
                """,
                (approved_role, admin_id, user_id),
            )

            cur.execute(
                """
                UPDATE verification_requests
                SET status = 'approved',
                    reviewed_by = %s,
                    reviewed_at = NOW(),
                    review_notes = 'Approved by administrator'
                WHERE user_id = %s AND status = 'pending'
                """,
                (admin_id, user_id),
            )

            if comm_id:
                cur.execute(
                    """
                    INSERT INTO community_members (community_id, user_id, role, status, joined_at)
                    VALUES (%s, %s, 'member', 'active', NOW())
                    ON CONFLICT (community_id, user_id) DO UPDATE SET status = 'active'
                    """,
                    (comm_id, user_id),
                )

            message = f"Verification request approved for {username}"
        else:
            cur.execute(
                """
                UPDATE users 
                SET verification_status = 'rejected',
                    role = 'unverified',
                    verified_by = %s,
                    verified_at = NOW()
                WHERE user_id = %s
                """,
                (admin_id, user_id),
            )

            cur.execute(
                """
                UPDATE verification_requests
                SET status = 'rejected',
                    reviewed_by = %s,
                    reviewed_at = NOW(),
                    review_notes = 'Request rejected by administrator'
                WHERE user_id = %s AND status = 'pending'
                """,
                (admin_id, user_id),
            )
            message = f"Verification request rejected for {username}"
        db.commit()
        return {
            "success": True,
            "message": message,
            "action": action,
            "user_id": user_id,
        }
    except Exception as e:
        db.rollback()
        import logging
        logging.error(f"Error handling verification request: {e}")
        return ({"success": False, "message": "Internal server error"}, 500)
    finally:
        cur.close()
