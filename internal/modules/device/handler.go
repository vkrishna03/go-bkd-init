package device

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

func (h *Handler) Register(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), userID, req)
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

func (h *Handler) ListOnline(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	resp, err := h.svc.ListOnline(c.Request.Context(), userID)
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

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid device id"))
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), userID, deviceID)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid device id"))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), userID, deviceID, req)
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

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid device id"))
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), userID, deviceID, req.IsOnline); err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "status updated"})
}

func (h *Handler) Heartbeat(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid device id"))
		return
	}

	if err := h.svc.Heartbeat(c.Request.Context(), userID, deviceID); err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "heartbeat received"})
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		apperr.Response(c, apperr.ErrUnauthorized)
		return
	}

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid device id"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, deviceID); err != nil {
		apperr.Response(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Setup registers device routes
func Setup(api *gin.RouterGroup, db *sql.DB, jwtSecret string) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	r := api.Group("/devices")
	r.Use(middleware.Auth(jwtSecret))

	r.GET("", h.List)
	r.GET("/online", h.ListOnline)
	r.POST("", h.Register)
	r.GET("/:id", h.Get)
	r.PUT("/:id", h.Update)
	r.PUT("/:id/status", h.UpdateStatus)
	r.POST("/:id/heartbeat", h.Heartbeat)
	r.DELETE("/:id", h.Delete)
}
