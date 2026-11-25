package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WSHandler adapts the framework-agnostic WebSocket server to Gin HTTP routes.
type WSHandler struct {
	server http.Handler
}

// NewWSHandler creates a new WSHandler delegating WebSocket upgrade requests to server.
func NewWSHandler(server http.Handler) *WSHandler {
	return &WSHandler{server: server}
}

// Connect handles incoming WebSocket handshake requests on GET /api/v1/ws.
func (h *WSHandler) Connect(c *gin.Context) {
	if h == nil || h.server == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	h.server.ServeHTTP(c.Writer, c.Request)
}
