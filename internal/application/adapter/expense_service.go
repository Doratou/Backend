package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type ExpenseService interface {
	Create(ctx context.Context, userID, groupID int64, req dto.SaveExpenseRequest) (*dto.ExpenseResponse, error)
	List(ctx context.Context, userID, groupID int64, req dto.ListExpensesRequest) (*dto.ExpenseListResponse, error)
	Get(ctx context.Context, userID, groupID, expenseID int64) (*dto.ExpenseResponse, error)
	Update(ctx context.Context, userID, groupID, expenseID int64, req dto.SaveExpenseRequest) (*dto.ExpenseResponse, error)
	Patch(ctx context.Context, userID, groupID, expenseID int64, req dto.PatchExpenseRequest) (*dto.ExpenseResponse, error)
	Delete(ctx context.Context, userID, groupID, expenseID int64) error
	ListParticipants(ctx context.Context, userID, groupID int64, req dto.ListExpensesRequest) (*dto.ExpenseParticipantListResponse, error)
}
