package user

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperr "github.com/vkrishna03/go-bkd-init/internal/errors"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	users, err := h.svc.List(c.Request.Context())
	if err != nil {
		apperr.Response(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid user id: must be a number"))
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), int32(id))
	if err != nil {
		apperr.Response(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid request: %s", err.Error()))
		return
	}

	user, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		apperr.Response(c, apperr.Wrap(apperr.ErrValidation, "invalid user id: must be a number"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), int32(id)); err != nil {
		apperr.Response(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Setup wires dependencies and registers routes
func Setup(api *gin.RouterGroup, db *sql.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	r := api.Group("/users")
	r.GET("", h.List)
	r.GET("/:id", h.Get)
	r.POST("", h.Create)
	r.DELETE("/:id", h.Delete)
}
