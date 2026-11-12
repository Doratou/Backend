package handler

import (
	"errors"
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type LuckyDrawHandler struct{ drawService adapter.LuckyDrawService }

func NewLuckyDrawHandler(drawService adapter.LuckyDrawService) *LuckyDrawHandler {
	return &LuckyDrawHandler{drawService: drawService}
}

func (h *LuckyDrawHandler) Create(c *gin.Context) {
	userID, groupID, ok := h.getGroupIdentity(c)
	if !ok {
		return
	}
	var req dto.CreateLuckyDrawRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256<<10)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.drawService.Create(c.Request.Context(), userID, groupID, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	success(c, http.StatusCreated, res)
}

func (h *LuckyDrawHandler) List(c *gin.Context) {
	userID, groupID, ok := h.getGroupIdentity(c)
	if !ok {
		return
	}
	var req dto.ListLuckyDrawsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.drawService.List(c.Request.Context(), userID, groupID, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *LuckyDrawHandler) Get(c *gin.Context) {
	userID, groupID, drawID, ok := h.getRequestIdentity(c)
	if !ok {
		return
	}
	res, err := h.drawService.Get(c.Request.Context(), userID, groupID, drawID)
	if err != nil {
		h.fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *LuckyDrawHandler) Update(c *gin.Context) {
	userID, groupID, drawID, ok := h.getRequestIdentity(c)
	if !ok {
		return
	}
	var req dto.UpdateLuckyDrawRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256<<10)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.drawService.Update(c.Request.Context(), userID, groupID, drawID, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *LuckyDrawHandler) Run(c *gin.Context) {
	userID, groupID, drawID, ok := h.getRequestIdentity(c)
	if !ok {
		return
	}
	res, err := h.drawService.Run(c.Request.Context(), userID, groupID, drawID)
	if err != nil {
		h.fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *LuckyDrawHandler) getGroupIdentity(c *gin.Context) (int64, int64, bool) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		h.fail(c, exception.ErrMissingToken)
		return 0, 0, false
	}
	if userID <= 0 {
		h.fail(c, exception.ErrTokenInvalid)
		return 0, 0, false
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil || groupID <= 0 {
		h.fail(c, exception.ErrBadRequest)
		return 0, 0, false
	}
	return userID, groupID, true
}

func (h *LuckyDrawHandler) getRequestIdentity(c *gin.Context) (int64, int64, int64, bool) {
	userID, groupID, ok := h.getGroupIdentity(c)
	if !ok {
		return 0, 0, 0, false
	}
	drawID, err := parseIDParam(c, "drawID")
	if err != nil || drawID <= 0 {
		h.fail(c, exception.ErrBadRequest)
		return 0, 0, 0, false
	}
	return userID, groupID, drawID, true
}

func (h *LuckyDrawHandler) fail(c *gin.Context, err error) {
	var appErr *exception.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus >= http.StatusInternalServerError {
		failLogged(c, err)
		return
	}
	fail(c, err)
}
