package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vkrishna03/go-bkd-init/internal/config"
	"github.com/vkrishna03/go-bkd-init/internal/middleware"
)

type Server struct {
	router *gin.Engine
	http   *http.Server
	cfg    *config.Config
}

func New(cfg *config.Config) *Server {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health endpoints
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return &Server{
		router: r,
		cfg:    cfg,
	}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Run() error {
	s.http = &http.Server{
		Addr:    ":" + s.cfg.Server.Port,
		Handler: s.router,
	}

	go func() {
		slog.Info("server starting", "port", s.cfg.Server.Port)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.http.Shutdown(ctx)
}
