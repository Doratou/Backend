package DB

import (
	"context"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
)

type PollVoteSummary struct {
	OptionCounts map[int64]int64
	MyOptionID   *int64
}

type PollRepository interface {
	CreateWithOptions(ctx context.Context, poll *entity.Poll) error
	FindByID(ctx context.Context, groupID, pollID int64) (*entity.Poll, error)
	FindByGroupID(ctx context.Context, groupID int64, status *enum.PollStatus, limit, offset int) ([]*entity.Poll, int64, error)
	FindGroupByID(ctx context.Context, groupID int64) (*entity.Group, error)
	FindActiveMemberByUserID(ctx context.Context, groupID, userID int64) (*entity.GroupMember, error)
	FindOptionByID(ctx context.Context, pollID, optionID int64) (*entity.PollOption, error)
	FindVoteByMemberID(ctx context.Context, pollID, memberID int64) (*entity.PollVote, error)
	UpsertVote(ctx context.Context, vote *entity.PollVote) error
	DeleteVote(ctx context.Context, pollID, memberID int64) error
	CountVotesByOption(ctx context.Context, pollID int64) (map[int64]int64, error)
	FindVoteSummaries(ctx context.Context, groupID, memberID int64, pollIDs []int64) (map[int64]PollVoteSummary, error)
	FindVotesByPollID(ctx context.Context, pollID int64, limit, offset int) ([]*entity.PollVote, int64, error)
	Close(ctx context.Context, poll *entity.Poll, expectedStatus enum.PollStatus) error
	DeleteByID(ctx context.Context, groupID, pollID int64) error
}
