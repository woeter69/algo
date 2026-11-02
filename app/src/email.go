package main

import (
	"fmt"
	"html"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

func escapeHTML(text string) string {
	return html.EscapeString(text)
}

func renderEmailWrapper(title, contentHTML, preheader, badgeText, badgeBg, badgeColor string) string {
	if badgeBg == "" {
		badgeBg = "#10b981"
	}
	if badgeColor == "" {
		badgeColor = "#ffffff"
	}

	badgeCell := ""
	if badgeText != "" {
		badgeCell = fmt.Sprintf(`
                  <td align="right" valign="top">
                    <span style="display: inline-block; background-color: %s; color: %s; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.8px; padding: 6px 14px; border-radius: 20px; box-shadow: 0 2px 6px rgba(0,0,0,0.15);">
                      %s
                    </span>
                  </td>`, badgeBg, badgeColor, escapeHTML(badgeText))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
</head>
<body style="margin: 0; padding: 0; background-color: #f1f5f9; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; -webkit-font-smoothing: antialiased; color: #1e293b;">
  <div style="display: none; max-height: 0px; overflow: hidden; opacity: 0;">
    %s
  </div>
  <table width="100%%" border="0" cellpadding="0" cellspacing="0" style="background-color: #f1f5f9; padding: 30px 10px;">
    <tr>
      <td align="center">
        <table width="100%%" border="0" cellpadding="0" cellspacing="0" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 10px 25px rgba(0, 0, 0, 0.08); border: 1px solid #e2e8f0;">
          <tr>
            <td style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 34px 32px 28px 32px; text-align: left;">
              <table width="100%%" border="0" cellpadding="0" cellspacing="0">
                <tr>
                  <td>
                    <div style="display: inline-block;">
                      <span style="font-size: 28px; font-weight: 800; letter-spacing: -0.5px; color: #ffffff; text-decoration: none;">algo</span>
                      <span style="display: inline-block; width: 8px; height: 8px; background-color: #38bdf8; border-radius: 50%%; margin-left: 2px;"></span>
                    </div>
                    <div style="color: #e0e7ff; font-size: 13px; font-weight: 500; margin-top: 4px; letter-spacing: 0.2px;">
                      AlumniGo &bull; Smart Alumni & Student Network
                    </div>
                  </td>
                  %s
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding: 36px 32px 28px 32px; background-color: #ffffff;">
              %s
            </td>
          </tr>
          <tr>
            <td style="background-color: #f8fafc; padding: 24px 32px; border-top: 1px solid #f1f5f9; text-align: center;">
              <div style="font-size: 13px; font-weight: 600; color: #475569; margin-bottom: 8px;">
                AlumniGo Platform
              </div>
              <div style="font-size: 12px; color: #64748b; line-height: 1.6; margin-bottom: 14px;">
                Connecting students, alumni, and institutions for mentorship and career opportunities.
              </div>
              <div style="font-size: 12px; color: #94a3b8;">
                &copy; %d AlumniGo. All rights reserved.
              </div>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, escapeHTML(title), escapeHTML(preheader), badgeCell, contentHTML, time.Now().Year())
}

// sendEmailAsync sends email in a background goroutine using standard SMTP
func sendEmailAsync(toEmail, subject, plainText, htmlContent, replyTo string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ Panic in sendEmailAsync: %v", r)
			}
		}()

		smtpServer := os.Getenv("SMTP_SERVER")
		if smtpServer == "" {
			smtpServer = "smtp.gmail.com"
		}
		smtpPortStr := os.Getenv("SMTP_PORT")
		if smtpPortStr == "" {
			smtpPortStr = "587"
		}
		smtpPort, _ := strconv.Atoi(smtpPortStr)

		smtpUser := os.Getenv("SMTP_USERNAME")
		if smtpUser == "" {
			smtpUser = os.Getenv("EMAIL_USER")
		}
		smtpPass := os.Getenv("SMTP_PASSWORD")
		if smtpPass == "" {
			smtpPass = os.Getenv("EMAIL_PASS")
		}
		fromEmail := os.Getenv("FROM_EMAIL")
		if fromEmail == "" {
			fromEmail = smtpUser
		}
		if fromEmail == "" {
			fromEmail = "alumnigo.sih@gmail.com"
		}

		if smtpUser == "" || smtpPass == "" {
			log.Printf("⚠️ SMTP credentials not configured; skipping email to %s for subject: %s", toEmail, subject)
			return
		}

		boundary := "boundary-" + strconv.FormatInt(time.Now().UnixNano(), 16)
		headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n",
			fromEmail, toEmail, subject, boundary)
		if replyTo != "" {
			headers += fmt.Sprintf("Reply-To: %s\r\n", replyTo)
		}
		headers += "\r\n"

		body := fmt.Sprintf("--%s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n\r\n", boundary, plainText)
		body += fmt.Sprintf("--%s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s\r\n\r\n--%s--\r\n", boundary, htmlContent, boundary)

		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpServer)
		addr := fmt.Sprintf("%s:%d", smtpServer, smtpPort)

		err := smtp.SendMail(addr, auth, fromEmail, []string{toEmail}, []byte(headers+body))
		if err != nil {
			log.Printf("❌ Failed to send email to %s: %v", toEmail, err)
		} else {
			log.Printf("✅ Email successfully sent to %s (subject: %s)", toEmail, subject)
		}
	}()
}

// BuildPasswordResetEmail builds the password reset email content and returns (subject, plainText, htmlBody).
func BuildPasswordResetEmail(userName, resetURL, resetToken string) (subject, plainText, htmlBody string) {
	subject = "Reset Your ALGO Password"
	greeting := "Hi there,"
	if userName != "" {
		greeting = fmt.Sprintf("Hi %s,", userName)
	}

	plainText = fmt.Sprintf(`%s

We received a request to reset your password for your ALGO account.

Visit this link to reset your password:
%s

Or use this token: %s

This link expires in 1 hour. If you did not request this, please ignore this email.

Best regards,
The ALGO Team`, greeting, resetURL, resetToken)

	contentHTML := fmt.Sprintf(`
      <h2 style="margin: 0 0 16px; color: #1e293b; font-size: 22px;">Password Reset Request</h2>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">%s</p>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">We received a request to reset the password for your AlumniGo account. Click the button below to choose a new password.</p>
      <div style="text-align: center; margin: 30px 0;">
        <a href="%s" style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: #ffffff; text-decoration: none; padding: 14px 32px; border-radius: 8px; font-weight: 600; display: inline-block;">Reset My Password</a>
      </div>
      <p style="color: #64748b; font-size: 13px;">Or copy and paste this URL into your browser:<br><a href="%s" style="color: #6366f1;">%s</a></p>
      <p style="color: #94a3b8; font-size: 12px; margin-top: 24px;">This link will expire in 1 hour.</p>
    `, escapeHTML(greeting), resetURL, resetURL, resetURL)

	htmlBody = renderEmailWrapper(subject, contentHTML, "Reset your AlumniGo password", "Security", "#f59e0b", "#ffffff")
	return
}

// BuildInquiryConfirmationEmail builds the inquiry confirmation email content and returns (subject, plainText, htmlBody).
func BuildInquiryConfirmationEmail(fullName, subjectStr, messageStr, submittedAt string, inquiryID int) (subject, plainText, htmlBody string) {
	subject = fmt.Sprintf("[AlumniGo] We've received your message: %s", subjectStr)

	plainText = fmt.Sprintf(`Hi %s,

Thank you for reaching out to AlumniGo. We have received your inquiry and assigned it ticket #%d.

Subject: %s
Submitted At: %s
Message:
%s

Our team will review your inquiry and get back to you shortly.

Best regards,
The ALGO Team`, fullName, inquiryID, subjectStr, submittedAt, messageStr)

	contentHTML := fmt.Sprintf(`
      <h2 style="margin: 0 0 16px; color: #1e293b; font-size: 22px;">Message Received</h2>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">Hi <strong>%s</strong>,</p>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">Thank you for getting in touch with the AlumniGo team. We have received your query and logged it under ticket <strong>#%d</strong>.</p>
      <div style="background-color: #f8fafc; border-left: 4px solid #6366f1; padding: 16px 20px; border-radius: 0 8px 8px 0; margin: 20px 0;">
        <div style="font-weight: 600; color: #334155; margin-bottom: 6px;">Subject: %s</div>
        <div style="color: #64748b; font-size: 14px; line-height: 1.5;">%s</div>
        <div style="color: #94a3b8; font-size: 12px; margin-top: 10px;">Submitted: %s</div>
      </div>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">Our community administrators will review your message and reply as soon as possible.</p>
    `, escapeHTML(fullName), inquiryID, escapeHTML(subjectStr), escapeHTML(messageStr), submittedAt)

	htmlBody = renderEmailWrapper(subject, contentHTML, "We've received your message", "Message Received", "#6366f1", "#ffffff")
	return
}

// BuildInquiryResolvedEmail builds the inquiry resolved email content and returns (subject, plainText, htmlBody).
func BuildInquiryResolvedEmail(fullName, subjectStr, resolutionNotes, originalMessage, resolvedAt, resolverName string) (subject, plainText, htmlBody string) {
	subject = fmt.Sprintf("[AlumniGo Support] Your inquiry has been resolved: %s", subjectStr)

	plainText = fmt.Sprintf(`Hi %s,

Your contact inquiry regarding "%s" has been marked as resolved by %s on %s.

Resolution Notes:
%s

Original Message:
%s

If you have further questions, feel free to contact us via our contact page.

Best regards,
The ALGO Team`, fullName, subjectStr, resolverName, resolvedAt, resolutionNotes, originalMessage)

	contentHTML := fmt.Sprintf(`
      <h2 style="margin: 0 0 16px; color: #065f46; font-size: 22px;">Inquiry Resolved</h2>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">Hi <strong>%s</strong>,</p>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">Your contact inquiry regarding "<strong>%s</strong>" has been addressed and marked as <strong>Resolved</strong> by %s on %s.</p>
      <div style="background-color: #ecfdf5; border-left: 4px solid #10b981; padding: 16px 20px; border-radius: 0 8px 8px 0; margin: 20px 0;">
        <div style="font-weight: 700; color: #065f46; margin-bottom: 6px;">Resolution Notes:</div>
        <div style="color: #047857; font-size: 14px; line-height: 1.5;">%s</div>
      </div>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">If you need further assistance, you can always submit a new inquiry via <a href="#" style="color: #6366f1;">our contact page</a>.</p>
    `, escapeHTML(fullName), escapeHTML(subjectStr), escapeHTML(resolverName), resolvedAt, escapeHTML(resolutionNotes))

	htmlBody = renderEmailWrapper(subject, contentHTML, "Your inquiry has been resolved", "Resolved", "#10b981", "#ffffff")
	return
}

// BuildInquiryReceivedEmail builds the admin notification email for a new contact inquiry and returns (subject, plainText, htmlBody).
func BuildInquiryReceivedEmail(fullName, email, phone, subjectStr, message, submittedAt string) (subject, plainText, htmlBody string) {
	subject = fmt.Sprintf("[AlumniGo Contact Inquiry] New message from %s: %s", fullName, subjectStr)

	plainText = fmt.Sprintf(`New contact inquiry received on AlumniGo.

From: %s
Email: %s
Phone: %s
Subject: %s
Submitted At: %s

Message:
%s`, fullName, email, phone, subjectStr, submittedAt, message)

	contentHTML := fmt.Sprintf(`
      <h2 style="margin: 0 0 16px; color: #1e293b; font-size: 22px;">New Contact Inquiry</h2>
      <p style="color: #475569; font-size: 15px; line-height: 1.6;">A new contact inquiry has been submitted by <strong>%s</strong>.</p>
      <div style="background-color: #f8fafc; border-left: 4px solid #6366f1; padding: 16px 20px; border-radius: 0 8px 8px 0; margin: 20px 0;">
        <div style="margin-bottom: 8px;"><strong style="color: #334155;">Name:</strong> <span style="color: #475569;">%s</span></div>
        <div style="margin-bottom: 8px;"><strong style="color: #334155;">Email:</strong> <a href="mailto:%s" style="color: #6366f1;">%s</a></div>
        <div style="margin-bottom: 8px;"><strong style="color: #334155;">Phone:</strong> <span style="color: #475569;">%s</span></div>
        <div style="margin-bottom: 8px;"><strong style="color: #334155;">Subject:</strong> <span style="color: #475569;">%s</span></div>
        <div style="color: #94a3b8; font-size: 12px; margin-bottom: 10px;">Submitted: %s</div>
      </div>
      <div style="background-color: #f1f5f9; padding: 16px 20px; border-radius: 8px; margin: 20px 0;">
        <div style="font-weight: 600; color: #334155; margin-bottom: 8px;">Message:</div>
        <div style="color: #475569; font-size: 14px; line-height: 1.6;">%s</div>
      </div>
    `, escapeHTML(fullName), escapeHTML(fullName), escapeHTML(email), escapeHTML(email), escapeHTML(phone), escapeHTML(subjectStr), submittedAt, escapeHTML(message))

	htmlBody = renderEmailWrapper(subject, contentHTML, "New contact inquiry received", "New Inquiry", "#6366f1", "#ffffff")
	return
}

// SendPasswordResetEmail sends password reset email
func SendPasswordResetEmail(toEmail, resetToken, userName string) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	resetURL := fmt.Sprintf("%s/reset-password/%s", baseURL, resetToken)

	subject, plainText, htmlBody := BuildPasswordResetEmail(userName, resetURL, resetToken)
	sendEmailAsync(toEmail, subject, plainText, htmlBody, "")
}

// SendInquiryConfirmationEmail sends auto-reply when contact form is submitted
func SendInquiryConfirmationEmail(toEmail, fullName, subjectStr, messageStr string, inquiryID int) {
	submittedAt := FormatISTTime(time.Now(), "%Y-%m-%d %I:%M %p IST")
	subject, plainText, htmlBody := BuildInquiryConfirmationEmail(fullName, subjectStr, messageStr, submittedAt, inquiryID)
	replyTo := os.Getenv("CONTACT_RECEIVER_EMAIL")
	if replyTo == "" {
		replyTo = "alumnigo.sih@gmail.com"
	}
	sendEmailAsync(toEmail, subject, plainText, htmlBody, replyTo)
}

// SendInquiryResolvedEmail sends notification when contact inquiry is resolved
func SendInquiryResolvedEmail(toEmail, fullName, subjectStr, resolutionNotes, originalMessage, resolverName string) {
	resolvedAt := FormatISTTime(time.Now(), "%Y-%m-%d %I:%M %p IST")
	subject, plainText, htmlBody := BuildInquiryResolvedEmail(fullName, subjectStr, resolutionNotes, originalMessage, resolvedAt, resolverName)
	replyTo := os.Getenv("CONTACT_RECEIVER_EMAIL")
	if replyTo == "" {
		replyTo = "alumnigo.sih@gmail.com"
	}
	sendEmailAsync(toEmail, subject, plainText, htmlBody, replyTo)
}
