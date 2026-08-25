package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ItemAttributes holds AI-inferred clothing metadata.
type ItemAttributes struct {
	Name             string   `json:"name"`
	Category         string   `json:"category"`
	MainColor        string   `json:"main_color"`
	SecondaryColors  []string `json:"secondary_colors"`
	Pattern          string   `json:"pattern"`
	Material         string   `json:"material"`
	Fit              string   `json:"fit"`
	Formality        int      `json:"formality"`
	SeasonTags       []string `json:"season_tags"`
	RainOK           bool     `json:"rain_ok"`
	ActivityTags     []string `json:"activity_tags"`
	VibeTags         []string `json:"vibe_tags"`
}

const analyzePrompt = `Analyze this clothing item image. Return ONLY valid JSON with these fields:
name, category (one of: tshirt, shirt, pants, shorts, jacket, shoes, accessory),
main_color, secondary_colors (array), pattern, material, fit,
formality (1-10), season_tags (array), rain_ok (boolean),
activity_tags (array), vibe_tags (array).`

// AnalyzeItem infers clothing attributes from an image. Uses OpenAI vision when
// a client is available; otherwise falls back to filename/category heuristics.
func AnalyzeItem(ctx context.Context, client *Client, imagePath string) (ItemAttributes, error) {
	if client != nil {
		attrs, err := analyzeWithVision(ctx, client, imagePath)
		if err == nil {
			return attrs, nil
		}
		// Fall through to heuristics on vision failure.
	}
	return heuristicAnalyze(imagePath), nil
}

func analyzeWithVision(ctx context.Context, client *Client, imagePath string) (ItemAttributes, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return ItemAttributes{}, err
	}

	mime := mimeFromExt(filepath.Ext(imagePath))
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, b64)

	content := []map[string]any{
		{"type": "text", "text": analyzePrompt},
		{"type": "image_url", "image_url": map[string]string{"url": dataURL}},
	}

	text, err := client.ChatCompletion(ctx, "gpt-4o-mini", []chatMessage{
		{Role: "user", Content: content},
	}, 800)
	if err != nil {
		return ItemAttributes{}, err
	}

	text = extractJSON(text)
	var attrs ItemAttributes
	if err := json.Unmarshal([]byte(text), &attrs); err != nil {
		return ItemAttributes{}, fmt.Errorf("parse vision response: %w", err)
	}
	normalizeAttributes(&attrs)
	return attrs, nil
}

func heuristicAnalyze(imagePath string) ItemAttributes {
	base := strings.ToLower(strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath)))
	category := inferCategory(base)
	color := inferColor(base)
	formality := categoryFormality(category)
	season := inferSeason(base, category)
	material := inferMaterial(category, base)

	return ItemAttributes{
		Name:            titleCase(defaultName(category, base)),
		Category:        category,
		MainColor:       color,
		SecondaryColors: secondaryFromBase(base, color),
		Pattern:         inferPattern(base),
		Material:        material,
		Fit:             inferFit(category, base),
		Formality:       formality,
		SeasonTags:      season,
		RainOK:          rainOK(category, material),
		ActivityTags:    activityTags(category, base),
		VibeTags:        vibeTags(category, base),
	}
}

func inferCategory(base string) string {
	rules := []struct {
		keywords []string
		category string
	}{
		{[]string{"tshirt", "tee", "t-shirt", "tank"}, "tshirt"},
		{[]string{"shirt", "blouse", "polo", "button"}, "shirt"},
		{[]string{"pant", "jean", "trouser", "chino", "legging"}, "pants"},
		{[]string{"short"}, "shorts"},
		{[]string{"jacket", "coat", "hoodie", "sweater", "cardigan", "blazer"}, "jacket"},
		{[]string{"shoe", "sneaker", "boot", "sandal", "loafer"}, "shoes"},
		{[]string{"hat", "belt", "scarf", "bag", "watch", "accessory"}, "accessory"},
	}
	for _, rule := range rules {
		for _, kw := range rule.keywords {
			if strings.Contains(base, kw) {
				return rule.category
			}
		}
	}
	return "tshirt"
}

func inferColor(base string) string {
	colors := []string{
		"black", "white", "navy", "blue", "red", "green", "beige", "gray", "grey",
		"brown", "pink", "yellow", "orange", "purple", "cream", "olive", "tan",
	}
	for _, c := range colors {
		if strings.Contains(base, c) {
			if c == "grey" {
				return "gray"
			}
			return c
		}
	}
	return "neutral"
}

func secondaryFromBase(base, main string) []string {
	var out []string
	colors := []string{"black", "white", "navy", "blue", "red", "green", "beige", "gray", "brown"}
	for _, c := range colors {
		if strings.Contains(base, c) && c != main {
			out = append(out, c)
		}
	}
	return out
}

func inferPattern(base string) string {
	patterns := map[string]string{
		"stripe": "striped", "striped": "striped",
		"plaid": "plaid", "check": "checkered", "checkered": "checkered",
		"floral": "floral", "print": "printed", "dot": "dotted",
		"solid": "solid", "plain": "solid",
	}
	for k, v := range patterns {
		if strings.Contains(base, k) {
			return v
		}
	}
	return "solid"
}

func inferMaterial(category, base string) string {
	materials := map[string]string{
		"denim": "denim", "jean": "denim", "cotton": "cotton", "linen": "linen",
		"wool": "wool", "leather": "leather", "synthetic": "synthetic",
		"polyester": "polyester", "silk": "silk", "nylon": "nylon",
	}
	for k, v := range materials {
		if strings.Contains(base, k) {
			return v
		}
	}
	switch category {
	case "jacket":
		return "cotton blend"
	case "shoes":
		return "synthetic"
	default:
		return "cotton"
	}
}

func inferFit(category, base string) string {
	if strings.Contains(base, "slim") || strings.Contains(base, "fitted") {
		return "slim"
	}
	if strings.Contains(base, "relaxed") || strings.Contains(base, "oversized") {
		return "relaxed"
	}
	switch category {
	case "pants", "shorts":
		return "regular"
	case "jacket":
		return "regular"
	default:
		return "regular"
	}
}

func categoryFormality(category string) int {
	switch category {
	case "tshirt", "shorts":
		return 3
	case "pants", "shoes":
		return 5
	case "shirt":
		return 6
	case "jacket":
		return 7
	case "accessory":
		return 4
	default:
		return 5
	}
}

func inferSeason(base, category string) []string {
	if strings.Contains(base, "winter") || strings.Contains(base, "wool") {
		return []string{"winter", "fall"}
	}
	if strings.Contains(base, "summer") || category == "shorts" {
		return []string{"summer", "spring"}
	}
	return []string{"spring", "summer", "fall"}
}

func rainOK(category, material string) bool {
	if category == "jacket" {
		return strings.Contains(material, "nylon") || strings.Contains(material, "synthetic")
	}
	return category != "shoes" || strings.Contains(material, "leather")
}

func activityTags(category, base string) []string {
	tags := []string{"casual"}
	if strings.Contains(base, "gym") || strings.Contains(base, "sport") {
		tags = append(tags, "active")
	}
	if category == "shirt" || category == "jacket" {
		tags = append(tags, "dining")
	}
	return tags
}

func vibeTags(category, base string) []string {
	if strings.Contains(base, "minimal") {
		return []string{"minimal", "clean"}
	}
	switch category {
	case "shirt", "jacket":
		return []string{"polished", "smart-casual"}
	case "tshirt", "shorts":
		return []string{"relaxed", "easygoing"}
	default:
		return []string{"versatile", "everyday"}
	}
}

func defaultName(category, base string) string {
	if base != "" && base != "image" && base != "photo" && base != "img" {
		return strings.ReplaceAll(base, "_", " ")
	}
	switch category {
	case "tshirt":
		return "T-shirt"
	case "shirt":
		return "Shirt"
	case "pants":
		return "Pants"
	case "shorts":
		return "Shorts"
	case "jacket":
		return "Jacket"
	case "shoes":
		return "Shoes"
	default:
		return "Clothing item"
	}
}

func normalizeAttributes(a *ItemAttributes) {
	a.Category = strings.ToLower(strings.TrimSpace(a.Category))
	if a.Formality < 1 {
		a.Formality = 1
	}
	if a.Formality > 10 {
		a.Formality = 10
	}
	if a.Name == "" {
		a.Name = defaultName(a.Category, "")
	}
}

func mimeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) >= 2 {
			s = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
