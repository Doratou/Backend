package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
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

// failLogged records the underlying error for group and invite endpoints while
// keeping internal details out of the HTTP response.
func failLogged(c *gin.Context, err error) {
	var appErr *exception.AppError
	if !errors.As(err, &appErr) {
		appErr = exception.ErrInternal
	}
	userID, _ := tokencontext.GetUserID(c)
	log.Printf("[API ERROR] method=%s route=%s status=%d code=%s user_id=%d error=%v",
		c.Request.Method, c.FullPath(), appErr.HTTPStatus, appErr.Code, userID, err)
	fail(c, err)
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
