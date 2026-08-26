package db

import (
	"database/sql"
	"fmt"
)

const postgresSchema = `
CREATE TABLE IF NOT EXISTS profiles (
  id TEXT PRIMARY KEY DEFAULT 'local',
  display_name TEXT,
  style_primary TEXT DEFAULT 'Casual',
  style_secondary TEXT DEFAULT '[]',
  likes TEXT DEFAULT '{}',
  dislikes TEXT DEFAULT '{}',
  comfort_bias INTEGER DEFAULT 7 CHECK (comfort_bias BETWEEN 1 AND 10),
  photo_look_bias INTEGER DEFAULT 5 CHECK (photo_look_bias BETWEEN 1 AND 10),
  no_repeat_top_days INTEGER DEFAULT 3,
  body_notes TEXT,
  default_formality INTEGER DEFAULT 5 CHECK (default_formality BETWEEN 1 AND 10),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS items (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL DEFAULT 'local',
  storage_path TEXT NOT NULL,
  category TEXT NOT NULL CHECK (category IN ('tshirt', 'shirt', 'pants', 'shorts', 'jacket', 'shoes', 'accessory')),
  subcategory TEXT,
  name TEXT,
  main_color TEXT,
  secondary_colors TEXT DEFAULT '[]',
  pattern TEXT,
  material TEXT,
  fit TEXT,
  formality INTEGER DEFAULT 5 CHECK (formality BETWEEN 1 AND 10),
  season_tags TEXT DEFAULT '[]',
  rain_ok BOOLEAN DEFAULT FALSE,
  activity_tags TEXT DEFAULT '[]',
  vibe_tags TEXT DEFAULT '[]',
  pair_hints TEXT DEFAULT '{}',
  ai_raw TEXT,
  user_corrected BOOLEAN DEFAULT FALSE,
  status TEXT DEFAULT 'active' CHECK (status IN ('active', 'dirty', 'packed', 'retired')),
  wear_count INTEGER DEFAULT 0,
  last_worn_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS items_user_id_idx ON items(user_id);
CREATE INDEX IF NOT EXISTS items_category_idx ON items(user_id, category);

CREATE TABLE IF NOT EXISTS outfit_sets (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL DEFAULT 'local',
  trip_id TEXT,
  plan_kind TEXT DEFAULT 'today',
  day_index INTEGER,
  label TEXT,
  item_ids TEXT NOT NULL DEFAULT '[]',
  why TEXT,
  score REAL,
  variant TEXT DEFAULT 'base',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wear_events (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL DEFAULT 'local',
  outfit_set_id TEXT,
  item_ids TEXT NOT NULL DEFAULT '[]',
  worn_at TIMESTAMPTZ DEFAULT NOW(),
  source TEXT DEFAULT 'recommended',
  accepted BOOLEAN DEFAULT TRUE,
  user_rating INTEGER CHECK (user_rating IS NULL OR (user_rating BETWEEN 1 AND 5))
);

CREATE TABLE IF NOT EXISTS fit_checks (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL DEFAULT 'local',
  photo_path TEXT NOT NULL,
  outfit_set_id TEXT,
  item_ids TEXT DEFAULT '[]',
  score REAL,
  critique TEXT,
  suggested_swaps TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
`

func MigratePostgres(db *sql.DB) error {
	if _, err := db.Exec(postgresSchema); err != nil {
		return fmt.Errorf("migrate postgres schema: %w", err)
	}
	_, err := db.Exec(`
		INSERT INTO profiles (id, style_primary, comfort_bias, photo_look_bias, no_repeat_top_days)
		VALUES ($1, 'Casual', 7, 5, 3)
		ON CONFLICT (id) DO NOTHING
	`, DefaultUserID)
	if err != nil {
		return fmt.Errorf("seed profile: %w", err)
	}
	return nil
}
