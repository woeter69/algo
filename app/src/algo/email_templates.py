"""
HTML and Plaintext Email Templates for AlumniGo (ALGO).
Designed with modern typography, clean theme-matching gradient banners,
subtle pill indicators, and robust email-client compatibility.
"""

import html


def _escape(text):
    if text is None:
        return ""
    return html.escape(str(text))


def render_email_wrapper(
    title: str,
    content_html: str,
    preheader: str = "",
    badge_text: str = "",
    badge_bg: str = "#10b981",
    badge_color: str = "#ffffff",
) -> str:
    """Master responsive HTML email container with branded ALGO banner and footer."""
    badge_cell = ""
    if badge_text:
        badge_cell = f"""
                  <td align="right" valign="top">
                    <!-- Status / Type Badge -->
                    <span style="display: inline-block; background-color: {badge_bg}; color: {badge_color}; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.8px; padding: 6px 14px; border-radius: 20px; box-shadow: 0 2px 6px rgba(0,0,0,0.15);">
                      {_escape(badge_text)}
                    </span>
                  </td>"""

    return f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{_escape(title)}</title>
</head>
<body style="margin: 0; padding: 0; background-color: #f1f5f9; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; -webkit-font-smoothing: antialiased; color: #1e293b;">
  <!-- Preview Preheader -->
  <div style="display: none; max-height: 0px; overflow: hidden; opacity: 0;">
    {_escape(preheader)}
  </div>

  <table width="100%" border="0" cellpadding="0" cellspacing="0" style="background-color: #f1f5f9; padding: 30px 10px;">
    <tr>
      <td align="center">
        <!-- Main Card Container -->
        <table width="100%" border="0" cellpadding="0" cellspacing="0" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 10px 25px rgba(0, 0, 0, 0.08); border: 1px solid #e2e8f0;">
          
          <!-- BRAND HEADER BANNER -->
          <tr>
            <td style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 34px 32px 28px 32px; text-align: left;">
              <table width="100%" border="0" cellpadding="0" cellspacing="0">
                <tr>
                  <td>
                    <!-- Logo -->
                    <div style="display: inline-block;">
                      <span style="font-size: 28px; font-weight: 800; letter-spacing: -0.5px; color: #ffffff; text-decoration: none;">algo</span>
                      <span style="display: inline-block; width: 8px; height: 8px; background-color: #38bdf8; border-radius: 50%; margin-left: 2px;"></span>
                    </div>
                    <div style="color: #e0e7ff; font-size: 13px; font-weight: 500; margin-top: 4px; letter-spacing: 0.2px;">
                      AlumniGo &bull; Smart Alumni & Student Network
                    </div>
                  </td>
                  {badge_cell}
                </tr>
              </table>
            </td>
          </tr>

          <!-- MAIN CONTENT BODY -->
          <tr>
            <td style="padding: 36px 32px 28px 32px; background-color: #ffffff;">
              {content_html}
            </td>
          </tr>

          <!-- FOOTER -->
          <tr>
            <td style="background-color: #f8fafc; padding: 24px 32px; border-top: 1px solid #f1f5f9; text-align: center;">
              <div style="font-size: 13px; font-weight: 600; color: #475569; margin-bottom: 8px;">
                AlumniGo Platform
              </div>
              <div style="font-size: 12px; color: #64748b; line-height: 1.6; margin-bottom: 14px;">
                Connecting students, alumni, and institutions for mentorship and career opportunities.
              </div>
              <div style="font-size: 12px; color: #94a3b8;">
                &copy; 2026 AlumniGo. All rights reserved.
              </div>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>"""


# --- 1. Contact Inquiry Received (To Admin) ---

def build_inquiry_received_email(
    full_name: str,
    email: str,
    phone: str,
    subject: str,
    message: str,
    submitted_at: str,
    admin_dashboard_url: str = "http://localhost:5000/admin_dashboard",
) -> tuple[str, str, str]:
    """
    Build email subject, plain text body, and styled HTML body for incoming inquiry notification.
    """
    email_subject = f"[AlumniGo Contact Inquiry] {subject} - {full_name}"
    preheader = f"New inquiry from {full_name} regarding {subject}"

    plain_body = f"""New Contact Inquiry Received via AlumniGo Portal

From: {full_name}
Email: {email}
Phone: {phone if phone else 'Not provided'}
Subject: {subject}
Submitted At: {submitted_at}

Message:
----------------------------------------------------------------------
{message}
----------------------------------------------------------------------

Review and resolve this inquiry in the Admin Dashboard:
{admin_dashboard_url}
"""

    phone_display = _escape(phone) if phone else "<span style='color: #94a3b8; font-style: italic;'>Not provided</span>"

    content_html = f"""
      <h2 style="margin: 0 0 8px 0; color: #0f172a; font-size: 20px; font-weight: 700;">
        New Contact Inquiry Received
      </h2>
      <p style="margin: 0 0 24px 0; color: #64748b; font-size: 14px; line-height: 1.5;">
        A user has submitted a question or feedback through the AlumniGo Contact Us page.
      </p>

      <!-- Sender Details Box -->
      <table width="100%" border="0" cellpadding="0" cellspacing="0" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0; margin-bottom: 24px;">
        <tr>
          <td style="padding: 16px 20px;">
            <table width="100%" border="0" cellpadding="0" cellspacing="0">
              <tr>
                <td style="padding: 6px 0; width: 110px; font-size: 13px; font-weight: 600; color: #64748b;">Sender:</td>
                <td style="padding: 6px 0; font-size: 14px; font-weight: 600; color: #0f172a;">{_escape(full_name)}</td>
              </tr>
              <tr>
                <td style="padding: 6px 0; font-size: 13px; font-weight: 600; color: #64748b;">Email:</td>
                <td style="padding: 6px 0; font-size: 14px; color: #2563eb;"><a href="mailto:{_escape(email)}" style="color: #2563eb; text-decoration: none;">{_escape(email)}</a></td>
              </tr>
              <tr>
                <td style="padding: 6px 0; font-size: 13px; font-weight: 600; color: #64748b;">Phone:</td>
                <td style="padding: 6px 0; font-size: 14px; color: #334155;">{phone_display}</td>
              </tr>
              <tr>
                <td style="padding: 6px 0; font-size: 13px; font-weight: 600; color: #64748b;">Subject:</td>
                <td style="padding: 6px 0; font-size: 14px; font-weight: 600; color: #0f172a;">{_escape(subject)}</td>
              </tr>
              <tr>
                <td style="padding: 6px 0; font-size: 13px; font-weight: 600; color: #64748b;">Received:</td>
                <td style="padding: 6px 0; font-size: 13px; color: #64748b;">{_escape(submitted_at)}</td>
              </tr>
            </table>
          </td>
        </tr>
      </table>

      <!-- Message Card -->
      <div style="font-size: 13px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; color: #475569; margin-bottom: 8px;">
        Inquiry Message
      </div>
      <div style="background-color: #f1f5f9; border-left: 4px solid #667eea; border-radius: 0 10px 10px 0; padding: 18px 20px; font-size: 14px; line-height: 1.6; color: #1e293b; white-space: pre-wrap; margin-bottom: 28px;">
{_escape(message)}
      </div>

      <!-- Action Button -->
      <table width="100%" border="0" cellpadding="0" cellspacing="0" style="text-align: center;">
        <tr>
          <td align="center">
            <a href="{admin_dashboard_url}" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; font-size: 14px; font-weight: 600; padding: 13px 28px; border-radius: 10px; box-shadow: 0 4px 12px rgba(102, 126, 234, 0.35);">
              Manage in Admin Dashboard &rarr;
            </a>
          </td>
        </tr>
      </table>
    """

    html_body = render_email_wrapper(
        title=f"New Contact Inquiry: {subject}",
        badge_text="Contact Inquiry",
        badge_bg="#2563eb",
        badge_color="#ffffff",
        content_html=content_html,
        preheader=preheader,
    )

    return email_subject, plain_body, html_body


# --- 2. Inquiry Resolved (To Querier) ---

def build_inquiry_resolved_email(
    full_name: str,
    subject: str,
    resolution_notes: str,
    original_message: str,
    resolved_at: str,
    resolver_name: str = "Admin",
    contact_url: str = "http://localhost:5000/contact",
) -> tuple[str, str, str]:
    """
    Build email subject, plain text body, and styled HTML body for inquiry resolution notification.
    """
    email_subject = f"[AlumniGo Support] Your inquiry has been resolved: {subject}"
    preheader = f"Your inquiry regarding '{subject}' has been resolved by our team."

    plain_body = f"""Dear {full_name},

Thank you for reaching out to AlumniGo. Your inquiry regarding "{subject}" has been reviewed and resolved by our administration team.

Resolution Note / Response:
----------------------------------------------------------------------
{resolution_notes}
----------------------------------------------------------------------

Resolved At: {resolved_at}
Resolved By: {resolver_name}

Your Original Message:
----------------------------------------------------------------------
{original_message}
----------------------------------------------------------------------

If you have further questions or require additional assistance, please feel free to reach back out to us at:
{contact_url}

Best regards,
The AlumniGo Administration Team
"""

    content_html = f"""
      <div style="margin-bottom: 24px;">
        <h2 style="margin: 0 0 8px 0; color: #0f172a; font-size: 22px; font-weight: 700;">
          Hi {_escape(full_name)},
        </h2>
        <p style="margin: 0; color: #475569; font-size: 15px; line-height: 1.6;">
          Your recent inquiry regarding <strong style="color: #0f172a;">"{_escape(subject)}"</strong> has been reviewed and resolved by our administration team.
        </p>
      </div>

      <!-- Admin Response Section -->
      <div style="margin-bottom: 8px;">
        <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.6px; color: #15803d; margin-bottom: 3px;">
          Admin Response / Resolution Note
        </div>
        <div style="font-size: 12px; color: #16a34a; font-weight: 500;">
          Resolved on <strong>{_escape(resolved_at)}</strong> by <strong>{_escape(resolver_name)}</strong>
        </div>
      </div>

      <!-- Resolution Response Box (Response Only) -->
      <div style="background: linear-gradient(180deg, #f0fdf4 0%, #ecfdf5 100%); border: 1px solid #bbf7d0; border-radius: 12px; padding: 18px 20px; font-size: 15px; line-height: 1.6; color: #166534; font-weight: 500; white-space: pre-wrap; margin-bottom: 26px;">
{_escape(resolution_notes)}
      </div>

      <!-- Original Message Recap -->
      <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; color: #64748b; margin-bottom: 8px;">
        Your Original Message
      </div>
      <div style="background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 10px; padding: 16px 18px; font-size: 13px; line-height: 1.6; color: #64748b; white-space: pre-wrap; margin-bottom: 28px;">
{_escape(original_message)}
      </div>

      <!-- Support / Contact Prompt -->
      <table width="100%" border="0" cellpadding="0" cellspacing="0" style="text-align: center;">
        <tr>
          <td align="center">
            <p style="margin: 0 0 16px 0; color: #64748b; font-size: 13px;">
              Have further questions or need additional support?
            </p>
            <a href="{contact_url}" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; font-size: 14px; font-weight: 600; padding: 12px 26px; border-radius: 10px; box-shadow: 0 4px 12px rgba(102, 126, 234, 0.35);">
              Contact AlumniGo Support &rarr;
            </a>
          </td>
        </tr>
      </table>
    """

    html_body = render_email_wrapper(
        title=f"Inquiry Resolved: {subject}",
        badge_text="Inquiry Resolved",
        badge_bg="#10b981",
        badge_color="#ffffff",
        content_html=content_html,
        preheader=preheader,
    )

    return email_subject, plain_body, html_body


# --- 3. Password Reset Email ---

def build_password_reset_email(
    user_name: str,
    reset_url: str,
    reset_token: str,
) -> tuple[str, str, str]:
    """
    Build email subject, plain text body, and styled HTML body for password reset.
    """
    email_subject = "Reset Your ALGO Password"
    greeting = f"Hi {user_name}," if user_name else "Hi there,"
    preheader = "Instructions to reset your ALGO account password."

    plain_body = f"""{greeting}

We received a request to reset your password for your ALGO account.

To reset your password, visit this link:
{reset_url}

Or use this reset token: {reset_token}

This link will expire in 1 hour for security reasons.
If you didn't request this password reset, please ignore this email.

Best regards,
The ALGO Team
"""

    content_html = f"""
      <h2 style="margin: 0 0 8px 0; color: #0f172a; font-size: 20px; font-weight: 700;">
        {_escape(greeting)}
      </h2>
      <p style="margin: 0 0 20px 0; color: #475569; font-size: 15px; line-height: 1.6;">
        We received a request to reset the password for your AlumniGo (ALGO) account. Click the button below to choose a new password:
      </p>

      <!-- Action Button -->
      <table width="100%" border="0" cellpadding="0" cellspacing="0" style="margin: 28px 0; text-align: center;">
        <tr>
          <td align="center">
            <a href="{reset_url}" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; font-size: 15px; font-weight: 600; padding: 14px 32px; border-radius: 10px; box-shadow: 0 4px 14px rgba(102, 126, 234, 0.4);">
              Reset My Password &rarr;
            </a>
          </td>
        </tr>
      </table>

      <!-- Security / Token Box -->
      <div style="background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 10px; padding: 16px 20px; font-size: 13px; line-height: 1.6; color: #64748b; margin-bottom: 20px;">
        <div style="font-weight: 600; color: #0f172a; margin-bottom: 6px;">Need to enter the token manually?</div>
        <div style="font-family: monospace; font-size: 14px; color: #4338ca; background: #e0e7ff; padding: 6px 10px; border-radius: 6px; display: inline-block;">
          {_escape(reset_token)}
        </div>
        <div style="margin-top: 10px; font-size: 12px; color: #94a3b8;">
          This link and token will expire in <strong>1 hour</strong>. If you did not request a password reset, you can safely ignore this email.
        </div>
      </div>
    """

    html_body = render_email_wrapper(
        title="Reset Your ALGO Password",
        badge_text="Security",
        badge_bg="#6366f1",
        badge_color="#ffffff",
        content_html=content_html,
        preheader=preheader,
    )

    return email_subject, plain_body, html_body


# --- 4. Password Changed Notification ---

def build_password_changed_email(
    user_name: str,
    contact_url: str = "http://localhost:5000/contact",
) -> tuple[str, str, str]:
    """
    Build email subject, plain text body, and styled HTML body for password changed notification.
    """
    email_subject = "Your ALGO Password Has Been Changed"
    greeting = f"Hi {user_name}," if user_name else "Hi there,"
    preheader = "Your AlumniGo account password was successfully updated."

    plain_body = f"""{greeting}

Your ALGO account password has been successfully changed.
If you didn't make this change, please contact our support team immediately:
{contact_url}

Best regards,
The ALGO Team
"""

    content_html = f"""
      <h2 style="margin: 0 0 8px 0; color: #0f172a; font-size: 20px; font-weight: 700;">
        {_escape(greeting)}
      </h2>
      <p style="margin: 0 0 20px 0; color: #475569; font-size: 15px; line-height: 1.6;">
        This email confirms that your AlumniGo (ALGO) account password was recently changed.
      </p>

      <div style="background-color: #fef2f2; border: 1px solid #fecaca; border-radius: 12px; padding: 18px 20px; font-size: 13px; line-height: 1.6; color: #991b1b; margin-bottom: 24px;">
        <strong>Did not make this change?</strong><br />
        If you did not initiate this change, your account may be compromised. Please reach out to our team immediately so we can secure your account.
      </div>

      <table width="100%" border="0" cellpadding="0" cellspacing="0" style="text-align: center;">
        <tr>
          <td align="center">
            <a href="{contact_url}" style="display: inline-block; background-color: #ef4444; color: #ffffff; text-decoration: none; font-size: 14px; font-weight: 600; padding: 12px 24px; border-radius: 8px;">
              Contact Support Immediately &rarr;
            </a>
          </td>
        </tr>
      </table>
    """

    html_body = render_email_wrapper(
        title="Your ALGO Password Has Been Changed",
        badge_text="Security Alert",
        badge_bg="#ef4444",
        badge_color="#ffffff",
        content_html=content_html,
        preheader=preheader,
    )

    return email_subject, plain_body, html_body


# --- 5. Inquiry Received Auto-Confirmation (To Querier) ---

def build_inquiry_confirmation_email(
    full_name: str,
    subject: str,
    message: str,
    submitted_at: str,
    inquiry_id: int | None = None,
    home_url: str = "http://localhost:5000",
) -> tuple[str, str, str]:
    """
    Build email subject, plain text body, and styled HTML body for instant auto-responder
    confirming query receipt to the user who submitted the contact form.
    """
    email_subject = f"[AlumniGo] We've received your message: {subject}"
    preheader = f"Hi {full_name}, thank you for contacting AlumniGo. Our team has received your inquiry."
    ref_display = f"#{inquiry_id}" if inquiry_id else "Pending"

    plain_body = f"""Dear {full_name},

Thank you for reaching out to AlumniGo!

We have received your message regarding "{subject}" (Reference ID: {ref_display}) and our team has noted your query.

Submitted On: {submitted_at}

Summary of your message:
----------------------------------------------------------------------
{message}
----------------------------------------------------------------------

What to expect next:
- A member of our administration or support team will review your query.
- You will receive a resolution response directly at this email address.

In the meantime, feel free to explore our platform:
{home_url}

Best regards,
The AlumniGo Support Team
"""

    content_html = f"""
      <div style="margin-bottom: 24px;">
        <h2 style="margin: 0 0 8px 0; color: #0f172a; font-size: 22px; font-weight: 700;">
          Hi {_escape(full_name)},
        </h2>
        <p style="margin: 0; color: #475569; font-size: 15px; line-height: 1.6;">
          Thank you for reaching out to us. We have received your inquiry regarding <strong style="color: #0f172a;">"{_escape(subject)}"</strong> and our team has logged it into our support queue.
        </p>
      </div>

      <!-- Ticket / Info Box -->
      <table width="100%" border="0" cellpadding="0" cellspacing="0" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0; margin-bottom: 24px;">
        <tr>
          <td style="padding: 16px 20px;">
            <table width="100%" border="0" cellpadding="0" cellspacing="0">
              <tr>
                <td style="padding: 5px 0; width: 120px; font-size: 13px; font-weight: 600; color: #64748b;">Reference ID:</td>
                <td style="padding: 5px 0; font-size: 14px; font-weight: 700; color: #4338ca;">{_escape(ref_display)}</td>
              </tr>
              <tr>
                <td style="padding: 5px 0; font-size: 13px; font-weight: 600; color: #64748b;">Status:</td>
                <td style="padding: 5px 0; font-size: 13px; font-weight: 600; color: #d97706;">In Queue / Pending Review</td>
              </tr>
              <tr>
                <td style="padding: 5px 0; font-size: 13px; font-weight: 600; color: #64748b;">Submitted On:</td>
                <td style="padding: 5px 0; font-size: 13px; color: #64748b;">{_escape(submitted_at)}</td>
              </tr>
            </table>
          </td>
        </tr>
      </table>

      <!-- Next Steps Callout -->
      <div style="background-color: #f0fdf4; border-left: 4px solid #10b981; border-radius: 0 10px 10px 0; padding: 16px 18px; font-size: 13px; line-height: 1.6; color: #166534; margin-bottom: 24px;">
        <strong>What happens next?</strong><br />
        Our college administration and support staff actively review incoming messages. We will contact you or post a resolution note directly to this email address soon.
      </div>

      <!-- Message Recap -->
      <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; color: #64748b; margin-bottom: 8px;">
        Copy of Your Submitted Message
      </div>
      <div style="background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 10px; padding: 16px 18px; font-size: 13px; line-height: 1.6; color: #475569; white-space: pre-wrap; margin-bottom: 28px;">
{_escape(message)}
      </div>

      <!-- Button -->
      <table width="100%" border="0" cellpadding="0" cellspacing="0" style="text-align: center;">
        <tr>
          <td align="center">
            <a href="{home_url}" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; font-size: 14px; font-weight: 600; padding: 12px 26px; border-radius: 10px; box-shadow: 0 4px 12px rgba(102, 126, 234, 0.35);">
              Visit AlumniGo Platform &rarr;
            </a>
          </td>
        </tr>
      </table>
    """

    html_body = render_email_wrapper(
        title=f"We've Received Your Message: {subject}",
        badge_text="Message Received",
        badge_bg="#6366f1",
        badge_color="#ffffff",
        content_html=content_html,
        preheader=preheader,
    )

    return email_subject, plain_body, html_body
