from flask import Blueprint, request, jsonify, session
from psycopg2.extras import RealDictCursor
import logging
import requests

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

channels_bp = Blueprint("channels", __name__)

# Import database connection from connection module
from algo.db import get_db


# Function to broadcast messages to Go WebSocket server
def broadcast_to_go_websocket(channel_id, message_data):
    """Send message to Go WebSocket server for real-time broadcasting"""
    try:
        # Go WebSocket server internal API endpoint
        go_websocket_url = "http://localhost:8080/api/broadcast"

        broadcast_payload = {
            "type": "new_message",
            "channel_id": channel_id,
            "message": message_data,
        }

        # Send to Go server with short timeout
        response = requests.post(
            go_websocket_url,
            json=broadcast_payload,
            timeout=1.0,  # Short timeout so we don't block the response
        )

        if response.status_code == 200:
            logger.info(
                f"✅ Broadcasted message to Go WebSocket for channel {channel_id}"
            )
        else:
            logger.warning(f"⚠️ Go WebSocket returned status {response.status_code}")

    except requests.exceptions.RequestException as e:
        logger.warning(f"⚠️ Failed to reach Go WebSocket server: {e}")
        # Don't raise - we don't want to fail the message send if WebSocket is down


def get_db_connection():
    return get_db()


def require_auth(f):
    """Decorator to require authentication"""

    def decorated_function(*args, **kwargs):
        if "user_id" not in session:
            return jsonify({"error": "Authentication required"}), 401
        return f(*args, **kwargs)

    decorated_function.__name__ = f.__name__
    return decorated_function


def require_community_member(f):
    """Decorator to require community membership"""

    def decorated_function(*args, **kwargs):
        if "user_id" not in session:
            return jsonify({"error": "Authentication required"}), 401

        # Get community_id from URL kwargs, request JSON, or query params
        community_id = kwargs.get("community_id")
        if not community_id:
            community_id = request.json.get("community_id") if request.is_json else None
        if not community_id:
            community_id = request.args.get("community_id")

        if not community_id:
            return jsonify({"error": "Community ID required"}), 400

        conn = get_db_connection()
        cursor = conn.cursor()
        try:
            cursor.execute(
                """
                SELECT role, status FROM community_members 
                WHERE community_id = %s AND user_id = %s AND status = 'active'
            """,
                (community_id, session["user_id"]),
            )

            member = cursor.fetchone()
            if not member:
                return jsonify({"error": "Not a member of this community"}), 403

            request.user_community_role = member[0]
            return f(*args, **kwargs)
        finally:
            cursor.close()

    decorated_function.__name__ = f.__name__
    return decorated_function


# Community Management Routes


@channels_bp.route("/communities", methods=["GET"])
@require_auth
def get_communities():
    """Get all communities user is a member of"""
    conn = get_db_connection()
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    try:
        cursor.execute(
            """
            SELECT c.*, cm.role, cm.joined_at
            FROM communities c
            JOIN community_members cm ON c.community_id = cm.community_id
            WHERE cm.user_id = %s AND cm.status = 'active'
            ORDER BY c.name
        """,
            (session["user_id"],),
        )

        communities = cursor.fetchall()
        return jsonify({"communities": communities}), 200

    except Exception as e:
        logger.error(f"Error fetching communities: {e}")
        return jsonify({"error": "Failed to fetch communities"}), 500
    finally:
        cursor.close()


@channels_bp.route("/communities/discover", methods=["GET"])
@require_auth
def discover_communities():
    """Get all available communities that user can join or is a member of"""
    user_id = session["user_id"]
    conn = get_db_connection()
    cursor = conn.cursor(cursor_factory=RealDictCursor)
    try:
        cursor.execute(
            """
            SELECT c.community_id, c.name, c.college_code, c.location, c.description,
                   cm.status as membership_status, cm.role as user_role,
                   (SELECT COUNT(*) FROM community_members WHERE community_id = c.community_id AND status = 'active') as member_count
            FROM communities c
            LEFT JOIN community_members cm ON c.community_id = cm.community_id AND cm.user_id = %s
            ORDER BY c.name
            """,
            (user_id,),
        )
        communities = cursor.fetchall()
        return jsonify({"communities": communities}), 200
    except Exception as e:
        logger.error(f"Error discovering communities: {e}")
        return jsonify({"error": "Failed to discover communities"}), 500
    finally:
        cursor.close()


@channels_bp.route("/communities/<int:community_id>/join", methods=["POST"])
@require_auth
def join_community_direct(community_id):
    """Directly join a community"""
    user_id = session["user_id"]
    conn = get_db_connection()
    cursor = conn.cursor()
    try:
        cursor.execute("SELECT community_id, name FROM communities WHERE community_id = %s", (community_id,))
        comm = cursor.fetchone()
        if not comm:
            return jsonify({"error": "Community not found"}), 404

        cursor.execute("""
            INSERT INTO community_members (community_id, user_id, role, status)
            VALUES (%s, %s, 'member', 'active')
            ON CONFLICT (community_id, user_id) DO UPDATE SET status = 'active'
        """, (community_id, user_id))

        cursor.execute("UPDATE users SET community_id = COALESCE(community_id, %s) WHERE user_id = %s", (community_id, user_id))
        conn.commit()
        return jsonify({"success": True, "message": f"Successfully joined {comm[1]}!"}), 200
    except Exception as e:
        conn.rollback()
        logger.error(f"Error joining community: {e}")
        return jsonify({"error": "Failed to join community"}), 500
    finally:
        cursor.close()


@channels_bp.route("/communities/<int:community_id>/leave", methods=["POST"])
@require_auth
def leave_community(community_id):
    """Leave a community"""
    user_id = session["user_id"]
    conn = get_db_connection()
    cursor = conn.cursor()
    try:
        cursor.execute("DELETE FROM community_members WHERE community_id = %s AND user_id = %s", (community_id, user_id))
        conn.commit()
        return jsonify({"success": True, "message": "Left community successfully"}), 200
    except Exception as e:
        conn.rollback()
        logger.error(f"Error leaving community: {e}")
        return jsonify({"error": "Failed to leave community"}), 500
    finally:
        cursor.close()


@channels_bp.route("/communities/<int:community_id>/join-request", methods=["POST"])
@require_auth
def request_join_community(community_id):
    """Request to join a community"""
    data = request.get_json()
    message = data.get("message", "")

    conn = get_db_connection()
    cursor = conn.cursor()

    try:
        # Check if community exists
        cursor.execute(
            "SELECT name FROM communities WHERE community_id = %s", (community_id,)
        )
        community = cursor.fetchone()
        if not community:
            return jsonify({"error": "Community not found"}), 404

        # Check if already a member
        cursor.execute(
            """
            SELECT status FROM community_members 
            WHERE community_id = %s AND user_id = %s
        """,
            (community_id, session["user_id"]),
        )

        existing_member = cursor.fetchone()
        if existing_member:
            if existing_member[0] == "active":
                return jsonify({"error": "Already a member of this community"}), 400
            elif existing_member[0] == "banned":
                return jsonify({"error": "You are banned from this community"}), 403

        # Check if already requested
        cursor.execute(
            """
            SELECT status FROM community_join_requests 
            WHERE community_id = %s AND user_id = %s
        """,
            (community_id, session["user_id"]),
        )

        existing_request = cursor.fetchone()
        if existing_request and existing_request[0] == "pending":
            return jsonify({"error": "Join request already pending"}), 400

        # Create join request
        cursor.execute(
            """
            INSERT INTO community_join_requests (community_id, user_id, message)
            VALUES (%s, %s, %s)
            ON CONFLICT (community_id, user_id) 
            DO UPDATE SET message = %s, status = 'pending', requested_at = CURRENT_TIMESTAMP
            RETURNING request_id
        """,
            (community_id, session["user_id"], message, message),
        )

        request_id = cursor.fetchone()[0]
        conn.commit()

        return (
            jsonify(
                {
                    "message": "Join request submitted successfully",
                    "request_id": request_id,
                }
            ),
            201,
        )

    except Exception as e:
        conn.rollback()
        logger.error(f"Error creating join request: {e}")
        return jsonify({"error": "Failed to submit join request"}), 500
    finally:
        cursor.close()


@channels_bp.route("/communities/<int:community_id>/invitations", methods=["POST"])
@require_auth
@require_community_member
def invite_user_to_community(community_id):
    """Invite a user to join the community (admin/moderator only)"""
    if request.user_community_role not in ["admin", "moderator"]:
        return jsonify({"error": "Insufficient permissions"}), 403

    data = request.get_json()
    invited_user_id = data.get("user_id")
    message = data.get("message", "")

    if not invited_user_id:
        return jsonify({"error": "User ID required"}), 400

    conn = get_db_connection()
    cursor = conn.cursor()

    try:
        # Check if user exists
        cursor.execute(
            "SELECT username FROM users WHERE user_id = %s", (invited_user_id,)
        )
        user = cursor.fetchone()
        if not user:
            return jsonify({"error": "User not found"}), 404

        # Check if already a member
        cursor.execute(
            """
            SELECT status FROM community_members 
            WHERE community_id = %s AND user_id = %s
        """,
            (community_id, invited_user_id),
        )

        existing_member = cursor.fetchone()
        if existing_member and existing_member[0] == "active":
            return jsonify({"error": "User is already a member"}), 400

        # Create invitation
        cursor.execute(
            """
            INSERT INTO community_invitations (community_id, invited_user_id, invited_by, message)
            VALUES (%s, %s, %s, %s)
            ON CONFLICT (community_id, invited_user_id) 
            DO UPDATE SET invited_by = %s, message = %s, status = 'pending', 
                         created_at = CURRENT_TIMESTAMP, expires_at = CURRENT_TIMESTAMP + INTERVAL '7 days'
            RETURNING invitation_id
        """,
            (
                community_id,
                invited_user_id,
                session["user_id"],
                message,
                session["user_id"],
                message,
            ),
        )

        invitation_id = cursor.fetchone()[0]
        conn.commit()

        return (
            jsonify(
                {
                    "message": "Invitation sent successfully",
                    "invitation_id": invitation_id,
                }
            ),
            201,
        )

    except Exception as e:
        conn.rollback()
        logger.error(f"Error creating invitation: {e}")
        return jsonify({"error": "Failed to send invitation"}), 500
    finally:
        cursor.close()


@channels_bp.route("/communities/<int:community_id>/members", methods=["GET"])
@require_auth
@require_community_member
def get_community_members(community_id):
    """Get members of a community"""
    conn = get_db_connection()
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    try:
        cursor.execute(
            """
            SELECT 
                u.user_id,
                u.firstname,
                u.lastname,
                u.username,
                u.pfp_path,
                u.department,
                u.role as user_role,
                cm.role as community_role,
                cm.joined_at,
                CASE WHEN u.last_login > NOW() - INTERVAL '5 minutes' THEN true ELSE false END as is_online
            FROM community_members cm
            JOIN users u ON cm.user_id = u.user_id
            WHERE cm.community_id = %s AND cm.status = 'active'
            ORDER BY 
                CASE cm.role 
                    WHEN 'admin' THEN 1 
                    WHEN 'moderator' THEN 2 
                    ELSE 3 
                END,
                u.firstname, u.lastname
        """,
            (community_id,),
        )

        members = cursor.fetchall()

        # Format the response
        formatted_members = []
        for member in members:
            formatted_members.append(
                {
                    "user_id": member["user_id"],
                    "firstname": member["firstname"],
                    "lastname": member["lastname"],
                    "username": member["username"],
                    "pfp_path": member["pfp_path"],
                    "department": member["department"],
                    "user_role": member["user_role"],
                    "role": member["community_role"],
                    "joined_at": (
                        member["joined_at"].isoformat() if member["joined_at"] else None
                    ),
                    "is_online": member["is_online"],
                }
            )

        return jsonify({"members": formatted_members}), 200

    except Exception as e:
        logger.error(f"Error fetching community members: {e}")
        return jsonify({"error": "Failed to fetch members"}), 500
    finally:
        cursor.close()


# Channel Management Routes


@channels_bp.route("/communities/<int:community_id>/channels", methods=["GET"])
@require_auth
@require_community_member
def get_channels(community_id):
    """Get all channels in a community"""
    conn = get_db_connection()
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    try:
        cursor.execute(
            """
            SELECT c.*, u.username as created_by_username
            FROM channels c
            LEFT JOIN users u ON c.created_by = u.user_id
            WHERE c.community_id = %s AND c.is_active = true
            ORDER BY c.position_order, c.created_at
        """,
            (community_id,),
        )

        channels = cursor.fetchall()
        return jsonify({"channels": channels}), 200

    except Exception as e:
        logger.error(f"Error fetching channels: {e}")
        return jsonify({"error": "Failed to fetch channels"}), 500
    finally:
        cursor.close()


@channels_bp.route("/communities/<int:community_id>/channels", methods=["POST"])
@require_auth
@require_community_member
def create_channel(community_id):
    """Create a new channel"""
    if request.user_community_role not in ["admin", "moderator", "member"] and session.get("role") != "admin":
        return jsonify({"error": "Insufficient permissions"}), 403

    data = request.get_json()
    name = data.get("name", "").strip()
    description = data.get("description", "")
    channel_type = data.get("channel_type", "text")
    is_private = data.get("is_private", False)

    if not name:
        return jsonify({"error": "Channel name required"}), 400

    conn = get_db_connection()
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    try:
        # Check if channel name already exists in community
        cursor.execute(
            """
            SELECT channel_id FROM channels 
            WHERE community_id = %s AND LOWER(name) = LOWER(%s) AND is_active = true
        """,
            (community_id, name),
        )

        if cursor.fetchone():
            return jsonify({"error": "Channel name already exists"}), 400

        # Create channel
        cursor.execute(
            """
            INSERT INTO channels (community_id, name, description, channel_type, is_private, created_by)
            VALUES (%s, %s, %s, %s, %s, %s)
            RETURNING *
        """,
            (
                community_id,
                name,
                description,
                channel_type,
                is_private,
                session["user_id"],
            ),
        )

        channel = cursor.fetchone()
        conn.commit()

        return (
            jsonify(
                {"message": "Channel created successfully", "channel": dict(channel)}
            ),
            201,
        )

    except Exception as e:
        conn.rollback()
        logger.error(f"Error creating channel: {e}")
        return jsonify({"error": "Failed to create channel"}), 500
    finally:
        cursor.close()


@channels_bp.route("/channels/<int:channel_id>/messages", methods=["GET"])
@require_auth
def get_channel_messages(channel_id):
    """Get messages from a channel"""
    page = request.args.get("page", 1, type=int)
    limit = min(request.args.get("limit", 50, type=int), 100)
    offset = (page - 1) * limit

    conn = get_db_connection()
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    try:
        # Check if user has access to channel
        cursor.execute(
            """
            SELECT c.community_id, c.is_private
            FROM channels c
            WHERE c.channel_id = %s AND c.is_active = true
        """,
            (channel_id,),
        )

        channel = cursor.fetchone()
        if not channel:
            return jsonify({"error": "Channel not found"}), 404

        # Check community membership
        cursor.execute(
            """
            SELECT role FROM community_members 
            WHERE community_id = %s AND user_id = %s AND status = 'active'
        """,
            (channel["community_id"], session["user_id"]),
        )

        if not cursor.fetchone():
            return jsonify({"error": "Access denied"}), 403

        # If private channel, check channel membership
        if channel["is_private"]:
            cursor.execute(
                """
                SELECT 1 FROM channel_members 
                WHERE channel_id = %s AND user_id = %s
            """,
                (channel_id, session["user_id"]),
            )

            if not cursor.fetchone():
                return jsonify({"error": "Access denied to private channel"}), 403

        # Get messages
        cursor.execute(
            """
            SELECT 
                cm.*,
                u.username,
                u.pfp_path,
                reply_msg.content as reply_content,
                reply_user.username as reply_username
            FROM channel_messages cm
            JOIN users u ON cm.user_id = u.user_id
            LEFT JOIN channel_messages reply_msg ON cm.reply_to_message_id = reply_msg.message_id
            LEFT JOIN users reply_user ON reply_msg.user_id = reply_user.user_id
            WHERE cm.channel_id = %s AND cm.is_deleted = false
            ORDER BY cm.created_at DESC
            LIMIT %s OFFSET %s
        """,
            (channel_id, limit, offset),
        )

        messages = cursor.fetchall()

        # Get reaction counts for messages
        if messages:
            message_ids = [msg["message_id"] for msg in messages]
            cursor.execute(
                """
                SELECT message_id, emoji, COUNT(*) as count
                FROM message_reactions
                WHERE message_id = ANY(%s)
                GROUP BY message_id, emoji
            """,
                (message_ids,),
            )

            reactions = cursor.fetchall()
            reaction_dict = {}
            for reaction in reactions:
                msg_id = reaction["message_id"]
                if msg_id not in reaction_dict:
                    reaction_dict[msg_id] = []
                reaction_dict[msg_id].append(
                    {"emoji": reaction["emoji"], "count": reaction["count"]}
                )

            # Add reactions to messages
            for message in messages:
                message["reactions"] = reaction_dict.get(message["message_id"], [])

        return (
            jsonify(
                {
                    "messages": [dict(msg) for msg in reversed(messages)],
                    "page": page,
                    "limit": limit,
                }
            ),
            200,
        )

    except Exception as e:
        logger.error(f"Error fetching messages: {e}")
        return jsonify({"error": "Failed to fetch messages"}), 500
    finally:
        cursor.close()


@channels_bp.route("/channels/<int:channel_id>/messages", methods=["POST"])
@require_auth
def send_message(channel_id):
    """Send a message to a channel"""
    data = request.get_json()
    content = data.get("content", "").strip()
    message_type = data.get("message_type", "text")
    reply_to_message_id = data.get("reply_to_message_id")

    if not content:
        return jsonify({"error": "Message content required"}), 400

    conn = get_db_connection()
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    try:
        # Check if user has access to channel
        cursor.execute(
            """
            SELECT c.community_id, c.is_private
            FROM channels c
            WHERE c.channel_id = %s AND c.is_active = true
        """,
            (channel_id,),
        )

        channel = cursor.fetchone()
        if not channel:
            return jsonify({"error": "Channel not found"}), 404

        # Check community membership
        cursor.execute(
            """
            SELECT role FROM community_members 
            WHERE community_id = %s AND user_id = %s AND status = 'active'
        """,
            (channel["community_id"], session["user_id"]),
        )

        member = cursor.fetchone()
        if not member:
            return jsonify({"error": "Access denied"}), 403

        # If private channel, check channel membership and permissions
        if channel["is_private"]:
            cursor.execute(
                """
                SELECT can_send_messages FROM channel_members 
                WHERE channel_id = %s AND user_id = %s
            """,
                (channel_id, session["user_id"]),
            )

            channel_member = cursor.fetchone()
            if not channel_member:
                return jsonify({"error": "Access denied to private channel"}), 403
            if not channel_member["can_send_messages"]:
                return (
                    jsonify({"error": "You cannot send messages in this channel"}),
                    403,
                )

        # Send message
        cursor.execute(
            """
            INSERT INTO channel_messages (channel_id, user_id, content, message_type, reply_to_message_id)
            VALUES (%s, %s, %s, %s, %s)
            RETURNING *
        """,
            (
                channel_id,
                session["user_id"],
                content,
                message_type,
                reply_to_message_id,
            ),
        )

        message = cursor.fetchone()
        conn.commit()

        # Get user info for response
        cursor.execute(
            """
            SELECT username, pfp_path FROM users WHERE user_id = %s
        """,
            (session["user_id"],),
        )
        user_info = cursor.fetchone()

        response_message = dict(message)
        response_message.update(
            {
                "username": user_info["username"],
                "pfp_path": user_info["pfp_path"],
                "reactions": [],
            }
        )

        # Convert datetime objects to strings for JSON serialization
        broadcast_message = dict(response_message)
        for key, value in broadcast_message.items():
            if hasattr(value, "isoformat"):  # datetime object
                broadcast_message[key] = value.isoformat()

        # Trigger Go WebSocket broadcast for real-time delivery
        try:
            broadcast_to_go_websocket(channel_id, broadcast_message)
        except Exception as e:
            logger.warning(f"Failed to broadcast to WebSocket: {e}")
            # Don't fail the request if WebSocket broadcast fails

        return (
            jsonify({"message": "Message sent successfully", "data": response_message}),
            201,
        )

    except Exception as e:
        conn.rollback()
        logger.error(f"Error sending message: {e}")
        return jsonify({"error": "Failed to send message"}), 500
    finally:
        cursor.close()


@channels_bp.route("/channels/<int:channel_id>/members", methods=["GET"])
@require_auth
def get_channel_members(channel_id):
    """Get members of a specific channel using channel_members or community_members"""
    try:
        conn = get_db_connection()
        cursor = conn.cursor(cursor_factory=RealDictCursor)

        # Check channel info
        cursor.execute("SELECT community_id, is_private FROM channels WHERE channel_id = %s", (channel_id,))
        channel = cursor.fetchone()
        if not channel:
            return jsonify({"error": "Channel not found"}), 404

        community_id = channel["community_id"]
        is_private = channel.get("is_private", False)

        if is_private:
            query = """
                SELECT 
                    u.user_id, 
                    u.username, 
                    u.firstname, 
                    u.lastname, 
                    u.pfp_path,
                    cm.role,
                    cm.joined_at,
                    CASE WHEN u.last_login > NOW() - INTERVAL '15 minutes' THEN true ELSE false END as is_online
                FROM channel_members cm
                JOIN users u ON cm.user_id = u.user_id
                WHERE cm.channel_id = %s
                ORDER BY is_online DESC, u.username ASC
            """
            cursor.execute(query, (channel_id,))
        else:
            query = """
                SELECT 
                    u.user_id, 
                    u.username, 
                    u.firstname, 
                    u.lastname, 
                    u.pfp_path,
                    cm.role,
                    cm.joined_at,
                    CASE WHEN u.last_login > NOW() - INTERVAL '15 minutes' THEN true ELSE false END as is_online
                FROM community_members cm
                JOIN users u ON cm.user_id = u.user_id
                WHERE cm.community_id = %s AND cm.status = 'active'
                ORDER BY is_online DESC, u.username ASC
            """
            cursor.execute(query, (community_id,))

        members = cursor.fetchall()

        # Convert to list of dictionaries
        members_list = []
        for member in members:
            members_list.append(
                {
                    "user_id": member["user_id"],
                    "username": member["username"],
                    "firstname": member["firstname"] or "",
                    "lastname": member["lastname"] or "",
                    "pfp_path": member["pfp_path"] or "",
                    "role": member["role"] or "member",
                    "joined_at": (
                        member["joined_at"].isoformat() if member["joined_at"] else None
                    ),
                    "is_online": bool(member["is_online"]),
                }
            )

        logger.info(
            f"✅ Retrieved {len(members_list)} members for channel {channel_id}"
        )
        return jsonify(
            {"success": True, "members": members_list, "count": len(members_list)}
        )

    except Exception as e:
        logger.error(f"❌ Error getting channel members: {e}")
        return jsonify({"error": "Failed to get channel members"}), 500
    finally:
        cursor.close()


@channels_bp.route("/messages/<int:message_id>/reactions", methods=["POST"])
@require_auth
def add_reaction(message_id):
    """Add a reaction to a message"""
    data = request.get_json()
    emoji = data.get("emoji", "").strip()

    if not emoji:
        return jsonify({"error": "Emoji required"}), 400

    conn = get_db_connection()
    cursor = conn.cursor()

    try:
        # Check if message exists and user has access
        cursor.execute(
            """
            SELECT cm.channel_id, c.community_id, c.is_private
            FROM channel_messages cm
            JOIN channels c ON cm.channel_id = c.channel_id
            WHERE cm.message_id = %s AND cm.is_deleted = false
        """,
            (message_id,),
        )

        message_info = cursor.fetchone()
        if not message_info:
            return jsonify({"error": "Message not found"}), 404

        channel_id, community_id, is_private = message_info

        # Check community membership
        cursor.execute(
            """
            SELECT role FROM community_members 
            WHERE community_id = %s AND user_id = %s AND status = 'active'
        """,
            (community_id, session["user_id"]),
        )

        if not cursor.fetchone():
            return jsonify({"error": "Access denied"}), 403

        # If private channel, check channel membership
        if is_private:
            cursor.execute(
                """
                SELECT 1 FROM channel_members 
                WHERE channel_id = %s AND user_id = %s
            """,
                (channel_id, session["user_id"]),
            )

            if not cursor.fetchone():
                return jsonify({"error": "Access denied to private channel"}), 403

        # Add or remove reaction
        cursor.execute(
            """
            INSERT INTO message_reactions (message_id, user_id, emoji)
            VALUES (%s, %s, %s)
            ON CONFLICT (message_id, user_id, emoji) DO NOTHING
            RETURNING reaction_id
        """,
            (message_id, session["user_id"], emoji),
        )

        result = cursor.fetchone()
        if result:
            conn.commit()
            return jsonify({"message": "Reaction added"}), 201
        else:
            # Reaction already exists, remove it
            cursor.execute(
                """
                DELETE FROM message_reactions 
                WHERE message_id = %s AND user_id = %s AND emoji = %s
            """,
                (message_id, session["user_id"], emoji),
            )
            conn.commit()
            return jsonify({"message": "Reaction removed"}), 200

    except Exception as e:
        conn.rollback()
        logger.error(f"Error managing reaction: {e}")
        return jsonify({"error": "Failed to manage reaction"}), 500
    finally:
        cursor.close()
