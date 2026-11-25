package websocket

import "errors"

var (
	// ErrHubClosed indicates that the WebSocket hub has been shut down and no longer accepts operations.
	ErrHubClosed = errors.New("websocket: hub is closed")

	// ErrConnectionNotFound indicates that the target WebSocket connection ID does not exist in the hub.
	ErrConnectionNotFound = errors.New("websocket: connection not found")

	// ErrClientClosed indicates that the target client connection has already been closed.
	ErrClientClosed = errors.New("websocket: client connection is closed")

	// ErrSendBufferFull indicates that a client's outbound queue has reached capacity.
	ErrSendBufferFull = errors.New("websocket: client send buffer is full")

	// ErrInvalidEnvelope indicates that an incoming or outgoing WebSocket envelope is malformed.
	ErrInvalidEnvelope = errors.New("websocket: invalid envelope")

	// ErrInvalidChannel indicates that a channel/room name is empty or invalid.
	ErrInvalidChannel = errors.New("websocket: invalid channel name")

	// ErrTooManyConnections indicates that the user exceeded the maximum allowed concurrent connections.
	ErrTooManyConnections = errors.New("websocket: too many active connections for user")

	// ErrUnauthorized indicates that the WebSocket handshake failed authentication.
	ErrUnauthorized = errors.New("websocket: unauthorized")
)
