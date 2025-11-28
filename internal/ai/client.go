package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type Client struct {
	ApiKey  string
	ApiBase string
	Model   string
	History []Message // Keep chat history
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

	return &Client{
		ApiKey:  apiKey,
		ApiBase: strings.TrimRight(apiBase, "/"),
		Model:   model,
		History: []Message{},
	}
}

// Mode defines whether we want a command or a chat response
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
		systemPrompt = `You are a professional developer tool that generates Git commit messages.
Your task is to analyze the provided "git diff" output and generate a concise, standard commit message.
RULES:
1. Use Conventional Commits format (e.g., "feat: add login", "fix: typo in readme").
2. Subject line must be under 50 chars.
3. Return ONLY the commit message text. No explanations.
4. If the diff is empty or unclear, describe what you see concisely.`
	} else {
		systemPrompt = `You are MarsX, an intelligent coding assistant.
You can explain git diffs, answer technical questions, or help with coding tasks.
Be concise and helpful. Use Markdown.`
	}

	// Construct messages
	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}
	// Append history (TODO: Limit history size if needed)
	messages = append(messages, c.History...)
	// Append current user query
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

	// Update history
	c.History = append(c.History, Message{Role: "user", Content: userQuery})
	c.History = append(c.History, Message{Role: "assistant", Content: result})

	return result, nil
}
