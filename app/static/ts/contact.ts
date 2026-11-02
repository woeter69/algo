// @ts-nocheck
// Mobile Navigation
const hamburger = document.querySelector(".hamburger");
const navLinks = document.querySelector(".nav-links");
const links = document.querySelectorAll(".nav-links li");

hamburger.addEventListener("click", () => {
  navLinks.classList.toggle("open");
  links.forEach((link, index) => {
    if (link.style.animation) {
      link.style.animation = "";
    } else {
      link.style.animation = `navLinkFade 0.5s ease forwards ${index / 7 + 0.3}s`;
    }
  });
  hamburger.classList.toggle("toggle");
});

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

// === EXTRACTED INLINE SCRIPTS ===

// Chatbase Integration


// Form Submission using AJAX
document.getElementById("contactForm").addEventListener("submit", function (e) {
  e.preventDefault(); // prevent normal POST

  const formData = new FormData(this);
  const actionUrl = this.action; // Use the form's action attribute (set in HTML)

  fetch(actionUrl, {
    method: "POST",
    body: formData,
  }).then((res) => {
    if (res.ok) {
      const successMessage = document.getElementById("successMessage");
      successMessage.style.display = "block";
      this.reset();
      setTimeout(() => {
        successMessage.style.display = "none";
      }, 5000);
    }
  });
});
