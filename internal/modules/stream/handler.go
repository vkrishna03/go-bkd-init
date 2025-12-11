package stream

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperr "github.com/vkrishna03/streamz/internal/errors"
	"github.com/vkrishna03/streamz/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Start(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	var req StartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	resp, err := h.svc.Start(c.Request.Context(), userID, req)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	resp, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListActive(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	resp, err := h.svc.ListActive(c.Request.Context(), userID)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	streamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid stream id"))
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), userID, streamID)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	streamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid stream id"))
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), userID, streamID, req); err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "status updated"})
}

func (h *Handler) UpdateLatency(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	streamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid stream id"))
		return
	}

	var req UpdateLatencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	if err := h.svc.UpdateLatency(c.Request.Context(), userID, streamID, req.LatencyMs); err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "latency updated"})
}

func (h *Handler) UpdateConnectionType(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	streamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid stream id"))
		return
	}

	var req UpdateConnectionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	if err := h.svc.UpdateConnectionType(c.Request.Context(), userID, streamID, req); err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "connection type updated"})
}

func (h *Handler) End(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	streamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid stream id"))
		return
	}

	if err := h.svc.End(c.Request.Context(), userID, streamID); err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "stream ended"})
}

// Setup registers stream routes
func Setup(api *gin.RouterGroup, db *sql.DB, jwtSecret string) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	r := api.Group("/streams")
	r.Use(middleware.Auth(jwtSecret))

	r.GET("", h.List)
	r.GET("/active", h.ListActive)
	r.POST("", h.Start)
	r.GET("/:id", h.Get)
	r.PUT("/:id/status", h.UpdateStatus)
	r.PUT("/:id/latency", h.UpdateLatency)
	r.PUT("/:id/connection-type", h.UpdateConnectionType)
	r.DELETE("/:id", h.End)
}
