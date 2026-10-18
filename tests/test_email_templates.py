import os
import sys
import pytest

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../app/src')))

from algo.email_templates import (
    build_inquiry_received_email,
    build_inquiry_resolved_email,
    build_password_reset_email,
    build_password_changed_email,
)


def test_build_inquiry_received_email():
    subject, plain, html = build_inquiry_received_email(
        full_name="Alice Smith",
        email="alice@example.com",
        phone="9876543210",
        subject="Project Collaboration",
        message="Can I participate in SIH?",
        submitted_at="2026-09-30 06:00 PM IST",
    )
    assert "[AlumniGo Contact Inquiry]" in subject
    assert "Alice Smith" in plain
    assert "Can I participate in SIH?" in plain
    assert "algo" in html
    assert "linear-gradient(135deg, #667eea 0%, #764ba2 100%)" in html
    assert "Alice Smith" in html
    assert "mailto:alice@example.com" in html


def test_build_inquiry_resolved_email():
    subject, plain, html = build_inquiry_resolved_email(
        full_name="Bob Jones",
        subject="Verification Request",
        resolution_notes="Your account has been verified.",
        original_message="Please verify my account.",
        resolved_at="2026-09-30 06:05 PM IST",
        resolver_name="Super Admin",
    )
    assert "[AlumniGo Support]" in subject
    assert "Your account has been verified." in plain
    assert "Super Admin" in plain
    assert "Inquiry Resolved" in html
    assert "#10b981" in html  # Header banner badge color
    assert "Admin Response / Resolution Note" in html
    assert "Resolved on" in html
    assert "Super Admin" in html
    assert "Your account has been verified." in html
    assert "Please verify my account." in html


def test_build_password_reset_and_changed_emails():
    s1, p1, h1 = build_password_reset_email(
        user_name="Charlie",
        reset_url="http://localhost:5000/reset-password/tok123",
        reset_token="tok123",
    )
    assert "Reset Your ALGO Password" in s1
    assert "tok123" in p1
    assert "tok123" in h1
    assert "Reset My Password" in h1

    s2, p2, h2 = build_password_changed_email(user_name="Charlie")
    assert "Your ALGO Password Has Been Changed" in s2
    assert "ALGO account password has been successfully changed" in p2
    assert "Security Alert" in h2


def test_build_inquiry_confirmation_email():
    from algo.email_templates import build_inquiry_confirmation_email

    subject, plain, html = build_inquiry_confirmation_email(
        full_name="Dave Wilson",
        subject="Internship Guidance",
        message="Looking for alumni at tech companies.",
        submitted_at="2026-09-30 06:15 PM IST",
        inquiry_id=42,
    )
    assert "[AlumniGo] We've received your message" in subject
    assert "Dave Wilson" in plain
    assert "#42" in plain
    assert "Dave Wilson" in html
    assert "#42" in html
    assert "Looking for alumni at tech companies." in html
    assert "Message Received" in html

