package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ashokparihar/fitcheck/pkg/db"
	"github.com/google/uuid"
)

// PostgresStore persists data in Supabase Postgres.
type PostgresStore struct {
	db     *sql.DB
	userID string
}

func NewPostgres(database *sql.DB) *PostgresStore {
	return &PostgresStore{db: database, userID: db.DefaultUserID}
}

func (s *PostgresStore) ListItems(category string) []Item {
	q := `SELECT id, user_id, storage_path, category, subcategory, name, main_color,
		secondary_colors, pattern, material, fit, formality, season_tags, rain_ok,
		activity_tags, vibe_tags, pair_hints, ai_raw, user_corrected, status,
		wear_count, last_worn_at, created_at, updated_at
		FROM items WHERE user_id = $1`
	args := []any{s.userID}
	if category != "" && category != "all" {
		q += ` AND category = $2`
		args = append(args, category)
	}
	q += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		item, err := scanPostgresItem(rows)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (s *PostgresStore) GetItem(id string) (Item, error) {
	row := s.db.QueryRow(`SELECT id, user_id, storage_path, category, subcategory, name, main_color,
		secondary_colors, pattern, material, fit, formality, season_tags, rain_ok,
		activity_tags, vibe_tags, pair_hints, ai_raw, user_corrected, status,
		wear_count, last_worn_at, created_at, updated_at
		FROM items WHERE user_id = $1 AND id = $2`, s.userID, id)
	item, err := scanPostgresItemRow(row)
	if err == sql.ErrNoRows {
		return Item{}, fmt.Errorf("item not found")
	}
	return item, err
}

func (s *PostgresStore) AddItem(name, category, storagePath string) (Item, error) {
	return s.AddItemWithAttrs(name, category, storagePath, ItemAttrs{Name: name, Category: category})
}

func (s *PostgresStore) AddItemWithAttrs(name, category, storagePath string, attrs ItemAttrs) (Item, error) {
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
	if attrs.MainColor == "" {
		attrs.MainColor = "Unknown"
	}
	if attrs.Formality == 0 {
		attrs.Formality = 5
	}

	id := uuid.New().String()
	_, err := s.db.Exec(`
		INSERT INTO items (id, user_id, storage_path, category, name, main_color,
			secondary_colors, pattern, material, fit, formality, season_tags, rain_ok,
			activity_tags, vibe_tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`, id, s.userID, storagePath, category, name, attrs.MainColor,
		db.MarshalJSONStrings(attrs.SecondaryColors), attrs.Pattern, attrs.Material,
		attrs.Fit, attrs.Formality, db.MarshalJSONStrings(attrs.SeasonTags), attrs.RainOK,
		db.MarshalJSONStrings(attrs.ActivityTags), db.MarshalJSONStrings(attrs.VibeTags))
	if err != nil {
		return Item{}, err
	}
	return s.GetItem(id)
}

func (s *PostgresStore) UpdateItem(item Item) error {
	_, err := s.db.Exec(`
		UPDATE items SET
			name = COALESCE(NULLIF($1, ''), name),
			category = COALESCE(NULLIF($2, ''), category),
			main_color = COALESCE(NULLIF($3, ''), main_color),
			pattern = COALESCE(NULLIF($4, ''), pattern),
			material = COALESCE(NULLIF($5, ''), material),
			fit = COALESCE(NULLIF($6, ''), fit),
			formality = CASE WHEN $7 > 0 THEN $7 ELSE formality END,
			season_tags = CASE WHEN $8 != '[]' THEN $8 ELSE season_tags END,
			activity_tags = CASE WHEN $9 != '[]' THEN $9 ELSE activity_tags END,
			user_corrected = TRUE,
			updated_at = NOW()
		WHERE user_id = $10 AND id = $11
	`, item.Name, item.Category, item.MainColor, item.Pattern, item.Material, item.Fit,
		item.Formality, db.MarshalJSONStrings(item.SeasonTags),
		db.MarshalJSONStrings(item.ActivityTags), s.userID, item.ID)
	return err
}

func (s *PostgresStore) GetProfile() StyleProfile {
	var p StyleProfile
	var likes, dislikes sql.NullString
	err := s.db.QueryRow(`
		SELECT style_primary, likes, dislikes, comfort_bias, photo_look_bias, no_repeat_top_days
		FROM profiles WHERE id = $1
	`, s.userID).Scan(&p.StylePrimary, &likes, &dislikes, &p.ComfortBias, &p.PhotoLookBias, &p.NoRepeatTopDays)
	if err != nil {
		return StyleProfile{StylePrimary: "Casual", ComfortBias: 7, PhotoLookBias: 5, NoRepeatTopDays: 3}
	}
	p.Likes = likes.String
	p.Dislikes = dislikes.String
	return p
}

func (s *PostgresStore) SaveProfile(p StyleProfile) error {
	_, err := s.db.Exec(`
		INSERT INTO profiles (id, style_primary, likes, dislikes, comfort_bias, photo_look_bias, no_repeat_top_days)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			style_primary = EXCLUDED.style_primary,
			likes = EXCLUDED.likes,
			dislikes = EXCLUDED.dislikes,
			comfort_bias = EXCLUDED.comfort_bias,
			photo_look_bias = EXCLUDED.photo_look_bias,
			no_repeat_top_days = EXCLUDED.no_repeat_top_days,
			updated_at = NOW()
	`, s.userID, p.StylePrimary, p.Likes, p.Dislikes, p.ComfortBias, p.PhotoLookBias, p.NoRepeatTopDays)
	return err
}

func (s *PostgresStore) SaveOutfitSet(ctx context.Context, outfit Outfit) (Outfit, error) {
	if outfit.ID == "" {
		outfit.ID = uuid.New().String()
	}
	itemIDs, _ := json.Marshal(outfit.ItemIDs)
	_, err := s.db.Exec(`
		INSERT INTO outfit_sets (id, user_id, plan_kind, label, item_ids, why, score)
		VALUES ($1, $2, 'today', $3, $4, $5, $6)
	`, outfit.ID, s.userID, outfit.Label, string(itemIDs), outfit.Why, outfit.Score)
	return outfit, err
}

func (s *PostgresStore) LogWear(outfitID string, itemIDs []string, source string) error {
	if source == "" {
		source = "recommended"
	}
	if len(itemIDs) == 0 && outfitID != "" {
		var idsJSON string
		_ = s.db.QueryRow(`SELECT item_ids FROM outfit_sets WHERE user_id = $1 AND id = $2`, s.userID, outfitID).Scan(&idsJSON)
		_ = json.Unmarshal([]byte(idsJSON), &itemIDs)
	}
	ids, _ := json.Marshal(itemIDs)
	_, err := s.db.Exec(`
		INSERT INTO wear_events (id, user_id, outfit_set_id, item_ids, source, accepted)
		VALUES ($1, $2, $3, $4, $5, TRUE)
	`, uuid.New().String(), s.userID, nullStr(outfitID), string(ids), source)
	if err != nil {
		return err
	}
	for _, id := range itemIDs {
		_, _ = s.db.Exec(`UPDATE items SET wear_count = wear_count + 1, last_worn_at = NOW() WHERE user_id = $1 AND id = $2`, s.userID, id)
	}
	return nil
}

func (s *PostgresStore) ListWearEvents(limit int) []WearEvent {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, outfit_set_id, item_ids, worn_at, source
		FROM wear_events WHERE user_id = $1 ORDER BY worn_at DESC LIMIT $2
	`, s.userID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var events []WearEvent
	for rows.Next() {
		var e WearEvent
		var outfitID sql.NullString
		var idsJSON string
		if err := rows.Scan(&e.ID, &outfitID, &idsJSON, &e.WornAt, &e.Source); err != nil {
			continue
		}
		e.OutfitSetID = outfitID.String
		_ = json.Unmarshal([]byte(idsJSON), &e.ItemIDs)
		events = append(events, e)
	}
	return events
}

func (s *PostgresStore) SaveFitCheck(photoPath string, itemIDs []string, score float64, critique string) error {
	ids, _ := json.Marshal(itemIDs)
	_, err := s.db.Exec(`
		INSERT INTO fit_checks (id, user_id, photo_path, item_ids, score, critique)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New().String(), s.userID, photoPath, string(ids), score, critique)
	return err
}

type pgScanner interface {
	Scan(dest ...any) error
}

func scanPostgresItem(rows pgScanner) (Item, error) {
	return scanPostgresItemFields(rows)
}

func scanPostgresItemRow(row *sql.Row) (Item, error) {
	return scanPostgresItemFields(row)
}

func scanPostgresItemFields(sc pgScanner) (Item, error) {
	var item Item
	var userID string
	var subcategory, name, mainColor, pattern, material, fit, aiRaw sql.NullString
	var secondaryColors, seasonTags, activityTags, vibeTags, pairHints sql.NullString
	var userCorrected bool
	var lastWorn sql.NullTime
	var updatedAt time.Time

	err := sc.Scan(
		&item.ID, &userID, &item.StoragePath, &item.Category,
		&subcategory, &name, &mainColor,
		&secondaryColors, &pattern, &material, &fit,
		&item.Formality, &seasonTags, &item.RainOK,
		&activityTags, &vibeTags, &pairHints,
		&aiRaw, &userCorrected, &item.Status,
		&item.WearCount, &lastWorn, &item.CreatedAt, &updatedAt,
	)
	if err != nil {
		return Item{}, err
	}
	_ = userID
	_ = userCorrected
	_ = lastWorn
	_ = updatedAt
	_ = subcategory
	_ = pairHints
	_ = aiRaw

	item.Name = coalesceStr(name, defaultName(item.Category))
	item.MainColor = coalesceStr(mainColor, "Unknown")
	item.Pattern = pattern.String
	item.Material = material.String
	item.Fit = fit.String
	item.SecondaryColors = db.UnmarshalJSONStrings(secondaryColors.String)
	item.SeasonTags = db.UnmarshalJSONStrings(seasonTags.String)
	item.ActivityTags = db.UnmarshalJSONStrings(activityTags.String)
	item.VibeTags = db.UnmarshalJSONStrings(vibeTags.String)
	item.Emoji = categoryEmoji(item.Category)
	return item, nil
}

func coalesceStr(v sql.NullString, fallback string) string {
	if v.Valid && v.String != "" {
		return v.String
	}
	return fallback
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
