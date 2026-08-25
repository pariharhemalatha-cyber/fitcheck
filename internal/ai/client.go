package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openAIBaseURL = "https://api.openai.com/v1"

// Client wraps OpenAI HTTP calls. Returns nil when no API key is configured.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a configured client or an error if OPENAI_API_KEY is empty.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai api key not configured")
	}
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}, nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletion sends a chat request and returns the assistant text.
func (c *Client) ChatCompletion(ctx context.Context, model string, messages []chatMessage, maxTokens int) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: 0.4,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("openai: %s", out.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("openai status %d: %s", resp.StatusCode, string(raw))
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response")
	}
	return out.Choices[0].Message.Content, nil
}
