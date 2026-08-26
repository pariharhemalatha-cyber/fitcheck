package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Memory is an in-memory Store for tests.
type Memory struct {
	mu      sync.RWMutex
	items   []Item
	profile StyleProfile
}

func NewMemory() *Memory {
	return &Memory{
		profile: StyleProfile{
			StylePrimary:    "Casual",
			ComfortBias:     7,
			PhotoLookBias:   5,
			NoRepeatTopDays: 3,
		},
	}
}

func (s *Memory) ListItems(category string) []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if category == "" || category == "all" {
		out := make([]Item, len(s.items))
		copy(out, s.items)
		return out
	}
	var filtered []Item
	for _, item := range s.items {
		if item.Category == category {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (s *Memory) GetItem(id string) (Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.ID == id {
			return item, nil
		}
	}
	return Item{}, fmt.Errorf("item not found")
}

func (s *Memory) AddItem(name, category, storagePath string) (Item, error) {
	return s.AddItemWithAttrs(name, category, storagePath, ItemAttrs{Name: name, Category: category})
}

func (s *Memory) AddItemWithAttrs(name, category, storagePath string, attrs ItemAttrs) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if name == "" {
		name = attrs.Name
	}
	if name == "" {
		name = defaultName(category)
	}
	item := Item{
		ID:              uuid.New().String()[:8],
		Name:            name,
		Category:        category,
		MainColor:       attrs.MainColor,
		SecondaryColors: attrs.SecondaryColors,
		Pattern:         attrs.Pattern,
		Material:        attrs.Material,
		Fit:             attrs.Fit,
		Formality:       attrs.Formality,
		SeasonTags:      attrs.SeasonTags,
		RainOK:          attrs.RainOK,
		ActivityTags:    attrs.ActivityTags,
		VibeTags:        attrs.VibeTags,
		StoragePath:     storagePath,
		Emoji:           categoryEmoji(category),
		Status:          "active",
		CreatedAt:       time.Now(),
	}
	s.items = append(s.items, item)
	return item, nil
}

func (s *Memory) UpdateItem(item Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.items {
		if existing.ID == item.ID {
			s.items[i] = item
			return nil
		}
	}
	return fmt.Errorf("item not found")
}

func (s *Memory) GetProfile() StyleProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile
}

func (s *Memory) SaveProfile(p StyleProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profile = p
	return nil
}

func (s *Memory) SaveOutfitSet(ctx context.Context, outfit Outfit) (Outfit, error) {
	if outfit.ID == "" {
		outfit.ID = uuid.New().String()[:8]
	}
	return outfit, nil
}

func (s *Memory) LogWear(outfitID string, itemIDs []string, source string) error {
	return nil
}

func (s *Memory) ListWearEvents(limit int) []WearEvent {
	return nil
}

func (s *Memory) SaveFitCheck(photoPath string, itemIDs []string, score float64, critique string) error {
	return nil
}
