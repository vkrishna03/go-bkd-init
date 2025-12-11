package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	apperr "github.com/vkrishna03/streamz/internal/errors"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	resp, err := h.svc.Refresh(c.Request.Context(), req)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	_ = h.svc.Logout(c.Request.Context(), req.RefreshToken)
	c.JSON(http.StatusOK, MessageResponse{Message: "Logged out successfully"})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	resp, err := h.svc.ForgotPassword(c.Request.Context(), req)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	resp, err := h.svc.ResetPassword(c.Request.Context(), req)
	if err != nil {
		apperr.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Setup registers auth routes
func Setup(api *gin.RouterGroup, db *sql.DB, cfg Config) {
	repo := NewRepository(db)
	svc := NewService(repo, cfg)
	h := NewHandler(svc)

	r := api.Group("/auth")
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
	r.POST("/logout", h.Logout)
	r.POST("/forgot-password", h.ForgotPassword)
	r.POST("/reset-password", h.ResetPassword)
}
