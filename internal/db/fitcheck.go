package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func CreateFitCheck(db *sql.DB, userID, id, photoPath string, itemIDs []string, score float64, critique string) error {
	if userID == "" {
		userID = DefaultUserID
	}
	itemIDsJSON, _ := json.Marshal(itemIDs)
	_, err := db.Exec(`
		INSERT INTO fit_checks (id, user_id, photo_path, item_ids, score, critique)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, userID, photoPath, string(itemIDsJSON), score, critique)
	if err != nil {
		return fmt.Errorf("create fit check: %w", err)
	}
	return nil
}
