package store

import (
	"context"
	"time"
)

type Item struct {
	ID              string
	Name            string
	Category        string
	MainColor       string
	SecondaryColors []string
	Pattern         string
	Material        string
	Fit             string
	Formality       int
	SeasonTags      []string
	RainOK          bool
	ActivityTags    []string
	VibeTags        []string
	StoragePath     string
	Emoji           string
	Status          string
	WearCount       int
	CreatedAt       time.Time
}

type StyleProfile struct {
	StylePrimary    string
	Likes           string
	Dislikes        string
	ComfortBias     int
	PhotoLookBias   int
	NoRepeatTopDays int
}

type OutfitItem struct {
	ID        string
	Name      string
	Emoji     string
	ImageURL  string
	Category  string
	MainColor string
}

type Outfit struct {
	ID      string
	Label   string
	Score   float64
	Why     string
	Items   []OutfitItem
	ItemIDs []string
}

type WearEvent struct {
	ID          string
	OutfitSetID string
	ItemIDs     []string
	WornAt      time.Time
	Source      string
}

type PackingCategory struct {
	Label string
	Items []string
}

type DayOutfit struct {
	Day         int
	Label       string
	Description string
	ItemIDs     []string
}

type ItemAttrs struct {
	Name            string
	Category        string
	MainColor       string
	SecondaryColors []string
	Pattern         string
	Material        string
	Fit             string
	Formality       int
	SeasonTags      []string
	RainOK          bool
	ActivityTags    []string
	VibeTags        []string
}

// Store is the persistence surface used by handlers and services.
type Store interface {
	ListItems(category string) []Item
	GetItem(id string) (Item, error)
	AddItem(name, category, storagePath string) (Item, error)
	AddItemWithAttrs(name, category, storagePath string, attrs ItemAttrs) (Item, error)
	UpdateItem(item Item) error
	GetProfile() StyleProfile
	SaveProfile(p StyleProfile) error
	SaveOutfitSet(ctx context.Context, outfit Outfit) (Outfit, error)
	LogWear(outfitID string, itemIDs []string, source string) error
	ListWearEvents(limit int) []WearEvent
	SaveFitCheck(photoPath string, itemIDs []string, score float64, critique string) error
}
