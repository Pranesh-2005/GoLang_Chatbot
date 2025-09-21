package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, using system environment")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("Set your OPENROUTER_API_KEY in .env or as environment variable")
		return
	}

	reqBody := Request{
		Model: "deepseek/deepseek-chat-v3.1:free", // change if needed
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello! How are you?"},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	// measure time
	start := time.Now()

	req, _ := http.NewRequest("POST", API_URL, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	elapsed := time.Since(start) // time taken

	var response Response
	json.Unmarshal(body, &response)

	if len(response.Choices) > 0 {
		fmt.Println("Assistant:", response.Choices[0].Message.Content)
	}
	fmt.Println("Go Latency:", elapsed)
}
