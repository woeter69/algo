from flask import Blueprint, render_template, redirect, url_for, request, flash, session
import secrets
import datetime
from werkzeug.security import generate_password_hash, check_password_hash
from algo.db import get_db

bp = Blueprint('auth', __name__, url_prefix='/auth')

@bp.route('/register', methods=('GET', 'POST'))
def register():
    if 'user_id' in session:
        return redirect(url_for('dashboard.user_dashboard'))

    if request.method == 'POST':
        firstname = request.form['firstname']
        lastname = request.form['lastname']
        email = request.form['email']
        username = request.form['username']
        password = request.form['password']

        db = get_db()
        cur = db.cursor()
        error = None

        cur.execute('SELECT user_id FROM users WHERE email = %s', (email,))
        if cur.fetchone() is not None:
            error = f"Email {email} is already registered."
        
        cur.execute('SELECT user_id FROM users WHERE username = %s', (username,))
        if cur.fetchone() is not None:
            error = f"Username {username} is already taken."

        if error is None:
            hashed_password = generate_password_hash(password)
            
            session['register_data'] = {
                "firstname": firstname,
                "lastname": lastname,
                "email": email,
                "username": username,
                "password": hashed_password,
            }
            
            token = secrets.token_urlsafe(32)
            expiry = datetime.datetime.utcnow() + datetime.timedelta(minutes=15)

            cur.execute(
                "INSERT INTO verification_tokens (email, token, expiry) VALUES (%s, %s, %s)",
                (email, token, expiry)
            )
            db.commit()
            
            cur.close()
            return redirect(url_for('auth.check_email'))

        flash(error)
        cur.close()

    return render_template('register.html')

@bp.route('/login', methods=('GET', 'POST'))
def login():
    if 'user_id' in session:
        return redirect(url_for('dashboard.user_dashboard'))

    if request.method == 'POST':
        email_or_username = request.form.get('email') or request.form.get('username') or ''
        password = request.form.get('password', '')
        db = get_db()
        error = None
        
        cur = db.cursor()
        cur.execute(
            'SELECT user_id, password, login_count, username, pfp_path, role FROM users WHERE email=%s OR username=%s',
            (email_or_username, email_or_username)
        )
        user = cur.fetchone()
        cur.close()

        if user is None:
            error = 'Incorrect email or username.'
        elif not check_password_hash(user[1], password):
            error = 'Incorrect password.'

        if error is None:
            session.clear()
            session['user_id'] = user[0]
            session['username'] = user[3]
            session['pfp_path'] = user[4]
            session['role'] = user[5]
            
            cur = db.cursor()
            cur.execute(
                'UPDATE users SET last_login=%s, login_count=login_count+1 WHERE user_id=%s',
                (datetime.datetime.utcnow(), user[0])
            )
            db.commit()
            cur.close()

            next_url = request.args.get('next')
            if user[2] == 0:
                if next_url:
                    session['post_profile_redirect'] = next_url
                return redirect(url_for('profile.complete_profile'))
            
            if next_url:
                return redirect(next_url)
            return redirect(url_for('dashboard.user_dashboard'))

        flash(error)

    return render_template('login.html')

@bp.route('/logout')
def logout():
    session.clear()
    flash('You have been logged out.', 'success')
    return redirect(url_for('core.home'))

@bp.route('/check_email')
def check_email():
    return render_template('check_email.html')

@bp.route('/verify/<token>')
def verify(token):
    db = get_db()
    cur = db.cursor()

    cur.execute("SELECT * FROM verification_tokens where token=%s", (token,))
    row = cur.fetchone()

    if not row:
        cur.close()
        return render_template("token_invalid.html")

    id, email, db_token, expiry = row

    if datetime.datetime.utcnow() > expiry:
        cur.close()
        return render_template("token_expired.html")

    register_data = session.get("register_data")
    if not register_data or register_data['email'] != email:
        flash("Verification failed. Please try registering again.", "error")
        return redirect(url_for('auth.register'))

    try:
        cur.execute(
            """
            INSERT INTO users (firstname, lastname, email, username, password, verified)
            VALUES (%s, %s, %s, %s, %s, %s)
            """,
            (
                register_data["firstname"],
                register_data["lastname"],
                register_data["email"],
                register_data["username"],
                register_data["password"],
                True,
            ),
        )
        cur.execute("DELETE FROM verification_tokens WHERE token=%s", (token,))
        db.commit()
    except Exception as e:
        db.rollback()
        flash("An error occurred during final registration. It's possible the user already exists.", "error")
        return redirect(url_for('auth.register'))
    finally:
        cur.close()

    session.pop("register_data", None)
    return redirect(url_for('auth.confirmation'))

@bp.route('/confirmation')
def confirmation():
    return render_template('confirmation.html')
