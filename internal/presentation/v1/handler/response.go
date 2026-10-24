package handler

import (
	"errors"
	"strconv"

	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/gin-gonic/gin"
)

// success writes the uniform success envelope used across handlers.
func success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

// fail writes the uniform error envelope derived from an *exception.AppError.
// Any unknown error type is normalized to ErrInternal so internal details are never leaked.
func fail(c *gin.Context, err error) {
	var appErr *exception.AppError
	if !errors.As(err, &appErr) {
		appErr = exception.ErrInternal
	}

	c.JSON(appErr.HTTPStatus, gin.H{
		"success": false,
		"code":    appErr.Code,
		"error":   appErr.Message,
	})
}

// parseIDParam parses an int64 route parameter (e.g. :id),
// returning ErrBadRequest on invalid input.
func parseIDParam(c *gin.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		return 0, exception.Wrap(exception.ErrBadRequest, err)
	}
	return id, nil
}
