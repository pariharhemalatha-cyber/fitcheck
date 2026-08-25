package outfit

import (
	"fmt"
	"sort"

	"github.com/ashokparihar/fitcheck/internal/weather"
)

// PackingList groups item IDs by category for trip packing.
type PackingList struct {
	Categories []PackingCategory `json:"categories"`
	ItemIDs    []string          `json:"item_ids"`
}

// PackingCategory is a labeled group of packed item IDs.
type PackingCategory struct {
	Label   string   `json:"label"`
	ItemIDs []string `json:"item_ids"`
}

// DayPackingOutfit maps a trip day to a full outfit using real item IDs.
type DayPackingOutfit struct {
	Day     int      `json:"day"`
	ItemIDs []string `json:"item_ids"`
	Label   string   `json:"label"`
}

// PackingSolution is the output of the packing solver.
type PackingSolution struct {
	Packing     PackingList        `json:"packing"`
	DayOutfits  []DayPackingOutfit `json:"day_outfits"`
	TotalPieces int                `json:"total_pieces"`
}

// SolvePacking selects a minimal wardrobe and day-by-day outfits for a trip.
func SolvePacking(
	items []Item,
	days int,
	laundry bool,
	luggage string,
	forecast []weather.DailyForecast,
) PackingSolution {
	if days < 1 {
		days = 1
	}

	capacity := luggageCapacity(luggage)
	plan := Plan{Days: days, Formality: "casual", LookGoal: "balanced"}
	profile := Profile{ComfortBias: 7, PhotoLookBias: 5, NoRepeatTopDays: 3}
	if laundry {
		profile.NoRepeatTopDays = max(1, (days+1)/2)
	}

	filtered := FilterItems(items, forecast, plan.Formality, profile)
	if len(filtered) == 0 {
		filtered = items
	}

	dayOutfits := BuildOutfits(filtered, forecast, plan, profile)
	selected := selectPackingItems(filtered, dayOutfits, days, laundry, capacity, forecast)

	packing := buildPackingList(selected)
	allIDs := uniqueIDs(selected)

	return PackingSolution{
		Packing:     packing,
		DayOutfits:  toDayOutfits(dayOutfits, days),
		TotalPieces: len(allIDs),
	}
}

func luggageCapacity(luggage string) int {
	switch luggage {
	case "checked", "checked_bag", "large":
		return 40
	case "personal", "personal_item", "small":
		return 12
	default: // carry_on
		return 20
	}
}

func selectPackingItems(
	all []Item,
	outfits []BuiltOutfit,
	days int,
	laundry bool,
	capacity int,
	forecast []weather.DailyForecast,
) []Item {
	needed := map[string]int{
		"tshirt": 1, "shirt": 1, "pants": 1, "shorts": 1,
		"shoes": 1, "jacket": 0,
	}

	topCount := days
	if laundry {
		topCount = (days + 1) / 2
		if topCount < 2 {
			topCount = 2
		}
	}
	needed["tshirt"] = topCount
	needed["shirt"] = min(2, topCount)

	avgTemp, rainy := summarizeWeather(forecast)
	_ = avgTemp
	if rainy > 0 {
		needed["jacket"] = 1
	}
	if avgTemp >= 24 {
		needed["shorts"] = 1
		needed["pants"] = 1
	} else {
		needed["pants"] = min(2, max(1, days/3+1))
		needed["shorts"] = 0
	}

	// Always include items referenced in day outfits.
	idSet := map[string]Item{}
	for _, o := range outfits {
		for _, id := range o.ItemIDs {
			for _, item := range all {
				if item.ID == id {
					idSet[id] = item
				}
			}
		}
	}

	byCategory := groupByCategory(all)
	for cat, count := range needed {
		if count == 0 {
			continue
		}
		picked := 0
		for _, item := range byCategory[cat] {
			if picked >= count {
				break
			}
			if _, exists := idSet[item.ID]; !exists {
				idSet[item.ID] = item
				picked++
			}
		}
	}

	selected := make([]Item, 0, len(idSet))
	for _, item := range idSet {
		selected = append(selected, item)
	}

	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Category == selected[j].Category {
			return selected[i].Name < selected[j].Name
		}
		return selected[i].Category < selected[j].Category
	})

	if len(selected) > capacity {
		selected = trimToCapacity(selected, outfits, capacity)
	}
	return selected
}

func trimToCapacity(items []Item, outfits []BuiltOutfit, capacity int) []Item {
	required := map[string]bool{}
	for _, o := range outfits {
		for _, id := range o.ItemIDs {
			required[id] = true
		}
	}

	var keep, optional []Item
	for _, item := range items {
		if required[item.ID] {
			keep = append(keep, item)
		} else {
			optional = append(optional, item)
		}
	}

	for len(keep) < capacity && len(optional) > 0 {
		keep = append(keep, optional[0])
		optional = optional[1:]
	}
	if len(keep) > capacity {
		return keep[:capacity]
	}
	return keep
}

func buildPackingList(items []Item) PackingList {
	labels := map[string]string{
		"tshirt": "Tops", "shirt": "Tops", "pants": "Bottoms", "shorts": "Bottoms",
		"jacket": "Layers", "shoes": "Shoes", "accessory": "Accessories",
	}
	grouped := map[string][]string{}
	var allIDs []string

	for _, item := range items {
		label := labels[item.Category]
		if label == "" {
			label = "Other"
		}
		grouped[label] = append(grouped[label], item.ID)
		allIDs = append(allIDs, item.ID)
	}

	order := []string{"Tops", "Bottoms", "Shoes", "Layers", "Accessories", "Other"}
	var categories []PackingCategory
	for _, label := range order {
		ids, ok := grouped[label]
		if !ok || len(ids) == 0 {
			continue
		}
		categories = append(categories, PackingCategory{
			Label:   label,
			ItemIDs: ids,
		})
	}

	return PackingList{
		Categories: categories,
		ItemIDs:    uniqueStrings(allIDs),
	}
}

func toDayOutfits(built []BuiltOutfit, days int) []DayPackingOutfit {
	if len(built) == 0 {
		return nil
	}
	n := days
	if len(built) < n {
		n = len(built)
	}
	out := make([]DayPackingOutfit, n)
	for i := 0; i < n; i++ {
		out[i] = DayPackingOutfit{
			Day:     i + 1,
			ItemIDs: built[i].ItemIDs,
			Label:   built[i].Label,
		}
	}
	return out
}

func uniqueIDs(items []Item) []string {
	seen := map[string]bool{}
	var out []string
	for _, item := range items {
		if !seen[item.ID] {
			seen[item.ID] = true
			out = append(out, item.ID)
		}
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// FormatPackingSummary returns a human-readable packing count string.
func FormatPackingSummary(s PackingSolution) string {
	return fmt.Sprintf("%d pieces across %d categories for %d days",
		s.TotalPieces, len(s.Packing.Categories), len(s.DayOutfits))
}
