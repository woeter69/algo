package main

import (
	"strings"
	"testing"
)

func TestBuildInquiryReceivedEmail(t *testing.T) {
	subject, plain, html := BuildInquiryReceivedEmail(
		"Alice Smith",
		"alice@example.com",
		"9876543210",
		"Project Collaboration",
		"Can I participate in SIH?",
		"2026-09-30 06:00 PM IST",
	)

	if !strings.Contains(subject, "[AlumniGo Contact Inquiry]") {
		t.Errorf("expected subject to contain '[AlumniGo Contact Inquiry]', got %s", subject)
	}
	if !strings.Contains(plain, "Alice Smith") {
		t.Errorf("expected plain body to contain 'Alice Smith'")
	}
	if !strings.Contains(plain, "Can I participate in SIH?") {
		t.Errorf("expected plain body to contain 'Can I participate in SIH?'")
	}
	if !strings.Contains(html, "Alice Smith") {
		t.Errorf("expected html body to contain 'Alice Smith'")
	}
	if !strings.Contains(html, "mailto:alice@example.com") {
		t.Errorf("expected html body to contain 'mailto:alice@example.com'")
	}
}

func TestBuildInquiryResolvedEmail(t *testing.T) {
	subject, plain, html := BuildInquiryResolvedEmail(
		"Bob Jones",
		"Verification Request",
		"Your account has been verified.",
		"Please verify my account.",
		"2026-09-30 06:05 PM IST",
		"Super Admin",
	)

	if !strings.Contains(subject, "[AlumniGo Support]") {
		t.Errorf("expected subject to contain '[AlumniGo Support]', got %s", subject)
	}
	if !strings.Contains(plain, "Your account has been verified.") {
		t.Errorf("expected plain body to contain resolution notes")
	}
	if !strings.Contains(plain, "Super Admin") {
		t.Errorf("expected plain body to contain resolver name")
	}
	if !strings.Contains(html, "Inquiry Resolved") {
		t.Errorf("expected html body to contain 'Inquiry Resolved'")
	}
	if !strings.Contains(html, "#10b981") {
		t.Errorf("expected html body to contain badge color")
	}
}

func TestBuildPasswordResetEmail(t *testing.T) {
	subject, plain, html := BuildPasswordResetEmail(
		"Charlie",
		"http://localhost:8080/reset-password/tok123",
		"tok123",
	)

	if !strings.Contains(subject, "Reset Your ALGO Password") {
		t.Errorf("expected subject to contain 'Reset Your ALGO Password', got %s", subject)
	}
	if !strings.Contains(plain, "tok123") {
		t.Errorf("expected plain body to contain token")
	}
	if !strings.Contains(html, "Reset My Password") {
		t.Errorf("expected html body to contain button")
	}
}

func TestBuildInquiryConfirmationEmail(t *testing.T) {
	subject, plain, html := BuildInquiryConfirmationEmail(
		"Dave Wilson",
		"Internship Guidance",
		"Looking for alumni at tech companies.",
		"2026-09-30 06:15 PM IST",
		42,
	)

	if !strings.Contains(subject, "[AlumniGo] We've received your message") {
		t.Errorf("expected subject to contain title")
	}
	if !strings.Contains(plain, "Dave Wilson") {
		t.Errorf("expected plain body to contain name")
	}
	if !strings.Contains(plain, "#42") {
		t.Errorf("expected plain body to contain inquiry ID")
	}
	if !strings.Contains(html, "#42") {
		t.Errorf("expected html body to contain inquiry ID")
	}
	if !strings.Contains(html, "Message Received") {
		t.Errorf("expected html body to contain 'Message Received'")
	}
}
