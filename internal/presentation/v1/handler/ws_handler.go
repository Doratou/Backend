package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Doratou/Backend/internal/domain/exception"
	domain_ws "github.com/Doratou/Backend/internal/domain/websocket"
	"github.com/gin-gonic/gin"
)

// WSHandler adapts the framework-agnostic WebSocket server and monitor to Gin HTTP routes.
type WSHandler struct {
	server  http.Handler
	monitor domain_ws.Monitor
}

// NewWSHandler creates a new WSHandler delegating WebSocket upgrades to server and monitoring queries to monitor.
func NewWSHandler(server http.Handler, monitor domain_ws.Monitor) *WSHandler {
	return &WSHandler{
		server:  server,
		monitor: monitor,
	}
}

// Connect handles incoming WebSocket handshake requests on GET /api/v1/ws.
func (h *WSHandler) Connect(c *gin.Context) {
	if h == nil || h.server == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	h.server.ServeHTTP(c.Writer, c.Request)
}

// Stats handles GET /api/v1/admin/ws/stats and returns current WebSocket hub operational metrics.
func (h *WSHandler) Stats(c *gin.Context) {
	if h == nil || h.monitor == nil {
		fail(c, exception.ErrInternal)
		return
	}
	success(c, http.StatusOK, h.monitor.Stats())
}

// Snapshot handles GET /api/v1/admin/ws/snapshot and returns a filtered, bounded point-in-time snapshot
// of active channels and connected WebSocket sessions.
func (h *WSHandler) Snapshot(c *gin.Context) {
	if h == nil || h.monitor == nil {
		fail(c, exception.ErrInternal)
		return
	}

	var filter domain_ws.SnapshotFilter
	if rawUserID := strings.TrimSpace(c.Query("user_id")); rawUserID != "" {
		uid, err := strconv.ParseInt(rawUserID, 10, 64)
		if err != nil || uid <= 0 {
			fail(c, exception.Wrap(exception.ErrBadRequest, err))
			return
		}
		filter.UserID = uid
	}

	filter.Channel = strings.TrimSpace(c.Query("channel"))

	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			fail(c, exception.Wrap(exception.ErrBadRequest, err))
			return
		}
		filter.Limit = limit
	}

	success(c, http.StatusOK, h.monitor.Snapshot(filter))
}
