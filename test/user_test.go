package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreateUserValidation(t *testing.T) {
	r := gin.New()

	r.POST("/users", func(c *gin.Context) {
		var req struct {
			Name  string `json:"name" binding:"required"`
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": req.Name, "email": req.Email})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       map[string]string{"name": "John", "email": "john@example.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing name",
			body:       map[string]string{"email": "john@example.com"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email",
			body:       map[string]string{"name": "John", "email": "not-an-email"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
