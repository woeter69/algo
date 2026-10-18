// Requests TypeScript - Connection Requests Management
// Following TypeScript-first approach for the AlumniGo project

interface ConnectionRequest {
  connection_id: number;
  user_id: number;
  name: string;
  username: string;
  avatar?: string;
  message?: string;
  created_at: string;
}

interface Connection {
  user_id: number;
  name: string;
  username: string;
  avatar?: string;
  connected_at: string;
}

interface RequestsData {
  pendingRequests: ConnectionRequest[];
  sentRequests: ConnectionRequest[];
  connections: Connection[];
  connectionsCount: number;
}

// Prevent multiple initializations
if ((window as any).requestsInitialized) {
  console.log("Requests already initialized, skipping...");
} else {
  (window as any).requestsInitialized = true;

  document.addEventListener("DOMContentLoaded", function () {
    console.log("🔔 Requests.ts initializing - DOM ready");

    // Get data from template
    const requestsData: RequestsData = (window as any).requestsData || {
      pendingRequests: [],
      sentRequests: [],
      connections: [],
      connectionsCount: 0,
    };

    console.log("Requests data:", requestsData);

    // Initialize components
    initializeTabs();
    initializeProfilePictures();
    initializeMobileNavigation();

    // Tab switching functionality
    function initializeTabs(): void {
      const tabButtons =
        document.querySelectorAll<HTMLButtonElement>(".tab-btn");
      const tabPanes = document.querySelectorAll<HTMLElement>(".tab-pane");

      tabButtons.forEach((button) => {
        button.addEventListener("click", () => {
          const targetTab = button.getAttribute("data-tab");

          // Remove active class from all tabs and panes
          tabButtons.forEach((btn) => btn.classList.remove("active"));
          tabPanes.forEach((pane) => pane.classList.remove("active"));

          // Add active class to clicked tab and corresponding pane
          button.classList.add("active");
          const targetPane = document.getElementById(targetTab);
          if (targetPane) {
            targetPane.classList.add("active");
          }
        });
      });
    }

    // Initialize profile pictures with fallbacks
    function initializeProfilePictures(): void {
      const avatars = document.querySelectorAll<HTMLElement>(".user-avatar");

      avatars.forEach((avatar) => {
        const img = avatar.querySelector("img");
        const fallback = avatar.querySelector<HTMLElement>(".avatar-fallback");

        if (img && fallback) {
          img.addEventListener("load", function () {
            if (
              this.src &&
              this.src !== "" &&
              !this.src.includes("placeholder")
            ) {
              fallback.style.display = "none";
              this.style.display = "block";
            }
          });

          img.addEventListener("error", function () {
            this.style.display = "none";
            fallback.style.display = "flex";
          });

          // Check if image is already loaded or has no src
          if (!img.src || img.src === "" || img.src.includes("placeholder")) {
            img.style.display = "none";
            fallback.style.display = "flex";
          }
        }
      });
    }

    // Mobile navigation
    function initializeMobileNavigation(): void {
      const hamburger = document.querySelector<HTMLElement>(".hamburger");
      const navLinks = document.querySelector<HTMLElement>(".nav-links");

      if (hamburger && navLinks) {
        hamburger.addEventListener("click", () => {
          hamburger.classList.toggle("toggle");
          navLinks.classList.toggle("open");
        });
      }
    }

    // Global functions for button actions
    (window as any).acceptRequest = async function (
      connectionId: number,
    ): Promise<void> {
      await handleConnectionResponse(connectionId, "accept");
    };

    (window as any).declineRequest = async function (
      connectionId: number,
    ): Promise<void> {
      await handleConnectionResponse(connectionId, "reject");
    };

    (window as any).cancelRequest = async function (
      connectionId: number,
    ): Promise<void> {
      await handleCancelRequest(connectionId);
    };

    (window as any).viewProfile = function (username: string): void {
      window.location.href = `/profile/${username}`;
    };

    (window as any).startChat = function (username: string): void {
      window.location.href = `/chat/${username}`;
    };

    // Handle accept/decline connection requests
    async function handleConnectionResponse(
      connectionId: number,
      action: "accept" | "reject",
    ): Promise<void> {
      try {
        showLoading(true);

        const response = await fetch("/api/respond_connection_request", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            connection_id: connectionId,
            action: action,
          }),
        });

        const result = await response.json();

        if (result.success) {
          // Remove the request card from DOM
          const requestCard = document.querySelector(
            `[data-request-id="${connectionId}"]`,
          );
          if (requestCard) {
            requestCard.remove();
          }

          // Update counters
          updateCounters();

          // Show success message
          const message =
            action === "accept"
              ? "Connection request accepted!"
              : "Connection request declined";
          showNotification(message, "success");

          // If accepted, refresh the page to update connections tab
          if (action === "accept") {
            setTimeout(() => {
              window.location.reload();
            }, 1500);
          }
        } else {
          showNotification(
            result.message || `Failed to ${action} connection request`,
            "error",
          );
        }
      } catch (error) {
        console.error(`Error ${action}ing connection request:`, error);
        showNotification("Network error. Please try again.", "error");
      } finally {
        showLoading(false);
      }
    }

    // Handle cancel sent request
    async function handleCancelRequest(connectionId: number): Promise<void> {
      try {
        showLoading(true);

        const response = await fetch("/api/cancel_connection_request", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            connection_id: connectionId,
          }),
        });

        const result = await response.json();

        if (result.success) {
          // Remove the request card from DOM
          const requestCard = document.querySelector(
            `[data-request-id="${connectionId}"]`,
          );
          if (requestCard) {
            requestCard.remove();
          }

          // Update counters
          updateCounters();

          showNotification("Connection request cancelled", "info");
        } else {
          showNotification(
            result.message || "Failed to cancel connection request",
            "error",
          );
        }
      } catch (error) {
        console.error("Error cancelling connection request:", error);
        showNotification("Network error. Please try again.", "error");
      } finally {
        showLoading(false);
      }
    }

    // Update counters after actions
    function updateCounters(): void {
      const incomingCards = document.querySelectorAll(
        "#incomingRequests .request-card",
      ).length;
      const sentCards = document.querySelectorAll(
        "#sentRequests .request-card",
      ).length;
      const connectionCards = document.querySelectorAll(
        "#myConnections .request-card",
      ).length;

      // Update badges
      const incomingBadge = document.getElementById("incomingBadge");
      const sentBadge = document.getElementById("sentBadge");
      const connectionsBadge = document.getElementById("connectionsBadge");

      if (incomingBadge) incomingBadge.textContent = incomingCards.toString();
      if (sentBadge) sentBadge.textContent = sentCards.toString();
      if (connectionsBadge)
        connectionsBadge.textContent = connectionCards.toString();

      // Update stat cards
      const pendingCount = document.getElementById("pendingCount");
      const sentCount = document.getElementById("sentCount");
      const connectionsCount = document.getElementById("connectionsCount");

      if (pendingCount) pendingCount.textContent = incomingCards.toString();
      if (sentCount) sentCount.textContent = sentCards.toString();
      if (connectionsCount)
        connectionsCount.textContent = connectionCards.toString();
    }

    // Show/hide loading overlay
    function showLoading(show: boolean): void {
      const loadingOverlay = document.getElementById("loadingOverlay");
      if (loadingOverlay) {
        loadingOverlay.style.display = show ? "flex" : "none";
      }
    }

    // Notification system
    function showNotification(
      message: string,
      type: "success" | "error" | "info" | "warning" = "info",
    ): void {
      const notification = document.createElement("div");
      notification.className = `notification notification-${type}`;

      const icons = {
        success: "fas fa-check-circle",
        error: "fas fa-exclamation-circle",
        info: "fas fa-info-circle",
        warning: "fas fa-exclamation-triangle",
      };

      const colors = {
        success: "#10b981",
        error: "#ef4444",
        info: "#6D28D9",
        warning: "#f59e0b",
      };

      notification.innerHTML = `
                <i class="${icons[type]}"></i>
                <span>${escapeHtml(message)}</span>
            `;

      notification.style.cssText = `
                position: fixed; top: 20px; right: 20px; z-index: 10000;
                background: ${colors[type]}; color: white; padding: 1rem 1.5rem;
                border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.15);
                display: flex; align-items: center; gap: 0.5rem;
                font-weight: 500; max-width: 400px;
                transform: translateX(100%); transition: transform 0.3s ease;
            `;

      document.body.appendChild(notification);

      // Animate in
      requestAnimationFrame(() => {
        notification.style.transform = "translateX(0)";
      });

      // Auto remove
      setTimeout(() => {
        notification.style.transform = "translateX(100%)";
        setTimeout(() => notification.remove(), 300);
      }, 4000);
    }

    // Utility function to escape HTML
    function escapeHtml(text: string): string {
      const div = document.createElement("div");
      div.textContent = text;
      return div.innerHTML;
    }

    // Header scroll effect
    const header = document.querySelector<HTMLElement>("header");
    if (header) {
      window.addEventListener("scroll", function () {
        if (window.scrollY > 100) {
          header.classList.add("scrolled");
        } else {
          header.classList.remove("scrolled");
        }
      });
    }

    console.log("✅ Requests TypeScript initialization complete");
  });
}

// Export types for potential external use
export { ConnectionRequest, Connection, RequestsData };
