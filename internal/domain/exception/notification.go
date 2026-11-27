package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainNotificationNotFound = errors.New("notification not found")
)

// Notification domain and application errors
var (
	ErrNotificationNotFound = &AppError{
		Code:       "NOTIFICATION_NOT_FOUND",
		Message:    "notification not found",
		HTTPStatus: http.StatusNotFound,
		Err:        ErrDomainNotificationNotFound,
	}
)
