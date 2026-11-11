package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type LuckyDrawCandidateInput struct {
	Title  string `json:"title" binding:"required"`
	Weight int64  `json:"weight" binding:"required,min=1,max=1000000"`
}

type CreateLuckyDrawRequest struct {
	Title      string                    `json:"title" binding:"required"`
	Candidates []LuckyDrawCandidateInput `json:"candidates" binding:"required,min=2,max=100,dive"`
}

type UpdateLuckyDrawRequest struct {
	Title      *string                    `json:"title" binding:"omitempty"`
	Candidates *[]LuckyDrawCandidateInput `json:"candidates" binding:"omitempty,min=2,max=100,dive"`
}

type ListLuckyDrawsRequest struct {
	Status enum.LuckyDrawStatus `form:"status" binding:"omitempty,oneof=DRAFT COMPLETED"`
	Page   int                  `form:"page,default=1" binding:"min=1"`
	Limit  int                  `form:"limit,default=20" binding:"min=1,max=100"`
}

type LuckyDrawResult struct {
	Index  int    `json:"index"`
	Title  string `json:"title"`
	Weight int64  `json:"weight"`
}

type LuckyDrawListItem struct {
	ID                int64                `json:"id"`
	GroupID           int64                `json:"group_id"`
	CreatedByMemberID int64                `json:"created_by_member_id"`
	Title             string               `json:"title"`
	Status            enum.LuckyDrawStatus `json:"status"`
	IsFrozen          bool                 `json:"is_frozen"`
	Result            *LuckyDrawResult     `json:"result"`
	CompletedAt       *time.Time           `json:"completed_at"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}

type LuckyDrawResponse struct {
	LuckyDrawListItem
	Candidates  []LuckyDrawCandidateResponse `json:"candidates"`
	TotalWeight int64                        `json:"total_weight"`
}

type LuckyDrawCandidateResponse struct {
	Title  string `json:"title"`
	Weight int64  `json:"weight"`
}

func ToLuckyDrawListItem(draw *entity.LuckyDraw, frozen bool) LuckyDrawListItem {
	item := LuckyDrawListItem{ID: draw.ID, GroupID: draw.GroupID, CreatedByMemberID: draw.CreatedByMemberID,
		Title: draw.Title, Status: draw.Status, IsFrozen: frozen,
		CreatedAt: draw.CreatedAt.UTC(), UpdatedAt: draw.UpdatedAt.UTC()}
	if draw.CompletedAt != nil {
		completedAt := draw.CompletedAt.UTC()
		item.CompletedAt = &completedAt
	}
	if draw.ResultIndex != nil && *draw.ResultIndex >= 0 && *draw.ResultIndex < len(draw.Candidates) {
		candidate := draw.Candidates[*draw.ResultIndex]
		item.Result = &LuckyDrawResult{Index: *draw.ResultIndex, Title: candidate.Title, Weight: candidate.Weight}
	}
	return item
}

func ToLuckyDrawResponse(draw *entity.LuckyDraw, frozen bool) LuckyDrawResponse {
	response := LuckyDrawResponse{LuckyDrawListItem: ToLuckyDrawListItem(draw, frozen),
		Candidates: make([]LuckyDrawCandidateResponse, 0, len(draw.Candidates))}
	for _, candidate := range draw.Candidates {
		response.Candidates = append(response.Candidates, LuckyDrawCandidateResponse{Title: candidate.Title, Weight: candidate.Weight})
		response.TotalWeight += candidate.Weight
	}
	return response
}
