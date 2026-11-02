// Utility functions for the chat UI

export function escapeHtml(text: string): string {
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

export function createMessageElement(
  text: string,
  isSent: boolean = false,
  options?: {
    senderUsername?: string;
    senderPfp?: string;
    otherUserPfp?: string;
  },
): HTMLElement {
  const messageDiv = document.createElement("div");
  messageDiv.className = `message ${isSent ? "sent" : "received"}`;

  const currentTime = new Date().toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
  });

  if (isSent) {
    messageDiv.innerHTML = `
            <div class="message-content">
                <div class="message-bubble">
                    <p>${escapeHtml(text)}</p>
                </div>
                <div class="message-time">${currentTime}</div>
            </div>
        `;
  } else {
    const avatarSrc = options?.otherUserPfp || "https://via.placeholder.com/50";
    messageDiv.innerHTML = `
            <div class="message-avatar">
                <img src="${escapeHtml(avatarSrc)}" alt="">
            </div>
            <div class="message-content">
                <div class="message-bubble">
                    <p>${escapeHtml(text)}</p>
                </div>
                <div class="message-time">${currentTime}</div>
            </div>
        `;
  }

  return messageDiv;
}

export function scrollToBottom(messagesArea: HTMLElement | null): void {
  if (!messagesArea) return;
  requestAnimationFrame(() => {
    messagesArea.scrollTop = messagesArea.scrollHeight;
  });
}

export function showTypingIndicator(typingIndicator: HTMLElement | null): void {
  if (typingIndicator) {
    typingIndicator.style.display = "flex";
  }
}

export function hideTypingIndicator(typingIndicator: HTMLElement | null): void {
  if (typingIndicator) {
    typingIndicator.style.display = "none";
  }
}
