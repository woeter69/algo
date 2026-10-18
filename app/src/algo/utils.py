"""
Utility functions for the AlumniGo application.

Consolidated module containing:
- URL safety checks
- Email validation (student email checking)
- File operations (loading cities data)
- Image uploading to external services
- Timezone conversions and datetime formatting
- User-related utilities (room ID generation, avatar creation)
- Password reset email utilities
"""

import re
import os
import base64
import json
import requests
import pytz
import urllib.parse
import smtplib
import logging
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from urllib.parse import urlparse
from flask import request
from dotenv import load_dotenv

load_dotenv()

# Set of test emails loaded from environment variable for development/testing
TEST_EMAILS = set(os.getenv("TEST_EMAILS", "").split(","))

# Indian Standard Time timezone
IST = pytz.timezone("Asia/Kolkata")

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


# --- URL Safety ---

def is_safe_url(target):
    """Check if the target URL is safe for redirecting."""
    if not target:
        return False
    parsed = urlparse(target)
    if not parsed.netloc and not parsed.scheme:
        return True
    try:
        ref_url = urlparse(request.host_url)
        return parsed.netloc == ref_url.netloc and parsed.scheme in ("http", "https")
    except RuntimeError:
        return not parsed.netloc and not parsed.scheme


# --- Email Validation ---

def is_student_email(email):
    """Check if an email address belongs to a student (educational institution)."""
    if email in TEST_EMAILS:
        return True
    pattern = r".+@.+\.(edu|ac\.in|ac\.uk|edu\.au|ac\.ca|ac\.nz|ac\.jp|ac\.kr|ac\.us|college\.edu|uni\.edu)$"
    return re.match(pattern, email) is not None


# --- File / Data Operations ---

def load_cities():
    """Load cities data from the JSON file."""
    BASE_DIR = os.path.dirname(os.path.abspath(__file__))
    DATA_PATH = os.path.join(BASE_DIR, "..", "..", "static", "data", "cities-name-list.json")
    with open(DATA_PATH, encoding="utf-8") as f:
        return json.load(f)


def upload_to_imgbb(file, api_key):
    """Upload an image file to ImgBB image hosting service."""
    payload = {"key": api_key, "image": base64.b64encode(file.read()).decode("utf-8")}
    response = requests.post("https://api.imgbb.com/1/upload", data=payload)
    if response.status_code == 200:
        return response.json()["data"]["url"]
    return None


# --- Timezone Helpers ---

def to_ist(utc_datetime):
    """Convert a UTC datetime to Indian Standard Time (IST)."""
    if utc_datetime is None:
        return None
    if utc_datetime.tzinfo is None:
        utc_datetime = pytz.utc.localize(utc_datetime)
    return utc_datetime.astimezone(IST)


def format_ist_time(utc_datetime, format_str="%I:%M %p"):
    """Format a UTC datetime as an IST time string."""
    if utc_datetime is None:
        return "now"
    ist_time = to_ist(utc_datetime)
    return ist_time.strftime(format_str)


def format_utc_timestamp(utc_datetime):
    """Convert a UTC datetime to a Unix timestamp in milliseconds."""
    if utc_datetime is None:
        return None
    if utc_datetime.tzinfo is None:
        utc_datetime = pytz.utc.localize(utc_datetime)
    return int(utc_datetime.timestamp() * 1000)


# --- User Utilities ---

def get_room_id(user1, user2):
    """Generate a consistent room ID for two users."""
    return f"room_{min(user1, user2)}_{max(user1, user2)}"


def generate_default_avatar(name):
    """Generate a default avatar URL for users without profile pictures."""
    if not name:
        return "https://i.ibb.co/QDy827D/default-avatar.png"

    clean_name = name.strip()
    words = clean_name.split()
    initials = "".join([word[0].upper() for word in words if word])[:2]

    colors = [
        "6D28D9", "3B82F6", "10B981", "F59E0B", "EF4444",
        "8B5CF6", "06B6D4", "F97316", "EC4899", "84CC16",
    ]
    hash_value = sum(ord(char) for char in clean_name)
    background_color = colors[hash_value % len(colors)]

    return f"https://ui-avatars.com/api/?name={urllib.parse.quote(initials)}&background={background_color}&color=fff&size=80&font-size=0.6&bold=true"


# --- Password Reset Emails ---

def create_password_reset_email_content(reset_token, user_name=None, to_email=None):
    """Create password reset email content."""
    base_url = os.getenv("BASE_URL", "http://localhost:5000")
    reset_url = f"{base_url}/reset-password/{reset_token}"
    subject = "Reset Your ALGO Password"
    greeting = f"Hi {user_name}," if user_name else "Hi there,"
    text_body = f"""
{greeting}

We received a request to reset your password for your ALGO account.

To reset your password, visit this link:
{reset_url}

Or use this reset token: {reset_token}

This link will expire in 1 hour for security reasons.
If you didn't request this password reset, please ignore this email.

Best regards,
The ALGO Team
    """
    return subject, text_body


def create_password_changed_email_content(user_name=None, to_email=None):
    """Create password changed notification email content."""
    subject = "Your ALGO Password Has Been Changed"
    greeting = f"Hi {user_name}," if user_name else "Hi there,"
    text_body = f"""
{greeting}

Your ALGO account password has been successfully changed.
If you didn't make this change, please contact our support team immediately.

Best regards,
The ALGO Team
    """
    return subject, text_body


def send_password_reset_email(to_email, reset_token, user_name=None):
    """Send password reset email with reset link and styled HTML layout."""
    try:
        smtp_server = os.getenv("SMTP_SERVER", "smtp.gmail.com")
        smtp_port = int(os.getenv("SMTP_PORT", "587"))
        smtp_username = os.getenv("SMTP_USERNAME") or os.getenv("EMAIL_USER")
        smtp_password = os.getenv("SMTP_PASSWORD") or os.getenv("EMAIL_PASS")
        from_email = os.getenv("FROM_EMAIL", smtp_username)
        base_url = os.getenv("BASE_URL", "http://localhost:5000")

        if not smtp_username or not smtp_password:
            logger.error("SMTP credentials not configured")
            return False

        from algo.email_templates import build_password_reset_email
        reset_url = f"{base_url}/reset-password/{reset_token}"
        email_subject, plain_body, html_body = build_password_reset_email(
            user_name=user_name,
            reset_url=reset_url,
            reset_token=reset_token,
        )

        msg = MIMEMultipart("alternative")
        msg["From"] = from_email
        msg["To"] = to_email
        msg["Subject"] = email_subject
        msg.attach(MIMEText(plain_body, "plain", "utf-8"))
        msg.attach(MIMEText(html_body, "html", "utf-8"))

        with smtplib.SMTP(smtp_server, smtp_port) as server:
            server.starttls()
            server.login(smtp_username, smtp_password)
            server.send_message(msg)

        logger.info(f"Password reset email sent to {to_email}")
        return True
    except Exception as e:
        logger.error(f"Failed to send password reset email to {to_email}: {str(e)}")
        return False


def send_password_changed_notification(to_email, user_name=None):
    """Send notification email when password is successfully changed."""
    try:
        smtp_server = os.getenv("SMTP_SERVER", "smtp.gmail.com")
        smtp_port = int(os.getenv("SMTP_PORT", "587"))
        smtp_username = os.getenv("SMTP_USERNAME") or os.getenv("EMAIL_USER")
        smtp_password = os.getenv("SMTP_PASSWORD") or os.getenv("EMAIL_PASS")
        from_email = os.getenv("FROM_EMAIL", smtp_username)
        base_url = os.getenv("BASE_URL", "http://localhost:5000")

        if not smtp_username or not smtp_password:
            logger.error("SMTP credentials not configured")
            return False

        from algo.email_templates import build_password_changed_email
        contact_url = f"{base_url}/contact"
        email_subject, plain_body, html_body = build_password_changed_email(
            user_name=user_name,
            contact_url=contact_url,
        )

        msg = MIMEMultipart("alternative")
        msg["From"] = from_email
        msg["To"] = to_email
        msg["Subject"] = email_subject
        msg.attach(MIMEText(plain_body, "plain", "utf-8"))
        msg.attach(MIMEText(html_body, "html", "utf-8"))

        with smtplib.SMTP(smtp_server, smtp_port) as server:
            server.starttls()
            server.login(smtp_username, smtp_password)
            server.send_message(msg)

        logger.info(f"Password changed notification sent to {to_email}")
        return True
    except Exception as e:
        logger.error(f"Failed to send password changed notification to {to_email}: {str(e)}")
        return False


# --- Contact Inquiry Emails ---

def send_contact_inquiry_email(full_name, email, phone, subject, message):
    """
    Send incoming contact inquiry email to the admin inbox (alumnigo.sih@gmail.com).
    Runs asynchronously in a background thread to prevent UI latency.
    """
    import threading
    from datetime import datetime
    from algo.email_templates import build_inquiry_received_email

    def _send():
        try:
            smtp_server = os.getenv("SMTP_SERVER", "smtp.gmail.com")
            smtp_port = int(os.getenv("SMTP_PORT", "587"))
            smtp_username = os.getenv("SMTP_USERNAME") or os.getenv("EMAIL_USER")
            smtp_password = os.getenv("SMTP_PASSWORD") or os.getenv("EMAIL_PASS")
            to_email = os.getenv("CONTACT_RECEIVER_EMAIL", "alumnigo.sih@gmail.com")
            from_email = os.getenv("FROM_EMAIL", smtp_username or to_email)
            base_url = os.getenv("BASE_URL", "http://localhost:5000")

            if not smtp_username or not smtp_password:
                logger.warning("SMTP credentials not configured; skipping inquiry email sending.")
                return False

            now_ist = format_ist_time(datetime.utcnow(), "%Y-%m-%d %I:%M %p IST")
            admin_dash_url = f"{base_url}/admin_dashboard"

            email_subject, plain_body, html_body = build_inquiry_received_email(
                full_name=full_name,
                email=email,
                phone=phone,
                subject=subject,
                message=message,
                submitted_at=now_ist,
                admin_dashboard_url=admin_dash_url,
            )

            msg = MIMEMultipart("alternative")
            msg["From"] = from_email
            msg["To"] = to_email
            msg["Reply-To"] = email
            msg["Subject"] = email_subject
            msg.attach(MIMEText(plain_body, "plain", "utf-8"))
            msg.attach(MIMEText(html_body, "html", "utf-8"))

            with smtplib.SMTP(smtp_server, smtp_port) as server:
                server.starttls()
                server.login(smtp_username, smtp_password)
                server.send_message(msg)

            logger.info(f"Contact inquiry from {email} successfully forwarded to {to_email}")
            return True
        except Exception as e:
            logger.error(f"Failed to forward contact inquiry from {email}: {str(e)}")
            return False

    t = threading.Thread(target=_send, daemon=True)
    t.start()


def send_inquiry_confirmation_email(
    to_email: str,
    full_name: str,
    subject: str,
    message: str,
    inquiry_id: int | None = None,
) -> None:
    """
    Send an automatic confirmation email to the user acknowledging receipt of their query.
    Runs asynchronously in a daemon thread so it never delays the HTTP request.
    """
    import threading
    from datetime import datetime
    from algo.email_templates import build_inquiry_confirmation_email

    def _send():
        try:
            smtp_server = os.getenv("SMTP_SERVER", "smtp.gmail.com")
            smtp_port = int(os.getenv("SMTP_PORT", "587"))
            smtp_username = os.getenv("SMTP_USERNAME") or os.getenv("EMAIL_USER")
            smtp_password = os.getenv("SMTP_PASSWORD") or os.getenv("EMAIL_PASS")
            from_email = os.getenv("FROM_EMAIL", smtp_username or "alumnigo.sih@gmail.com")
            base_url = os.getenv("BASE_URL", "http://localhost:5000")

            if not smtp_username or not smtp_password:
                logger.warning("SMTP credentials not configured; skipping inquiry confirmation email.")
                return False

            now_ist = format_ist_time(datetime.utcnow(), "%Y-%m-%d %I:%M %p IST")
            email_subject, plain_body, html_body = build_inquiry_confirmation_email(
                full_name=full_name,
                subject=subject,
                message=message,
                submitted_at=now_ist,
                inquiry_id=inquiry_id,
                home_url=base_url,
            )

            reply_to = os.getenv("CONTACT_RECEIVER_EMAIL", "alumnigo.sih@gmail.com")

            msg = MIMEMultipart("alternative")
            msg["From"] = from_email
            msg["To"] = to_email
            msg["Reply-To"] = reply_to
            msg["Subject"] = email_subject
            msg.attach(MIMEText(plain_body, "plain", "utf-8"))
            msg.attach(MIMEText(html_body, "html", "utf-8"))

            with smtplib.SMTP(smtp_server, smtp_port) as server:
                server.starttls()
                server.login(smtp_username, smtp_password)
                server.send_message(msg)

            logger.info(f"Auto-confirmation email sent to {to_email} for inquiry #{inquiry_id}")
            return True
        except Exception as e:
            logger.error(f"Failed to send auto-confirmation email to {to_email}: {str(e)}")
            return False

    t = threading.Thread(target=_send, daemon=True)
    t.start()


def send_inquiry_resolved_email(
    to_email: str,
    full_name: str,
    subject: str,
    resolution_notes: str,
    original_message: str = "",
    resolver_name: str = "Admin",
) -> None:
    """Send an asynchronous branded HTML email notification to the user whose contact inquiry was resolved."""
    import smtplib
    from email.mime.text import MIMEText
    from email.mime.multipart import MIMEMultipart
    import threading
    from datetime import datetime
    from algo.email_templates import build_inquiry_resolved_email

    def _send():
        try:
            smtp_server = os.getenv("SMTP_SERVER", "smtp.gmail.com")
            smtp_port = int(os.getenv("SMTP_PORT", "587"))
            smtp_username = os.getenv("SMTP_USERNAME") or os.getenv("EMAIL_USER")
            smtp_password = os.getenv("SMTP_PASSWORD") or os.getenv("EMAIL_PASS")
            from_email = os.getenv("FROM_EMAIL", smtp_username or "alumnigo.sih@gmail.com")
            base_url = os.getenv("BASE_URL", "http://localhost:5000")

            if not smtp_username or not smtp_password:
                logger.warning("SMTP credentials not configured; skipping inquiry resolved email sending.")
                return False

            now_ist = format_ist_time(datetime.utcnow(), "%Y-%m-%d %I:%M %p IST")
            contact_url = f"{base_url}/contact"

            email_subject, plain_body, html_body = build_inquiry_resolved_email(
                full_name=full_name,
                subject=subject,
                resolution_notes=resolution_notes,
                original_message=original_message,
                resolved_at=now_ist,
                resolver_name=resolver_name,
                contact_url=contact_url,
            )

            reply_to = os.getenv("CONTACT_RECEIVER_EMAIL", "alumnigo.sih@gmail.com")

            msg = MIMEMultipart("alternative")
            msg["From"] = from_email
            msg["To"] = to_email
            msg["Reply-To"] = reply_to
            msg["Subject"] = email_subject
            msg.attach(MIMEText(plain_body, "plain", "utf-8"))
            msg.attach(MIMEText(html_body, "html", "utf-8"))

            with smtplib.SMTP(smtp_server, smtp_port) as server:
                server.starttls()
                server.login(smtp_username, smtp_password)
                server.send_message(msg)

            logger.info(f"Resolution notification sent successfully to {to_email} for query '{subject}'")
            return True
        except Exception as e:
            logger.error(f"Failed to send resolution notification to {to_email}: {str(e)}")
            return False

    t = threading.Thread(target=_send, daemon=True)
    t.start()

