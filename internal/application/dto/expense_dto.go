package dto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type ExpenseShareRequest struct {
	MemberID int64           `json:"member_id" binding:"required,gt=0"`
	Value    json.RawMessage `json:"value"`
}

type SaveExpenseRequest struct {
	Title            string                `json:"title" binding:"required,min=1,max=200"`
	Description      *string               `json:"description"`
	TotalAmountToman int64                 `json:"total_amount_toman" binding:"required,gt=0"`
	SplitType        enum.ExpenseSplitType `json:"split_type" binding:"required,oneof=EQUAL EXACT PERCENTAGE WEIGHTED"`
	OccurredAt       *time.Time            `json:"occurred_at"`
	Shares           []ExpenseShareRequest `json:"shares" binding:"required,min=1,dive"`
	AttachmentIDs    *[]int64              `json:"attachment_ids" binding:"omitempty,dive,gt=0"`
}

type ExpenseField[T any] struct {
	Present bool
	Null    bool
	Value   T
}

func (f *ExpenseField[T]) UnmarshalJSON(raw []byte) error {
	f.Present = true
	f.Null = bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
	if f.Null {
		return nil
	}
	return json.Unmarshal(raw, &f.Value)
}

type PatchExpenseRequest struct {
	Title            ExpenseField[string]                `json:"title"`
	Description      ExpenseField[string]                `json:"description"`
	TotalAmountToman ExpenseField[int64]                 `json:"total_amount_toman"`
	SplitType        ExpenseField[enum.ExpenseSplitType] `json:"split_type"`
	OccurredAt       ExpenseField[time.Time]             `json:"occurred_at"`
	Shares           ExpenseField[[]ExpenseShareRequest] `json:"shares"`
	AttachmentIDs    ExpenseField[[]int64]               `json:"attachment_ids"`
}

type ExpensePaginationRequest struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}

type ListExpensesRequest struct {
	ExpensePaginationRequest
	Scope enum.ExpenseListScope `form:"scope,default=all" binding:"oneof=all created_by_me owed_by_me"`
}

type ExpenseParticipantResponse struct {
	MemberID           int64                  `json:"member_id"`
	UserID             int64                  `json:"user_id"`
	Name               string                 `json:"name"`
	AvatarAttachmentID *int64                 `json:"avatar_attachment_id"`
	AvatarURL          *string                `json:"avatar_url"`
	Status             enum.GroupMemberStatus `json:"status"`
	JoinedAt           time.Time              `json:"joined_at"`
	LeftAt             *time.Time             `json:"left_at"`
}

type ExpenseShareResponse struct {
	MemberID         int64       `json:"member_id"`
	Value            json.Number `json:"value"`
	ShareAmountToman int64       `json:"share_amount_toman"`
}

type ExpenseResponse struct {
	ID                 int64                       `json:"id"`
	GroupID            int64                       `json:"group_id"`
	CreatedByMemberID  int64                       `json:"created_by_member_id"`
	PaidByMemberID     int64                       `json:"paid_by_member_id"`
	Title              string                      `json:"title"`
	Description        *string                     `json:"description"`
	TotalAmountToman   int64                       `json:"total_amount_toman"`
	ExpenseType        enum.ExpenseType            `json:"expense_type"`
	SplitType          enum.ExpenseSplitType       `json:"split_type"`
	OccurredAt         *time.Time                  `json:"occurred_at"`
	CreatedAt          time.Time                   `json:"created_at"`
	UpdatedAt          time.Time                   `json:"updated_at"`
	IsFrozen           bool                        `json:"is_frozen"`
	CanEdit            bool                        `json:"can_edit"`
	CanDelete          bool                        `json:"can_delete"`
	Shares             []ExpenseShareResponse      `json:"shares"`
	MyShareAmountToman int64                       `json:"my_share_amount_toman"`
	AttachmentIDs      []int64                     `json:"attachment_ids"`
	Attachments        []ExpenseAttachmentResponse `json:"attachments"`
}

type ExpenseAttachmentResponse struct {
	ID               int64   `json:"id"`
	OriginalFilename *string `json:"original_filename"`
	MimeType         *string `json:"mime_type"`
	FileSizeBytes    *int64  `json:"file_size_bytes"`
	AccessURL        *string `json:"access_url"`
}

type ExpensePaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type ExpenseListResponse struct {
	Items      []ExpenseResponse         `json:"items"`
	Pagination ExpensePaginationResponse `json:"pagination"`
}

type ExpenseParticipantListResponse struct {
	Items      []ExpenseParticipantResponse `json:"items"`
	Pagination ExpensePaginationResponse    `json:"pagination"`
}

func ToExpenseResponse(expense *entity.Expense, group *entity.Group, member *entity.GroupMember) ExpenseResponse {
	frozen := group.Status != enum.GroupStatusActive
	editable := !frozen && (expense.CreatedByMemberID == member.ID || group.CreatedByUserID == member.UserID)
	response := ExpenseResponse{
		ID: expense.ID, GroupID: expense.GroupID, CreatedByMemberID: expense.CreatedByMemberID,
		PaidByMemberID: expense.CreatedByMemberID, Title: expense.Title, Description: expense.Description,
		TotalAmountToman: expense.TotalAmountToman, ExpenseType: expense.ExpenseType, SplitType: expense.SplitType,
		OccurredAt: expense.OccurredAt, CreatedAt: expense.CreatedAt, UpdatedAt: expense.UpdatedAt,
		IsFrozen: frozen, CanEdit: editable, CanDelete: editable,
		Shares: make([]ExpenseShareResponse, 0, len(expense.Shares)), AttachmentIDs: make([]int64, 0, len(expense.Attachments)),
		Attachments: make([]ExpenseAttachmentResponse, 0, len(expense.Attachments)),
	}
	for _, share := range expense.Shares {
		if share.DeletedAt.Valid {
			continue
		}
		response.Shares = append(response.Shares, ExpenseShareResponse{MemberID: share.MemberID, Value: ExpenseSplitJSONValue(expense.SplitType, share.SplitValue), ShareAmountToman: share.ShareAmountToman})
		if share.MemberID == member.ID {
			response.MyShareAmountToman = share.ShareAmountToman
		}
	}
	for _, attachment := range expense.Attachments {
		if !attachment.DeletedAt.Valid && attachment.Type == enum.AttachmentTypeExpenseReceipt {
			response.AttachmentIDs = append(response.AttachmentIDs, attachment.ID)
			response.Attachments = append(response.Attachments, ExpenseAttachmentResponse{ID: attachment.ID, OriginalFilename: attachment.OriginalFilename, MimeType: attachment.MimeType, FileSizeBytes: attachment.FileSizeBytes})
		}
	}
	return response
}

func ExpenseSplitJSONValue(splitType enum.ExpenseSplitType, value int64) json.Number {
	if splitType == enum.ExpenseSplitTypePercentage {
		return json.Number(fmt.Sprintf("%d.%02d", value/100, value%100))
	}
	return json.Number(strconv.FormatInt(value, 10))
}

func ToExpenseParticipantResponse(member *entity.GroupMember) ExpenseParticipantResponse {
	response := ExpenseParticipantResponse{MemberID: member.ID, UserID: member.UserID, Status: member.Status, JoinedAt: member.JoinedAt, LeftAt: member.LeftAt}
	if member.User != nil {
		response.Name = member.User.Name
		response.AvatarAttachmentID = member.User.AvatarAttachmentID
	}
	return response
}
