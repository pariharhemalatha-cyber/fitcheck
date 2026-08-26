package db

import (
	"database/sql"
	"fmt"
	"time"
)

type ProfileRow struct {
	ID               string
	DisplayName      sql.NullString
	StylePrimary     string
	StyleSecondary   string
	Likes            string
	Dislikes         string
	ComfortBias      int
	PhotoLookBias    int
	NoRepeatTopDays  int
	BodyNotes        sql.NullString
	DefaultFormality int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func GetProfile(db *sql.DB, userID string) (ProfileRow, error) {
	if userID == "" {
		userID = DefaultUserID
	}

	var p ProfileRow
	var createdAt, updatedAt string
	err := db.QueryRow(`
		SELECT id, display_name, style_primary, style_secondary, likes, dislikes,
			comfort_bias, photo_look_bias, no_repeat_top_days, body_notes,
			default_formality, created_at, updated_at
		FROM profiles WHERE id = ?
	`, userID).Scan(
		&p.ID, &p.DisplayName, &p.StylePrimary, &p.StyleSecondary,
		&p.Likes, &p.Dislikes, &p.ComfortBias, &p.PhotoLookBias,
		&p.NoRepeatTopDays, &p.BodyNotes, &p.DefaultFormality,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return ProfileRow{}, fmt.Errorf("get profile: %w", err)
	}
	p.CreatedAt, _ = parseSQLiteTime(createdAt)
	p.UpdatedAt, _ = parseSQLiteTime(updatedAt)
	return p, nil
}

type SaveProfileParams struct {
	UserID          string
	StylePrimary    string
	Likes           string
	Dislikes        string
	ComfortBias     int
	PhotoLookBias   int
	NoRepeatTopDays int
}

func SaveProfile(db *sql.DB, p SaveProfileParams) (ProfileRow, error) {
	userID := p.UserID
	if userID == "" {
		userID = DefaultUserID
	}

	_, err := db.Exec(`
		INSERT INTO profiles (id, style_primary, likes, dislikes, comfort_bias, photo_look_bias, no_repeat_top_days)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			style_primary = excluded.style_primary,
			likes = excluded.likes,
			dislikes = excluded.dislikes,
			comfort_bias = excluded.comfort_bias,
			photo_look_bias = excluded.photo_look_bias,
			no_repeat_top_days = excluded.no_repeat_top_days,
			updated_at = datetime('now')
	`, userID, p.StylePrimary, p.Likes, p.Dislikes, p.ComfortBias, p.PhotoLookBias, p.NoRepeatTopDays)
	if err != nil {
		return ProfileRow{}, fmt.Errorf("save profile: %w", err)
	}
	return GetProfile(db, userID)
}
