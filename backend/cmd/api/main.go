package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"taskmanager/internal/config"
	"taskmanager/internal/database"
	"taskmanager/internal/server"
	"taskmanager/internal/task"
)

// @title		Task Management API
// @version	1.0
// @BasePath	/api
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
	// The cache is optional: when Redis is down, fail fast and fall back to MySQL.
	rdb := redis.NewClient(&redis.Options{
		Addr:          cfg.RedisAddr,
		DialTimeout:   200 * time.Millisecond,
		ReadTimeout:   200 * time.Millisecond,
		WriteTimeout:  200 * time.Millisecond,
		MaxRetries:    -1,
		DialerRetries: 1, // default is 5 attempts per dial
	})
	handler := task.NewHandler(task.NewService(repo, task.NewRedisListCache(rdb)))
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
