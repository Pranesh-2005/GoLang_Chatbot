const API_BASE_URL = 'https://gochatbotbackend.onrender.com';
let isLoading = false;

// DOM Elements
const chatMessages = document.getElementById('chatMessages');
const messageInput = document.getElementById('messageInput');
const sendButton = document.getElementById('sendButton');
const statusElement = document.getElementById('status');
const typingIndicator = document.getElementById('typingIndicator');

// Auto-resize textarea
messageInput.addEventListener('input', function() {
    this.style.height = 'auto';
    this.style.height = Math.min(this.scrollHeight, 100) + 'px';
});

// Send message on Enter (Shift+Enter for new line)
messageInput.addEventListener('keydown', function(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        sendMessage();
    }
});

// Check server status
async function checkStatus() {
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        if (response.ok) {
            statusElement.textContent = 'Online';
            statusElement.className = 'status online';
        } else {
            throw new Error('Server error');
        }
    } catch (error) {
        statusElement.textContent = 'Offline';
        statusElement.className = 'status offline';
    }
}

// Add message to chat
function addMessage(content, role, time = null) {
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${role}`;
    const timeString = time || new Date().toLocaleTimeString();
    // Parse markdown content for assistant messages, keep plain text for user messages
    const parsedContent = role === 'assistant' ? marked.parse(content) : content;
    messageDiv.innerHTML = `
        <div class="message-content">${parsedContent}</div>
        <div class="message-time">${timeString}</div>
    `;
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Show/hide typing indicator
function showTyping(show) {
    typingIndicator.className = show ? 'typing-indicator show' : 'typing-indicator';
    if (show) {
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
}

// Show error message
function showError(message) {
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-message';
    errorDiv.textContent = `Error: ${message}`;
    chatMessages.appendChild(errorDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
    setTimeout(() => { errorDiv.remove(); }, 5000);
}

// Send message
async function sendMessage() {
    const message = messageInput.value.trim();
    if (!message || isLoading) return;

    addMessage(message, 'user');
    messageInput.value = '';
    messageInput.style.height = 'auto';

    isLoading = true;
    sendButton.disabled = true;
    sendButton.textContent = 'Sending...';
    showTyping(true);

    try {
        const response = await fetch(`${API_BASE_URL}/chat`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ prompt: message })
        });

        const data = await response.json();

        if (response.ok) {
            const timeString = `${new Date().toLocaleTimeString()} (${data.latency.toFixed(2)}s)`;
            addMessage(data.answer, 'assistant', timeString);
        } else {
            showError('Failed to get response from AI');
        }
    } catch (error) {
        showError('Failed to connect to server. Please check if the backend is running.');
        console.error('Error:', error);
    } finally {
        isLoading = false;
        sendButton.disabled = false;
        sendButton.textContent = 'Send';
        showTyping(false);
        messageInput.focus();
    }
}

// Reset conversation (clear chat locally)
function resetConversation() {
    if (!confirm('Are you sure you want to clear the chat?')) return;

    chatMessages.innerHTML = `
        <div class="message assistant">
            <div class="message-content">
                Hello! I'm your AI assistant. How can I help you today?
            </div>
            <div class="message-time">Just now</div>
        </div>
    `;
}

// Initialize
document.addEventListener('DOMContentLoaded', function() {
    checkStatus();
    messageInput.focus();
    setInterval(checkStatus, 30000);
});