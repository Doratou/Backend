package dto

import (
	"encoding/json"
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type LuckyDrawCandidateInput struct {
	Type   enum.LuckyDrawCandidateType `json:"type" binding:"required,oneof=USER TEXT"`
	Value  json.RawMessage             `json:"value" binding:"required"`
	Weight *entity.LuckyDrawPercentage `json:"weight"`
}

type CreateLuckyDrawRequest struct {
	Title      string                    `json:"title" binding:"required"`
	DrawType   enum.LuckyDrawType        `json:"draw_type" binding:"required,oneof=EQUAL WEIGHTED"`
	Candidates []LuckyDrawCandidateInput `json:"candidates" binding:"required,min=2,max=100,dive"`
}

type UpdateLuckyDrawRequest struct {
	Title      *string                    `json:"title" binding:"omitempty"`
	DrawType   *enum.LuckyDrawType        `json:"draw_type" binding:"omitempty,oneof=EQUAL WEIGHTED"`
	Candidates *[]LuckyDrawCandidateInput `json:"candidates" binding:"omitempty,min=2,max=100,dive"`
}

type ListLuckyDrawsRequest struct {
	Status enum.LuckyDrawStatus `form:"status" binding:"omitempty,oneof=DRAFT COMPLETED"`
	Page   int                  `form:"page,default=1" binding:"min=1"`
	Limit  int                  `form:"limit,default=20" binding:"min=1,max=100"`
}

type LuckyDrawResult struct {
	Index int `json:"index"`
	LuckyDrawCandidateResponse
}

type LuckyDrawListItem struct {
	ID                int64                `json:"id"`
	GroupID           int64                `json:"group_id"`
	CreatedByMemberID int64                `json:"created_by_member_id"`
	Title             string               `json:"title"`
	DrawType          enum.LuckyDrawType   `json:"draw_type"`
	CandidateCount    int                  `json:"candidate_count"`
	Status            enum.LuckyDrawStatus `json:"status"`
	IsFrozen          bool                 `json:"is_frozen"`
	CanEdit           bool                 `json:"can_edit"`
	CanRun            bool                 `json:"can_run"`
	Result            *LuckyDrawResult     `json:"result"`
	CompletedAt       *time.Time           `json:"completed_at"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}

type LuckyDrawResponse struct {
	LuckyDrawListItem
	Candidates  []LuckyDrawCandidateResponse `json:"candidates"`
	TotalWeight entity.LuckyDrawPercentage   `json:"total_weight"`
}

type LuckyDrawCandidateResponse struct {
	Type   enum.LuckyDrawCandidateType `json:"type"`
	Value  json.RawMessage             `json:"value"`
	Weight entity.LuckyDrawPercentage  `json:"weight"`
}

func ToLuckyDrawListItem(draw *entity.LuckyDraw, frozen bool, memberID int64) LuckyDrawListItem {
	open := draw.Status == enum.LuckyDrawStatusDraft && !frozen
	item := LuckyDrawListItem{ID: draw.ID, GroupID: draw.GroupID, CreatedByMemberID: draw.CreatedByMemberID,
		Title: draw.Title, DrawType: draw.DrawType, CandidateCount: len(draw.Candidates), Status: draw.Status, IsFrozen: frozen,
		CanEdit: open && memberID == draw.CreatedByMemberID, CanRun: open,
		CreatedAt: draw.CreatedAt.UTC(), UpdatedAt: draw.UpdatedAt.UTC()}
	if draw.CompletedAt != nil {
		completedAt := draw.CompletedAt.UTC()
		item.CompletedAt = &completedAt
	}
	if draw.ResultIndex != nil && *draw.ResultIndex >= 0 && *draw.ResultIndex < len(draw.Candidates) {
		candidate := draw.Candidates[*draw.ResultIndex]
		item.Result = &LuckyDrawResult{Index: *draw.ResultIndex, LuckyDrawCandidateResponse: toLuckyDrawCandidateResponse(candidate)}
	}
	return item
}

func ToLuckyDrawResponse(draw *entity.LuckyDraw, frozen bool, memberID int64) LuckyDrawResponse {
	response := LuckyDrawResponse{LuckyDrawListItem: ToLuckyDrawListItem(draw, frozen, memberID),
		Candidates: make([]LuckyDrawCandidateResponse, 0, len(draw.Candidates))}
	for _, candidate := range draw.Candidates {
		response.Candidates = append(response.Candidates, toLuckyDrawCandidateResponse(candidate))
		response.TotalWeight += candidate.Weight
	}
	return response
}

func toLuckyDrawCandidateResponse(candidate entity.LuckyDrawCandidate) LuckyDrawCandidateResponse {
	return LuckyDrawCandidateResponse{Type: candidate.Type, Value: candidate.Value, Weight: candidate.Weight}
}
