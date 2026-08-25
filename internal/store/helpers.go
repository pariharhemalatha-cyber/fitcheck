package store

import (
	"fmt"

	"github.com/google/uuid"
)

func generateOutfitsFromItems(items []Item, days int) []Outfit {
	if len(items) == 0 {
		return nil
	}

	outfits := []Outfit{
		{
			ID:    uuid.New().String(),
			Label: "Day 1 — Travel + sightseeing",
			Score: 8.7,
			Why:   "Comfortable for walking, appropriate for warm weather, and neutral colors work well together.",
			Items: pickItems(items, []string{"tshirt", "pants", "shoes"}),
		},
		{
			ID:    uuid.New().String(),
			Label: "Day 2 — City + dinner",
			Score: 8.9,
			Why:   "Slightly more dressed-up for dinner while still comfortable.",
			Items: pickItems(items, []string{"shirt", "pants", "shoes"}),
		},
	}

	if days <= 1 {
		return outfits[:1]
	}
	return outfits
}

func generateTripPlan(days int, laundry string) ([]PackingCategory, []DayOutfit) {
	topCount := days
	if laundry == "yes" {
		topCount = (days + 1) / 2
		if topCount < 2 {
			topCount = 2
		}
	}

	packing := []PackingCategory{
		{Label: "Tops", Items: []string{
			stringCount(topCount, "T-shirt"),
			"1 shirt",
		}},
		{Label: "Bottoms", Items: []string{
			"2 pants",
			"1 shorts",
		}},
		{Label: "Shoes", Items: []string{
			"White sneakers",
			"Casual shoes",
		}},
		{Label: "Layers", Items: []string{
			"1 jacket",
		}},
	}

	var daysOut []DayOutfit
	for i := 1; i <= days; i++ {
		daysOut = append(daysOut, DayOutfit{
			Day:         i,
			Description: dayDescription(i),
		})
	}

	return packing, daysOut
}

func pickItems(items []Item, categories []string) []OutfitItem {
	var result []OutfitItem
	for _, cat := range categories {
		for _, item := range items {
			if item.Category == cat {
				result = append(result, OutfitItem{Name: item.Name, Emoji: item.Emoji})
				break
			}
		}
		if len(result) == len(categories) {
			break
		}
	}
	if len(result) == 0 && len(items) > 0 {
		for i, item := range items {
			if i >= 3 {
				break
			}
			result = append(result, OutfitItem{Name: item.Name, Emoji: item.Emoji})
		}
	}
	return result
}

func defaultName(category string) string {
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
		return "Item"
	}
}

func categoryEmoji(category string) string {
	switch category {
	case "tshirt":
		return "👕"
	case "shirt":
		return "👔"
	case "pants":
		return "👖"
	case "shorts":
		return "🩳"
	case "jacket":
		return "🧥"
	case "shoes":
		return "👟"
	default:
		return "🧢"
	}
}

func stringCount(n int, label string) string {
	if n == 1 {
		return "1 " + label
	}
	return fmt.Sprintf("%d %ss", n, label)
}

func dayDescription(day int) string {
	descriptions := []string{
		"Black T-shirt + beige pants + sneakers",
		"Blue shirt + black pants + sneakers",
		"White T-shirt + shorts + sneakers",
		"Green T-shirt + beige pants + casual shoes",
		"Black shirt + black pants + sneakers",
	}
	if day-1 < len(descriptions) {
		return descriptions[day-1]
	}
	return "Casual outfit from your closet"
}
