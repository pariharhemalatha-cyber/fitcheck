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

const (
	ProviderGemini = "gemini"
	ProviderOpenAI = "openai"

	geminiOpenAIBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
	openAIBaseURL       = "https://api.openai.com/v1"

	DefaultGeminiVisionModel = "gemini-2.5-flash"
	DefaultGeminiTextModel   = "gemini-2.5-flash"
	DefaultOpenAIVisionModel = "gpt-4o-mini"
	DefaultOpenAITextModel   = "gpt-4o-mini"
)

// Client wraps OpenAI-compatible chat API calls (Gemini, OpenAI, etc.).
type Client struct {
	Provider    string
	apiKey      string
	baseURL     string
	visionModel string
	textModel   string
	httpClient  *http.Client
}

type ClientConfig struct {
	Provider    string
	APIKey      string
	BaseURL     string
	VisionModel string
	TextModel   string
}

// NewClient creates a client from explicit config.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key not configured")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = openAIBaseURL
	}
	if cfg.VisionModel == "" {
		cfg.VisionModel = DefaultOpenAIVisionModel
	}
	if cfg.TextModel == "" {
		cfg.TextModel = cfg.VisionModel
	}
	if cfg.Provider == "" {
		cfg.Provider = ProviderOpenAI
	}
	return &Client{
		Provider:    cfg.Provider,
		apiKey:      cfg.APIKey,
		baseURL:     cfg.BaseURL,
		visionModel: cfg.VisionModel,
		textModel:   cfg.TextModel,
		httpClient:  &http.Client{Timeout: 90 * time.Second},
	}, nil
}

// NewGeminiClient creates a Google Gemini client (free tier, vision + text).
func NewGeminiClient(apiKey string) (*Client, error) {
	return NewClient(ClientConfig{
		Provider:    ProviderGemini,
		APIKey:      apiKey,
		BaseURL:     geminiOpenAIBaseURL,
		VisionModel: DefaultGeminiVisionModel,
		TextModel:   DefaultGeminiTextModel,
	})
}

// NewOpenAIClient creates an OpenAI client (paid).
func NewOpenAIClient(apiKey string) (*Client, error) {
	return NewClient(ClientConfig{
		Provider:    ProviderOpenAI,
		APIKey:      apiKey,
		BaseURL:     openAIBaseURL,
		VisionModel: DefaultOpenAIVisionModel,
		TextModel:   DefaultOpenAITextModel,
	})
}

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type chatRequest struct {
	Model           string        `json:"model"`
	Messages        []chatMessage `json:"messages"`
	MaxTokens       int           `json:"max_tokens,omitempty"`
	Temperature     float64       `json:"temperature,omitempty"`
	ReasoningEffort string        `json:"reasoning_effort,omitempty"` // Gemini 2.5: "none" avoids empty output on low max_tokens
}

type chatResponse struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletion sends a chat request with an explicit model.
func (c *Client) ChatCompletion(ctx context.Context, model string, messages []chatMessage, maxTokens int) (string, error) {
	if model == "" {
		model = c.textModel
	}
	reqBody := chatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: 0.4,
	}
	// Gemini 2.5 counts internal "thinking" tokens against max_tokens; disable for JSON/vision tasks.
	if c.Provider == ProviderGemini {
		reqBody.ReasoningEffort = "none"
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
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

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("%s status %d: %s", c.Provider, resp.StatusCode, parseAPIError(raw))
	}

	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		// Gemini sometimes returns errors as a JSON array.
		if msg := parseAPIError(raw); msg != "" {
			return "", fmt.Errorf("%s: %s", c.Provider, msg)
		}
		return "", fmt.Errorf("decode ai response: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("%s: %s", c.Provider, out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("%s: empty response", c.Provider)
	}
	content := out.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("%s: empty message content (finish_reason=%s)", c.Provider, out.Choices[0].FinishReason)
	}
	return content, nil
}

func parseAPIError(raw []byte) string {
	var obj struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &obj) == nil && obj.Error.Message != "" {
		return obj.Error.Message
	}
	var arr []struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 && arr[0].Error.Message != "" {
		return arr[0].Error.Message
	}
	return string(raw)
}

// VisionCompletion uses the configured vision model.
func (c *Client) VisionCompletion(ctx context.Context, messages []chatMessage, maxTokens int) (string, error) {
	return c.ChatCompletion(ctx, c.visionModel, messages, maxTokens)
}

// TextCompletion uses the configured text model.
func (c *Client) TextCompletion(ctx context.Context, messages []chatMessage, maxTokens int) (string, error) {
	return c.ChatCompletion(ctx, c.textModel, messages, maxTokens)
}
