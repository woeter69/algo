from flask import Blueprint, render_template, session, redirect, url_for, flash, request, jsonify, current_app
from algo.db import get_db
from algo.auth.decorators import login_required
import datetime
from algo import utils

bp = Blueprint('settings', __name__)

@bp.route("/settings")
@login_required
def settings_page():
    """User settings page"""
    try:
        user_id = session["user_id"]
        db = get_db()
        cur = db.cursor()
        cur.execute(
            """
            SELECT user_id, firstname, lastname, email, username, dob,
                   university_name, department, college, graduation_year, current_city, 
                   pfp_path, role, community_id, last_login, login_count,
                   bio, profile_visibility, email_notifications, job_alerts,
                   linkedin, github, twitter, website, phone
            FROM users 
            WHERE user_id = %s
            """,
            (user_id,),
        )
        user_data = cur.fetchone()
        if user_data:
            user_dict = {
                "user_id": user_data[0],
                "firstname": user_data[1],
                "lastname": user_data[2],
                "email": user_data[3],
                "username": user_data[4],
                "dob": user_data[5],
                "university_name": user_data[6],
                "department": user_data[7],
                "college": user_data[8],
                "graduation_year": user_data[9],
                "current_city": user_data[10],
                "pfp_path": user_data[11],
                "role": user_data[12],
                "community_id": user_data[13],
                "last_login": user_data[14],
                "login_count": user_data[15],
                "bio": user_data[16] or "",
                "profile_visibility": user_data[17] or "public",
                "email_notifications": True if user_data[18] is None else bool(user_data[18]),
                "job_alerts": True if user_data[19] is None else bool(user_data[19]),
                "linkedin": user_data[20] or "",
                "github": user_data[21] or "",
                "twitter": user_data[22] or "",
                "website": user_data[23] or "",
                "phone": user_data[24] or "",
                "show_email": False,
                "allow_messages": True,
            }
        else:
            user_dict = {}
        return render_template("settings.html", user_data=user_dict)
    except Exception as e:
        current_app.logger.error(f"Error loading settings: {str(e)}")
        flash("Error loading settings page")
        return redirect(url_for("dashboard.user_dashboard"))
    finally:
        cur.close()

@bp.route("/api/update_account", methods=["POST"])
@login_required
def update_account():
    try:
        data = request.get_json()
        user_id = session["user_id"]
        db = get_db()
        cur = db.cursor()
        if data.get("username"):
            cur.execute(
                "SELECT user_id FROM users WHERE username = %s AND user_id != %s",
                (data["username"], user_id),
            )
            if cur.fetchone():
                return ({"success": False, "message": "Username already taken"}, 400)
        update_fields = []
        update_values = []
        if data.get("username"):
            update_fields.append("username = %s")
            update_values.append(data["username"])
        if data.get("firstName"):
            update_fields.append("firstname = %s")
            update_values.append(data["firstName"])
        if data.get("lastName"):
            update_fields.append("lastname = %s")
            update_values.append(data["lastName"])
        if data.get("university"):
            update_fields.append("university_name = %s")
            update_values.append(data["university"])
        if data.get("department"):
            update_fields.append("department = %s")
            update_values.append(data["department"])
        if data.get("college"):
            update_fields.append("college = %s")
            update_values.append(data["college"])
        if data.get("graduationYear"):
            update_fields.append("graduation_year = %s")
            update_values.append(data["graduationYear"])
        if data.get("currentCity"):
            update_fields.append("current_city = %s")
            update_values.append(data["currentCity"])
        if update_fields:
            update_values.append(user_id)
            query = f"UPDATE users SET {', '.join(update_fields)} WHERE user_id = %s"
            cur.execute(query, update_values)
            db.commit()
            if data.get("username"):
                session["username"] = data["username"]
        return {"success": True, "message": "Account updated successfully"}
    except Exception as e:
        return ({"success": False, "message": "Error updating account"}, 500)
    finally:
        cur.close()

@bp.route("/api/update_privacy", methods=["POST"])
@login_required
def update_privacy():
    try:
        data = request.get_json(silent=True) or request.form or {}
        user_id = session["user_id"]
        visibility = data.get("profileVisibility") or data.get("profile_visibility", "public")
        db = get_db()
        cur = db.cursor()
        cur.execute(
            """
            UPDATE users SET 
                profile_visibility = %s
            WHERE user_id = %s
            """,
            (visibility, user_id),
        )
        db.commit()
        return jsonify({"success": True, "message": "Privacy settings updated successfully"})
    except Exception as e:
        return jsonify({"success": False, "message": "Error updating privacy settings"}), 500
    finally:
        cur.close()

@bp.route("/api/update_preferences", methods=["POST"])
@login_required
def update_preferences():
    try:
        data = request.get_json()
        return {"success": True, "message": "Preferences updated successfully"}
    except Exception as e:
        return ({"success": False, "message": "Error updating preferences"}, 500)

@bp.route("/api/update_notifications", methods=["POST"])
@login_required
def update_notifications():
    try:
        data = request.get_json(silent=True) or request.form or {}
        user_id = session["user_id"]
        email_notifications = data.get("email_notifications", data.get("emailNotifications", True))
        job_alerts = data.get("job_alerts", data.get("jobAlerts", True))
        db = get_db()
        cur = db.cursor()
        email_str = "true" if email_notifications in [True, "true", "True", "t", 1, "1"] else "false"
        job_str = "true" if job_alerts in [True, "true", "True", "t", 1, "1"] else "false"
        cur.execute(
            """
            UPDATE users 
            SET email_notifications = %s,
                job_alerts = %s
            WHERE user_id = %s
            """,
            (email_str, job_str, user_id),
        )
        db.commit()
        return {"success": True, "message": "Notifications updated successfully"}
    except Exception as e:
        return ({"success": False, "message": "Error updating notifications"}, 500)
    finally:
        cur.close()

@bp.route("/api/change_password", methods=["POST"])
@login_required
def change_password():
    try:
        from werkzeug.security import check_password_hash, generate_password_hash
        data = request.get_json()
        current_password = data.get("currentPassword")
        new_password = data.get("newPassword")
        user_id = session["user_id"]
        
        db = get_db()
        cur = db.cursor()
        cur.execute("SELECT password FROM users WHERE user_id = %s", (user_id,))
        result = cur.fetchone()
        
        if not result or not check_password_hash(result[0], current_password):
            return ({"success": False, "message": "Incorrect current password"}, 400)
            
        hashed_password = generate_password_hash(new_password)
        cur.execute(
            "UPDATE users SET password = %s WHERE user_id = %s",
            (hashed_password, user_id),
        )
        db.commit()
        return {"success": True, "message": "Password changed successfully"}
    except Exception as e:
        return ({"success": False, "message": "Error changing password"}, 500)
    finally:
        cur.close()

@bp.route("/api/update_profile_picture", methods=["POST"])
@login_required
def update_profile_picture():
    try:
        import os
        from algo import validators
        if "pfp" not in request.files:
            return ({"success": False, "message": "No image provided"}, 400)
            
        file = request.files["pfp"]
        if file.filename == "":
            return ({"success": False, "message": "No file selected"}, 400)
            
        if file and validators.allowed_file(file.filename):
            pfp_url = utils.upload_to_imgbb(file, os.getenv("PFP_API"))
            if pfp_url:
                user_id = session["user_id"]
                db = get_db()
                cur = db.cursor()
                cur.execute(
                    "UPDATE users SET pfp_path = %s WHERE user_id = %s",
                    (pfp_url, user_id),
                )
                db.commit()
                session["pfp_path"] = pfp_url
                return {"success": True, "pfp_url": pfp_url}
        return ({"success": False, "message": "Failed to upload image"}, 400)
    except Exception as e:
        return ({"success": False, "message": "Error updating profile picture"}, 500)

@bp.route("/api/deactivate_account", methods=["POST"])
@login_required
def deactivate_account():
    try:
        user_id = session["user_id"]
        db = get_db()
        cur = db.cursor()
        cur.execute("UPDATE users SET verified = FALSE WHERE user_id = %s", (user_id,))
        db.commit()
        session.clear()
        return {"success": True, "message": "Account deactivated successfully"}
    except Exception as e:
        return {"success": False, "message": "Error deactivating account"}, 500
    finally:
        cur.close()

@bp.route("/api/delete_account", methods=["POST"])
@login_required
def delete_account():
    try:
        user_id = session["user_id"]
        db = get_db()
        cur = db.cursor()
        cur.execute("DELETE FROM user_interests WHERE user_id = %s", (user_id,))
        cur.execute(
            "DELETE FROM connections WHERE user_id = %s OR con_user_id = %s",
            (user_id, user_id),
        )
        cur.execute(
            "DELETE FROM messages WHERE sender_id = %s OR receiver_id = %s",
            (user_id, user_id),
        )
        cur.execute("DELETE FROM education_details WHERE user_id = %s", (user_id,))
        cur.execute("DELETE FROM work_experience WHERE user_id = %s", (user_id,))
        cur.execute("DELETE FROM users WHERE user_id = %s", (user_id,))
        db.commit()
        session.clear()
        return {"success": True, "message": "Account deleted successfully"}
    except Exception as e:
        return {"success": False, "message": "Error deleting account"}, 500
    finally:
        cur.close()

@bp.route("/api/logout_all_sessions", methods=["POST"])
@login_required
def logout_all_sessions():
    try:
        session.clear()
        return {"success": True, "message": "Logged out from all sessions"}
    except Exception as e:
        return {"success": False, "message": "Error logging out sessions"}, 500

@bp.route("/api/export_data", methods=["POST"])
@login_required
def export_data():
    """Export user data"""
    import datetime
    import json
    from flask import current_app
    try:
        user_id = session["user_id"]
        db = get_db()
        cur = db.cursor()
        cur.execute(
            """
            SELECT firstname, lastname, email, username, dob,
                   university_name, college, graduation_year, current_city,
                   registration_date, role
            FROM users WHERE user_id = %s
            """,
            (user_id,),
        )
        user_data = cur.fetchone()
        cur.execute(
            """
            SELECT u.firstname, u.lastname, u.email
            FROM connections c
            JOIN users u ON (c.con_user_id = u.user_id OR c.user_id = u.user_id)
            WHERE (c.user_id = %s OR c.con_user_id = %s) 
            AND c.status = 'accepted' AND u.user_id != %s
            """,
            (user_id, user_id, user_id),
        )
        connections = cur.fetchall()
        export_data = {
            "user_info": {
                "firstname": user_data[0] if user_data else None,
                "lastname": user_data[1] if user_data else None,
                "email": user_data[2] if user_data else None,
                "username": user_data[3] if user_data else None,
                "dob": str(user_data[4]) if user_data and user_data[4] else None,
                "university": user_data[5] if user_data else None,
                "college": user_data[6] if user_data else None,
                "graduation_year": user_data[7] if user_data else None,
                "city": user_data[8] if user_data else None,
                "registration_date": (
                    str(user_data[9]) if user_data and user_data[9] else None
                ),
                "role": user_data[10] if user_data else None,
            },
            "connections": [
                {
                    "name": f"{conn[0]} {conn[1]}",
                    "email": conn[2],
                    "connected_date": None,
                }
                for conn in connections
            ],
            "export_date": datetime.datetime.utcnow().isoformat(),
        }

        json_data = json.dumps(export_data, indent=2)
        response = current_app.response_class(
            response=json_data,
            status=200,
            mimetype="application/json",
            headers={
                "Content-Disposition": "attachment; filename=algo_data_export.json"
            },
        )
        return response
    except Exception as e:
        return ({"success": False, "message": "Error exporting data"}, 500)
    finally:
        cur.close()
