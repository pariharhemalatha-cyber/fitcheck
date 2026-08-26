package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ashokparihar/fitcheck/pkg/config"
	"github.com/ashokparihar/fitcheck/pkg/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	app, err := server.Bootstrap(cfg)
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	defer func() {
		if err := app.Cleanup(); err != nil {
			log.Printf("cleanup: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("FitCheck listening on %s", addr)

	if err := http.ListenAndServe(addr, app.Handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
