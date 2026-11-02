// @ts-nocheck
// Mobile Navigation
const hamburger = document.querySelector(".hamburger");
const navLinks = document.querySelector(".nav-links");
const links = document.querySelectorAll(".nav-links li");

hamburger.addEventListener("click", () => {
  // Toggle Nav
  navLinks.classList.toggle("open");

  // Animate Links
  links.forEach((link, index) => {
    if (link.style.animation) {
      link.style.animation = "";
    } else {
      link.style.animation = `navLinkFade 0.5s ease forwards ${index / 7 + 0.3}s`;
    }
  });

  // Hamburger Animation
  hamburger.classList.toggle("toggle");
});

// Close mobile menu when clicking on a nav link
links.forEach((link) => {
  link.addEventListener("click", () => {
    navLinks.classList.remove("open");
    hamburger.classList.remove("toggle");
    links.forEach((link) => {
      link.style.animation = "";
    });
  });
});

// Scroll Header Effect
window.addEventListener("scroll", () => {
  const header = document.querySelector("header");
  header.classList.toggle("scrolled", window.scrollY > 50);
});

// Profile Picture Upload Functionality
const profilePictureInput = document.getElementById("profile-picture");
const profilePreview = document.getElementById("profile-preview");

if (profilePictureInput && profilePreview) {
  // Click on preview to trigger file input
  profilePreview.addEventListener("click", () => {
    profilePictureInput.click();
  });

  // Handle file selection and preview
  profilePictureInput.addEventListener("change", function (e) {
    const file = e.target.files[0];

    if (file) {
      // Validate file type
      const validTypes = ["image/jpeg", "image/jpg", "image/png", "image/gif"];
      if (!validTypes.includes(file.type)) {
        alert("Please select a valid image file (JPG, PNG, or GIF)");
        return;
      }

      // Validate file size (5MB max)
      const maxSize = 5 * 1024 * 1024; // 5MB in bytes
      if (file.size > maxSize) {
        alert("File size must be less than 5MB");
        return;
      }

      // Create preview
      const reader = new FileReader();
      reader.onload = function (e) {
        profilePreview.innerHTML = `<img src="${e.target.result}" alt="Profile Preview">`;
        profilePreview.classList.add("has-image");
      };
      reader.readAsDataURL(file);
    }
  });

  // Allow drag and drop
  profilePreview.addEventListener("dragover", (e) => {
    e.preventDefault();
    profilePreview.style.borderColor = "var(--primary)";
    profilePreview.style.backgroundColor = "rgba(109, 40, 217, 0.1)";
  });

  profilePreview.addEventListener("dragleave", (e) => {
    e.preventDefault();
    profilePreview.style.borderColor = "#e2e8f0";
    profilePreview.style.backgroundColor = "#f8fafc";
  });

  profilePreview.addEventListener("drop", (e) => {
    e.preventDefault();
    profilePreview.style.borderColor = "#e2e8f0";
    profilePreview.style.backgroundColor = "#f8fafc";

    const files = e.dataTransfer.files;
    if (files.length > 0) {
      profilePictureInput.files = files;
      profilePictureInput.dispatchEvent(new Event("change"));
    }
  });
}

// Form Validation
const form = document.querySelector(".profile-form");
const requiredFields = document.querySelectorAll(
  "input[required], select[required]",
);

// Add validation on form submit
if (form) {
  form.addEventListener("submit", function (e) {
    let isValid = true;

    // Validate required fields
    requiredFields.forEach((field) => {
      if (!field.value.trim()) {
        showFieldError(field, "This field is required");
        isValid = false;
      } else {
        clearFieldError(field);
      }
    });

    // Validate email format for social links
    const urlFields = ["linkedin", "github", "website"];
    urlFields.forEach((fieldName) => {
      const field = document.getElementById(fieldName);
      if (field && field.value && !isValidUrl(field.value)) {
        showFieldError(field, "Please enter a valid URL");
        isValid = false;
      }
    });

    // Validate phone number format
    const phoneField = document.getElementById("phone");
    if (phoneField && phoneField.value && !isValidPhone(phoneField.value)) {
      showFieldError(phoneField, "Please enter a valid phone number");
      isValid = false;
    }

    // Validate GPA range
    const gpaField = document.getElementById("gpa");
    if (gpaField && gpaField.value) {
      const gpa = parseFloat(gpaField.value);
      if (gpa < 0 || gpa > 4) {
        showFieldError(gpaField, "GPA must be between 0.0 and 4.0");
        isValid = false;
      }
    }

    if (!isValid) {
      e.preventDefault();
      // Scroll to first error
      const firstError = document.querySelector(".field-error");
      if (firstError) {
        firstError.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }
  });
}

// Real-time validation
requiredFields.forEach((field) => {
  field.addEventListener("blur", function () {
    if (!this.value.trim()) {
      showFieldError(this, "This field is required");
    } else {
      clearFieldError(this);
    }
  });

  field.addEventListener("input", function () {
    if (this.classList.contains("error")) {
      clearFieldError(this);
    }
  });
});

// Helper functions
function showFieldError(field, message) {
  clearFieldError(field);
  field.classList.add("error");

  const errorDiv = document.createElement("div");
  errorDiv.className = "field-error";
  errorDiv.textContent = message;

  field.parentNode.appendChild(errorDiv);
}

function clearFieldError(field) {
  field.classList.remove("error");
  const existingError = field.parentNode.querySelector(".field-error");
  if (existingError) {
    existingError.remove();
  }
}

function isValidUrl(string) {
  try {
    new URL(string);
    return true;
  } catch (_) {
    return false;
  }
}

function isValidPhone(phone) {
  const phoneRegex = /^[\+]?[1-9][\d]{0,15}$/;
  return phoneRegex.test(phone.replace(/[\s\-\(\)]/g, ""));
}

// Add error styles
const style = document.createElement("style");
style.textContent = `
    .form-group input.error,
    .form-group select.error,
    .form-group textarea.error {
        border-color: #ef4444;
        box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
    }
    
    .field-error {
        color: #ef4444;
        font-size: 0.85rem;
        margin-top: 0.25rem;
        display: block;
    }
`;
document.head.appendChild(style);

// Chatbase Integration

