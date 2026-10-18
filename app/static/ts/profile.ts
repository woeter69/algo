// @ts-nocheck
// Profile JavaScript functionality
document.addEventListener("DOMContentLoaded", function () {
  // Mobile navigation
  const hamburger = document.querySelector(".hamburger");
  const navLinks = document.querySelector(".nav-links");

  if (hamburger && navLinks) {
    hamburger.addEventListener("click", function () {
      hamburger.classList.toggle("toggle");
      navLinks.classList.toggle("open");
    });
  }

  // Modal functionality
  const editBtn = document.getElementById(
    "editProfileBtn",
  ) as HTMLButtonElement | null;
  const modal = document.getElementById("editModal") as HTMLElement | null;
  const closeBtn = document.getElementById(
    "closeModal",
  ) as HTMLButtonElement | null;
  const cancelBtn = document.getElementById(
    "cancelBtn",
  ) as HTMLButtonElement | null;
  const profileForm = document.getElementById(
    "profileForm",
  ) as HTMLFormElement | null;

  // Open modal
  if (editBtn && modal) {
    editBtn.addEventListener("click", function () {
      modal.classList.add("active");
      document.body.style.overflow = "hidden";
    });
  }

  // Close modal function
  function closeModal() {
    if (modal) {
      modal.classList.remove("active");
      document.body.style.overflow = "auto";
    }
  }

  // Close modal events
  if (closeBtn) {
    closeBtn.addEventListener("click", closeModal);
  }

  if (cancelBtn) {
    cancelBtn.addEventListener("click", closeModal);
  }

  // Close modal when clicking outside
  if (modal) {
    modal.addEventListener("click", function (e) {
      if (e.target === modal) {
        closeModal();
      }
    });
  }

  // Profile picture upload
  const profilePictureUpload = document.getElementById("profilePictureUpload");
  const profilePicture = document.getElementById("profilePicture");
  const avatarPlaceholder = document.getElementById("avatarPlaceholder");

  if (profilePictureUpload && profilePicture) {
    profilePictureUpload.addEventListener("change", function (e) {
      const file = e.target.files[0];
      if (file) {
        // Validate file size (5MB max)
        if (file.size > 5 * 1024 * 1024) {
          showNotification("File size must be less than 5MB", "error");
          return;
        }

        // Validate file type
        if (!file.type.startsWith("image/")) {
          showNotification("Please select a valid image file", "error");
          return;
        }

        const reader = new FileReader();
        reader.onload = function (e) {
          const profileAvatarContainer = document.getElementById(
            "profileAvatarContainer",
          );
          profilePicture.src = e.target.result;
          profilePicture.style.display = "block";
          if (avatarPlaceholder) {
            avatarPlaceholder.style.display = "none";
          }
          if (profileAvatarContainer) {
            profileAvatarContainer.classList.remove("has-placeholder");
          }
        };
        reader.readAsDataURL(file);
      }
    });
  }

  // Initialize profile picture state - placeholder is shown by default in HTML
  console.log("Profile picture placeholder should be visible by default");

  // Tab functionality
  const tabBtns = document.querySelectorAll(".tab-btn");
  const tabContents = document.querySelectorAll(".tab-content");

  tabBtns.forEach((btn) => {
    btn.addEventListener("click", function () {
      const targetTab = this.getAttribute("data-tab");

      // Remove active class from all tabs and contents
      tabBtns.forEach((b) => b.classList.remove("active"));
      tabContents.forEach((c) => c.classList.remove("active"));

      // Add active class to clicked tab and corresponding content
      this.classList.add("active");
      document.getElementById(targetTab + "-tab").classList.add("active");
    });
  });

  // Form submission
  if (profileForm) {
    profileForm.addEventListener("submit", function (e) {
      e.preventDefault();

      // Validate form
      if (!validateForm()) {
        showNotification("Please fix the errors before submitting", "error");
        return;
      }

      // Send form data to backend
      const formData = new FormData(profileForm);
      const saveBtn = profileForm.querySelector('button[type="submit"]') as HTMLButtonElement | null;
      if (saveBtn) {
        saveBtn.disabled = true;
        saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Saving...';
      }

      fetch("/api/profile/update", {
        method: "POST",
        body: formData,
      })
        .then((res) => res.json())
        .then((result) => {
          if (result.success) {
            showNotification(result.message || "Profile updated successfully!", "success");
            closeModal();
            setTimeout(() => {
              window.location.reload();
            }, 700);
          } else {
            showNotification(result.error || "Failed to update profile", "error");
          }
        })
        .catch((err) => {
          console.error("Profile update error:", err);
          showNotification("An error occurred while updating profile", "error");
        })
        .finally(() => {
          if (saveBtn) {
            saveBtn.disabled = false;
            saveBtn.innerHTML = '<i class="fas fa-save"></i> Save Changes';
          }
        });
    });
  }

  // Update profile display function
  function updateProfileDisplay(data) {
    // Update name
    const displayFirstName = document.getElementById("displayFirstName");
    const displayLastName = document.getElementById("displayLastName");
    if (displayFirstName) displayFirstName.textContent = data.firstName;
    if (displayLastName) displayLastName.textContent = data.lastName;

    // Update username
    const displayUsername = document.getElementById("displayUsername");
    const displayUsernameCard = document.getElementById("displayUsernameCard");
    if (displayUsername) displayUsername.textContent = data.username;
    if (displayUsernameCard)
      displayUsernameCard.textContent = "@" + data.username;

    // Update email
    const displayEmail = document.getElementById("displayEmail");
    if (displayEmail) displayEmail.textContent = data.email;

    // Update phone
    const displayPhone = document.getElementById("displayPhone");
    if (displayPhone) displayPhone.textContent = data.phone || "Not provided";

    // Update current city
    const displayCurrentCity = document.getElementById("displayCurrentCity");
    const displayCurrentCityCard = document.getElementById(
      "displayCurrentCityCard",
    );
    if (displayCurrentCity) displayCurrentCity.textContent = data.currentCity;
    if (displayCurrentCityCard)
      displayCurrentCityCard.textContent = data.currentCity;

    // Update university
    const displayUniversityName = document.getElementById(
      "displayUniversityName",
    );
    const displayUniversityNameCard = document.getElementById(
      "displayUniversityNameCard",
    );
    if (displayUniversityName)
      displayUniversityName.textContent = data.universityName;
    if (displayUniversityNameCard)
      displayUniversityNameCard.textContent = data.universityName;

    // Update graduation year
    const displayGraduationYear = document.getElementById(
      "displayGraduationYear",
    );
    if (displayGraduationYear)
      displayGraduationYear.textContent =
        data.graduationYear || "Not specified";

    // Update bio
    const displayBio = document.getElementById("displayBio");
    if (displayBio) displayBio.textContent = data.bio || "No bio provided yet.";

    // Update interests
    const interestsContainer = document.getElementById("interestsContainer");
    if (interestsContainer && data.interests) {
      const interestsArray = data.interests
        .split(",")
        .map((interest) => interest.trim())
        .filter((interest) => interest);
      interestsContainer.innerHTML = "";

      interestsArray.forEach((interest) => {
        const tag = document.createElement("span");
        tag.className = "interest-tag";
        tag.textContent = interest;
        interestsContainer.appendChild(tag);
      });
    }

    // Update skills
    const skillsContainer = document.getElementById("skillsContainer");
    if (skillsContainer && data.skills) {
      const skillsArray = data.skills
        .split(",")
        .map((skill) => skill.trim())
        .filter((skill) => skill);
      skillsContainer.innerHTML = "";

      skillsArray.forEach((skill) => {
        const tag = document.createElement("span");
        tag.className = "skill-tag";
        tag.textContent = skill;
        skillsContainer.appendChild(tag);
      });
    }

    // Update social links
    const displayLinkedIn = document.getElementById("displayLinkedIn");
    const displayGitHub = document.getElementById("displayGitHub");
    const displayTwitter = document.getElementById("displayTwitter");
    const displayWebsite = document.getElementById("displayWebsite");

    if (displayLinkedIn) {
      displayLinkedIn.textContent = data.linkedIn
        ? data.linkedIn.replace("https://", "").replace("http://", "")
        : "Not provided";
    }
    if (displayGitHub) {
      displayGitHub.textContent = data.github
        ? data.github.replace("https://", "").replace("http://", "")
        : "Not provided";
    }
    if (displayTwitter) {
      displayTwitter.textContent = data.twitter || "Not provided";
    }
    if (displayWebsite) {
      displayWebsite.textContent = data.website
        ? data.website.replace("https://", "").replace("http://", "")
        : "Not provided";
    }
  }

  // Notification function
  function showNotification(message, type = "info") {
    // Create notification element
    const notification = document.createElement("div");
    notification.className = `notification notification-${type}`;

    let icon = "fas fa-info-circle";
    let bgColor = "#6D28D9";

    if (type === "success") {
      icon = "fas fa-check-circle";
      bgColor = "#10b981";
    } else if (type === "error") {
      icon = "fas fa-exclamation-circle";
      bgColor = "#ef4444";
    }

    notification.innerHTML = `
            <div class="notification-content">
                <i class="${icon}"></i>
                <span>${message}</span>
            </div>
        `;

    // Add styles
    notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${bgColor};
            color: white;
            padding: 1rem 1.5rem;
            border-radius: 8px;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
            z-index: 3000;
            transform: translateX(100%);
            transition: all 0.3s ease;
            max-width: 400px;
            word-wrap: break-word;
        `;

    // Add to body
    document.body.appendChild(notification);

    // Animate in
    setTimeout(() => {
      notification.style.transform = "translateX(0)";
    }, 100);

    // Remove after 4 seconds (longer for error messages)
    const duration = type === "error" ? 5000 : 3000;
    setTimeout(() => {
      notification.style.transform = "translateX(100%)";
      setTimeout(() => {
        if (notification.parentNode) {
          notification.parentNode.removeChild(notification);
        }
      }, 300);
    }, duration);
  }

  // Profile avatar click to open file upload
  const profileAvatarContainer = document.getElementById(
    "profileAvatarContainer",
  );
  if (profileAvatarContainer && profilePictureUpload) {
    profileAvatarContainer.addEventListener("click", function () {
      profilePictureUpload.click();
    });
  }

  // Header scroll effect
  const header = document.querySelector("header");
  if (header) {
    window.addEventListener("scroll", function () {
      if (window.scrollY > 100) {
        header.classList.add("scrolled");
      } else {
        header.classList.remove("scrolled");
      }
    });
  }

  // Smooth scrolling for navigation links
  const navLinksElements = document.querySelectorAll(
    ".nav-links a[data-section]",
  );
  navLinksElements.forEach((link) => {
    link.addEventListener("click", function (e) {
      e.preventDefault();
      const section = this.getAttribute("data-section");

      // You can add navigation logic here
      console.log("Navigate to:", section);

      // Close mobile menu if open
      if (navLinks && navLinks.classList.contains("open")) {
        hamburger.classList.remove("toggle");
        navLinks.classList.remove("open");
      }
    });
  });

  // Form validation
  const inputs = document.querySelectorAll("#profileForm input[required]");
  inputs.forEach((input) => {
    input.addEventListener("blur", function () {
      validateField(this);
    });

    input.addEventListener("input", function () {
      if (this.classList.contains("error")) {
        validateField(this);
      }
    });
  });

  function validateField(field) {
    const value = field.value.trim();
    const fieldName = field.getAttribute("name");
    let isValid = true;
    let errorMessage = "";

    // Remove existing error
    field.classList.remove("error");
    const existingError = field.parentNode.querySelector(".error-message");
    if (existingError) {
      existingError.remove();
    }

    // Skip validation for optional fields that are empty
    if (!value && !field.hasAttribute("required")) {
      return true;
    }

    // Validation rules
    if (!value && field.hasAttribute("required")) {
      isValid = false;
      errorMessage = "This field is required";
    } else if (fieldName === "email") {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(value)) {
        isValid = false;
        errorMessage = "Please enter a valid email address";
      }
    } else if (fieldName === "username") {
      const usernameRegex = /^[a-zA-Z0-9_]{3,20}$/;
      if (!usernameRegex.test(value)) {
        isValid = false;
        errorMessage =
          "Username must be 3-20 characters, letters, numbers, and underscores only";
      }
    } else if (fieldName === "phone" && value) {
      const phoneRegex = /^[\+]?[1-9][\d]{0,15}$/;
      if (!phoneRegex.test(value.replace(/[\s\-\(\)]/g, ""))) {
        isValid = false;
        errorMessage = "Please enter a valid phone number";
      }
    } else if (
      (fieldName === "linkedIn" ||
        fieldName === "github" ||
        fieldName === "website") &&
      value
    ) {
      const urlRegex = /^https?:\/\/.+/;
      if (!urlRegex.test(value)) {
        isValid = false;
        errorMessage =
          "Please enter a valid URL starting with http:// or https://";
      }
    } else if (fieldName === "gpa" && value) {
      const gpa = parseFloat(value);
      if (isNaN(gpa) || gpa < 0 || gpa > 4) {
        isValid = false;
        errorMessage = "GPA must be between 0.0 and 4.0";
      }
    }

    if (!isValid) {
      field.classList.add("error");
      const errorDiv = document.createElement("div");
      errorDiv.className = "error-message";
      errorDiv.textContent = errorMessage;
      errorDiv.style.cssText =
        "color: #ef4444; font-size: 0.8rem; margin-top: 0.25rem;";
      field.parentNode.appendChild(errorDiv);
    }

    return isValid;
  }

  // Validate entire form
  function validateForm() {
    const inputs = document.querySelectorAll(
      "#profileForm input, #profileForm textarea",
    );
    let isFormValid = true;

    inputs.forEach((input) => {
      if (!validateField(input)) {
        isFormValid = false;
      }
    });

    return isFormValid;
  }

  // Validate password change
  function validatePasswordChange(
    currentPassword,
    newPassword,
    confirmPassword,
  ) {
    if (!currentPassword) {
      showNotification(
        "Current password is required to change password",
        "error",
      );
      return false;
    }

    if (!newPassword) {
      showNotification("New password is required", "error");
      return false;
    }

    if (newPassword.length < 8) {
      showNotification(
        "New password must be at least 8 characters long",
        "error",
      );
      return false;
    }

    if (newPassword !== confirmPassword) {
      showNotification("New passwords do not match", "error");
      return false;
    }

    // Check password strength
    const hasUpperCase = /[A-Z]/.test(newPassword);
    const hasLowerCase = /[a-z]/.test(newPassword);
    const hasNumbers = /\d/.test(newPassword);
    const hasSpecialChar = /[!@#$%^&*(),.?":{}|<>]/.test(newPassword);

    if (!hasUpperCase || !hasLowerCase || !hasNumbers) {
      showNotification(
        "Password must contain uppercase, lowercase, and numbers",
        "error",
      );
      return false;
    }

    return true;
  }

  // Add error styles
  const style = document.createElement("style");
  style.textContent = `
        .form-group input.error,
        .form-group textarea.error {
            border-color: #ef4444;
            box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
        }
    `;
  document.head.appendChild(style);
});
