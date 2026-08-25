package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type OutfitSetRow struct {
	ID        string
	UserID    string
	TripID    sql.NullString
	PlanKind  string
	DayIndex  sql.NullInt64
	Label     sql.NullString
	ItemIDs   []string
	Why       sql.NullString
	Score     sql.NullFloat64
	Variant   string
	CreatedAt time.Time
}

type WearEventRow struct {
	ID          string
	UserID      string
	OutfitSetID sql.NullString
	ItemIDs     []string
	WornAt      time.Time
	Source      string
	Accepted    bool
	UserRating  sql.NullInt64
}

func CreateOutfitSet(db *sql.DB, row OutfitSetRow) (OutfitSetRow, error) {
	userID := row.UserID
	if userID == "" {
		userID = DefaultUserID
	}
	itemIDsJSON, err := json.Marshal(row.ItemIDs)
	if err != nil {
		return OutfitSetRow{}, fmt.Errorf("marshal item ids: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO outfit_sets (id, user_id, trip_id, plan_kind, day_index, label, item_ids, why, score, variant)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, row.ID, userID, nullString(row.TripID), row.PlanKind, nullInt64(row.DayIndex),
		nullString(row.Label), string(itemIDsJSON), nullString(row.Why),
		nullFloat64(row.Score), row.Variant)
	if err != nil {
		return OutfitSetRow{}, fmt.Errorf("create outfit set: %w", err)
	}
	return GetOutfitSet(db, userID, row.ID)
}

func GetOutfitSet(db *sql.DB, userID, id string) (OutfitSetRow, error) {
	if userID == "" {
		userID = DefaultUserID
	}

	var row OutfitSetRow
	var itemIDsJSON string
	var createdAt string
	err := db.QueryRow(`
		SELECT id, user_id, trip_id, plan_kind, day_index, label, item_ids, why, score, variant, created_at
		FROM outfit_sets WHERE user_id = ? AND id = ?
	`, userID, id).Scan(
		&row.ID, &row.UserID, &row.TripID, &row.PlanKind, &row.DayIndex,
		&row.Label, &itemIDsJSON, &row.Why, &row.Score, &row.Variant, &createdAt,
	)
	if err != nil {
		return OutfitSetRow{}, fmt.Errorf("get outfit set: %w", err)
	}
	if err := json.Unmarshal([]byte(itemIDsJSON), &row.ItemIDs); err != nil {
		row.ItemIDs = []string{}
	}
	row.CreatedAt, _ = parseSQLiteTime(createdAt)
	return row, nil
}

func ListOutfitSets(db *sql.DB, userID string) ([]OutfitSetRow, error) {
	if userID == "" {
		userID = DefaultUserID
	}

	rows, err := db.Query(`
		SELECT id, user_id, trip_id, plan_kind, day_index, label, item_ids, why, score, variant, created_at
		FROM outfit_sets WHERE user_id = ? ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list outfit sets: %w", err)
	}
	defer rows.Close()

	var sets []OutfitSetRow
	for rows.Next() {
		var row OutfitSetRow
		var itemIDsJSON string
		var createdAt string
		if err := rows.Scan(
			&row.ID, &row.UserID, &row.TripID, &row.PlanKind, &row.DayIndex,
			&row.Label, &itemIDsJSON, &row.Why, &row.Score, &row.Variant, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan outfit set: %w", err)
		}
		_ = json.Unmarshal([]byte(itemIDsJSON), &row.ItemIDs)
		row.CreatedAt, _ = parseSQLiteTime(createdAt)
		sets = append(sets, row)
	}
	return sets, rows.Err()
}

type CreateWearEventParams struct {
	ID          string
	UserID      string
	OutfitSetID string
	ItemIDs     []string
	Source      string
	Accepted    bool
	UserRating  *int
}

func CreateWearEvent(db *sql.DB, p CreateWearEventParams) (WearEventRow, error) {
	userID := p.UserID
	if userID == "" {
		userID = DefaultUserID
	}
	if p.Source == "" {
		p.Source = "recommended"
	}

	itemIDsJSON, err := json.Marshal(p.ItemIDs)
	if err != nil {
		return WearEventRow{}, fmt.Errorf("marshal item ids: %w", err)
	}

	accepted := 0
	if p.Accepted {
		accepted = 1
	}

	var rating any
	if p.UserRating != nil {
		rating = *p.UserRating
	}

	_, err = db.Exec(`
		INSERT INTO wear_events (id, user_id, outfit_set_id, item_ids, source, accepted, user_rating)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, p.ID, userID, nullIfEmpty(p.OutfitSetID), string(itemIDsJSON), p.Source, accepted, rating)
	if err != nil {
		return WearEventRow{}, fmt.Errorf("create wear event: %w", err)
	}

	// Bump wear counts on items
	for _, itemID := range p.ItemIDs {
		_, _ = db.Exec(`
			UPDATE items SET wear_count = wear_count + 1, last_worn_at = datetime('now'), updated_at = datetime('now')
			WHERE user_id = ? AND id = ?
		`, userID, itemID)
	}

	return GetWearEvent(db, userID, p.ID)
}

func GetWearEvent(db *sql.DB, userID, id string) (WearEventRow, error) {
	if userID == "" {
		userID = DefaultUserID
	}

	var row WearEventRow
	var itemIDsJSON string
	var wornAt string
	var accepted int
	err := db.QueryRow(`
		SELECT id, user_id, outfit_set_id, item_ids, worn_at, source, accepted, user_rating
		FROM wear_events WHERE user_id = ? AND id = ?
	`, userID, id).Scan(
		&row.ID, &row.UserID, &row.OutfitSetID, &itemIDsJSON,
		&wornAt, &row.Source, &accepted, &row.UserRating,
	)
	if err != nil {
		return WearEventRow{}, fmt.Errorf("get wear event: %w", err)
	}
	_ = json.Unmarshal([]byte(itemIDsJSON), &row.ItemIDs)
	row.Accepted = accepted != 0
	row.WornAt, _ = parseSQLiteTime(wornAt)
	return row, nil
}

func ListWearEvents(db *sql.DB, userID string, limit int) ([]WearEventRow, error) {
	if userID == "" {
		userID = DefaultUserID
	}
	if limit <= 0 {
		limit = 50
	}

	rows, err := db.Query(`
		SELECT id, user_id, outfit_set_id, item_ids, worn_at, source, accepted, user_rating
		FROM wear_events WHERE user_id = ? ORDER BY worn_at DESC LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list wear events: %w", err)
	}
	defer rows.Close()

	var events []WearEventRow
	for rows.Next() {
		var row WearEventRow
		var itemIDsJSON string
		var wornAt string
		var accepted int
		if err := rows.Scan(
			&row.ID, &row.UserID, &row.OutfitSetID, &itemIDsJSON,
			&wornAt, &row.Source, &accepted, &row.UserRating,
		); err != nil {
			return nil, fmt.Errorf("scan wear event: %w", err)
		}
		_ = json.Unmarshal([]byte(itemIDsJSON), &row.ItemIDs)
		row.Accepted = accepted != 0
		row.WornAt, _ = parseSQLiteTime(wornAt)
		events = append(events, row)
	}
	return events, rows.Err()
}

func nullString(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

func nullInt64(v sql.NullInt64) any {
	if v.Valid {
		return v.Int64
	}
	return nil
}

func nullFloat64(v sql.NullFloat64) any {
	if v.Valid {
		return v.Float64
	}
	return nil
}
