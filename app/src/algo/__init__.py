import os
import sys
from flask import Flask
from flask_bcrypt import Bcrypt

# Add the parent directory to the path to allow for package imports
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))

# Import db and other components that will be initialized
from algo import db
# Import blueprints that will be registered
# from .blueprints import core, auth, profile, profile, admin, connections, chat, settings
from .blueprints import core, auth, profile, dashboard, connections, settings, chat, communities, channels

# Initialize extensions without an app
bcrypt = Bcrypt()

def create_app(test_config=None):
    """Create and configure an instance of the Flask application."""
    app = Flask(__name__, instance_relative_config=True,
                template_folder="../../templates",
                static_folder="../../static")

    # Load default configuration
    app.config.from_mapping(
        SECRET_KEY=os.environ.get("SECRET_KEY", "dev-secret"),
        # Other default configs
    )

    if test_config is None:
        # Load the instance config, if it exists, when not testing
        app.config.from_pyfile("config.py", silent=True)
    else:
        # Load the test config if passed in
        app.config.update(test_config)

    # Ensure the instance folder exists
    try:
        os.makedirs(app.instance_path)
    except OSError:
        pass

    # Initialize Flask extensions
    bcrypt.init_app(app)
    db.init_app(app)

    # Register blueprints
    app.register_blueprint(core.bp)
    app.register_blueprint(auth.bp)
    app.register_blueprint(profile.bp)
    app.register_blueprint(dashboard.bp)
    app.register_blueprint(connections.bp)
    app.register_blueprint(settings.bp)
    app.register_blueprint(chat.bp)
    app.register_blueprint(communities.bp)
    app.register_blueprint(channels.channels_bp, url_prefix="/api")
    app.register_blueprint(channels.channels_bp, name="channels_direct")

    from algo import utils
    @app.context_processor
    def inject_time_functions():
        return {
            "to_ist": utils.to_ist,
            "format_ist_time": utils.format_ist_time,
            "generate_default_avatar": utils.generate_default_avatar,
        }

    # A simple hello route to test app factory
    @app.route('/hello')
    def hello():
        return 'Hello, World!'

    @app.route('/forgot_password', methods=['GET', 'POST'])
    def forgot_password():
        from flask import render_template, request, flash
        if request.method == 'POST':
            flash("If that account exists, password reset instructions have been sent.", "info")
        return render_template("forgot_password.html")

    # Backward compatibility endpoint aliases
    app.add_url_rule('/create_community', endpoint='create_community', view_func=communities.create_community, methods=['GET', 'POST'])
    app.add_url_rule('/user_dashboard', endpoint='user_dashboard', view_func=dashboard.user_dashboard, methods=['GET'])
    app.add_url_rule('/admin_dashboard', endpoint='admin_dashboard', view_func=dashboard.admin_dashboard, methods=['GET'])
    app.add_url_rule('/verification_request', endpoint='verification_request', view_func=dashboard.verification_request, methods=['GET', 'POST'])

    return app
