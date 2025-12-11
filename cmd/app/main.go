package main

import (
	"log/slog"
	"os"

	"github.com/vkrishna03/streamz/internal/config"
	"github.com/vkrishna03/streamz/internal/database"
	"github.com/vkrishna03/streamz/internal/modules/auth"
	"github.com/vkrishna03/streamz/internal/modules/device"
	"github.com/vkrishna03/streamz/internal/modules/stream"
	"github.com/vkrishna03/streamz/internal/server"
)

func main() {
	// Config
	cfg := config.Load()

	// Database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	// Server
	srv := server.New(cfg)

	// Module routes
	api := srv.Router().Group("/api/v1")

	// Auth module (public routes)
	auth.Setup(api, db, auth.Config{
		JWTSecret:        cfg.JWT.Secret,
		JWTExpiry:        cfg.JWT.Expiry,
		RefreshExpiry:    cfg.JWT.RefreshExpiry,
		PasswordResetExp: cfg.JWT.PasswordResetExp,
	})

	// Device module (protected routes)
	device.Setup(api, db, cfg.JWT.Secret)

	// Stream module (protected routes)
	stream.Setup(api, db, cfg.JWT.Secret)

	// Run
	if err := srv.Run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
