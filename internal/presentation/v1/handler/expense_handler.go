package handler

import (
	"net/http"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type ExpenseHandler struct{ expenseService adapter.ExpenseService }

func NewExpenseHandler(expenseService adapter.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	userID, groupID, ok := h.groupIdentity(c)
	if !ok {
		return
	}
	var req dto.SaveExpenseRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.expenseService.Create(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusCreated, res)
}

func (h *ExpenseHandler) List(c *gin.Context) {
	userID, groupID, ok := h.groupIdentity(c)
	if !ok {
		return
	}
	var req dto.ListExpensesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.expenseService.List(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *ExpenseHandler) Get(c *gin.Context) {
	userID, groupID, expenseID, ok := h.expenseIdentity(c)
	if !ok {
		return
	}
	res, err := h.expenseService.Get(c.Request.Context(), userID, groupID, expenseID)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	userID, groupID, expenseID, ok := h.expenseIdentity(c)
	if !ok {
		return
	}
	var req dto.SaveExpenseRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.expenseService.Update(c.Request.Context(), userID, groupID, expenseID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	userID, groupID, expenseID, ok := h.expenseIdentity(c)
	if !ok {
		return
	}
	if err := h.expenseService.Delete(c.Request.Context(), userID, groupID, expenseID); err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, gin.H{"message": "expense deleted successfully"})
}

func (h *ExpenseHandler) Patch(c *gin.Context) {
	userID, groupID, expenseID, ok := h.expenseIdentity(c)
	if !ok {
		return
	}
	var req dto.PatchExpenseRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.expenseService.Patch(c.Request.Context(), userID, groupID, expenseID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *ExpenseHandler) ListParticipants(c *gin.Context) {
	userID, groupID, ok := h.groupIdentity(c)
	if !ok {
		return
	}
	var req dto.ExpensePaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		fail(c, exception.Wrap(exception.ErrBadRequest, err))
		return
	}
	res, err := h.expenseService.ListParticipants(c.Request.Context(), userID, groupID, req)
	if err != nil {
		fail(c, err)
		return
	}
	success(c, http.StatusOK, res)
}

func (h *ExpenseHandler) groupIdentity(c *gin.Context) (int64, int64, bool) {
	userID, ok := tokencontext.GetUserID(c)
	if !ok {
		fail(c, exception.ErrMissingToken)
		return 0, 0, false
	}
	if userID <= 0 {
		fail(c, exception.ErrTokenInvalid)
		return 0, 0, false
	}
	groupID, err := parseIDParam(c, "groupID")
	if err != nil || groupID <= 0 {
		fail(c, exception.ErrBadRequest)
		return 0, 0, false
	}
	return userID, groupID, true
}

func (h *ExpenseHandler) expenseIdentity(c *gin.Context) (int64, int64, int64, bool) {
	userID, groupID, ok := h.groupIdentity(c)
	if !ok {
		return 0, 0, 0, false
	}
	expenseID, err := parseIDParam(c, "expenseID")
	if err != nil || expenseID <= 0 {
		fail(c, exception.ErrBadRequest)
		return 0, 0, 0, false
	}
	return userID, groupID, expenseID, true
}
