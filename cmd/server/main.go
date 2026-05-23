package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"backend-dev-val/internal/di"

	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_USER", "postgres"),
		env("DB_PASSWORD", "postgres"),
		env("DB_NAME", "backend_dev_val"),
		env("DB_SSLMODE", "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		return
	}
	slog.Info("connected to database")

	engine := di.InitializeApp(db)

	addr := ":" + env("PORT", "8080")
	slog.Info("starting server", "addr", addr)
	if err := engine.Run(addr); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
