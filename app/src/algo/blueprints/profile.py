from flask import (
    Blueprint, flash, redirect, render_template, request, session, url_for, current_app, jsonify
)
import os
import datetime
from algo.db import get_db
from algo.auth.decorators import login_required
from algo.utils import is_safe_url
from algo import utils
from algo import validators

bp = Blueprint('profile', __name__)

@bp.route('/profile')
@bp.route('/')
@login_required
def profile_redirect():
    username = session.get('username') or session.get('user_id')
    return redirect(url_for('profile.user_profile', username=username))

@bp.route('/profile/<username>')
@bp.route('/<username>')
@login_required
def user_profile(username):
    if username == 'profile':
        return profile_redirect()
    db = get_db()
    cur = db.cursor()
    user_id = session.get('user_id')

    # Support looking up by user_id if integer or by username
    if username.isdigit():
        cur.execute(
            """
            SELECT user_id, firstname, lastname, email, username, dob, graduation_year, 
                   university_name, department, college, current_city, pfp_path, 
                   registration_date, role, enrollment_number, community_id, last_login, login_count,
                   bio, linkedin, github, twitter, website, phone
            FROM users 
            WHERE user_id = %s
        """,
            (int(username),),
        )
    else:
        cur.execute(
            """
            SELECT user_id, firstname, lastname, email, username, dob, graduation_year, 
                   university_name, department, college, current_city, pfp_path, 
                   registration_date, role, enrollment_number, community_id, last_login, login_count,
                   bio, linkedin, github, twitter, website, phone
            FROM users 
            WHERE username = %s
        """,
            (username,),
        )

    user_data = cur.fetchone()
    if not user_data:
        flash('User not found')
        return redirect(url_for('core.home'))

    cur.execute(
        """
        SELECT i.name 
        FROM interests i
        JOIN user_interests ui ON i.interest_id = ui.interest_id
        WHERE ui.user_id = %s
    """,
        (user_data[0],),
    )

    user_interests = [row[0] for row in cur.fetchall()]

    cur.execute(
        """
        SELECT degree_type, university_name, college_name, major, graduation_year, gpa, detail_id
        FROM education_details 
        WHERE user_id = %s
        ORDER BY graduation_year DESC NULLS LAST
    """,
        (user_data[0],),
    )

    education_rows = cur.fetchall()
    education_data = education_rows[0] if education_rows else None

    cur.execute(
        """
        SELECT company_name, job_title, join_year, leave_year, exp_id
        FROM work_experience 
        WHERE user_id = %s
        ORDER BY join_year DESC NULLS LAST
    """,
        (user_data[0],),
    )

    work_experience = cur.fetchall()

    cur.execute(
        """
        SELECT COUNT(*) 
        FROM connections 
        WHERE (user_id = %s OR con_user_id = %s) AND status = 'accepted'
    """,
        (user_data[0], user_data[0]),
    )

    connections_count = cur.fetchone()[0]
    
    # Get community name if exists
    community_name = None
    if user_data[15]:
        cur.execute(
            "SELECT name FROM communities WHERE community_id = %s", (user_data[15],)
        )
        community_result = cur.fetchone()
        if community_result:
            community_name = community_result[0]

    # Get current logged-in user's info for the dropdown
    current_user_info = None
    if user_id != user_data[0]:  # If viewing someone else's profile
        cur.execute(
            """
            SELECT user_id, firstname, lastname, email, username, pfp_path, role
            FROM users 
            WHERE user_id = %s
        """,
            (user_id,),
        )
        current_user_info = cur.fetchone()

    cur.close()

    user_social_links = {
        'linkedin': user_data[19] if len(user_data) > 19 else None,
        'github': user_data[20] if len(user_data) > 20 else None,
        'twitter': user_data[21] if len(user_data) > 21 else None,
        'website': user_data[22] if len(user_data) > 22 else None,
    }

    return render_template(
        'profile.html',
        user_data=user_data,
        user_interests=user_interests,
        education_data=education_data,
        education_list=education_rows,
        work_experience=work_experience,
        connections_count=connections_count,
        community_name=community_name,
        current_user_info=current_user_info,
        user_bio=user_data[18],
        user_skills=user_interests,
        user_social_links=user_social_links,
        user_phone=user_data[23] if len(user_data) > 23 else None,
    )

def normalize_degree_type(degree):
    if not degree:
        return "Bachelors"
    cleaned = degree.lower().replace(".", "").strip()
    valid_map = {
        "btech": "B Tech",
        "b tech": "B Tech",
        "mtech": "M Tech",
        "m tech": "M Tech",
        "be": "B.E.",
        "me": "M.E.",
        "bsc": "B.Sc.",
        "msc": "M.Sc.",
        "bca": "BCA",
        "mca": "MCA",
        "mba": "MBA",
        "bba": "BBA",
        "phd": "PHD",
        "doctorate": "Doctorate",
        "diploma": "Diploma",
        "masters": "Masters",
        "bachelors": "Bachelors",
    }
    for key, val in valid_map.items():
        if key in cleaned:
            return val
    return "Bachelors"


@bp.route("/api/profile/update", methods=["POST"])
@login_required
def update_profile():
    """Update profile information, bio, social links, education, and avatar with local fallback."""
    user_id = session.get("user_id")
    if not user_id:
        return jsonify({"error": "Unauthorized"}), 401

    db = get_db()
    cur = db.cursor()
    try:
        first_name = request.form.get("firstName")
        last_name = request.form.get("lastName")
        phone = request.form.get("phone")
        current_city = request.form.get("currentCity")
        bio = request.form.get("bio")
        university_name = request.form.get("universityName")
        grad_year = request.form.get("graduationYear", type=int)
        degree = request.form.get("degree")
        major = request.form.get("major")
        gpa = request.form.get("gpa")
        linkedin = request.form.get("linkedIn")
        github = request.form.get("github")
        twitter = request.form.get("twitter")
        website = request.form.get("website")

        # Profile Picture Upload
        pfp_url = None
        if "profilePicture" in request.files:
            file = request.files["profilePicture"]
            if file and file.filename:
                import time
                ext = file.filename.rsplit(".", 1)[-1].lower() if "." in file.filename else "jpg"
                if ext in ["jpg", "jpeg", "png", "gif", "webp"]:
                    filename = f"avatar_{user_id}_{int(time.time())}.{ext}"
                    filepath = os.path.join(current_app.root_path, "static", "uploads", "avatars", filename)
                    os.makedirs(os.path.dirname(filepath), exist_ok=True)
                    file.save(filepath)
                    pfp_url = f"/static/uploads/avatars/{filename}"
                    session["pfp_path"] = pfp_url

        # Build dynamic user update query
        update_fields = []
        update_values = []
        if first_name is not None:
            update_fields.append("firstname = %s")
            update_values.append(first_name.strip())
        if last_name is not None:
            update_fields.append("lastname = %s")
            update_values.append(last_name.strip())
        if phone is not None:
            update_fields.append("phone = %s")
            update_values.append(phone.strip())
        if current_city is not None:
            update_fields.append("current_city = %s")
            update_values.append(current_city.strip())
        if bio is not None:
            update_fields.append("bio = %s")
            update_values.append(bio.strip())
        if university_name is not None:
            update_fields.append("university_name = %s")
            update_values.append(university_name.strip())
        if grad_year is not None:
            update_fields.append("graduation_year = %s")
            update_values.append(grad_year)
        if major is not None:
            update_fields.append("department = %s")
            update_values.append(major.strip())
        if linkedin is not None:
            update_fields.append("linkedin = %s")
            update_values.append(linkedin.strip())
        if github is not None:
            update_fields.append("github = %s")
            update_values.append(github.strip())
        if twitter is not None:
            update_fields.append("twitter = %s")
            update_values.append(twitter.strip())
        if website is not None:
            update_fields.append("website = %s")
            update_values.append(website.strip())
        if pfp_url is not None:
            update_fields.append("pfp_path = %s")
            update_values.append(pfp_url)

        if update_fields:
            update_values.append(user_id)
            query = f"UPDATE users SET {', '.join(update_fields)} WHERE user_id = %s"
            cur.execute(query, tuple(update_values))

        # Update education details
        if university_name or degree or major or grad_year:
            degree = normalize_degree_type(degree)
            cur.execute("SELECT detail_id FROM education_details WHERE user_id = %s", (user_id,))
            ed_row = cur.fetchone()
            parsed_gpa = float(gpa) if gpa and gpa.replace('.', '', 1).isdigit() else None
            if ed_row:
                cur.execute("""
                    UPDATE education_details 
                    SET university_name = COALESCE(%s, university_name),
                        degree_type = COALESCE(%s, degree_type),
                        major = COALESCE(%s, major),
                        graduation_year = COALESCE(%s, graduation_year),
                        gpa = COALESCE(%s, gpa)
                    WHERE detail_id = %s
                """, (university_name, degree, major, grad_year, parsed_gpa, ed_row[0]))
            else:
                cur.execute("""
                    INSERT INTO education_details (user_id, university_name, degree_type, major, graduation_year, gpa)
                    VALUES (%s, %s, %s, %s, %s, %s)
                """, (user_id, university_name or "", degree or "Bachelors", major or "", grad_year, parsed_gpa))

        # Update interests
        interests_str = request.form.get("interests")
        if interests_str:
            interest_names = [i.strip() for i in interests_str.split(",") if i.strip()]
            for iname in interest_names:
                cur.execute("INSERT INTO interests (name) VALUES (%s) ON CONFLICT (name) DO NOTHING", (iname,))
                cur.execute("SELECT interest_id FROM interests WHERE name = %s", (iname,))
                irow = cur.fetchone()
                if irow:
                    cur.execute("INSERT INTO user_interests (user_id, interest_id) VALUES (%s, %s) ON CONFLICT DO NOTHING", (user_id, irow[0]))

        db.commit()
        return jsonify({
            "success": True,
            "message": "Profile updated successfully!",
            "pfp_path": pfp_url,
        })
    except Exception as e:
        db.rollback()
        import logging
        logging.error(f"Error updating profile: {e}")
        return jsonify({"error": "Failed to update profile"}), 500
    finally:
        cur.close()

@bp.route("/api/profile/experience", methods=["POST"])
@login_required
def add_experience():
    """Add a work experience entry"""
    user_id = session.get("user_id")
    data = request.get_json(silent=True) or request.form
    company_name = data.get("company_name", "").strip()
    job_title = data.get("job_title", "").strip()
    join_year = data.get("join_year")
    leave_year = data.get("leave_year")

    if not company_name or not job_title:
        return jsonify({"error": "Company and title are required"}), 400

    db = get_db()
    cur = db.cursor()
    try:
        j_yr = int(join_year) if join_year and str(join_year).isdigit() else None
        l_yr = int(leave_year) if leave_year and str(leave_year).isdigit() else None
        cur.execute("""
            INSERT INTO work_experience (user_id, company_name, job_title, join_year, leave_year)
            VALUES (%s, %s, %s, %s, %s)
            RETURNING exp_id
        """, (user_id, company_name, job_title, j_yr, l_yr))
        exp_id = cur.fetchone()[0]
        db.commit()
        return jsonify({"success": True, "exp_id": exp_id, "message": "Experience added!"}), 201
    except Exception as e:
        db.rollback()
        return jsonify({"error": "Failed to add experience"}), 500
    finally:
        cur.close()

@bp.route("/api/profile/experience/<int:exp_id>", methods=["DELETE"])
@login_required
def delete_experience(exp_id):
    """Delete a work experience entry"""
    user_id = session.get("user_id")
    db = get_db()
    cur = db.cursor()
    try:
        cur.execute("DELETE FROM work_experience WHERE exp_id = %s AND user_id = %s", (exp_id, user_id))
        db.commit()
        return jsonify({"success": True, "message": "Experience deleted!"})
    except Exception as e:
        db.rollback()
        return jsonify({"error": "Failed to delete experience"}), 500
    finally:
        cur.close()

@bp.route("/api/profile/education", methods=["POST"])
@login_required
def add_education():
    """Add an education entry"""
    user_id = session.get("user_id")
    data = request.get_json(silent=True) or request.form
    degree_type = data.get("degree_type", "").strip()
    university_name = data.get("university_name", "").strip()
    college_name = data.get("college_name", "").strip()
    major = data.get("major", "").strip()
    graduation_year = data.get("graduation_year")
    gpa = data.get("gpa")

    if not university_name and not college_name:
        return jsonify({"error": "Institution name is required"}), 400

    db = get_db()
    cur = db.cursor()
    try:
        degree_type = normalize_degree_type(degree_type)
        g_yr = int(graduation_year) if graduation_year and str(graduation_year).isdigit() else None
        parsed_gpa = float(gpa) if gpa and str(gpa).replace('.', '', 1).isdigit() else None
        cur.execute("""
            INSERT INTO education_details (user_id, degree_type, university_name, college_name, major, graduation_year, gpa)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            RETURNING detail_id
        """, (user_id, degree_type, university_name, college_name, major, g_yr, parsed_gpa))
        detail_id = cur.fetchone()[0]
        db.commit()
        return jsonify({"success": True, "detail_id": detail_id, "message": "Education added!"}), 201
    except Exception as e:
        db.rollback()
        return jsonify({"error": "Failed to add education"}), 500
    finally:
        cur.close()

@bp.route("/api/profile/education/<int:detail_id>", methods=["DELETE"])
@login_required
def delete_education(detail_id):
    """Delete an education entry"""
    user_id = session.get("user_id")
    db = get_db()
    cur = db.cursor()
    try:
        cur.execute("DELETE FROM education_details WHERE detail_id = %s AND user_id = %s", (detail_id, user_id))
        db.commit()
        return jsonify({"success": True, "message": "Education deleted!"})
    except Exception as e:
        db.rollback()
        return jsonify({"error": "Failed to delete education"}), 500
    finally:
        cur.close()


@bp.route("/complete_profile", methods=["GET", "POST"])
@login_required
def complete_profile():
    db = get_db()
    cur = db.cursor()
    try:
        cutoff_date = (
            datetime.datetime.utcnow()
            .date()
            .replace(year=datetime.datetime.utcnow().year - 16)
        )
        cities = utils.load_cities()
        if request.method == "POST":
            # Personal Information
            first_name = request.form.get("first_name", "")
            last_name = request.form.get("last_name", "")
            dob = request.form["dob"]
            bio = request.form.get("bio", "")

            # Role Information
            role = request.form.get("role")
            student_id = request.form.get("student_id", "")
            alumni_id = request.form.get("alumni_id", "")
            employee_id = request.form.get("employee_id", "")
            department_role = request.form.get("department_role", "")

            # Education Information
            uni_name = request.form["uni_name"]
            clg_name = request.form.get("clg_name", "")
            degree = request.form.get("degree", "")
            major = request.form.get("major", "")
            grad_year = request.form["grad_year"]
            gpa = request.form.get("gpa", "")

            # Location
            city = request.form["city"]

            # Work Experience
            company = request.form.get("company", "")
            job_title = request.form.get("job_title", "")
            join_year = request.form.get("join_year", "")
            leave_year = request.form.get("leave_year", "")

            # Skills and Interests
            skills = request.form.get("skills", "")
            interests = request.form.get("interests", "")

            # Social Links
            linkedin = request.form.get("linkedin", "")
            github = request.form.get("github", "")
            twitter = request.form.get("twitter", "")
            website = request.form.get("website", "")

            # Privacy Settings
            profile_visibility = request.form.get("profile_visibility", "")
            email_notifications = request.form.get("email_notifications", "")
            job_alerts = request.form.get("job_alerts", "")

            # Handle profile picture upload
            pfp_file = request.files.get("pfp")
            pfp_url = None
            if pfp_file and validators.allowed_file(pfp_file.filename):
                pfp_url = utils.upload_to_imgbb(pfp_file, os.getenv("PFP_API"))

            # Validate age
            dob_date = datetime.datetime.strptime(dob, "%Y-%m-%d").date()
            today = datetime.date.today()
            age = (
                today.year
                - dob_date.year
                - ((today.month, today.day) < (dob_date.month, dob_date.day))
            )

            if age < 16:
                return render_template(
                    "complete_profile.html",
                    cities=cities,
                    error="You must be atleast 16",
                )

            user_id = session.get("user_id")

            # Validate role selection
            if not role or role not in ["student", "alumni", "staff"]:
                return render_template(
                    "complete_profile.html",
                    cities=cities,
                    cutoff_date=cutoff_date,
                    error="Please select a valid role",
                )

            # Validate role-specific fields
            if role == "student" and not student_id:
                return render_template(
                    "complete_profile.html",
                    cities=cities,
                    cutoff_date=cutoff_date,
                    error="Student ID is required for student role",
                )
            elif role == "staff" and (not employee_id or not department_role):
                return render_template(
                    "complete_profile.html",
                    cities=cities,
                    cutoff_date=cutoff_date,
                    error="Employee ID and Department/Position are required for staff role",
                )

            # Update users table with basic info and role
            cur.execute(
                """
                UPDATE users SET 
                    firstname=%s, lastname=%s, dob=%s,
                    university_name=%s, college=%s, graduation_year=%s, current_city=%s, 
                    pfp_path=%s, role=%s, verification_status=%s
                WHERE user_id=%s
            """,
                (
                    first_name,
                    last_name,
                    dob,
                    uni_name,
                    clg_name,
                    grad_year,
                    city,
                    pfp_url,
                    role,
                    "verified" if role == "admin" else "pending",
                    user_id,
                ),
            )

            # Insert education details if provided
            if degree or major or gpa:
                # Check if education record exists
                cur.execute(
                    "SELECT COUNT(*) FROM education_details WHERE user_id = %s",
                    (user_id,),
                )
                if cur.fetchone()[0] > 0:
                    # Update existing record
                    cur.execute(
                        """
                        UPDATE education_details SET 
                            degree_type=%s, major=%s, university_name=%s, college_name=%s, 
                            graduation_year=%s, gpa=%s
                        WHERE user_id=%s
                    """,
                        (
                            degree,
                            major,
                            uni_name,
                            clg_name,
                            grad_year,
                            float(gpa) if gpa else None,
                            user_id,
                        ),
                    )
                else:
                    # Insert new record
                    cur.execute(
                        """
                        INSERT INTO education_details 
                        (user_id, degree_type, major, university_name, college_name, graduation_year, gpa)
                        VALUES (%s, %s, %s, %s, %s, %s, %s)
                    """,
                        (
                            user_id,
                            degree,
                            major,
                            uni_name,
                            clg_name,
                            grad_year,
                            float(gpa) if gpa else None,
                        ),
                    )

            # Insert work experience if provided (and not "None" or empty)
            if company and company.lower() not in ["none", "n/a", ""] and job_title:
                # Check if work experience record exists
                cur.execute(
                    "SELECT COUNT(*) FROM work_experience WHERE user_id = %s",
                    (user_id,),
                )
                if cur.fetchone()[0] > 0:
                    # Update existing record
                    cur.execute(
                        """
                        UPDATE work_experience SET 
                            company_name=%s, job_title=%s, join_year=%s, leave_year=%s
                        WHERE user_id=%s
                    """,
                        (
                            company,
                            job_title,
                            int(join_year) if join_year else None,
                            int(leave_year) if leave_year else None,
                            user_id,
                        ),
                    )
                else:
                    # Insert new record
                    cur.execute(
                        """
                        INSERT INTO work_experience 
                        (user_id, company_name, job_title, join_year, leave_year)
                        VALUES (%s, %s, %s, %s, %s)
                    """,
                        (
                            user_id,
                            company,
                            job_title,
                            int(join_year) if join_year else None,
                            int(leave_year) if leave_year else None,
                        ),
                    )

            # Update session with new profile picture and name
            if pfp_url:
                session["pfp_path"] = pfp_url
            if first_name and last_name:
                session["username"] = f"{first_name} {last_name}"

            # Update session with role information
            session["role"] = role
            session["verification_status"] = (
                "verified" if role == "admin" else "pending"
            )

            db.commit()

            # Redirect to interests page to complete profile setup
            return redirect(url_for("profile.interests"))
        return render_template(
            "complete_profile.html", cities=cities, cutoff_date=cutoff_date
        )
    except Exception as e:
        current_app.logger.error(f"Error during complete_profile: {str(e)}")
        flash("An unexpected error occurred. Please try again.")
        return render_template("complete_profile.html"), 500
    finally:
        cur.close()

@bp.route("/interests", methods=["GET", "POST"])
@login_required
def interests():
    db = get_db()
    cur = db.cursor()
    try:
        cur.execute("SELECT interest_id, name FROM interests ORDER BY name")
        db_interests = cur.fetchall()

        if request.method == "POST":
            selected_interests = request.form.getlist("interests")
            user_id = session["user_id"]

            cur.execute("DELETE FROM user_interests where user_id=%s", (user_id,))

            for interest_id in selected_interests:
                cur.execute(
                    "INSERT INTO user_interests (user_id,interest_id) VALUES (%s,%s) ",
                    (user_id, interest_id),
                )
            db.commit()

            # Check for saved redirect URL after profile completion
            post_profile_redirect = session.pop("post_profile_redirect", None)
            if post_profile_redirect and is_safe_url(post_profile_redirect):
                return redirect(post_profile_redirect)
            else:
                return redirect(url_for("dashboard.user_dashboard"))
        return render_template("interests.html", db_interests=db_interests)

    except Exception as e:
        current_app.logger.error(f"Error updating interests: {str(e)}")
        flash("There was an error saving your interests.")
        return redirect(url_for("core.home"))
    finally:
        cur.close()
