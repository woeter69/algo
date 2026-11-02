// @ts-nocheck
// Login Page TypeScript - Navigation and interactions
// Mobile Navigation with proper typing
const hamburger = document.querySelector(".hamburger");
const navLinks = document.querySelector(".nav-links");
const links = document.querySelectorAll(".nav-links li");
if (hamburger && navLinks) {
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
}
// Close mobile menu when clicking on a nav link
links.forEach((link) => {
  link.addEventListener("click", () => {
    if (navLinks && hamburger) {
      navLinks.classList.remove("open");
      hamburger.classList.remove("toggle");
      links.forEach((link) => {
        link.style.animation = "";
      });
    }
  });
});
// Scroll Header Effect
window.addEventListener("scroll", () => {
  const header = document.querySelector("header");
  if (header) {
    header.classList.toggle("scrolled", window.scrollY > 50);
  }
});
function validateLoginForm(formData) {
  const { email, password } = formData;
  if (!email || !email.includes("@")) {
    console.error("Invalid email address");
    return false;
  }
  if (!password || password.length < 6) {
    console.error("Password must be at least 6 characters");
    return false;
  }
  return true;
}
// Enhanced form handling
const loginForm = document.querySelector("form");
if (loginForm) {
  loginForm.addEventListener("submit", (e) => {
    const formData = new FormData(loginForm);
    const loginData = {
      email: formData.get("email") || "",
      password: formData.get("password") || "",
    };
    if (!validateLoginForm(loginData)) {
      e.preventDefault();
      // Show error message to user
      console.log("Form validation failed");
    }
  });
}

export { validateLoginForm };
//# sourceMappingURL=login.js.map
