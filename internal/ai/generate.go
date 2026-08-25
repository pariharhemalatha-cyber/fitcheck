package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ashokparihar/fitcheck/internal/outfit"
	"github.com/ashokparihar/fitcheck/internal/weather"
)

// GeneratedOutfit is an AI- or rule-produced outfit recommendation.
type GeneratedOutfit struct {
	ItemIDs []string `json:"item_ids"`
	Label   string   `json:"label"`
	Why     string   `json:"why"`
	Score   float64  `json:"score"`
}

// GenerateOutfits builds outfit recommendations. Uses OpenAI when a client is
// available; otherwise delegates to the rule-based outfit generator.
func GenerateOutfits(
	ctx context.Context,
	client *Client,
	profile outfit.Profile,
	items []outfit.Item,
	forecast []weather.DailyForecast,
	plan outfit.Plan,
) ([]GeneratedOutfit, error) {
	filtered := outfit.FilterItems(items, forecast, plan.Formality, profile)
	if len(filtered) == 0 {
		return nil, nil
	}

	if client == nil {
		return ruleBasedOutfits(profile, filtered, forecast, plan), nil
	}

	outfits, err := generateWithAI(ctx, client, profile, filtered, forecast, plan)
	if err != nil {
		return ruleBasedOutfits(profile, filtered, forecast, plan), nil
	}
	return outfits, nil
}

func generateWithAI(
	ctx context.Context,
	client *Client,
	profile outfit.Profile,
	items []outfit.Item,
	forecast []weather.DailyForecast,
	plan outfit.Plan,
) ([]GeneratedOutfit, error) {
	itemSummaries := make([]map[string]any, len(items))
	for i, item := range items {
		itemSummaries[i] = map[string]any{
			"id":            item.ID,
			"name":          item.Name,
			"category":      item.Category,
			"main_color":    item.MainColor,
			"formality":     item.Formality,
			"activity_tags": item.ActivityTags,
		}
	}

	weatherSummaries := make([]map[string]any, len(forecast))
	for i, day := range forecast {
		weatherSummaries[i] = map[string]any{
			"date":               day.Date,
			"temp_high_c":        day.TempHighC,
			"temp_low_c":         day.TempLowC,
			"precip_probability": day.PrecipProbability,
		}
	}

	prompt := fmt.Sprintf(`You are a personal stylist. Given closet items, weather, and trip plan, suggest %d outfits.
Return ONLY a JSON array of objects with: item_ids (array of item id strings from the closet),
label (short day/occasion title), why (1-2 sentences), score (0-10 number).

Profile: style=%s comfort_bias=%d photo_look_bias=%d
Plan: location=%s days=%d activities=%s formality=%s look_goal=%s

Items: %s
Weather: %s`,
		plan.Days,
		profile.StylePrimary,
		profile.ComfortBias,
		profile.PhotoLookBias,
		plan.Location,
		plan.Days,
		strings.Join(plan.Activities, ", "),
		plan.Formality,
		plan.LookGoal,
		mustJSON(itemSummaries),
		mustJSON(weatherSummaries),
	)

	text, err := client.ChatCompletion(ctx, "gpt-4o-mini", []chatMessage{
		{Role: "user", Content: prompt},
	}, 2000)
	if err != nil {
		return nil, err
	}

	text = extractJSON(text)
	if !strings.HasPrefix(strings.TrimSpace(text), "[") {
		// Sometimes model wraps in an object.
		var wrapper struct {
			Outfits []GeneratedOutfit `json:"outfits"`
		}
		if err := json.Unmarshal([]byte(text), &wrapper); err == nil && len(wrapper.Outfits) > 0 {
			return wrapper.Outfits, nil
		}
	}

	var outfits []GeneratedOutfit
	if err := json.Unmarshal([]byte(text), &outfits); err != nil {
		return nil, fmt.Errorf("parse outfit response: %w", err)
	}
	return outfits, nil
}

func ruleBasedOutfits(
	profile outfit.Profile,
	items []outfit.Item,
	forecast []weather.DailyForecast,
	plan outfit.Plan,
) []GeneratedOutfit {
	built := outfit.BuildOutfits(items, forecast, plan, profile)
	out := make([]GeneratedOutfit, len(built))
	for i, o := range built {
		out[i] = GeneratedOutfit{
			ItemIDs: o.ItemIDs,
			Label:   o.Label,
			Why:     o.Why,
			Score:   o.Score,
		}
	}
	return out
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}
