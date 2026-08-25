package server

import (
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

func New(h *handlers.Handler, uploads *storage.Local) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

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

	if uploads != nil {
		r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploads.Dir()))))
	}

	return r
}

func NewWithSQLite(cfg *config.Config) (http.Handler, func() error, error) {
	database, err := db.Open(cfg.SQLitePath)
	if err != nil {
		return nil, nil, err
	}

	uploads, err := storage.NewLocal(storage.UploadDir)
	if err != nil {
		_ = database.Close()
		return nil, nil, err
	}

	var aiClient *ai.Client
	if cfg.OpenAIAPIKey != "" {
		aiClient, err = ai.NewClient(cfg.OpenAIAPIKey)
		if err != nil {
			log.Printf("OpenAI not configured: %v (using rule-based fallbacks)", err)
		} else {
			log.Printf("OpenAI vision enabled")
		}
	} else {
		log.Printf("No OPENAI_API_KEY — using heuristic analysis and rule-based outfits")
	}

	s := store.NewSQLite(database)
	h := handlers.New(s, uploads, aiClient)
	cleanup := func() error { return database.Close() }
	return New(h, uploads), cleanup, nil
}
