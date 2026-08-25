package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	SQLitePath             string
	SupabaseURL            string
	SupabaseAnonKey        string
	SupabaseServiceRoleKey string
	DatabaseURL            string
	OpenAIAPIKey           string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "fitcheck.db"
	}

	return &Config{
		Port:                   port,
		SQLitePath:             sqlitePath,
		SupabaseURL:            os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:        os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:           os.Getenv("OPENAI_API_KEY"),
	}, nil
}
