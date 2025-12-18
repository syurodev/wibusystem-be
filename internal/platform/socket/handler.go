package socket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"system/internal/app/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Create a custom CheckOrigin function
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now (dev mode)
		// TODO: Restrict origin in production
		return true
	},
}

// Handler handles websocket requests
type Handler struct {
	hub    *Hub
	logger *zap.Logger
}

// NewHandler creates a new websocket handler
func NewHandler(hub *Hub, logger *zap.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// HandleWebSocket handles websocket connections
// @Summary WebSocket endpoint
// @Tags Socket
// @Security BearerAuth
// @Router /ws [get]
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Authenticate user
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		// Try to get token from query param for WebSocket
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		// TODO: Validate token manually if needed, but optimally middleware should handle it
		// For now assume middleware handled it or we rejected earlier
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	client := &Client{
		hub:    h.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		logger: h.logger,
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
