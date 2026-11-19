package handler

import (
	"strconv"

	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/gin-gonic/gin"
)

// parsePagination reads the shared page/limit query parameters and converts
// them to the limit/offset values expected by repositories.
func parsePagination(c *gin.Context) (page, limit, offset int, err error) {
	page, err = strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		return 0, 0, 0, exception.ErrBadRequest
	}
	limit, err = strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, 0, exception.ErrBadRequest
	}
	if page-1 > int(^uint(0)>>1)/limit {
		return 0, 0, 0, exception.ErrBadRequest
	}
	return page, limit, (page - 1) * limit, nil
}
