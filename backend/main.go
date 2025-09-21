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

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Response struct {
	Choices []Choice `json:"choices"`
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to %s", r.Method, r.URL.Path)

	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse user input
	var input struct {
		Prompt string `json:"prompt"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("User prompt: %s", input.Prompt)

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Printf("API key not set")
		http.Error(w, "API key not set", http.StatusInternalServerError)
		return
	}

	reqBody := Request{
		Model: "x-ai/grok-4-fast:free",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: input.Prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		http.Error(w, "Error preparing request", http.StatusInternalServerError)
		return
	}

	start := time.Now()

	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to OpenRouter: %v", err)
		http.Error(w, "Error contacting AI service", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	log.Printf("OpenRouter response status: %d", res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(start)

	if res.StatusCode != http.StatusOK {
		log.Printf("OpenRouter error response: %s", string(body))
		http.Error(w, "AI service error", http.StatusInternalServerError)
		return
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		http.Error(w, "Error parsing response", http.StatusInternalServerError)
		return
	}

	output := map[string]interface{}{
		"answer":  "",
		"latency": elapsed.Seconds(),
	}

	if len(response.Choices) > 0 {
		output["answer"] = response.Choices[0].Message.Content
		log.Printf("AI response: %s", response.Choices[0].Message.Content)
	} else {
		log.Printf("No choices in response")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Check if API key is set
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is not set")
	}

	// Setup routes with CORS middleware
	http.HandleFunc("/chat", corsMiddleware(chatHandler))

	// Health check endpoint
	http.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Endpoints:")
	fmt.Println("  POST /chat - Chat with AI")
	fmt.Println("  GET /health - Health check")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
