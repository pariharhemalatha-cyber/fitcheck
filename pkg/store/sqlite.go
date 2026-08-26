package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ashokparihar/fitcheck/pkg/db"
	"github.com/google/uuid"
)

// SQLiteStore persists data in SQLite via the internal/db layer.
type SQLiteStore struct {
	db     *sql.DB
	userID string
}

func NewSQLite(database *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: database, userID: db.DefaultUserID}
}

func (s *SQLiteStore) ListItems(category string) []Item {
	rows, err := db.ListItems(s.db, s.userID, category)
	if err != nil {
		return nil
	}
	items := make([]Item, len(rows))
	for i, row := range rows {
		items[i] = itemFromRow(row)
	}
	return items
}

func (s *SQLiteStore) GetItem(id string) (Item, error) {
	row, err := db.GetItem(s.db, s.userID, id)
	if err != nil {
		return Item{}, err
	}
	return itemFromRow(row), nil
}

func (s *SQLiteStore) AddItem(name, category, storagePath string) (Item, error) {
	return s.AddItemWithAttrs(name, category, storagePath, ItemAttrs{
		Name:     name,
		Category: category,
	})
}

func (s *SQLiteStore) AddItemWithAttrs(name, category, storagePath string, attrs ItemAttrs) (Item, error) {
	if storagePath == "" {
		return Item{}, fmt.Errorf("storage path required")
	}
	if category == "" {
		category = attrs.Category
	}
	if category == "" {
		category = "tshirt"
	}
	if name == "" {
		name = attrs.Name
	}
	if name == "" {
		name = defaultName(category)
	}

	row, err := db.CreateItem(s.db, db.CreateItemParams{
		ID:              uuid.New().String(),
		UserID:          s.userID,
		StoragePath:     storagePath,
		Category:        category,
		Name:            name,
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
	})
	if err != nil {
		return Item{}, err
	}
	return itemFromRow(row), nil
}

func (s *SQLiteStore) UpdateItem(item Item) error {
	_, err := db.UpdateItem(s.db, db.UpdateItemParams{
		ID:           item.ID,
		UserID:       s.userID,
		Name:         item.Name,
		Category:     item.Category,
		MainColor:    item.MainColor,
		Pattern:      item.Pattern,
		Material:     item.Material,
		Fit:          item.Fit,
		Formality:    item.Formality,
		SeasonTags:   item.SeasonTags,
		ActivityTags: item.ActivityTags,
		VibeTags:     item.VibeTags,
		StoragePath:  item.StoragePath,
		Status:       item.Status,
	})
	return err
}

func (s *SQLiteStore) GetProfile() StyleProfile {
	row, err := db.GetProfile(s.db, s.userID)
	if err != nil {
		return StyleProfile{
			StylePrimary:    "Casual",
			ComfortBias:     7,
			PhotoLookBias:   5,
			NoRepeatTopDays: 3,
		}
	}
	return profileFromRow(row)
}

func (s *SQLiteStore) SaveProfile(p StyleProfile) error {
	_, err := db.SaveProfile(s.db, db.SaveProfileParams{
		UserID:          s.userID,
		StylePrimary:    p.StylePrimary,
		Likes:           p.Likes,
		Dislikes:        p.Dislikes,
		ComfortBias:     p.ComfortBias,
		PhotoLookBias:   p.PhotoLookBias,
		NoRepeatTopDays: p.NoRepeatTopDays,
	})
	return err
}

func (s *SQLiteStore) SaveOutfitSet(ctx context.Context, outfit Outfit) (Outfit, error) {
	if outfit.ID == "" {
		outfit.ID = uuid.New().String()
	}
	row, err := db.CreateOutfitSet(s.db, db.OutfitSetRow{
		ID:       outfit.ID,
		UserID:   s.userID,
		PlanKind: "today",
		Label:    sql.NullString{String: outfit.Label, Valid: outfit.Label != ""},
		ItemIDs:  outfit.ItemIDs,
		Why:      sql.NullString{String: outfit.Why, Valid: outfit.Why != ""},
		Score:    sql.NullFloat64{Float64: outfit.Score, Valid: outfit.Score > 0},
	})
	if err != nil {
		return outfit, err
	}
	outfit.ID = row.ID
	return outfit, nil
}

func (s *SQLiteStore) LogWear(outfitID string, itemIDs []string, source string) error {
	if source == "" {
		source = "recommended"
	}
	if len(itemIDs) == 0 && outfitID != "" {
		row, err := db.GetOutfitSet(s.db, s.userID, outfitID)
		if err == nil {
			itemIDs = row.ItemIDs
		}
	}
	_, err := db.CreateWearEvent(s.db, db.CreateWearEventParams{
		ID:          uuid.New().String(),
		UserID:      s.userID,
		OutfitSetID: outfitID,
		ItemIDs:     itemIDs,
		Source:      source,
		Accepted:    true,
	})
	return err
}

func (s *SQLiteStore) ListWearEvents(limit int) []WearEvent {
	rows, err := db.ListWearEvents(s.db, s.userID, limit)
	if err != nil {
		return nil
	}
	events := make([]WearEvent, len(rows))
	for i, row := range rows {
		events[i] = WearEvent{
			ID:          row.ID,
			OutfitSetID: row.OutfitSetID.String,
			ItemIDs:     row.ItemIDs,
			WornAt:      row.WornAt,
			Source:      row.Source,
		}
	}
	return events
}

func (s *SQLiteStore) SaveFitCheck(photoPath string, itemIDs []string, score float64, critique string) error {
	return db.CreateFitCheck(s.db, s.userID, uuid.New().String(), photoPath, itemIDs, score, critique)
}

func itemFromRow(row db.ItemRow) Item {
	name := row.Name.String
	if name == "" {
		name = defaultName(row.Category)
	}
	mainColor := row.MainColor.String
	if mainColor == "" {
		mainColor = "Unknown"
	}
	return Item{
		ID:              row.ID,
		Name:            name,
		Category:        row.Category,
		MainColor:       mainColor,
		SecondaryColors: db.UnmarshalJSONStrings(row.SecondaryColors),
		Pattern:         row.Pattern.String,
		Material:        row.Material.String,
		Fit:             row.Fit.String,
		Formality:       row.Formality,
		SeasonTags:      db.UnmarshalJSONStrings(row.SeasonTags),
		RainOK:          row.RainOK,
		ActivityTags:    db.UnmarshalJSONStrings(row.ActivityTags),
		VibeTags:        db.UnmarshalJSONStrings(row.VibeTags),
		StoragePath:     row.StoragePath,
		Emoji:           categoryEmoji(row.Category),
		Status:          row.Status,
		WearCount:       row.WearCount,
		CreatedAt:       row.CreatedAt,
	}
}

func profileFromRow(row db.ProfileRow) StyleProfile {
	return StyleProfile{
		StylePrimary:    row.StylePrimary,
		Likes:           row.Likes,
		Dislikes:        row.Dislikes,
		ComfortBias:     row.ComfortBias,
		PhotoLookBias:   row.PhotoLookBias,
		NoRepeatTopDays: row.NoRepeatTopDays,
	}
}
