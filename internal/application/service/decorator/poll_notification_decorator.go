package decorator

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/broker/payload"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type pollNotificationDecorator struct {
	next      adapter.PollService
	notifSvc  adapter.NotificationService
	groupRepo domain_db.GroupRepository
	userRepo  domain_db.UserRepository
}

// NewPollNotificationDecorator wraps an adapter.PollService to dispatch group notifications
// when polls are created and closed.
func NewPollNotificationDecorator(
	next adapter.PollService,
	notifSvc adapter.NotificationService,
	groupRepo domain_db.GroupRepository,
	userRepo domain_db.UserRepository,
) adapter.PollService {
	return &pollNotificationDecorator{
		next:      next,
		notifSvc:  notifSvc,
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (d *pollNotificationDecorator) Create(ctx context.Context, userID, groupID int64, req dto.CreatePollRequest) (*dto.PollResponse, error) {
	res, err := d.next.Create(ctx, userID, groupID, req)
	if err != nil {
		return nil, err
	}

	groupName := resolveGroupTitle(ctx, d.groupRepo, groupID)
	creatorName, _ := resolveUserProfile(ctx, d.userRepo, userID)
	options := make([]string, len(res.Options))
	for i, opt := range res.Options {
		options[i] = opt.Title
	}

	dispatchGroupNotification(ctx, d.notifSvc, groupID, &userID, payload.WSEventPollCreated, payload.PollCreatedData{
		PollID:      res.ID,
		GroupID:     groupID,
		GroupName:   groupName,
		CreatorID:   strconv.FormatInt(userID, 10),
		CreatorName: creatorName,
		Question:    res.Question,
		Options:     options,
	})

	return res, nil
}

func (d *pollNotificationDecorator) Close(ctx context.Context, userID, groupID, pollID int64) (*dto.PollStatusResponse, error) {
	res, err := d.next.Close(ctx, userID, groupID, pollID)
	if err != nil {
		return nil, err
	}

	if d.notifSvc != nil {
		pollDetail, gErr := d.next.Get(ctx, userID, groupID, pollID)
		if gErr == nil && pollDetail != nil {
			groupName := resolveGroupTitle(ctx, d.groupRepo, groupID)
			totalVotes := int(pollDetail.TotalVotes)
			results := make([]payload.PollOptionResult, 0, len(pollDetail.Options))
			var maxVotes int64

			for _, opt := range pollDetail.Options {
				var pct float64
				if pollDetail.TotalVotes > 0 {
					pct = math.Round(float64(opt.VoteCount)*10000/float64(pollDetail.TotalVotes)) / 100
				}
				if opt.VoteCount > maxVotes {
					maxVotes = opt.VoteCount
				}
				results = append(results, payload.PollOptionResult{
					OptionID:   opt.ID,
					Text:       opt.Title,
					VoteCount:  int(opt.VoteCount),
					Percentage: pct,
				})
			}

			winners := make([]payload.PollOptionResult, 0)
			if maxVotes > 0 {
				for _, r := range results {
					if int64(r.VoteCount) == maxVotes {
						winners = append(winners, r)
					}
				}
			}

			closedAt := time.Now().UTC()
			if res.ClosedAt != nil {
				closedAt = *res.ClosedAt
			}

			dispatchGroupNotification(ctx, d.notifSvc, groupID, &userID, payload.WSEventPollClosed, payload.PollClosedData{
				PollID:         pollID,
				GroupID:        groupID,
				GroupName:      groupName,
				Question:       pollDetail.Question,
				TotalVotes:     totalVotes,
				Results:        results,
				WinningOptions: winners,
				ClosedAt:       closedAt,
			})
		}
	}

	return res, nil
}

func (d *pollNotificationDecorator) List(ctx context.Context, userID, groupID int64, req dto.ListPollsRequest) (*dto.PollListResponse, error) {
	return d.next.List(ctx, userID, groupID, req)
}

func (d *pollNotificationDecorator) Get(ctx context.Context, userID, groupID, pollID int64) (*dto.PollResponse, error) {
	return d.next.Get(ctx, userID, groupID, pollID)
}

func (d *pollNotificationDecorator) Vote(ctx context.Context, userID, groupID, pollID int64, req dto.VotePollRequest) (*dto.PollVoteResponse, error) {
	return d.next.Vote(ctx, userID, groupID, pollID, req)
}

func (d *pollNotificationDecorator) WithdrawVote(ctx context.Context, userID, groupID, pollID int64) error {
	return d.next.WithdrawVote(ctx, userID, groupID, pollID)
}

func (d *pollNotificationDecorator) ListVoters(ctx context.Context, userID, groupID, pollID int64, req dto.ListPollVotersRequest) (*dto.PollVoterListResponse, error) {
	return d.next.ListVoters(ctx, userID, groupID, pollID, req)
}

func (d *pollNotificationDecorator) Delete(ctx context.Context, userID, groupID, pollID int64) error {
	return d.next.Delete(ctx, userID, groupID, pollID)
}
