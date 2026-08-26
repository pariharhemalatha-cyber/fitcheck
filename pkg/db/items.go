package db

import (
	"database/sql"
	"fmt"
	"time"
)

type ItemRow struct {
	ID               string
	UserID           string
	StoragePath      string
	Category         string
	Subcategory      sql.NullString
	Name             sql.NullString
	MainColor        sql.NullString
	SecondaryColors  string
	Pattern          sql.NullString
	Material         sql.NullString
	Fit              sql.NullString
	Formality        int
	SeasonTags       string
	RainOK           bool
	ActivityTags     string
	VibeTags         string
	PairHints        string
	AIRaw            sql.NullString
	UserCorrected    bool
	Status           string
	WearCount        int
	LastWornAt       sql.NullTime
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

const itemColumns = `
	id, user_id, storage_path, category, subcategory, name, main_color,
	secondary_colors, pattern, material, fit, formality, season_tags, rain_ok,
	activity_tags, vibe_tags, pair_hints, ai_raw, user_corrected, status,
	wear_count, last_worn_at, created_at, updated_at
`

func scanItem(row scanner) (ItemRow, error) {
	var item ItemRow
	var rainOK int
	var userCorrected int
	var lastWornAt sql.NullString
	var createdAt, updatedAt string

	err := row.Scan(
		&item.ID, &item.UserID, &item.StoragePath, &item.Category,
		&item.Subcategory, &item.Name, &item.MainColor,
		&item.SecondaryColors, &item.Pattern, &item.Material, &item.Fit,
		&item.Formality, &item.SeasonTags, &rainOK,
		&item.ActivityTags, &item.VibeTags, &item.PairHints,
		&item.AIRaw, &userCorrected, &item.Status,
		&item.WearCount, &lastWornAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return ItemRow{}, err
	}

	item.RainOK = rainOK != 0
	item.UserCorrected = userCorrected != 0
	if lastWornAt.Valid {
		if t, err := time.Parse(time.RFC3339, lastWornAt.String); err == nil {
			item.LastWornAt = sql.NullTime{Time: t, Valid: true}
		}
	}
	item.CreatedAt, _ = parseSQLiteTime(createdAt)
	item.UpdatedAt, _ = parseSQLiteTime(updatedAt)
	return item, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func parseSQLiteTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02 15:04:05", s)
}

func ListItems(db *sql.DB, userID, category string) ([]ItemRow, error) {
	query := `SELECT ` + itemColumns + ` FROM items WHERE user_id = ?`
	args := []any{userID}

	if category != "" && category != "all" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []ItemRow
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func GetItem(db *sql.DB, userID, id string) (ItemRow, error) {
	row := db.QueryRow(`SELECT `+itemColumns+` FROM items WHERE user_id = ? AND id = ?`, userID, id)
	item, err := scanItem(row)
	if err == sql.ErrNoRows {
		return ItemRow{}, fmt.Errorf("item not found")
	}
	if err != nil {
		return ItemRow{}, fmt.Errorf("get item: %w", err)
	}
	return item, nil
}

type CreateItemParams struct {
	ID               string
	UserID           string
	StoragePath      string
	Category         string
	Name             string
	MainColor        string
	SecondaryColors  []string
	Pattern          string
	Material         string
	Fit              string
	Formality        int
	SeasonTags       []string
	RainOK           bool
	ActivityTags     []string
	VibeTags         []string
}

func CreateItem(db *sql.DB, p CreateItemParams) (ItemRow, error) {
	userID := p.UserID
	if userID == "" {
		userID = DefaultUserID
	}
	if p.MainColor == "" {
		p.MainColor = "Unknown"
	}
	if p.Formality == 0 {
		p.Formality = 5
	}

	rainOK := 0
	if p.RainOK {
		rainOK = 1
	}

	_, err := db.Exec(`
		INSERT INTO items (
			id, user_id, storage_path, category, name, main_color,
			secondary_colors, pattern, material, fit, formality,
			season_tags, rain_ok, activity_tags, vibe_tags
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.ID, userID, p.StoragePath, p.Category, nullIfEmpty(p.Name), p.MainColor,
		MarshalJSONStrings(p.SecondaryColors), nullIfEmpty(p.Pattern), nullIfEmpty(p.Material),
		nullIfEmpty(p.Fit), p.Formality, MarshalJSONStrings(p.SeasonTags), rainOK,
		MarshalJSONStrings(p.ActivityTags), MarshalJSONStrings(p.VibeTags))
	if err != nil {
		return ItemRow{}, fmt.Errorf("create item: %w", err)
	}
	return GetItem(db, userID, p.ID)
}

type UpdateItemParams struct {
	ID           string
	UserID       string
	Name         string
	Category     string
	MainColor    string
	Pattern      string
	Material     string
	Fit          string
	Formality    int
	SeasonTags   []string
	ActivityTags []string
	VibeTags     []string
	StoragePath  string
	Status       string
}

func UpdateItem(db *sql.DB, p UpdateItemParams) (ItemRow, error) {
	userID := p.UserID
	if userID == "" {
		userID = DefaultUserID
	}

	_, err := db.Exec(`
		UPDATE items SET
			name = COALESCE(NULLIF(?, ''), name),
			category = COALESCE(NULLIF(?, ''), category),
			main_color = COALESCE(NULLIF(?, ''), main_color),
			pattern = COALESCE(NULLIF(?, ''), pattern),
			material = COALESCE(NULLIF(?, ''), material),
			fit = COALESCE(NULLIF(?, ''), fit),
			formality = CASE WHEN ? > 0 THEN ? ELSE formality END,
			season_tags = CASE WHEN ? != '[]' THEN ? ELSE season_tags END,
			activity_tags = CASE WHEN ? != '[]' THEN ? ELSE activity_tags END,
			vibe_tags = CASE WHEN ? != '[]' THEN ? ELSE vibe_tags END,
			storage_path = COALESCE(NULLIF(?, ''), storage_path),
			status = COALESCE(NULLIF(?, ''), status),
			user_corrected = 1,
			updated_at = datetime('now')
		WHERE user_id = ? AND id = ?
	`, p.Name, p.Category, p.MainColor, p.Pattern, p.Material, p.Fit,
		p.Formality, p.Formality,
		MarshalJSONStrings(p.SeasonTags), MarshalJSONStrings(p.SeasonTags),
		MarshalJSONStrings(p.ActivityTags), MarshalJSONStrings(p.ActivityTags),
		MarshalJSONStrings(p.VibeTags), MarshalJSONStrings(p.VibeTags),
		p.StoragePath, p.Status, userID, p.ID)
	if err != nil {
		return ItemRow{}, fmt.Errorf("update item: %w", err)
	}
	return GetItem(db, userID, p.ID)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
