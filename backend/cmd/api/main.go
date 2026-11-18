package main

import (
	"log/slog"
	"net/http"
	"os"

	"taskmanager/internal/config"
	"taskmanager/internal/database"
	"taskmanager/internal/server"
	"taskmanager/internal/task"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.MySQLDSN)
	if err != nil {
		fatal("open database", err)
	}
	if err := database.Migrate(db); err != nil {
		fatal("migrate database", err)
	}

	repo := task.NewGormRepository(db)
	handler := task.NewHandler(task.NewService(repo))
	router := server.New(handler, cfg.CORSOrigins)

	slog.Info("api listening", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		fatal("serve", err)
	}
}

func fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}
