package ztocopy

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// TODO: Add handler methods
func (h *Handler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// Setup wires dependencies and registers routes
func Setup(api *gin.RouterGroup, db *sql.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	r := api.Group("/ztocopy") // TODO: Change route
	r.GET("/:id", h.Get)
}
