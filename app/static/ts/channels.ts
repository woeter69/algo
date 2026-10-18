// @ts-nocheck
// Get user data from template
const channelsData: any = (window as any).channelsData || {};
const currentUserId: number = (window as any).currentUserId;
const currentUsername: string = (window as any).currentUsername || "User";
const currentUserPfp: string = (window as any).currentUserPfp || "";
const selectedCommunityId: number = channelsData.selectedCommunityId || (channelsData.communities && channelsData.communities[0] ? channelsData.communities[0][0] : null);

console.log("🚀 Channels initializing with user:", {
  currentUserId,
  currentUsername,
  selectedCommunityId,
});

// Initialize Go WebSocket connection
const socket: any = (window as any).goSocket;
if (socket && currentUserId) {
  console.log("🔌 Connecting to Go WebSocket server...");
  socket.connect(currentUserId.toString(), currentUsername, currentUserPfp);
}

// UI Elements
const channels = document.querySelectorAll<HTMLElement>(".channel");
const currentChannelName = document.getElementById("currentChannelName") || document.querySelector(".current-channel-name");
const currentChannelIcon = document.querySelector(".current-channel-icon");
const channelDescription = document.getElementById("channelDescription");
const messageInput = document.getElementById("messageInput") as HTMLInputElement;
const sendBtn = document.getElementById("sendBtn");
const chatMessages = document.getElementById("chatMessages");
const membersList = document.getElementById("membersList");

// Current active channel
let currentChannelId: string | null = null;

// Tab switching
const tabBtns = document.querySelectorAll(".tab-btn");
const tabPanels = document.querySelectorAll(".tab-panel");

tabBtns.forEach((btn) => {
  btn.addEventListener("click", () => {
    const tabName = btn.getAttribute("data-tab");

    tabBtns.forEach((b) => b.classList.remove("active"));
    tabPanels.forEach((p) => p.classList.remove("active"));

    btn.classList.add("active");
    const targetPanel = document.getElementById(`${tabName}-panel`);
    if (targetPanel) {
      targetPanel.classList.add("active");
    }

    if (tabName === "connections" && currentChannelId) {
      loadChannelMembers(currentChannelId);
    }
  });
});

// Channel switching logic
function selectChannel(channelEl: HTMLElement) {
  channels.forEach((c) => c.classList.remove("active"));
  channelEl.classList.add("active");

  const channelNameEl = channelEl.querySelector(".channel-name");
  const channelName = channelNameEl ? channelNameEl.textContent?.trim() : "general";
  const channelIconEl = channelEl.querySelector(".channel-icon i");
  const channelIcon = channelIconEl ? channelIconEl.className : "fas fa-hashtag";
  const channelId = channelEl.getAttribute("data-channel-id") || channelName;

  // Update UI
  if (currentChannelName) {
    currentChannelName.textContent = `# ${channelName}`;
  }
  if (currentChannelIcon) {
    currentChannelIcon.innerHTML = `<i class="${channelIcon}"></i>`;
  }
  if (messageInput) {
    messageInput.placeholder = `Message #${channelName}`;
  }

  // Leave previous channel and join new one in Go WebSocket
  if (currentChannelId && socket) {
    socket.leaveChannel(currentChannelId);
  }

  currentChannelId = channelId;
  if (socket) {
    console.log(`🚪 Joining channel: ${channelId}`);
    socket.joinChannel(channelId);
  }

  // Load channel messages and members
  loadChannelMessages(channelId);
  loadChannelMembers(channelId);
}

channels.forEach((channel) => {
  channel.addEventListener("click", () => {
    selectChannel(channel);
  });
});

// Load channel messages from API
async function loadChannelMessages(channelId: string) {
  if (!chatMessages) return;

  chatMessages.innerHTML = `
    <div class="loading-messages" style="display: flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 2rem; color: #94a3b8;">
      <i class="fas fa-spinner fa-spin"></i>
      <span>Loading messages...</span>
    </div>
  `;

  try {
    const res = await fetch(`/channels/${channelId}/messages?limit=50`);
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}`);
    }
    const data = await res.json();
    const messages = data.messages || [];

    chatMessages.innerHTML = "";

    if (messages.length === 0) {
      chatMessages.innerHTML = `
        <div style="text-align: center; padding: 3rem 1rem; color: #94a3b8;">
          <div style="font-size: 2.5rem; margin-bottom: 0.5rem; opacity: 0.6;">
            <i class="fas fa-comments"></i>
          </div>
          <div style="font-weight: 600; font-size: 1rem; color: #475569;">No messages yet</div>
          <div style="font-size: 0.85rem; margin-top: 0.25rem;">Be the first to start the conversation!</div>
        </div>
      `;
      return;
    }

    messages.forEach((msg: any) => {
      const isSent = String(msg.user_id) === String(currentUserId);
      const timeStr = msg.created_at
        ? new Date(msg.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
        : "Just now";

      const messageDiv = document.createElement("div");
      messageDiv.className = `message ${isSent ? "sent" : "received"}`;
      messageDiv.setAttribute("data-message-id", msg.message_id || "");

      messageDiv.innerHTML = `
        <div class="message-avatar">
          ${msg.pfp_path ? `<img src="${msg.pfp_path}" alt="${msg.username || 'User'}">` : (msg.username ? msg.username[0].toUpperCase() : 'U')}
        </div>
        <div class="message-content">
          <div class="message-header">
            <span class="message-author">${msg.username || 'User'}</span>
            <span class="message-time">${timeStr}</span>
          </div>
          <div class="message-text">${escapeHtml(msg.content)}</div>
        </div>
      `;

      chatMessages.appendChild(messageDiv);
    });

    chatMessages.scrollTop = chatMessages.scrollHeight;
  } catch (err) {
    console.error("❌ Failed to load channel messages:", err);
    chatMessages.innerHTML = `
      <div style="text-align: center; padding: 2rem; color: #ef4444;">
        <i class="fas fa-exclamation-triangle"></i> Failed to load channel messages. Please refresh or try another channel.
      </div>
    `;
  }
}

// Load channel members from API
async function loadChannelMembers(channelId: string) {
  if (!membersList) return;

  membersList.innerHTML = `
    <div class="loading-members" style="display: flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 2rem; color: #94a3b8;">
      <i class="fas fa-spinner fa-spin"></i>
      <span>Loading members...</span>
    </div>
  `;

  try {
    const res = await fetch(`/channels/${channelId}/members`);
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}`);
    }
    const data = await res.json();
    const members = data.members || [];

    membersList.innerHTML = "";

    if (members.length === 0) {
      membersList.innerHTML = `
        <div style="text-align: center; padding: 2rem; color: #94a3b8; font-size: 0.9rem;">
          No members found in this channel.
        </div>
      `;
      return;
    }

    members.forEach((member: any) => {
      const item = document.createElement("div");
      item.className = "connection";
      item.style.cursor = "pointer";

      const displayName = [member.firstname, member.lastname].filter(Boolean).join(" ") || member.username;
      const initial = member.username ? member.username[0].toUpperCase() : "U";

      item.innerHTML = `
        <div class="connection-avatar-circle" style="background: var(--gradient); overflow: hidden;">
          ${member.pfp_path ? `<img src="${member.pfp_path}" alt="${member.username}" style="width: 100%; height: 100%; object-fit: cover;">` : initial}
        </div>
        <div class="connection-info">
          <div class="connection-name">${escapeHtml(displayName)}</div>
          <div class="connection-details">@${escapeHtml(member.username)} &bull; ${escapeHtml(member.role || 'Member')}</div>
        </div>
        <div class="connection-status ${member.is_online ? 'online' : 'offline'}" title="${member.is_online ? 'Online' : 'Offline'}"></div>
      `;

      item.addEventListener("click", () => {
        openProfileModal(member);
      });

      membersList.appendChild(item);
    });
  } catch (err) {
    console.error("❌ Failed to load channel members:", err);
    membersList.innerHTML = `
      <div style="text-align: center; padding: 2rem; color: #ef4444; font-size: 0.85rem;">
        Failed to load members.
      </div>
    `;
  }
}

// Profile Modal handler
const profileModal = document.getElementById("profileModal");
const closeProfileModalBtn = document.getElementById("closeProfileModal");
const closePreviewBtn = document.getElementById("closePreview");
const profileFromPreview = document.getElementById("profileFromPreview");
const connectFromPreview = document.getElementById("connectFromPreview");
const messageFromPreview = document.getElementById("messageFromPreview");
const profilePreview = document.getElementById("profilePreview");

let activePreviewMember: any = null;

function openProfileModal(member: any) {
  activePreviewMember = member;
  if (!profileModal || !profilePreview) return;

  const displayName = [member.firstname, member.lastname].filter(Boolean).join(" ") || member.username;
  const initial = member.username ? member.username[0].toUpperCase() : "U";

  profilePreview.innerHTML = `
    <div style="display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem;">
      <div style="width: 60px; height: 60px; border-radius: 50%; background: var(--gradient); display: flex; align-items: center; justify-content: center; color: white; font-size: 1.5rem; font-weight: 700; overflow: hidden;">
        ${member.pfp_path ? `<img src="${member.pfp_path}" alt="${member.username}" style="width: 100%; height: 100%; object-fit: cover;">` : initial}
      </div>
      <div>
        <h4 style="font-size: 1.1rem; font-weight: 700; color: var(--dark);">${escapeHtml(displayName)}</h4>
        <p style="color: #64748b; font-size: 0.85rem;">@${escapeHtml(member.username)}</p>
        <span style="display: inline-block; padding: 0.2rem 0.6rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; background: #ede9fe; color: #6d28d9; margin-top: 0.25rem;">
          ${escapeHtml(member.role || 'Member')}
        </span>
      </div>
    </div>
  `;

  if (profileFromPreview) {
    profileFromPreview.onclick = () => {
      window.location.href = `/profile/${member.user_id}`;
    };
  }
  if (messageFromPreview) {
    messageFromPreview.onclick = () => {
      window.location.href = `/chat/${encodeURIComponent(member.username)}`;
    };
  }
  if (connectFromPreview) {
    connectFromPreview.onclick = async () => {
      try {
        const res = await fetch("/api/connect", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ target_id: member.user_id }),
        });
        if (res.ok) {
          connectFromPreview.innerHTML = `<i class="fas fa-check"></i> Requested`;
          connectFromPreview.setAttribute("disabled", "true");
        }
      } catch (e) {
        console.error("Connect error:", e);
      }
    };
  }

  profileModal.classList.add("active");
}

function closeProfileModal() {
  if (profileModal) {
    profileModal.classList.remove("active");
  }
}

closeProfileModalBtn?.addEventListener("click", closeProfileModal);
closePreviewBtn?.addEventListener("click", closeProfileModal);
profileModal?.addEventListener("click", (e) => {
  if (e.target === profileModal) closeProfileModal();
});

// Create Channel Modal handler
const createChannelModal = document.getElementById("createChannelModal");
const openCreateChannelBtn = document.getElementById("openCreateChannelBtn");
const closeCreateChannelModal = document.getElementById("closeCreateChannelModal");
const cancelCreateChannel = document.getElementById("cancelCreateChannel");
const createChannelForm = document.getElementById("createChannelForm");
const createChannelError = document.getElementById("createChannelError");

function openCreateModal() {
  if (createChannelModal) {
    createChannelModal.classList.add("active");
    if (createChannelError) createChannelError.style.display = "none";
    const nameInput = document.getElementById("newChannelName") as HTMLInputElement;
    if (nameInput) {
      nameInput.value = "";
      nameInput.focus();
    }
  }
}

function closeCreateModal() {
  if (createChannelModal) {
    createChannelModal.classList.remove("active");
  }
}

openCreateChannelBtn?.addEventListener("click", openCreateModal);
closeCreateChannelModal?.addEventListener("click", closeCreateModal);
cancelCreateChannel?.addEventListener("click", closeCreateModal);
createChannelModal?.addEventListener("click", (e) => {
  if (e.target === createChannelModal) closeCreateModal();
});

createChannelForm?.addEventListener("submit", async (e) => {
  e.preventDefault();
  const nameInput = document.getElementById("newChannelName") as HTMLInputElement;
  const descInput = document.getElementById("newChannelDesc") as HTMLTextAreaElement;
  const privateInput = document.getElementById("newChannelPrivate") as HTMLInputElement;

  const name = nameInput.value.trim();
  const description = descInput ? descInput.value.trim() : "";
  const is_private = privateInput ? privateInput.checked : false;

  if (!name) return;
  if (!selectedCommunityId) {
    if (createChannelError) {
      createChannelError.textContent = "Please select a community first.";
      createChannelError.style.display = "block";
    }
    return;
  }

  try {
    const res = await fetch(`/communities/${selectedCommunityId}/channels`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name,
        description,
        is_private,
        channel_type: "text",
      }),
    });

    const data = await res.json();
    if (res.ok) {
      closeCreateModal();
      window.location.reload();
    } else {
      if (createChannelError) {
        createChannelError.textContent = data.error || "Failed to create channel.";
        createChannelError.style.display = "block";
      }
    }
  } catch (err) {
    console.error("Create channel error:", err);
    if (createChannelError) {
      createChannelError.textContent = "An error occurred while creating the channel.";
      createChannelError.style.display = "block";
    }
  }
});

// Explore Communities Modal handler
const exploreModal = document.getElementById("exploreCommunitiesModal");
const openExploreBtn = document.getElementById("openExploreCommunitiesBtn");
const explorePromptBtn = document.getElementById("explorePromptBtn");
const closeExploreModalBtn = document.getElementById("closeExploreCommunitiesModal");
const exploreList = document.getElementById("exploreCommunitiesList");

async function loadExploreCommunities() {
  if (!exploreList) return;
  exploreList.innerHTML = `
    <div style="text-align: center; padding: 2rem; color: #a0aec0;">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <p style="margin-top: 0.5rem;">Loading communities...</p>
    </div>
  `;

  try {
    const res = await fetch("/communities/discover");
    if (!res.ok) throw new Error("Failed to fetch communities");
    const data = await res.json();
    const communities = data.communities || [];

    if (communities.length === 0) {
      exploreList.innerHTML = `<div style="text-align: center; padding: 2rem; color: #718096;">No communities available yet.</div>`;
      return;
    }

    exploreList.innerHTML = communities
      .map((c: any) => {
        const isJoined = c.membership_status === "active";
        return `
          <div style="display: flex; justify-content: space-between; align-items: center; padding: 1rem; border: 1px solid #e2e8f0; border-radius: 12px; transition: all 0.2s ease;">
            <div>
              <div style="font-weight: 600; color: #2d3748; font-size: 0.95rem;">
                ${c.name} ${c.college_code ? `<span style="font-size: 0.75rem; background: #edf2f7; color: #4a5568; padding: 0.15rem 0.4rem; border-radius: 4px; margin-left: 0.4rem;">${c.college_code}</span>` : ""}
              </div>
              <div style="font-size: 0.8rem; color: #718096; margin-top: 0.2rem;">
                ${c.location ? `<i class="fas fa-map-marker-alt"></i> ${c.location} &bull; ` : ""}
                <i class="fas fa-users"></i> ${c.member_count || 0} members
              </div>
              ${c.description ? `<p style="font-size: 0.8rem; color: #4a5568; margin-top: 0.4rem;">${c.description}</p>` : ""}
            </div>
            <div>
              ${
                isJoined
                  ? `<button type="button" onclick="window.location.href='?community_id=${c.community_id}'" style="background: #edf2f7; color: #4a5568; border: none; padding: 0.45rem 0.9rem; border-radius: 8px; font-size: 0.85rem; font-weight: 600; cursor: pointer;">
                      <i class="fas fa-arrow-right"></i> Open
                    </button>`
                  : `<button type="button" class="join-community-btn" data-community-id="${c.community_id}" style="background: var(--gradient); color: white; border: none; padding: 0.45rem 1rem; border-radius: 8px; font-size: 0.85rem; font-weight: 600; cursor: pointer;">
                      <i class="fas fa-plus"></i> Join
                    </button>`
              }
            </div>
          </div>
        `;
      })
      .join("");

    const joinButtons = exploreList.querySelectorAll(".join-community-btn");
    joinButtons.forEach((btn) => {
      btn.addEventListener("click", async () => {
        const commId = btn.getAttribute("data-community-id");
        if (!commId) return;
        btn.setAttribute("disabled", "true");
        btn.innerHTML = `<i class="fas fa-spinner fa-spin"></i> Joining...`;

        try {
          const joinRes = await fetch(`/communities/${commId}/join`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
          });
          if (joinRes.ok) {
            btn.innerHTML = `<i class="fas fa-check"></i> Joined!`;
            setTimeout(() => {
              window.location.href = `?community_id=${commId}`;
            }, 600);
          } else {
            const errData = await joinRes.json();
            alert(errData.error || "Failed to join community");
            btn.removeAttribute("disabled");
            btn.innerHTML = `<i class="fas fa-plus"></i> Join`;
          }
        } catch (e) {
          console.error("Join error:", e);
          btn.removeAttribute("disabled");
          btn.innerHTML = `<i class="fas fa-plus"></i> Join`;
        }
      });
    });
  } catch (err) {
    console.error("Error loading communities:", err);
    exploreList.innerHTML = `<div style="text-align: center; padding: 2rem; color: #ef4444;">Failed to load communities. Please try again.</div>`;
  }
}

function openExploreModal() {
  if (exploreModal) {
    exploreModal.classList.add("active");
    loadExploreCommunities();
  }
}

function closeExploreModal() {
  if (exploreModal) {
    exploreModal.classList.remove("active");
  }
}

openExploreBtn?.addEventListener("click", openExploreModal);
explorePromptBtn?.addEventListener("click", openExploreModal);
closeExploreModalBtn?.addEventListener("click", closeExploreModal);
exploreModal?.addEventListener("click", (e) => {
  if (e.target === exploreModal) closeExploreModal();
});

// Message sending
function sendMessage() {
  if (!messageInput) return;
  const message = messageInput.value.trim();

  if (message && currentChannelId) {
    // Create optimistic message element
    const messageDiv = document.createElement("div");
    messageDiv.className = "message sent";
    messageDiv.innerHTML = `
      <div class="message-avatar">
        ${currentUserPfp ? `<img src="${currentUserPfp}" alt="${currentUsername}">` : (currentUsername ? currentUsername[0].toUpperCase() : 'U')}
      </div>
      <div class="message-content">
        <div class="message-header">
          <span class="message-author">${escapeHtml(currentUsername)}</span>
          <span class="message-time">Now</span>
        </div>
        <div class="message-text">${escapeHtml(message)}</div>
      </div>
    `;

    chatMessages?.appendChild(messageDiv);
    if (chatMessages) {
      chatMessages.scrollTop = chatMessages.scrollHeight;
    }
    messageInput.value = "";

    // Send through WebSocket using dedicated channel message method
    if (socket) {
      console.log(`📤 Sending message to channel ${currentChannelId}:`, message);
      socket.sendChannelMessage(message, currentChannelId);
    }
  } else if (!currentChannelId) {
    console.warn("⚠️ No channel selected");
  }
}

sendBtn?.addEventListener("click", sendMessage);
messageInput?.addEventListener("keypress", (e) => {
  if (e.key === "Enter") {
    sendMessage();
  }
});

// Display incoming channel message
function displayChannelMessage(data: any) {
  if (String(data.channel_id) !== String(currentChannelId)) {
    return; // Only show messages for currently selected channel
  }

  // Prevent duplicate rendering if this client sent the message optimistically
  if (String(data.user_id) === String(currentUserId)) {
    return;
  }

  const messageDiv = document.createElement("div");
  messageDiv.className = "message received";
  const timeStr = data.created_at
    ? new Date(data.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    : "Just now";

  messageDiv.innerHTML = `
    <div class="message-avatar">
      ${data.pfp_path ? `<img src="${data.pfp_path}" alt="${data.username || 'User'}">` : (data.username ? data.username[0].toUpperCase() : 'U')}
    </div>
    <div class="message-content">
      <div class="message-header">
        <span class="message-author">${escapeHtml(data.username || 'User')}</span>
        <span class="message-time">${timeStr}</span>
      </div>
      <div class="message-text">${escapeHtml(data.content)}</div>
    </div>
  `;

  chatMessages?.appendChild(messageDiv);
  if (chatMessages) {
    chatMessages.scrollTop = chatMessages.scrollHeight;
  }
}

// Helper: Escape HTML strings for XSS prevention
function escapeHtml(text: string): string {
  if (!text) return "";
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

// WebSocket event handlers
if (socket) {
  socket.on("connect", () => {
    console.log("✅ Connected to Go WebSocket server for channels!");
    // Select the first channel by default if none is currently selected
    if (!currentChannelId) {
      const firstChannel = document.querySelector<HTMLElement>(".channel");
      if (firstChannel) {
        selectChannel(firstChannel);
      }
    }
  });

  socket.on("disconnect", () => {
    console.log("❌ Disconnected from WebSocket server");
  });

  socket.on("error", (error: any) => {
    console.error("❌ WebSocket error:", error);
  });

  // Handle incoming channel messages
  socket.on("new_message", (data: any) => {
    console.log("📨 Received channel message:", data);
    displayChannelMessage(data);
  });

  // Handle user join/leave events
  socket.on("user_joined", (data: any) => {
    if (String(data.channel_id) === String(currentChannelId)) {
      const systemMsg = document.createElement("div");
      systemMsg.className = "system-message text-center py-2 text-xs text-gray-400";
      systemMsg.innerHTML = `<span>👋 ${escapeHtml(data.username || 'A user')} joined the channel</span>`;
      chatMessages?.appendChild(systemMsg);
      if (chatMessages) chatMessages.scrollTop = chatMessages.scrollHeight;
    }
  });

  socket.on("user_left", (data: any) => {
    if (String(data.channel_id) === String(currentChannelId)) {
      const systemMsg = document.createElement("div");
      systemMsg.className = "system-message text-center py-2 text-xs text-gray-400";
      systemMsg.innerHTML = `<span>👋 ${escapeHtml(data.username || 'A user')} left the channel</span>`;
      chatMessages?.appendChild(systemMsg);
      if (chatMessages) chatMessages.scrollTop = chatMessages.scrollHeight;
    }
  });

  // Handle typing indicators
  socket.on("user_typing", (data: any) => {
    if (String(data.channel_id) === String(currentChannelId) && String(data.user_id) !== String(currentUserId)) {
      showTypingIndicator(data.username);
    }
  });
}

// Show typing indicator
function showTypingIndicator(username: string) {
  const existingIndicator = document.querySelector(".typing-indicator");
  if (existingIndicator) {
    existingIndicator.remove();
  }

  const typingDiv = document.createElement("div");
  typingDiv.className = "typing-indicator";
  typingDiv.innerHTML = `<span>${escapeHtml(username)} is typing...</span>`;
  chatMessages?.appendChild(typingDiv);
  if (chatMessages) chatMessages.scrollTop = chatMessages.scrollHeight;

  setTimeout(() => {
    if (typingDiv.parentNode) {
      typingDiv.remove();
    }
  }, 3000);
}

// Typing debounce for channel input
let typingTimer: any = null;
messageInput?.addEventListener("input", () => {
  if (socket && currentChannelId) {
    socket.startTyping(currentChannelId);

    if (typingTimer) clearTimeout(typingTimer);
    typingTimer = setTimeout(() => {
      socket.stopTyping(currentChannelId);
    }, 1000);
  }
});

// Auto-select initial channel on page load if active
const activeChannel = document.querySelector<HTMLElement>(".channel.active") || document.querySelector<HTMLElement>(".channel");
if (activeChannel) {
  selectChannel(activeChannel);
}

console.log("✅ Channels full architecture initialized!");
