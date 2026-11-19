package dto

import (
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"time"
)

type CreateSettlementRequest struct {
	ToMemberID    int64      `json:"to_member_id" binding:"required,gt=0"`
	AmountToman   int64      `json:"amount_toman" binding:"required,gt=0"`
	PaidAt        *time.Time `json:"paid_at"`
	Note          *string    `json:"note" binding:"omitempty,max=2000"`
	AttachmentIDs []int64    `json:"attachment_ids" binding:"max=100,dive,gt=0"`
}

type ListSettlementsRequest struct {
	Page      int                      `form:"page,default=1" binding:"min=1"`
	Limit     int                      `form:"limit,default=20" binding:"min=1,max=100"`
	Status    *enum.SettlementStatus   `form:"status" binding:"omitempty,oneof=PENDING SUCCESS REJECT CANCELED"`
	Direction enum.SettlementDirection `form:"direction,default=all" binding:"oneof=all incoming outgoing"`
}

type SettlementMemberResponse struct {
	MemberID  int64   `json:"member_id"`
	UserID    int64   `json:"user_id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

type SettlementWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SettlementResponse struct {
	ID            int64                        `json:"id"`
	GroupID       int64                        `json:"group_id"`
	FromMemberID  int64                        `json:"from_member_id"`
	ToMemberID    int64                        `json:"to_member_id"`
	FromMember    SettlementMemberResponse     `json:"from_member"`
	ToMember      SettlementMemberResponse     `json:"to_member"`
	AmountToman   int64                        `json:"amount_toman"`
	Status        enum.SettlementStatus        `json:"status"`
	Note          *string                      `json:"note"`
	PaidAt        time.Time                    `json:"paid_at"`
	ConfirmedAt   *time.Time                   `json:"confirmed_at"`
	RejectedAt    *time.Time                   `json:"rejected_at"`
	CanceledAt    *time.Time                   `json:"canceled_at"`
	CancelReason  *enum.SettlementCancelReason `json:"cancel_reason"`
	CreatedAt     time.Time                    `json:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at"`
	IsFrozen      bool                         `json:"is_frozen"`
	IsExternal    bool                         `json:"is_external"`
	CanConfirm    bool                         `json:"can_confirm"`
	CanReject     bool                         `json:"can_reject"`
	CanDelete     bool                         `json:"can_delete"`
	AttachmentIDs []int64                      `json:"attachment_ids"`
	Attachments   []ExpenseAttachmentResponse  `json:"attachments"`
	Warnings      []SettlementWarning          `json:"warnings"`
}

type SettlementListResponse struct {
	Items      []SettlementResponse      `json:"items"`
	Pagination ExpensePaginationResponse `json:"pagination"`
}

func ToSettlementResponse(payment *entity.Settlement, group *entity.Group, member *entity.GroupMember) SettlementResponse {
	frozen := group.Status == enum.GroupStatusClosed
	pending := payment.Status == enum.SettlementStatusPending && !frozen
	res := SettlementResponse{ID: payment.ID, GroupID: payment.GroupID, FromMemberID: payment.FromMemberID, ToMemberID: payment.ToMemberID,
		AmountToman: payment.AmountToman, Status: payment.Status, Note: payment.Note, PaidAt: payment.PaidAt, ConfirmedAt: payment.ConfirmedAt, RejectedAt: payment.RejectedAt,
		CanceledAt: payment.CanceledAt, CancelReason: payment.CancelReason, CreatedAt: payment.CreatedAt, UpdatedAt: payment.UpdatedAt,
		IsFrozen: frozen, CanConfirm: pending && member.ID == payment.ToMemberID, CanReject: pending && member.ID == payment.ToMemberID, CanDelete: pending && member.ID == payment.FromMemberID,
		AttachmentIDs: make([]int64, 0, len(payment.Attachments)), Attachments: make([]ExpenseAttachmentResponse, 0, len(payment.Attachments)), Warnings: make([]SettlementWarning, 0)}
	for _, pair := range []struct {
		source *entity.GroupMember
		target *SettlementMemberResponse
	}{{payment.FromMember, &res.FromMember}, {payment.ToMember, &res.ToMember}} {
		if pair.source == nil {
			continue
		}
		pair.target.MemberID = pair.source.ID
		pair.target.UserID = pair.source.UserID
		if pair.source.User != nil {
			pair.target.Name = pair.source.User.Name
		}
		if pair.source.Status != enum.GroupMemberStatusActive || pair.source.LeftAt != nil {
			res.IsExternal = true
		}
	}
	for _, att := range payment.Attachments {
		if !att.DeletedAt.Valid && att.Type == enum.AttachmentTypeSettlementReceipt {
			res.AttachmentIDs = append(res.AttachmentIDs, att.ID)
			res.Attachments = append(res.Attachments, ExpenseAttachmentResponse{ID: att.ID, OriginalFilename: att.OriginalFilename, MimeType: att.MimeType, FileSizeBytes: att.FileSizeBytes})
		}
	}
	return res
}
