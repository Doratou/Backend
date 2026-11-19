package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type PollService interface {
	Create(ctx context.Context, userID, groupID int64, req dto.CreatePollRequest) (*dto.PollResponse, error)
	List(ctx context.Context, userID, groupID int64, req dto.ListPollsRequest) (*dto.PollListResponse, error)
	Get(ctx context.Context, userID, groupID, pollID int64) (*dto.PollResponse, error)
	Vote(ctx context.Context, userID, groupID, pollID int64, req dto.VotePollRequest) (*dto.PollVoteResponse, error)
	WithdrawVote(ctx context.Context, userID, groupID, pollID int64) error
	ListVoters(ctx context.Context, userID, groupID, pollID int64, req dto.ListPollVotersRequest) (*dto.PollVoterListResponse, error)
	Close(ctx context.Context, userID, groupID, pollID int64) (*dto.PollStatusResponse, error)
	Delete(ctx context.Context, userID, groupID, pollID int64) error
}
