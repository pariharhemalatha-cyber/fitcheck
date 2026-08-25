package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ashokparihar/fitcheck/internal/outfit"
)

// FitCheckResult holds a selfie-based outfit evaluation.
type FitCheckResult struct {
	Score          float64      `json:"score"`
	Critique       string       `json:"critique"`
	SuggestedSwaps []SwapSuggestion `json:"suggested_swaps"`
}

// SwapSuggestion recommends replacing one item with another from the closet.
type SwapSuggestion struct {
	FromItemID string `json:"from_item_id"`
	ToItemID   string `json:"to_item_id"`
	Reason     string `json:"reason"`
}

// AnalyzeFitCheck scores how well the worn outfit matches the planned items.
// Uses OpenAI vision when available; otherwise applies heuristic scoring.
func AnalyzeFitCheck(
	ctx context.Context,
	client *Client,
	selfiePath string,
	itemIDs []string,
	items []outfit.Item,
) (FitCheckResult, error) {
	if client != nil {
		result, err := fitCheckWithVision(ctx, client, selfiePath, itemIDs, items)
		if err == nil {
			return result, nil
		}
	}
	return heuristicFitCheck(itemIDs, items), nil
}

func fitCheckWithVision(
	ctx context.Context,
	client *Client,
	selfiePath string,
	itemIDs []string,
	items []outfit.Item,
) (FitCheckResult, error) {
	data, err := os.ReadFile(selfiePath)
	if err != nil {
		return FitCheckResult{}, err
	}

	mime := mimeFromExt(filepath.Ext(selfiePath))
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, b64)

	itemDesc := make([]map[string]string, 0, len(items))
	for _, item := range items {
		for _, id := range itemIDs {
			if item.ID == id {
				itemDesc = append(itemDesc, map[string]string{
					"id":       item.ID,
					"name":     item.Name,
					"category": item.Category,
					"color":    item.MainColor,
				})
			}
		}
	}

	prompt := fmt.Sprintf(`Rate this outfit selfie against the intended items. Return ONLY JSON:
{"score": 0-10, "critique": "2-3 sentences", "suggested_swaps": [{"from_item_id":"","to_item_id":"","reason":""}]}
Intended items: %s`, mustJSON(itemDesc))

	content := []map[string]any{
		{"type": "text", "text": prompt},
		{"type": "image_url", "image_url": map[string]string{"url": dataURL}},
	}

	text, err := client.ChatCompletion(ctx, "gpt-4o-mini", []chatMessage{
		{Role: "user", Content: content},
	}, 600)
	if err != nil {
		return FitCheckResult{}, err
	}

	text = extractJSON(text)
	var result FitCheckResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return FitCheckResult{}, fmt.Errorf("parse fitcheck response: %w", err)
	}
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > 10 {
		result.Score = 10
	}
	return result, nil
}

func heuristicFitCheck(itemIDs []string, items []outfit.Item) FitCheckResult {
	selected := make(map[string]outfit.Item)
	for _, item := range items {
		for _, id := range itemIDs {
			if item.ID == id {
				selected[id] = item
			}
		}
	}

	score := 7.0
	var notes []string
	categories := map[string]bool{}

	for _, item := range selected {
		categories[item.Category] = true
	}

	if !categories["shoes"] {
		score -= 1.5
		notes = append(notes, "No shoes detected in the planned outfit.")
	}
	if !categories["pants"] && !categories["shorts"] {
		score -= 1.0
		notes = append(notes, "Missing a bottom layer.")
	}
	if !categories["tshirt"] && !categories["shirt"] && !categories["jacket"] {
		score -= 1.0
		notes = append(notes, "Missing a top layer.")
	}

	// Color harmony bonus.
	colors := make([]string, 0, len(selected))
	for _, item := range selected {
		colors = append(colors, strings.ToLower(item.MainColor))
	}
	if hasNeutralBase(colors) {
		score += 0.5
		notes = append(notes, "Neutral palette works well together.")
	}

	formalitySpread := formalityRange(selected)
	if formalitySpread <= 2 {
		score += 0.5
	} else if formalitySpread > 4 {
		score -= 0.5
		notes = append(notes, "Formality levels feel mismatched.")
	}

	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}

	critique := "Solid foundation — the planned pieces form a coherent outfit."
	if len(notes) > 0 {
		critique = strings.Join(notes, " ")
	}

	swaps := suggestSwaps(selected, items)
	return FitCheckResult{
		Score:          roundScore(score),
		Critique:       critique,
		SuggestedSwaps: swaps,
	}
}

func hasNeutralBase(colors []string) bool {
	neutrals := map[string]bool{"black": true, "white": true, "gray": true, "beige": true, "navy": true, "neutral": true}
	for _, c := range colors {
		if neutrals[c] {
			return true
		}
	}
	return false
}

func formalityRange(selected map[string]outfit.Item) int {
	minF, maxF := 10, 0
	for _, item := range selected {
		if item.Formality < minF {
			minF = item.Formality
		}
		if item.Formality > maxF {
			maxF = item.Formality
		}
	}
	return maxF - minF
}

func suggestSwaps(selected map[string]outfit.Item, closet []outfit.Item) []SwapSuggestion {
	var swaps []SwapSuggestion
	for _, worn := range selected {
		for _, alt := range closet {
			if alt.ID == worn.ID || alt.Category != worn.Category {
				continue
			}
			if alt.Formality > worn.Formality && isNeutral(alt.MainColor) {
				swaps = append(swaps, SwapSuggestion{
					FromItemID: worn.ID,
					ToItemID:   alt.ID,
					Reason:     fmt.Sprintf("Swap to %s for a slightly sharper look.", alt.Name),
				})
				break
			}
		}
	}
	if len(swaps) > 2 {
		swaps = swaps[:2]
	}
	return swaps
}

func isNeutral(color string) bool {
	switch strings.ToLower(color) {
	case "black", "white", "gray", "beige", "navy", "neutral":
		return true
	default:
		return false
	}
}

func roundScore(s float64) float64 {
	return float64(int(s*10+0.5)) / 10
}
