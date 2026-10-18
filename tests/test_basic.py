import os
import sys
import pytest

# Add the app/src directory to the path so we can import the modules
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../app/src')))

from algo import create_app

@pytest.fixture
def app():
    """Create and configure a new app instance for each test."""
    flask_app = create_app({
        "TESTING": True,
        "DEBUG": False,
        "SECRET_KEY": "test-secret"
    })
    yield flask_app

@pytest.fixture
def client(app):
    """A test client for the app."""
    return app.test_client()

def test_home_page(client):
    """Test that the home page loads correctly."""
    response = client.get('/')
    assert response.status_code == 200
    # Assuming there's a reference to ALGO or AlumniGo in the home page
    assert b"algo" in response.data.lower()
