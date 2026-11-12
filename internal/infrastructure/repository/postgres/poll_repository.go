package postgres

import (
	"context"
	"errors"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pollRepository struct {
	db database.Database
}

func NewPollRepository(db database.Database) domain_repo.PollRepository {
	return &pollRepository{db: db}
}

func (r *pollRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *pollRepository) CreateWithOptions(ctx context.Context, poll *entity.Poll) error {
	if poll == nil || poll.GroupID <= 0 || poll.CreatedByMemberID <= 0 {
		return exception.ErrBadRequest
	}
	if len(poll.Options) < 2 {
		return exception.ErrPollInvalidOptions
	}
	for _, option := range poll.Options {
		if option == nil {
			return exception.ErrPollInvalidOptions
		}
	}

	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		group, err := r.findGroup(tx, poll.GroupID)
		if err != nil {
			return err
		}
		if group.Status != enum.GroupStatusActive {
			return exception.ErrPollGroupInactive
		}
		if err := r.requireActiveMember(tx, poll.GroupID, poll.CreatedByMemberID); err != nil {
			return err
		}
		if err := tx.Omit(clause.Associations).Create(poll).Error; err != nil {
			return err
		}
		for _, option := range poll.Options {
			option.PollID = poll.ID
		}
		return tx.Omit(clause.Associations).Create(&poll.Options).Error
	})
	return pollPersistenceError(err)
}

func (r *pollRepository) FindByID(ctx context.Context, groupID, pollID int64) (*entity.Poll, error) {
	db := r.getDB(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "polls"}}).
		Preload("Options", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("CreatedByMember.User.AvatarAttachment")
	return r.findPoll(db, groupID, pollID)
}

func (r *pollRepository) FindByGroupID(ctx context.Context, groupID int64, status *enum.PollStatus, limit, offset int) ([]*entity.Poll, int64, error) {
	polls := make([]*entity.Poll, 0)
	var total int64
	db := r.getDB(ctx).Model(&entity.Poll{}).
		Joins("JOIN groups ON groups.id = polls.group_id AND groups.deleted_at IS NULL").
		Where("polls.group_id = ?", groupID)
	if status != nil {
		db = db.Where("polls.status = ?", *status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, pollPersistenceError(err)
	}
	query := db.Select("polls.*").
		Preload("Options", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Order("polls.created_at DESC").Order("polls.id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Find(&polls).Error; err != nil {
		return nil, 0, pollPersistenceError(err)
	}
	return polls, total, nil
}

func (r *pollRepository) FindGroupByID(ctx context.Context, groupID int64) (*entity.Group, error) {
	return r.findGroup(r.getDB(ctx), groupID)
}

func (r *pollRepository) FindActiveMemberByUserID(ctx context.Context, groupID, userID int64) (*entity.GroupMember, error) {
	var member entity.GroupMember
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "SHARE"}).
		Preload("User.AvatarAttachment").
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, enum.GroupMemberStatusActive).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrGroupMemberNotFound
	}
	if err != nil {
		return nil, pollPersistenceError(err)
	}
	return &member, nil
}

func (r *pollRepository) FindOptionByID(ctx context.Context, pollID, optionID int64) (*entity.PollOption, error) {
	var option entity.PollOption
	err := r.getDB(ctx).Select("poll_options.*").
		Joins("JOIN polls ON polls.id = poll_options.poll_id AND polls.deleted_at IS NULL").
		Joins("JOIN groups ON groups.id = polls.group_id AND groups.deleted_at IS NULL").
		Where("poll_options.poll_id = ? AND poll_options.id = ?", pollID, optionID).First(&option).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrPollOptionNotFound
	}
	if err != nil {
		return nil, pollPersistenceError(err)
	}
	return &option, nil
}

func (r *pollRepository) FindVoteByMemberID(ctx context.Context, pollID, memberID int64) (*entity.PollVote, error) {
	var vote entity.PollVote
	err := r.voteQuery(r.getDB(ctx)).Select("poll_votes.*").
		Where("poll_votes.poll_id = ? AND poll_votes.member_id = ?", pollID, memberID).First(&vote).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrPollVoteNotFound
	}
	if err != nil {
		return nil, pollPersistenceError(err)
	}
	return &vote, nil
}

func (r *pollRepository) UpsertVote(ctx context.Context, vote *entity.PollVote) error {
	if vote == nil || vote.PollID <= 0 || vote.MemberID <= 0 || vote.OptionID <= 0 {
		return exception.ErrBadRequest
	}
	return r.withPollWrite(ctx, 0, vote.PollID, func(tx *gorm.DB, poll *entity.Poll) error {
		if poll.Status != enum.PollStatusActive {
			return exception.ErrPollClosed
		}
		if err := r.requireActiveMember(tx, poll.GroupID, vote.MemberID); err != nil {
			return err
		}
		var option entity.PollOption
		err := tx.Clauses(clause.Locking{Strength: "SHARE"}).
			Where("poll_id = ? AND id = ?", vote.PollID, vote.OptionID).First(&option).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.ErrPollOptionNotFound
		}
		if err != nil {
			return err
		}
		row := entity.PollVote{PollID: vote.PollID, MemberID: vote.MemberID, OptionID: vote.OptionID}
		err = tx.Omit(clause.Associations).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "poll_id"}, {Name: "member_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"option_id", "updated_at", "deleted_at"}),
		}, clause.Returning{}).Create(&row).Error
		if err != nil {
			return err
		}
		vote.BaseEntity = row.BaseEntity
		return nil
	})
}

func (r *pollRepository) DeleteVote(ctx context.Context, pollID, memberID int64) error {
	return r.withPollWrite(ctx, 0, pollID, func(tx *gorm.DB, poll *entity.Poll) error {
		if poll.Status != enum.PollStatusActive {
			return exception.ErrPollClosed
		}
		if err := r.requireActiveMember(tx, poll.GroupID, memberID); err != nil {
			return err
		}
		result := tx.Where("poll_id = ? AND member_id = ?", pollID, memberID).Delete(&entity.PollVote{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return exception.ErrPollVoteNotFound
		}
		return nil
	})
}

func (r *pollRepository) CountVotesByOption(ctx context.Context, pollID int64) (map[int64]int64, error) {
	db := r.getDB(ctx)
	if _, err := r.findPoll(db, 0, pollID); err != nil {
		return nil, err
	}
	var counts []struct {
		OptionID  int64
		VoteCount int64
	}
	err := r.voteQuery(db).Select("poll_votes.option_id, COUNT(*) AS vote_count").
		Where("poll_votes.poll_id = ?", pollID).Group("poll_votes.option_id").Scan(&counts).Error
	if err != nil {
		return nil, pollPersistenceError(err)
	}
	result := make(map[int64]int64, len(counts))
	for _, count := range counts {
		result[count.OptionID] = count.VoteCount
	}
	return result, nil
}

func (r *pollRepository) FindVoteSummaries(ctx context.Context, groupID, memberID int64, pollIDs []int64) (map[int64]domain_repo.PollVoteSummary, error) {
	if groupID <= 0 || memberID <= 0 {
		return nil, exception.ErrBadRequest
	}
	result := make(map[int64]domain_repo.PollVoteSummary, len(pollIDs))
	if len(pollIDs) == 0 {
		return result, nil
	}
	var counts []struct {
		PollID    int64
		OptionID  int64
		VoteCount int64
		IsMyVote  bool
	}
	err := r.voteQuery(r.getDB(ctx)).
		Select("poll_votes.poll_id, poll_votes.option_id, COUNT(*) AS vote_count, BOOL_OR(poll_votes.member_id = ?) AS is_my_vote", memberID).
		Where("polls.group_id = ? AND poll_votes.poll_id IN ?", groupID, pollIDs).
		Group("poll_votes.poll_id, poll_votes.option_id").Scan(&counts).Error
	if err != nil {
		return nil, pollPersistenceError(err)
	}
	for _, count := range counts {
		summary := result[count.PollID]
		if summary.OptionCounts == nil {
			summary.OptionCounts = make(map[int64]int64)
		}
		summary.OptionCounts[count.OptionID] = count.VoteCount
		if count.IsMyVote {
			optionID := count.OptionID
			summary.MyOptionID = &optionID
		}
		result[count.PollID] = summary
	}
	return result, nil
}

func (r *pollRepository) FindVotesByPollID(ctx context.Context, pollID int64, limit, offset int) ([]*entity.PollVote, int64, error) {
	db := r.getDB(ctx)
	poll, err := r.findPoll(db, 0, pollID)
	if err != nil {
		return nil, 0, err
	}
	if poll.Visibility != enum.PollVisibilityPublic {
		return nil, 0, exception.ErrPollPrivate
	}
	votes := make([]*entity.PollVote, 0)
	var total int64
	query := r.voteQuery(db).Where("poll_votes.poll_id = ?", pollID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, pollPersistenceError(err)
	}
	query = query.Select("poll_votes.*").Preload("Member.User.AvatarAttachment").
		Order("poll_votes.created_at ASC").Order("poll_votes.id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Find(&votes).Error; err != nil {
		return nil, 0, pollPersistenceError(err)
	}
	return votes, total, nil
}

func (r *pollRepository) Close(ctx context.Context, poll *entity.Poll, expectedStatus enum.PollStatus) error {
	if poll == nil || poll.ID <= 0 || poll.GroupID <= 0 {
		return exception.ErrBadRequest
	}
	if expectedStatus != enum.PollStatusActive {
		return exception.ErrPollClosed
	}
	return r.withPollWrite(ctx, poll.GroupID, poll.ID, func(tx *gorm.DB, stored *entity.Poll) error {
		if stored.Status != expectedStatus {
			return exception.ErrPollClosed
		}
		updatedAt := tx.NowFunc()
		closedAt := updatedAt
		if poll.ClosedAt != nil {
			closedAt = *poll.ClosedAt
		}
		result := tx.Model(&entity.Poll{}).Where("id = ? AND group_id = ? AND status = ?", poll.ID, poll.GroupID, expectedStatus).
			Updates(map[string]any{"status": enum.PollStatusInactive, "closed_at": closedAt, "updated_at": updatedAt})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return exception.ErrPollClosed
		}
		poll.Status = enum.PollStatusInactive
		poll.ClosedAt = &closedAt
		poll.UpdatedAt = updatedAt
		return nil
	})
}

func (r *pollRepository) DeleteByID(ctx context.Context, groupID, pollID int64) error {
	if groupID <= 0 || pollID <= 0 {
		return exception.ErrBadRequest
	}
	return r.withPollWrite(ctx, groupID, pollID, func(tx *gorm.DB, poll *entity.Poll) error {
		if err := tx.Where("poll_id = ?", poll.ID).Delete(&entity.PollVote{}).Error; err != nil {
			return err
		}
		if err := tx.Where("poll_id = ?", poll.ID).Delete(&entity.PollOption{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND group_id = ?", poll.ID, groupID).Delete(&entity.Poll{}).Error
	})
}

func (r *pollRepository) findGroup(db *gorm.DB, groupID int64) (*entity.Group, error) {
	var group entity.Group
	err := db.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", groupID).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrGroupNotFound
	}
	if err != nil {
		return nil, pollPersistenceError(err)
	}
	return &group, nil
}

func (r *pollRepository) findPoll(db *gorm.DB, groupID, pollID int64) (*entity.Poll, error) {
	var poll entity.Poll
	query := db.Select("polls.*").Joins("JOIN groups ON groups.id = polls.group_id AND groups.deleted_at IS NULL").
		Where("polls.id = ?", pollID)
	if groupID > 0 {
		query = query.Where("polls.group_id = ?", groupID)
	}
	err := query.First(&poll).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrPollNotFound
	}
	if err != nil {
		return nil, pollPersistenceError(err)
	}
	return &poll, nil
}

func (r *pollRepository) requireActiveMember(db *gorm.DB, groupID, memberID int64) error {
	var member entity.GroupMember
	err := db.Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ? AND group_id = ? AND status = ?", memberID, groupID, enum.GroupMemberStatusActive).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrPollForbidden
	}
	return err
}

func (r *pollRepository) withPollWrite(ctx context.Context, groupID, pollID int64, fn func(*gorm.DB, *entity.Poll) error) error {
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		poll, err := r.findPoll(tx, groupID, pollID)
		if err != nil {
			return err
		}
		group, err := r.findGroup(tx, poll.GroupID)
		if err != nil {
			return err
		}
		if group.Status != enum.GroupStatusActive {
			return exception.ErrPollGroupInactive
		}
		poll, err = r.findPoll(tx.Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "polls"}}), groupID, pollID)
		if err != nil {
			return err
		}
		return fn(tx, poll)
	})
	return pollPersistenceError(err)
}

func (r *pollRepository) voteQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&entity.PollVote{}).
		Joins("JOIN polls ON polls.id = poll_votes.poll_id AND polls.deleted_at IS NULL").
		Joins("JOIN groups ON groups.id = polls.group_id AND groups.deleted_at IS NULL").
		Joins("JOIN poll_options ON poll_options.poll_id = poll_votes.poll_id AND poll_options.id = poll_votes.option_id AND poll_options.deleted_at IS NULL")
}

func pollPersistenceError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *exception.AppError
	if errors.As(err, &appErr) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "idx_poll_options_active_title" {
				return exception.ErrPollDuplicateOptions
			}
		case "23503":
			switch pgErr.ConstraintName {
			case "polls_group_id_fkey":
				return exception.ErrGroupNotFound
			case "polls_created_by_member_id_fkey", "poll_votes_member_id_fkey":
				return exception.ErrGroupMemberNotFound
			case "poll_options_poll_id_fkey", "poll_votes_poll_id_fkey":
				return exception.ErrPollNotFound
			case "fk_poll_votes_poll_option":
				return exception.ErrPollOptionNotFound
			}
		case "23514":
			if pgErr.ConstraintName == "chk_poll_options_title_not_empty" {
				return exception.ErrPollInvalidOptions
			}
			return exception.ErrBadRequest
		case "22001", "22P02", "23502":
			return exception.ErrBadRequest
		}
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrForeignKeyViolated) {
		return exception.ErrBadRequest
	}
	return exception.Wrap(exception.ErrInternal, err)
}
