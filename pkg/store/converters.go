package store

import (
	"github.com/ashokparihar/fitcheck/pkg/outfit"
)

func ToOutfitItem(item Item) outfit.Item {
	return outfit.Item{
		ID:              item.ID,
		Name:            item.Name,
		Category:        item.Category,
		MainColor:       item.MainColor,
		SecondaryColors: item.SecondaryColors,
		Formality:       item.Formality,
		SeasonTags:      item.SeasonTags,
		RainOK:          item.RainOK,
		ActivityTags:    item.ActivityTags,
		VibeTags:        item.VibeTags,
	}
}

func ToOutfitItems(items []Item) []outfit.Item {
	out := make([]outfit.Item, len(items))
	for i, item := range items {
		out[i] = ToOutfitItem(item)
	}
	return out
}

func ToOutfitProfile(p StyleProfile) outfit.Profile {
	return outfit.Profile{
		StylePrimary:    p.StylePrimary,
		ComfortBias:     p.ComfortBias,
		PhotoLookBias:   p.PhotoLookBias,
		NoRepeatTopDays: p.NoRepeatTopDays,
	}
}

func ItemsByID(items []Item) map[string]Item {
	m := make(map[string]Item, len(items))
	for _, item := range items {
		m[item.ID] = item
	}
	return m
}

func OutfitItemsFromIDs(all []Item, ids []string) []OutfitItem {
	byID := ItemsByID(all)
	var out []OutfitItem
	for _, id := range ids {
		if item, ok := byID[id]; ok {
			out = append(out, OutfitItem{
				ID:        item.ID,
				Name:      item.Name,
				Emoji:     item.Emoji,
				ImageURL:  imageURLFromPath(item.StoragePath),
				Category:  item.Category,
				MainColor: item.MainColor,
			})
		}
	}
	return out
}

func imageURLFromPath(path string) string {
	if path == "" {
		return ""
	}
	return stringsTrimPrefix(path, "/uploads/")
}

func stringsTrimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}
