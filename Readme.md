# GoLang_Chatbot 🤖

A blazing fast chatbot powered by Go and OpenRouter's free LLM. GoLang_Chatbot brings together Go's performance on the backend and a clean, modern web frontend for a seamless conversational experience.

---

## 🚀 Introduction

GoLang_Chatbot is a lightweight web chatbot built with a Go backend and a simple HTML/JS/CSS frontend. It leverages the [openrouter.ai](https://openrouter.ai) API for language model capabilities, providing a free and easy-to-deploy chatbot for your personal use, demos, or as a starting point for more advanced projects.

---

## ✨ Features

- **Fast Go Backend**: Handles chat requests efficiently.
- **OpenRouter.ai Integration**: Uses free LLM models for responses.
- **Responsive Frontend**: Mobile-friendly and modern UI.
- **PWA Support**: Includes a manifest for installable experience.
- **Easy Deployment**: Minimal setup, deployable on platforms like Render.

---

## 🛠️ Installation

### Prerequisites

- [Go](https://golang.org/dl/) (v1.18+)
- [Node.js](https://nodejs.org/) and npm (if modifying frontend dependencies)
- [openrouter.ai](https://openrouter.ai) API key

### Steps

1. **Clone the Repository**
    ```bash
    git clone https://github.com/yourusername/GoLang_Chatbot.git
    cd GoLang_Chatbot
    ```

2. **Setup Backend**

    - Navigate to the backend directory:
      ```bash
      cd backend
      ```
    - Copy `.env.example` to `.env` and add your OpenRouter API key:
      ```
      OPENROUTER_API_KEY=your_api_key_here
      ```
    - Install dependencies (if any) and run the server:
      ```bash
      go run main.go
      ```

3. **Setup Frontend**

    - Serve the contents of the `frontend/` folder using any static server, or open `index.html` directly in your browser.
    - Make sure the backend URL in `script.js` matches your deployment.

---

## 💡 Usage

1. **Start the Backend**
    - Ensure the Go server is running and reachable.

2. **Open the Frontend**
    - Open `frontend/index.html` in your browser.
    - Start chatting! Your messages will be processed via OpenRouter's LLM.

> **Tip**: For production, deploy the backend (e.g., on [Render](https://render.com)) and host the frontend on any static hosting service.

---

## 🤝 Contributing

We welcome contributions! To get started:

1. Fork this repo.
2. Clone your fork and create a new branch.
3. Make your changes.
4. Submit a Pull Request (PR) describing your changes.

**Please follow conventional commit messages and ensure code is clean and documented.**

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

---

> _Made with ❤️ using Go and OpenRouter.ai_


## License
This project is licensed under the **MIT** License.

---
🔗 GitHub Repo: https://github.com/Pranesh-2005/GoLang_Chatbot