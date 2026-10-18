from flask import Blueprint, render_template, session, redirect, url_for, flash, request
from algo.db import get_db
from algo.auth.decorators import login_required
from algo.auth import user_roles
from algo import utils
import datetime

bp = Blueprint('connections', __name__)

@bp.route("/requests")
@login_required
def requests_page():
    """Requests page - manage connection requests"""
    user_id = session["user_id"]
    try:
        db = get_db()
        cur = db.cursor()
        cur.execute(
            "SELECT username, firstname, lastname, role, pfp_path FROM users WHERE user_id = %s",
            (user_id,),
        )
        current_user = cur.fetchone()
        cur.execute(
            """
            SELECT c.connection_id, c.user_id, c.request,
                   u.firstname, u.lastname, u.username, u.pfp_path
            FROM connections c
            JOIN users u ON c.user_id = u.user_id
            WHERE c.con_user_id = %s AND c.status = 'pending'
            ORDER BY c.connection_id DESC
            """,
            (user_id,),
        )
        pending_requests = []
        for row in cur.fetchall():
            pending_requests.append(
                {
                    "connection_id": row[0],
                    "user_id": row[1],
                    "message": row[2],
                    "created_at": "Recently",
                    "name": f"{row[3]} {row[4]}",
                    "username": row[5],
                    "avatar": row[6],
                }
            )
        cur.execute(
            """
            SELECT c.connection_id, c.con_user_id, c.request,
                   u.firstname, u.lastname, u.username, u.pfp_path
            FROM connections c
            JOIN users u ON c.con_user_id = u.user_id
            WHERE c.user_id = %s AND c.status = 'pending'
            ORDER BY c.connection_id DESC
            """,
            (user_id,),
        )
        sent_requests = []
        for row in cur.fetchall():
            sent_requests.append(
                {
                    "connection_id": row[0],
                    "user_id": row[1],
                    "message": row[2],
                    "created_at": "Recently",
                    "name": f"{row[3]} {row[4]}",
                    "username": row[5],
                    "avatar": row[6],
                }
            )
        cur.execute(
            """
            SELECT DISTINCT
                CASE 
                    WHEN c.user_id = %s THEN c.con_user_id 
                    ELSE c.user_id 
                END as other_user_id,
                CASE 
                    WHEN c.user_id = %s THEN u2.firstname 
                    ELSE u1.firstname 
                END as firstname,
                CASE 
                    WHEN c.user_id = %s THEN u2.lastname 
                    ELSE u1.lastname 
                END as lastname,
                CASE 
                    WHEN c.user_id = %s THEN u2.username 
                    ELSE u1.username 
                END as username,
                CASE 
                    WHEN c.user_id = %s THEN u2.pfp_path 
                    ELSE u1.pfp_path 
                END as avatar,
                c.connection_id
            FROM connections c
            JOIN users u1 ON c.user_id = u1.user_id
            JOIN users u2 ON c.con_user_id = u2.user_id
            WHERE (c.user_id = %s OR c.con_user_id = %s) AND c.status = 'accepted'
            ORDER BY c.connection_id DESC
            """,
            (user_id, user_id, user_id, user_id, user_id, user_id, user_id),
        )
        connections = []
        for row in cur.fetchall():
            connections.append(
                {
                    "user_id": row[0],
                    "name": f"{row[1]} {row[2]}",
                    "username": row[3],
                    "avatar": row[4],
                    "connected_at": "Recently",
                }
            )
        connections_count = len(connections)
        
        return render_template(
            "requests.html",
            pending_requests=pending_requests,
            sent_requests=sent_requests,
            connections=connections,
            connections_count=connections_count,
            current_user=current_user,
        )
    except Exception as e:
        flash("Error loading requests page")
        return redirect(url_for("dashboard.user_dashboard"))
    finally:
        cur.close()

@bp.route("/connect")
@login_required
@user_roles.verified_user_required
def connect():
    """Connect page - show alumni for networking"""
    user_id = session["user_id"]
    try:
        db = get_db()
        cur = db.cursor()
        cur.execute(
            "SELECT username, firstname, lastname, role, pfp_path, university_name, graduation_year, current_city FROM users WHERE user_id = %s",
            (user_id,),
        )
        current_user = cur.fetchone()
        cur.execute(
            """
            SELECT u.user_id, u.firstname, u.lastname, u.username, u.university_name, 
                   u.graduation_year, u.current_city, u.pfp_path, u.role,
                   we.company_name, we.job_title,
                   STRING_AGG(i.name, ', ') as interests
            FROM users u
            LEFT JOIN work_experience we ON u.user_id = we.user_id AND we.leave_year IS NULL
            LEFT JOIN user_interests ui ON u.user_id = ui.user_id
            LEFT JOIN interests i ON ui.interest_id = i.interest_id
            WHERE u.user_id != %s AND u.verification_status = 'verified'
            GROUP BY u.user_id, u.firstname, u.lastname, u.username, u.university_name, 
                     u.graduation_year, u.current_city, u.pfp_path, u.role,
                     we.company_name, we.job_title
            ORDER BY u.lastname, u.firstname
            """,
            (user_id,),
        )
        all_users = cur.fetchall()
        cur.execute(
            """
            SELECT con_user_id, status, connection_id FROM connections 
            WHERE user_id = %s
            """,
            (user_id,),
        )
        connections = {
            row[0]: {"status": row[1], "connection_id": row[2]}
            for row in cur.fetchall()
        }
        cur.execute(
            """
            SELECT user_id, status, connection_id FROM connections 
            WHERE con_user_id = %s
            """,
            (user_id,),
        )
        reverse_connections = {
            row[0]: {"status": row[1], "connection_id": row[2]}
            for row in cur.fetchall()
        }
        people_data = []
        for user in all_users:
            name = f"{user[1]} {user[2]}"
            avatar = user[7] if user[7] else utils.generate_default_avatar(name)
            user_data = {
                "id": user[0],
                "name": name,
                "username": user[3],
                "university": user[4] or "Not specified",
                "graduation_year": user[5] or "Not specified",
                "location": user[6] or "Not specified",
                "avatar": avatar,
                "role": user[8] or "unverified",
                "company": user[9] or "Not specified",
                "title": user[10] or "Not specified",
                "interests": user[11].split(", ") if user[11] else [],
                "connection_status": "none",
            }
            if user[0] in connections:
                user_data["connection_status"] = connections[user[0]]["status"]
                user_data["connection_id"] = connections[user[0]]["connection_id"]
            elif user[0] in reverse_connections:
                if reverse_connections[user[0]]["status"] == "accepted":
                    user_data["connection_status"] = "connected"
                elif reverse_connections[user[0]]["status"] == "pending":
                    user_data["connection_status"] = "received_request"
                    user_data["connection_id"] = reverse_connections[user[0]][
                        "connection_id"
                    ]
            people_data.append(user_data)
        universities = list(
            set(
                [
                    user["university"]
                    for user in people_data
                    if user["university"] != "Not specified"
                ]
            )
        )
        graduation_years = list(
            set(
                [
                    str(user["graduation_year"])
                    for user in people_data
                    if user["graduation_year"] != "Not specified"
                ]
            )
        )
        locations = list(
            set(
                [
                    user["location"]
                    for user in people_data
                    if user["location"] != "Not specified"
                ]
            )
        )
        universities.sort()
        graduation_years.sort(reverse=True)
        locations.sort()
        return render_template(
            "connect.html",
            people=people_data,
            universities=universities,
            graduation_years=graduation_years,
            locations=locations,
            current_user=current_user,
        )
    except Exception as e:
        import traceback
        traceback.print_exc()
        flash("Error loading connections page")
        return redirect(url_for("dashboard.user_dashboard"))
    finally:
        if 'cur' in locals() and cur is not None:
            cur.close()

@bp.route("/api/send_connection_request", methods=["POST"])
@login_required
@user_roles.verified_user_required
def send_connection_request():
    """Send a connection request to another user"""
    try:
        data = request.get_json()
        target_user_id = data.get("user_id")
        message = data.get("message", "")
        if not target_user_id:
            return ({"success": False, "message": "User ID is required"}, 400)
        user_id = session["user_id"]
        if user_id == target_user_id:
            return ({"success": False, "message": "Cannot connect to yourself"}, 400)
            
        db = get_db()
        cur = db.cursor()
        cur.execute(
            """
            SELECT connection_id, status FROM connections 
            WHERE (user_id = %s AND con_user_id = %s) OR (user_id = %s AND con_user_id = %s)
            """,
            (user_id, target_user_id, target_user_id, user_id),
        )
        existing_connection = cur.fetchone()
        if existing_connection:
            status = existing_connection[1]
            if status == "accepted":
                return ({"success": False, "message": "Already connected"}, 400)
            elif status == "pending":
                return (
                    {"success": False, "message": "Connection request already sent"},
                    400,
                )
        cur.execute(
            """
            INSERT INTO connections (user_id, con_user_id, request, status)
            VALUES (%s, %s, %s, 'pending')
            """,
            (user_id, target_user_id, message),
        )
        db.commit()
        cur.execute(
            "SELECT firstname, lastname FROM users WHERE user_id = %s",
            (target_user_id,),
        )
        target_user = cur.fetchone()
        target_name = f"{target_user[0]} {target_user[1]}" if target_user else "User"
        return {
            "success": True,
            "message": f"Connection request sent to {target_name}",
            "status": "pending",
        }
    except Exception as e:
        return ({"success": False, "message": "Error sending connection request"}, 500)
    finally:
        cur.close()

@bp.route("/api/respond_connection_request", methods=["POST"])
@login_required
def respond_connection_request():
    """Accept or reject a connection request"""
    try:
        data = request.get_json()
        connection_id = data.get("connection_id")
        action = data.get("action")
        if not connection_id or action not in ["accept", "reject"]:
            return ({"success": False, "message": "Invalid request"}, 400)
        user_id = session["user_id"]
        
        db = get_db()
        cur = db.cursor()
        cur.execute(
            """
            SELECT user_id, con_user_id, status FROM connections 
            WHERE connection_id = %s AND con_user_id = %s AND status = 'pending'
            """,
            (connection_id, user_id),
        )
        connection_record = cur.fetchone()
        if not connection_record:
            return ({"success": False, "message": "Connection request not found"}, 404)
        new_status = "accepted" if action == "accept" else "denied"
        cur.execute(
            """
            UPDATE connections SET status = %s WHERE connection_id = %s
            """,
            (new_status, connection_id),
        )
        db.commit()
        return {
            "success": True,
            "message": f"Connection request {action}ed",
            "status": new_status,
        }
    except Exception as e:
        return ({"success": False, "message": "Error processing request"}, 500)
    finally:
        cur.close()

@bp.route("/api/get_connection_requests")
@login_required
def get_connection_requests():
    """Get pending connection requests for current user"""
    try:
        user_id = session["user_id"]
        db = get_db()
        cur = db.cursor()
        cur.execute(
            """
            SELECT c.connection_id, c.user_id, c.request,
                   u.firstname, u.lastname, u.username, u.pfp_path
            FROM connections c
            JOIN users u ON c.user_id = u.user_id
            WHERE c.con_user_id = %s AND c.status = 'pending'
            ORDER BY c.connection_id DESC
            """,
            (user_id,),
        )
        requests = []
        for row in cur.fetchall():
            requests.append(
                {
                    "connection_id": row[0],
                    "user_id": row[1],
                    "message": row[2],
                    "created_at": "Recently",
                    "name": f"{row[3]} {row[4]}",
                    "username": row[5],
                    "avatar": row[6],
                }
            )
        return {"success": True, "requests": requests}
    except Exception as e:
        return ({"success": False, "message": "Error loading requests"}, 500)
    finally:
        cur.close()

@bp.route("/api/cancel_connection_request", methods=["POST"])
@login_required
def cancel_connection_request():
    """Cancel a sent connection request"""
    try:
        data = request.get_json()
        connection_id = data.get("connection_id")
        if not connection_id:
            return ({"success": False, "message": "Connection ID is required"}, 400)
        user_id = session["user_id"]
        db = get_db()
        cur = db.cursor()
        cur.execute(
            """
            SELECT user_id, con_user_id, status FROM connections 
            WHERE connection_id = %s AND user_id = %s AND status = 'pending'
            """,
            (connection_id, user_id),
        )
        connection = cur.fetchone()
        if not connection:
            return ({"success": False, "message": "Connection request not found"}, 404)
        cur.execute(
            """
            DELETE FROM connections WHERE connection_id = %s
            """,
            (connection_id,),
        )
        db.commit()
        return {"success": True, "message": "Connection request cancelled successfully"}
    except Exception as e:
        return ({"success": False, "message": "Error cancelling request"}, 500)
    finally:
        cur.close()
