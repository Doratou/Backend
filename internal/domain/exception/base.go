// Package exception defines the application's domain error vocabulary and uniform AppError envelopes.
package exception

import (
	"fmt"
)

// AppError is the standard error container throughout the application.
// Handlers translate AppErrors into JSON error envelopes without leaking internal details.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Wrap returns a new copy of sentinel with underlying err attached.
func Wrap(sentinel *AppError, err error) *AppError {
	if sentinel == nil {
		return nil
	}
	return &AppError{
		Code:       sentinel.Code,
		Message:    sentinel.Message,
		HTTPStatus: sentinel.HTTPStatus,
		Err:        err,
	}
}
