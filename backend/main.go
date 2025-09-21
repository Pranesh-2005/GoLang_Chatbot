package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/joho/godotenv"
)

const API_URL = "https://openrouter.ai/api/v1/chat/completions"

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Message string `json:"message"`
}

type ChatResponse struct {
    Response string `json:"response"`
    Error    string `json:"error,omitempty"`
    Time     string `json:"time"`
}

type OpenRouterRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
}

type Choice struct {
    Message Message `json:"message"`
}

type OpenRouterResponse struct {
    Choices []Choice `json:"choices"`
}

// Store conversation history (in production, use a database)
var conversationHistory = []Message{
    {Role: "system", Content: "You are a helpful assistant."},
}

var apiKey string

// Middleware to log requests like Flask
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Create a custom response writer to capture status code
        ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        // Log the incoming request
        log.Printf("[%s] %s %s %s", 
            start.Format("2006-01-02 15:04:05"), 
            r.Method, 
            r.RequestURI, 
            r.RemoteAddr)
        
        // Call the next handler
        next.ServeHTTP(ww, r)
        
        // Log the response
        duration := time.Since(start)
        log.Printf("[%s] %s %s - %d - %v", 
            time.Now().Format("2006-01-02 15:04:05"),
            r.Method, 
            r.RequestURI, 
            ww.statusCode,
            duration)
    }
}

// Custom response writer to capture status code
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func enableCORS(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var chatReq ChatRequest
    err := json.NewDecoder(r.Body).Decode(&chatReq)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        response := ChatResponse{Error: "Invalid JSON"}
        json.NewEncoder(w).Encode(response)
        return
    }

    if chatReq.Message == "" {
        w.WriteHeader(http.StatusBadRequest)
        response := ChatResponse{Error: "Message cannot be empty"}
        json.NewEncoder(w).Encode(response)
        return
    }

    // Log the user message
    log.Printf("User message: %s", chatReq.Message)

    // Add user message to conversation history
    conversationHistory = append(conversationHistory, Message{
        Role:    "user",
        Content: chatReq.Message,
    })

    // Prepare request to OpenRouter
    reqBody := OpenRouterRequest{
        Model:    "deepseek/deepseek-chat-v3.1:free",
        Messages: conversationHistory,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        response := ChatResponse{Error: "Failed to prepare request"}
        json.NewEncoder(w).Encode(response)
        return
    }

    // Measure response time
    start := time.Now()

    // Log OpenRouter API request
    log.Printf("Making request to OpenRouter API...")

    // Make request to OpenRouter
    req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(jsonData))
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        response := ChatResponse{Error: "Failed to create request"}
        json.NewEncoder(w).Encode(response)
        return
    }

    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 30 * time.Second}
    res, err := client.Do(req)
    if err != nil {
        log.Printf("OpenRouter API error: %v", err)
        w.WriteHeader(http.StatusBadGateway)
        response := ChatResponse{Error: "Failed to get response from AI"}
        json.NewEncoder(w).Encode(response)
        return
    }
    defer res.Body.Close()

    // Log OpenRouter API response status
    log.Printf("OpenRouter API response: %d %s", res.StatusCode, res.Status)

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        response := ChatResponse{Error: "Failed to read response"}
        json.NewEncoder(w).Encode(response)
        return
    }

    elapsed := time.Since(start)

    var openRouterResponse OpenRouterResponse
    err = json.Unmarshal(body, &openRouterResponse)
    if err != nil {
        log.Printf("Failed to parse OpenRouter response: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        response := ChatResponse{Error: "Failed to parse AI response"}
        json.NewEncoder(w).Encode(response)
        return
    }

    if len(openRouterResponse.Choices) == 0 {
        log.Printf("No choices in OpenRouter response")
        w.WriteHeader(http.StatusInternalServerError)
        response := ChatResponse{Error: "No response from AI"}
        json.NewEncoder(w).Encode(response)
        return
    }

    assistantMessage := openRouterResponse.Choices[0].Message.Content

    // Log the AI response
    log.Printf("AI response: %s", assistantMessage)

    // Add assistant response to conversation history
    conversationHistory = append(conversationHistory, Message{
        Role:    "assistant",
        Content: assistantMessage,
    })

    // Send response back to frontend
    response := ChatResponse{
        Response: assistantMessage,
        Time:     elapsed.String(),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Reset conversation history
    conversationHistory = []Message{
        {Role: "system", Content: "You are a helpful assistant."},
    }

    log.Printf("Conversation history reset")

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
}

func main() {
    // Load .env file
    err := godotenv.Load()
    if err != nil {
        fmt.Println("Warning: .env file not found, using system environment")
    }

    apiKey = os.Getenv("OPENROUTER_API_KEY")
    if apiKey == "" {
        fmt.Println("Error: Set your OPENROUTER_API_KEY in .env or as environment variable")
        return
    }

    // Setup routes with logging middleware
    http.HandleFunc("/api/chat", loggingMiddleware(chatHandler))
    http.HandleFunc("/api/health", loggingMiddleware(healthHandler))
    http.HandleFunc("/api/reset", loggingMiddleware(resetHandler))

    port := os.Getenv("PORT")
    if port == "" {
        port = "5001"
    }

    fmt.Printf("Server starting on port %s...\n", port)
    fmt.Println("Endpoints:")
    fmt.Println("  POST /api/chat - Send message to chatbot")
    fmt.Println("  GET  /api/health - Health check")
    fmt.Println("  POST /api/reset - Reset conversation")
    fmt.Println("\n--- Server Logs ---")

    err = http.ListenAndServe(":"+port, nil)
    if err != nil {
        fmt.Printf("Error starting server: %v\n", err)
    }
}