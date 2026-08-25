package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ashokparihar/fitcheck/internal/ai"
	"github.com/ashokparihar/fitcheck/internal/config"
	"github.com/ashokparihar/fitcheck/internal/db"
	"github.com/ashokparihar/fitcheck/internal/handlers"
	"github.com/ashokparihar/fitcheck/internal/storage"
	"github.com/ashokparihar/fitcheck/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	Handler http.Handler
	Cleanup func() error
}

func New(h *handlers.Handler, uploads storage.Storage, local *storage.Local) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(55 * time.Second))

	r.Get("/health", h.Health)
	r.Get("/", h.Home)

	r.Get("/closet", h.Closet)
	r.Post("/closet/items", h.AddItem)
	r.Get("/closet/items/{id}", h.ItemDetail)
	r.Post("/closet/items/{id}", h.ItemDetail)

	r.Get("/style", h.Style)
	r.Post("/style", h.Style)

	r.Get("/plan", h.Plan)
	r.Get("/outfits", h.Outfits)
	r.Post("/outfits/{id}/wear", h.WearOutfit)

	r.Get("/trip", h.Trip)

	r.Get("/fitcheck", h.FitCheck)
	r.Post("/fitcheck", h.FitCheck)

	if local != nil {
		r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(local.Dir()))))
	}

	return r
}

func Bootstrap(cfg *config.Config) (*App, error) {
	stor, err := storage.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("storage: %w", err)
	}

	if sb, ok := stor.(*storage.Supabase); ok {
		if err := sb.EnsureBucket(); err != nil {
			log.Printf("supabase bucket setup: %v", err)
		}
	}

	var local *storage.Local
	if !cfg.UseSupabaseStorage() {
		local, _ = storage.NewLocal(storage.UploadDir)
	}

	var s store.Store
	var cleanup func() error = func() error { return nil }

	if cfg.UsePostgres() {
		database, err := db.OpenPostgres(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("postgres: %w", err)
		}
		s = store.NewPostgres(database)
		cleanup = func() error { return database.Close() }
		log.Printf("Using Supabase Postgres")
	} else {
		database, err := db.Open(cfg.SQLitePath)
		if err != nil {
			return nil, fmt.Errorf("sqlite: %w", err)
		}
		s = store.NewSQLite(database)
		cleanup = func() error { return database.Close() }
		log.Printf("Using local SQLite (%s)", cfg.SQLitePath)
	}

	var aiClient *ai.Client
	aiClient, err = ai.NewClientFromConfig(cfg)
	if err != nil {
		log.Printf("AI setup failed: %v (using rule-based fallbacks)", err)
	}

	h := handlers.New(s, stor, aiClient)
	return &App{Handler: New(h, stor, local), Cleanup: cleanup}, nil
}

func NewWithSQLite(cfg *config.Config) (http.Handler, func() error, error) {
	app, err := Bootstrap(cfg)
	if err != nil {
		return nil, nil, err
	}
	return app.Handler, app.Cleanup, nil
}
