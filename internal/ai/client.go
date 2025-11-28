package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Client struct {
	ApiKey       string
	ApiBase      string
	Model        string
	History      []Message
	CommitPrompt string
	ChatPrompt   string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func NewClient() *Client {
	apiKey := viper.GetString("openai_api_key")
	apiBase := viper.GetString("openai_api_base")
	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}
	model := viper.GetString("openai_model")
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// Default Prompts (Hardcoded fallback)
	commitPrompt := fmt.Sprintf(`You are a commit message generator.
Output format: <type>(<scope>): <subject>
Language: Chinese (Simplified).
OS: %s
Return ONLY the commit message.`, runtime.GOOS)

	chatPrompt := fmt.Sprintf("You are MarsX helper. Language: Chinese (Simplified). OS: %s", runtime.GOOS)

	// Try to load from PROMPTS.md
	if content, err := os.ReadFile("PROMPTS.md"); err == nil {
		fullText := string(content)
		// Simple split by headers if possible, but for MVP, let's assume the file
		// is primarily for the System Prompt of the main task (Commit).
		// Or we can just use the whole file as context.
		// Let's stick to a simple logic: If file exists, use it as the BASE for commit generation.
		commitPrompt = fullText
	}

	return &Client{
		ApiKey:       apiKey,
		ApiBase:      strings.TrimRight(apiBase, "/"),
		Model:        model,
		History:      []Message{},
		CommitPrompt: commitPrompt,
		ChatPrompt:   chatPrompt,
	}
}

type Mode int

const (
	ModeCommand Mode = iota
	ModeChat
)

func (c *Client) SendRequest(userQuery string, mode Mode) (string, error) {
	if c.ApiKey == "" {
		return "", fmt.Errorf("API key not found")
	}

	var systemPrompt string
	if mode == ModeCommand {
		systemPrompt = c.CommitPrompt
	} else {
		systemPrompt = c.ChatPrompt
	}

	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}

	// Only append history for Chat mode to avoid polluting commit context
	if mode == ModeChat {
		messages = append(messages, c.History...)
	}

	messages = append(messages, Message{Role: "user", Content: userQuery})

	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.ApiBase+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	result := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// Update history only for chat mode
	if mode == ModeChat {
		c.History = append(c.History, Message{Role: "user", Content: userQuery})
		c.History = append(c.History, Message{Role: "assistant", Content: result})
	}

	return result, nil
}
