package ai

import (
	"log"

	"github.com/ashokparihar/fitcheck/internal/config"
)

// NewClientFromConfig picks the best available AI provider.
// Priority: Gemini (free) → OpenAI (paid) → nil (heuristics only).
func NewClientFromConfig(cfg *config.Config) (*Client, error) {
	if cfg.GeminiAPIKey != "" {
		client, err := NewGeminiClient(cfg.GeminiAPIKey)
		if err != nil {
			return nil, err
		}
		log.Printf("AI: Google Gemini (%s) — free tier, vision enabled", DefaultGeminiVisionModel)
		return client, nil
	}
	if cfg.OpenAIAPIKey != "" {
		client, err := NewOpenAIClient(cfg.OpenAIAPIKey)
		if err != nil {
			return nil, err
		}
		log.Printf("AI: OpenAI (%s)", DefaultOpenAIVisionModel)
		return client, nil
	}
	log.Printf("AI: no API key — using rule-based fallbacks (add GEMINI_API_KEY for free vision)")
	return nil, nil
}
