package webrtc

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vkrishna03/streamz/internal/config"
	"github.com/vkrishna03/streamz/internal/middleware"
)

// ICEServer represents a single ICE server configuration for WebRTC
type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// ICEServersResponse is the response for the ice-servers endpoint
type ICEServersResponse struct {
	ICEServers []ICEServer `json:"ice_servers"`
}

// Handler handles WebRTC configuration endpoints
type Handler struct {
	cfg config.ICEConfig
}

// NewHandler creates a new WebRTC handler
func NewHandler(cfg config.ICEConfig) *Handler {
	return &Handler{cfg: cfg}
}

// GetICEServers returns ICE server configuration for WebRTC connections
func (h *Handler) GetICEServers(c *gin.Context) {
	servers := make([]ICEServer, 0)

	// Add STUN servers (no credentials needed)
	for _, stun := range h.cfg.STUNServers {
		servers = append(servers, ICEServer{
			URLs: []string{stun},
		})
	}

	// Add TURN servers (with credentials)
	for _, turn := range h.cfg.TURNServers {
		servers = append(servers, ICEServer{
			URLs:       []string{turn},
			Username:   h.cfg.TURNUsername,
			Credential: h.cfg.TURNCredential,
		})
	}

	c.JSON(http.StatusOK, ICEServersResponse{
		ICEServers: servers,
	})
}

// Setup registers WebRTC routes
func Setup(router *gin.RouterGroup, cfg config.ICEConfig, jwtSecret string) {
	handler := NewHandler(cfg)

	webrtc := router.Group("/webrtc")
	webrtc.Use(middleware.Auth(jwtSecret))
	{
		webrtc.GET("/ice-servers", handler.GetICEServers)
	}
}
