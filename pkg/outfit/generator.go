package outfit

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ashokparihar/fitcheck/pkg/weather"
)

// Item represents a closet piece with attributes used for outfit logic.
type Item struct {
	ID              string
	Name            string
	Category        string
	MainColor       string
	SecondaryColors []string
	Formality       int
	SeasonTags      []string
	RainOK          bool
	ActivityTags    []string
	VibeTags        []string
}

// Profile captures user style preferences.
type Profile struct {
	StylePrimary    string
	ComfortBias     int
	PhotoLookBias   int
	NoRepeatTopDays int
}

// Plan describes the trip or styling session context.
type Plan struct {
	Location   string
	Days       int
	Activities []string
	Formality  string
	LookGoal   string
}

// BuiltOutfit is a rule-generated outfit with real item IDs.
type BuiltOutfit struct {
	ItemIDs []string
	Label   string
	Why     string
	Score   float64
}

var topCategories = []string{"tshirt", "shirt", "jacket"}
var bottomCategories = []string{"pants", "shorts"}
var requiredCategories = []string{"shoes"}

// FilterItems returns closet items appropriate for weather, formality, and profile.
func FilterItems(items []Item, forecast []weather.DailyForecast, formality string, profile Profile) []Item {
	targetFormality := formalityLevel(formality, profile)
	avgTemp, rainyDays := summarizeWeather(forecast)

	var filtered []Item
	for _, item := range items {
		if !matchesFormality(item, targetFormality) {
			continue
		}
		if !matchesWeather(item, avgTemp, rainyDays) {
			continue
		}
		if !matchesProfile(item, profile) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

// BuildOutfits composes day-by-day outfits from filtered items.
func BuildOutfits(items []Item, forecast []weather.DailyForecast, plan Plan, profile Profile) []BuiltOutfit {
	if len(items) == 0 {
		return nil
	}

	days := plan.Days
	if days < 1 {
		days = 1
	}
	if len(forecast) > 0 && len(forecast) < days {
		days = len(forecast)
	}

	byCategory := groupByCategory(items)
	usedTops := make(map[string]int)

	var outfits []BuiltOutfit
	for day := 0; day < days; day++ {
		var dayWeather weather.DailyForecast
		if day < len(forecast) {
			dayWeather = forecast[day]
		}

		top := pickTop(byCategory, dayWeather, usedTops, profile.NoRepeatTopDays)
		bottom := pickBottom(byCategory, dayWeather)
		shoes := pickFirst(byCategory["shoes"])
		layer := pickLayer(byCategory, dayWeather)

		var ids []string
		if top != nil {
			ids = append(ids, top.ID)
			usedTops[top.ID] = day
		}
		if bottom != nil {
			ids = append(ids, bottom.ID)
		}
		if shoes != nil {
			ids = append(ids, shoes.ID)
		}
		if layer != nil {
			ids = append(ids, layer.ID)
		}

		if len(ids) == 0 {
			continue
		}

		label := dayLabel(day+1, plan)
		why := explainOutfit(top, bottom, shoes, layer, dayWeather, plan)
		score := scoreOutfit(top, bottom, shoes, layer, dayWeather, profile)

		outfits = append(outfits, BuiltOutfit{
			ItemIDs: ids,
			Label:   label,
			Why:     why,
			Score:   score,
		})
	}
	return outfits
}

func formalityLevel(formality string, profile Profile) int {
	switch strings.ToLower(formality) {
	case "formal", "dressy":
		return 8
	case "smart", "smart_casual", "smart-casual":
		return 6
	case "casual":
		return 4
	default:
		if profile.PhotoLookBias >= 7 {
			return 6
		}
		return 5
	}
}

func summarizeWeather(forecast []weather.DailyForecast) (avgTemp float64, rainyDays int) {
	if len(forecast) == 0 {
		return 20, 0
	}
	var sum float64
	for _, d := range forecast {
		sum += (d.TempHighC + d.TempLowC) / 2
		if d.IsRainy() {
			rainyDays++
		}
	}
	return sum / float64(len(forecast)), rainyDays
}

func matchesFormality(item Item, target int) bool {
	diff := abs(item.Formality - target)
	return diff <= 3
}

func matchesWeather(item Item, avgTemp float64, rainyDays int) bool {
	if rainyDays > 0 && (item.Category == "shoes" || item.Category == "jacket") && !item.RainOK {
		// Still allow but prefer rain-ok items in picker; don't hard exclude all.
	}
	if avgTemp >= 26 && item.Category == "jacket" && !containsTag(item.SeasonTags, "summer") {
		return false
	}
	if avgTemp < 10 && item.Category == "shorts" {
		return false
	}
	return true
}

func matchesProfile(item Item, profile Profile) bool {
	style := strings.ToLower(profile.StylePrimary)
	for _, tag := range item.VibeTags {
		if strings.Contains(strings.ToLower(tag), style) {
			return true
		}
	}
	// Permissive default — most items pass unless strongly mismatched.
	if profile.ComfortBias >= 8 && item.Category == "shirt" && item.Formality >= 8 {
		return false
	}
	return true
}

func groupByCategory(items []Item) map[string][]Item {
	m := make(map[string][]Item)
	for _, item := range items {
		m[item.Category] = append(m[item.Category], item)
	}
	for cat := range m {
		sort.Slice(m[cat], func(i, j int) bool {
			return m[cat][i].Formality > m[cat][j].Formality
		})
	}
	return m
}

func pickTop(byCategory map[string][]Item, day weather.DailyForecast, used map[string]int, noRepeatDays int) *Item {
	candidates := append(append([]Item{}, byCategory["shirt"]...), byCategory["tshirt"]...)
	if len(candidates) == 0 {
		return nil
	}

	if day.IsHot() {
		candidates = preferCategory(candidates, "tshirt")
	}

	for i := range candidates {
		item := &candidates[i]
		if lastUsed, ok := used[item.ID]; ok && noRepeatDays > 0 {
			if len(used)-lastUsed < noRepeatDays {
				continue
			}
		}
		return item
	}
	return &candidates[0]
}

func pickBottom(byCategory map[string][]Item, day weather.DailyForecast) *Item {
	if day.IsHot() || day.TempHighC >= 24 {
		if item := pickFirst(byCategory["shorts"]); item != nil {
			return item
		}
	}
	return pickFirst(byCategory["pants"])
}

func pickLayer(byCategory map[string][]Item, day weather.DailyForecast) *Item {
	if day.IsCold() || day.IsRainy() {
		return pickFirst(byCategory["jacket"])
	}
	return nil
}

func pickFirst(items []Item) *Item {
	if len(items) == 0 {
		return nil
	}
	return &items[0]
}

func preferCategory(items []Item, category string) []Item {
	var preferred, rest []Item
	for _, item := range items {
		if item.Category == category {
			preferred = append(preferred, item)
		} else {
			rest = append(rest, item)
		}
	}
	return append(preferred, rest...)
}

func dayLabel(day int, plan Plan) string {
	activity := "exploring"
	if len(plan.Activities) > 0 {
		activity = plan.Activities[(day-1)%len(plan.Activities)]
	}
	if plan.Location != "" {
		return fmt.Sprintf("Day %d — %s in %s", day, activity, plan.Location)
	}
	return fmt.Sprintf("Day %d — %s", day, activity)
}

func explainOutfit(top, bottom, shoes, layer *Item, day weather.DailyForecast, plan Plan) string {
	parts := []string{"Comfortable combination from your closet"}
	if day.IsHot() {
		parts = append(parts, "light layers for warm weather")
	} else if day.IsCold() {
		parts = append(parts, "layered for cooler temps")
	}
	if day.IsRainy() {
		parts = append(parts, "rain-friendly pieces where possible")
	}
	if strings.ToLower(plan.LookGoal) == "photos" {
		parts = append(parts, "neutral tones photograph well")
	}
	return strings.Join(parts, ", ") + "."
}

func scoreOutfit(top, bottom, shoes, layer *Item, day weather.DailyForecast, profile Profile) float64 {
	score := 7.0
	if top != nil && bottom != nil && shoes != nil {
		score += 1.0
	}
	if layer != nil && (day.IsCold() || day.IsRainy()) {
		score += 0.5
	}
	if top != nil && bottom != nil && colorsHarmonize(top.MainColor, bottom.MainColor) {
		score += 0.5
	}
	comfort := float64(profile.ComfortBias) / 10
	score = score*0.7 + (score+comfort)*0.3
	if score > 10 {
		score = 10
	}
	return math.Round(score*10) / 10
}

func colorsHarmonize(a, b string) bool {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	if a == b {
		return true
	}
	neutrals := map[string]bool{"black": true, "white": true, "gray": true, "beige": true, "navy": true, "neutral": true}
	return neutrals[a] || neutrals[b]
}

func containsTag(tags []string, want string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// HasRequiredCategories reports whether items cover minimum outfit slots.
func HasRequiredCategories(items []Item) bool {
	cats := map[string]bool{}
	for _, item := range items {
		cats[item.Category] = true
	}
	for _, cat := range append(append(topCategories, bottomCategories...), requiredCategories...) {
		if !cats[cat] && (cat == "shoes" || cat == "pants" || cat == "tshirt" || cat == "shirt") {
			if cat == "pants" && cats["shorts"] {
				continue
			}
			if (cat == "tshirt" || cat == "shirt") && (cats["tshirt"] || cats["shirt"]) {
				continue
			}
			if cat == "shoes" {
				return false
			}
		}
	}
	return len(cats) > 0
}
