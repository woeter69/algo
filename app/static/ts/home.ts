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

// Section Navigation
const allSections = document.querySelectorAll("section");
const navItems = document.querySelectorAll("[data-section]");

// Hide all coming soon sections initially
document.querySelectorAll(".coming-soon").forEach((section) => {
  section.style.display = "none";
});

// Handle navigation clicks
navItems.forEach((item) => {
  item.addEventListener("click", (e) => {
    // Don't prevent default for login and register links
    const sectionName = item.getAttribute("data-section");
    if (sectionName === "login" || sectionName === "register") {
      return; // Let the link work normally
    }

    e.preventDefault();

    if (sectionName === "home") {
      // Show home section, hide others
      document.getElementById("home-section").style.display = "block";
      document.querySelector(".features").style.display = "block";
      document.querySelector(".cta").style.display = "block";
      document.querySelectorAll(".coming-soon").forEach((section) => {
        section.style.display = "none";
      });

      // Scroll to top
      window.scrollTo({
        top: 0,
        behavior: "smooth",
      });
    } else {
      // Hide main content, show the selected coming soon section
      document.getElementById("home-section").style.display = "none";
      document.querySelector(".features").style.display = "none";
      document.querySelector(".cta").style.display = "none";
      document.querySelectorAll(".coming-soon").forEach((section) => {
        section.style.display = "none";
      });

      const targetSection = document.getElementById(`${sectionName}-section`);
      if (targetSection) {
        targetSection.style.display = "flex";

        // Scroll to the section
        window.scrollTo({
          top: targetSection.offsetTop - 80,
          behavior: "smooth",
        });
      }
    }

    // Update active nav link
    navItems.forEach((navItem) => {
      navItem.classList.remove("active");
    });
    item.classList.add("active");
  });
});

// Initialize page - make sure home is shown
document.getElementById("home-section").style.display = "block";
document.querySelector(".features").style.display = "block";
document.querySelector(".cta").style.display = "block";

