from flask import Blueprint, render_template, request, flash, redirect, url_for

bp = Blueprint('core', __name__)

@bp.route('/')
def home():
    """Renders the main landing page."""
    return render_template('home.html')

@bp.route('/login', methods=['GET', 'POST'])
def login():
    """Forward to auth.login for both GET and POST."""
    from algo.blueprints.auth import login as auth_login
    return auth_login()

@bp.route('/register', methods=['GET', 'POST'])
def register():
    """Forward to auth.register for both GET and POST."""
    from algo.blueprints.auth import register as auth_register
    return auth_register()

@bp.route('/logout')
def logout():
    """Forward to auth.logout."""
    from algo.blueprints.auth import logout as auth_logout
    return auth_logout()

@bp.route('/about')
def about():
    """Renders the about us page."""
    return render_template('about.html')

@bp.route('/contact', methods=['GET', 'POST'])
def contact():
    """Render the contact page and handle form submission"""
    if request.method == 'POST':
        full_name = request.form.get('full_name', '').strip()
        email = request.form.get('email', '').strip()
        phone = request.form.get('phone', '').strip()
        subject = request.form.get('subject', '').strip()
        message = request.form.get('message', '').strip()

        if not full_name or not email or not message:
            flash('Please fill in all required fields.', 'error')
            return redirect(url_for('core.contact'))

        try:
            from algo.db import get_db
            from algo import utils
            db = get_db()
            cur = db.cursor()
            cur.execute("""
                INSERT INTO contacts (full_name, email, phone, subject, message, status)
                VALUES (%s, %s, %s, %s, %s, 'pending')
                RETURNING id;
            """, (full_name, email, phone, subject, message))
            inquiry_id_row = cur.fetchone()
            inquiry_id = inquiry_id_row[0] if inquiry_id_row else None
            db.commit()
            cur.close()

            # 1. Forward query to admin inbox (alumnigo.sih@gmail.com) asynchronously
            utils.send_contact_inquiry_email(full_name, email, phone, subject, message)

            # 2. Send instant confirmation acknowledgment to the querier
            utils.send_inquiry_confirmation_email(email, full_name, subject, message, inquiry_id)

            flash('Thank you for your message! A confirmation email has been sent to your address and our team will contact you soon.', 'success')
        except Exception as e:
            import logging
            logging.error(f"Error saving contact message: {e}")
            flash('An error occurred while saving your message. Please try again.', 'error')

        return redirect(url_for('core.contact'))
    return render_template('contact.html')

@bp.route("/thanks", methods=["GET", "POST"])
def thanks():
    if request.method == "POST":
        return redirect(url_for("core.home"))
    return render_template("thanks.html")

@bp.route("/recommendations")
def recommendations():
    return render_template("recommendations.html")
