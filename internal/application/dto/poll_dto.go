package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type CreatePollRequest struct {
	Question             string              `json:"question" binding:"required,min=1,max=500"`
	Visibility           enum.PollVisibility `json:"visibility" binding:"required,oneof=PUBLIC PRIVATE"`
	AllowMultipleChoices bool                `json:"allow_multiple_choices"`
	Options              []string            `json:"options" binding:"required,min=2,dive,required,min=1,max=255"`
}

type ListPollsRequest struct {
	Status string `form:"status" binding:"omitempty,oneof=ACTIVE INACTIVE"`
	Page   int    `form:"page,default=1" binding:"min=1"`
	Limit  int    `form:"limit,default=20" binding:"min=1,max=100"`
}

type VotePollRequest struct {
	OptionIDs []int64 `json:"option_ids" binding:"required,min=1,dive,gt=0"`
}

type ListPollVotersRequest struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}

type PollListItem struct {
	PollResponse
}

type PollOptionResponse struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	VoteCount int64  `json:"vote_count"`
}

type PollResponse struct {
	ID                   int64                `json:"id"`
	GroupID              int64                `json:"group_id"`
	CreatedByMemberID    int64                `json:"created_by_member_id"`
	Question             string               `json:"question"`
	Visibility           enum.PollVisibility  `json:"visibility"`
	AllowMultipleChoices bool                 `json:"allow_multiple_choices"`
	Status               enum.PollStatus      `json:"status"`
	IsFrozen             bool                 `json:"is_frozen"`
	ClosedAt             *time.Time           `json:"closed_at"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
	Options              []PollOptionResponse `json:"options"`
	TotalVotes           int64                `json:"total_votes"`
	TotalVoters          int64                `json:"total_voters"`
	MyOptionIDs          []int64              `json:"my_option_ids"`
}

type PollPaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type PollListResponse struct {
	Items      []PollListItem         `json:"items"`
	Pagination PollPaginationResponse `json:"pagination"`
}

type PollVoteResponse struct {
	PollID    int64                       `json:"poll_id"`
	OptionIDs []int64                     `json:"option_ids"`
	Votes     []PollVoteSelectionResponse `json:"votes"`
}

type PollVoteSelectionResponse struct {
	OptionID  int64     `json:"option_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PollVoterUserResponse struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	AvatarAttachmentID *int64 `json:"avatar_attachment_id"`
}

type PollVoterResponse struct {
	MemberID  int64                 `json:"member_id"`
	User      PollVoterUserResponse `json:"user"`
	OptionID  int64                 `json:"option_id"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type PollVoterListResponse struct {
	Items      []PollVoterResponse    `json:"items"`
	Pagination PollPaginationResponse `json:"pagination"`
}

type PollStatusResponse struct {
	ID       int64           `json:"id"`
	Status   enum.PollStatus `json:"status"`
	ClosedAt *time.Time      `json:"closed_at"`
}

func ToPollListItem(poll *entity.Poll, summary domain_db.PollVoteSummary, isFrozen bool) PollListItem {
	return PollListItem{PollResponse: ToPollResponse(poll, summary, isFrozen)}
}

func ToPollResponse(poll *entity.Poll, summary domain_db.PollVoteSummary, isFrozen bool) PollResponse {
	response := PollResponse{
		ID: poll.ID, GroupID: poll.GroupID, CreatedByMemberID: poll.CreatedByMemberID,
		Question: poll.Question, Visibility: poll.Visibility, Status: poll.Status,
		AllowMultipleChoices: poll.AllowMultipleChoices, TotalVoters: summary.TotalVoters,
		IsFrozen: isFrozen, ClosedAt: poll.ClosedAt, CreatedAt: poll.CreatedAt, UpdatedAt: poll.UpdatedAt,
		Options:     make([]PollOptionResponse, 0, len(poll.Options)),
		MyOptionIDs: append([]int64{}, summary.MyOptionIDs...),
	}
	for _, option := range poll.Options {
		if option.DeletedAt.Valid {
			continue
		}
		response.Options = append(response.Options, PollOptionResponse{
			ID: option.ID, Title: option.Title, VoteCount: summary.OptionCounts[option.ID],
		})
		response.TotalVotes += summary.OptionCounts[option.ID]
	}
	return response
}

func ToPollVoteResponse(pollID int64, votes []*entity.PollVote) PollVoteResponse {
	response := PollVoteResponse{PollID: pollID, OptionIDs: make([]int64, 0, len(votes)), Votes: make([]PollVoteSelectionResponse, 0, len(votes))}
	for _, vote := range votes {
		response.OptionIDs = append(response.OptionIDs, vote.OptionID)
		response.Votes = append(response.Votes, PollVoteSelectionResponse{OptionID: vote.OptionID, CreatedAt: vote.CreatedAt, UpdatedAt: vote.UpdatedAt})
	}
	return response
}

func ToPollVoterResponse(vote *entity.PollVote) PollVoterResponse {
	response := PollVoterResponse{MemberID: vote.MemberID, OptionID: vote.OptionID, CreatedAt: vote.CreatedAt, UpdatedAt: vote.UpdatedAt}
	if vote.Member != nil {
		response.User.ID = vote.Member.UserID
		if vote.Member.User != nil {
			response.User.Name = vote.Member.User.Name
			response.User.AvatarAttachmentID = vote.Member.User.AvatarAttachmentID
		}
	}
	return response
}

func ToPollStatusResponse(poll *entity.Poll) PollStatusResponse {
	return PollStatusResponse{ID: poll.ID, Status: poll.Status, ClosedAt: poll.ClosedAt}
}
