from flask import Blueprint, render_template, session, redirect, url_for, flash, request, jsonify, current_app
from algo.db import get_db
from algo.auth.decorators import login_required
from algo.auth import user_roles
from algo import utils
import datetime

bp = Blueprint('chat', __name__)

@bp.route("/chat")
@bp.route("/chat_list")
@login_required
@user_roles.verified_user_required
def chat_list():
    user_id = session["user_id"]
    db = get_db()
    cur = db.cursor()
    try:
        cur.execute(
            """
            SELECT DISTINCT other_user_id FROM (
                -- Users from message history
                SELECT CASE 
                    WHEN m.sender_id = %s THEN m.receiver_id 
                    ELSE m.sender_id 
                END as other_user_id
                FROM messages m
                WHERE m.sender_id = %s OR m.receiver_id = %s
                
                UNION
                
                -- Users from accepted connections
                SELECT CASE 
                    WHEN c.user_id = %s THEN c.con_user_id 
                    ELSE c.user_id 
                END as other_user_id
                FROM connections c
                WHERE (c.user_id = %s OR c.con_user_id = %s) AND c.status = 'accepted'
            ) AS combined_users
            """,
            (user_id, user_id, user_id, user_id, user_id, user_id),
        )
        user_ids = cur.fetchall()
        chat_history = []
        for (other_user_id,) in user_ids:
            cur.execute(
                "SELECT username, pfp_path, firstname, lastname FROM users WHERE user_id = %s",
                (other_user_id,),
            )
            user_data = cur.fetchone()
            if user_data:
                username, pfp_path, firstname, lastname = user_data
                if not pfp_path:
                    full_name = f"{firstname or ''} {lastname or ''}".strip()
                    pfp_path = utils.generate_default_avatar(full_name or username)
                cur.execute(
                    """
                    SELECT content, created_at 
                    FROM messages 
                    WHERE (sender_id = %s AND receiver_id = %s) 
                       OR (sender_id = %s AND receiver_id = %s)
                    ORDER BY created_at DESC 
                    LIMIT 1
                    """,
                    (user_id, other_user_id, other_user_id, user_id),
                )
                message_data = cur.fetchone()
                if message_data:
                    last_message = message_data[0]
                    last_message_time = message_data[1]
                else:
                    cur.execute(
                        """
                        SELECT 1 FROM connections 
                        WHERE ((user_id = %s AND con_user_id = %s) OR (user_id = %s AND con_user_id = %s)) 
                        AND status = 'accepted'
                        """,
                        (user_id, other_user_id, other_user_id, user_id),
                    )
                    if cur.fetchone():
                        last_message = "You're now connected! Start a conversation."
                        last_message_time = None
                    else:
                        last_message = None
                        last_message_time = None
                chat_history.append(
                    (other_user_id, username, pfp_path, last_message, last_message_time)
                )
        chat_history.sort(
            key=lambda x: x[4] if x[4] else datetime.datetime.min, reverse=True
        )
    except Exception as e:
        chat_history = []
    finally:
        cur.close()
    return render_template("chat_list.html", conversations=chat_history)

@bp.route("/api/online_status")
@login_required
def get_online_status():
    """Return list of currently active user IDs based on recent activity and session."""
    db = get_db()
    cur = db.cursor()
    try:
        cur.execute(
            "SELECT user_id FROM users WHERE last_login > NOW() - INTERVAL '15 minutes'"
        )
        rows = cur.fetchall()
        online_users = [r[0] for r in rows]
        if session.get("user_id") and session["user_id"] not in online_users:
            online_users.append(session["user_id"])
        return jsonify({"online_users": online_users, "success": True})
    except Exception as e:
        logger.error(f"Error fetching online status: {e}")
        return jsonify({"online_users": [session.get("user_id")] if session.get("user_id") else []})
    finally:
        cur.close()


@bp.route("/api/upload_file", methods=["POST"])
@bp.route("/api/upload_image", methods=["POST"])
@login_required
def upload_chat_file():
    """Upload chat attachment (image/document) and return accessible static URL."""
    try:
        file = request.files.get("file") or request.files.get("image")
        if not file or file.filename == "":
            return jsonify({"error": "No file provided"}), 400

        import os
        import uuid
        from werkzeug.utils import secure_filename

        orig_filename = secure_filename(file.filename) or "attachment"
        ext = orig_filename.rsplit(".", 1)[1].lower() if "." in orig_filename else "bin"
        unique_filename = f"{uuid.uuid4().hex}.{ext}"

        upload_dir = os.path.join(current_app.static_folder, "uploads", "chat")
        os.makedirs(upload_dir, exist_ok=True)
        file_path = os.path.join(upload_dir, unique_filename)
        file.save(file_path)

        file_url = f"/static/uploads/chat/{unique_filename}"
        return jsonify(
            {
                "success": True,
                "file_url": file_url,
                "image_url": file_url,
                "filename": orig_filename,
            }
        ), 200
    except Exception as e:
        logger.error(f"Chat file upload failed: {e}")
        return jsonify({"error": "Upload failed"}), 500


@bp.route("/api/search_users")
@login_required
def search_users():
    """Search registered users to initiate a new direct chat."""
    q = request.args.get("q", "").strip()
    user_id = session.get("user_id")
    db = get_db()
    cur = db.cursor()
    try:
        if q:
            query = """
                SELECT user_id, username, firstname, lastname, pfp_path, role, department, university_name
                FROM users
                WHERE user_id != %s AND (
                    username ILIKE %s OR firstname ILIKE %s OR lastname ILIKE %s OR email ILIKE %s
                )
                ORDER BY firstname ASC
                LIMIT 20
            """
            search_param = f"%{q}%"
            cur.execute(query, (user_id, search_param, search_param, search_param, search_param))
        else:
            query = """
                SELECT DISTINCT u.user_id, u.username, u.firstname, u.lastname, u.pfp_path, u.role, u.department, u.university_name
                FROM users u
                LEFT JOIN connections c ON (
                    (c.user_id = %s AND c.con_user_id = u.user_id) OR
                    (c.con_user_id = %s AND c.user_id = u.user_id)
                ) AND c.status = 'accepted'
                WHERE u.user_id != %s
                ORDER BY u.firstname ASC
                LIMIT 20
            """
            cur.execute(query, (user_id, user_id, user_id))

        users = []
        for r in cur.fetchall():
            users.append(
                {
                    "user_id": r[0],
                    "username": r[1],
                    "name": f"{r[2]} {r[3]}".strip() or r[1],
                    "pfp_path": r[4] or "/static/images/default_pfp.png",
                    "role": r[5].capitalize() if r[5] else "Member",
                    "headline": f"{r[6] or ''} at {r[7] or ''}".strip().strip("at").strip() or (r[5].capitalize() if r[5] else "AlumniGo Member"),
                }
            )
        return jsonify({"success": True, "users": users})
    except Exception as e:
        logger.error(f"Error searching users: {e}")
        return jsonify({"success": False, "users": []})
    finally:
        cur.close()

@bp.route("/chat/<username>")
@login_required
def chat(username):
    user_id = session["user_id"]
    db = get_db()
    cur = db.cursor()
    cur.execute(
        "SELECT user_id, username, pfp_path, firstname, lastname FROM users WHERE username=%s",
        (username,),
    )
    other_user_row = cur.fetchone()
    if not other_user_row:
        flash("User not found")
        return redirect(url_for("dashboard.user_dashboard"))
    other_user_id, other_user_name, other_user_pfp, other_firstname, other_lastname = (
        other_user_row
    )
    if not other_user_pfp:
        other_full_name = f"{other_firstname or ''} {other_lastname or ''}".strip()
        other_user_pfp = utils.generate_default_avatar(
            other_full_name or other_user_name
        )
    cur.execute(
        """
        SELECT m.sender_id, m.receiver_id, m.content, m.created_at, u.username
        FROM messages m
        JOIN users u ON m.sender_id = u.user_id
        WHERE (m.sender_id=%s AND m.receiver_id=%s) OR (m.sender_id=%s AND m.receiver_id=%s)
        ORDER BY m.created_at
        """,
        (user_id, other_user_id, other_user_id, user_id),
    )
    conversation = cur.fetchall()
    try:
        cur.execute(
            """
            SELECT DISTINCT 
                CASE 
                    WHEN m.sender_id = %s THEN m.receiver_id 
                    ELSE m.sender_id 
                END as other_user_id
            FROM messages m
            WHERE m.sender_id = %s OR m.receiver_id = %s
            """,
            (user_id, user_id, user_id),
        )
        user_ids = cur.fetchall()
        chat_history = []
        for (chat_user_id,) in user_ids:
            cur.execute(
                "SELECT username, pfp_path, firstname, lastname FROM users WHERE user_id = %s",
                (chat_user_id,),
            )
            user_data = cur.fetchone()
            if user_data:
                chat_username, pfp_path, chat_firstname, chat_lastname = user_data
                if not pfp_path:
                    chat_full_name = (
                        f"{chat_firstname or ''} {chat_lastname or ''}".strip()
                    )
                    pfp_path = utils.generate_default_avatar(
                        chat_full_name or chat_username
                    )
                cur.execute(
                    """
                    SELECT content, created_at 
                    FROM messages 
                    WHERE (sender_id = %s AND receiver_id = %s) 
                       OR (sender_id = %s AND receiver_id = %s)
                    ORDER BY created_at DESC 
                    LIMIT 1
                    """,
                    (user_id, chat_user_id, chat_user_id, user_id),
                )
                message_data = cur.fetchone()
                last_message = message_data[0] if message_data else None
                last_message_time = message_data[1] if message_data else None
                chat_history.append(
                    (
                        chat_user_id,
                        chat_username,
                        pfp_path,
                        last_message,
                        last_message_time,
                    )
                )
        chat_history.sort(
            key=lambda x: x[4] if x[4] else datetime.datetime.min, reverse=True
        )
    except Exception as e:
        chat_history = []
    finally:
        cur.close()
    return render_template(
        "chat.html",
        conversation=conversation,
        other_user_id=other_user_id or None,
        other_user_name=other_user_name or "",
        other_user_pfp=other_user_pfp or "",
        chat_history=chat_history or [],
        current_user_id=user_id or None,
    )
