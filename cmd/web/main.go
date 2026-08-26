package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ashokparihar/fitcheck/pkg/config"
	"github.com/ashokparihar/fitcheck/pkg/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	handler, cleanup, err := server.NewWithSQLite(cfg)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			log.Printf("db close: %v", err)
		}
	}()

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("FitCheck running at http://localhost%s (sqlite: %s)", addr, cfg.SQLitePath)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
