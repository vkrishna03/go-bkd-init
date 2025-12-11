package main

import (
	"log/slog"
	"os"

	"github.com/vkrishna03/go-bkd-init/internal/config"
	"github.com/vkrishna03/go-bkd-init/internal/database"
	"github.com/vkrishna03/go-bkd-init/internal/modules/user"
	"github.com/vkrishna03/go-bkd-init/internal/server"
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
	user.Setup(api, db)

	// Run
	if err := srv.Run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
