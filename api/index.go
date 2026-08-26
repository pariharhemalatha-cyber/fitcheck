package handler

import (
	"net/http"
	"sync"

	"github.com/ashokparihar/fitcheck/pkg/config"
	"github.com/ashokparihar/fitcheck/pkg/server"
)

var (
	initOnce sync.Once
	app      *server.App
	initErr  error
)

// Handler is the Vercel serverless entrypoint for all routes.
func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(func() {
		cfg, err := config.Load()
		if err != nil {
			initErr = err
			return
		}
		app, initErr = server.Bootstrap(cfg)
	})

	if initErr != nil {
		http.Error(w, "FitCheck failed to start: "+initErr.Error(), http.StatusInternalServerError)
		return
	}
	app.Handler.ServeHTTP(w, r)
}
