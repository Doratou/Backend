package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/finance"
	domainDB "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type financialReportRepository struct{ db database.Database }

func NewFinancialReportRepository(db database.Database) domainDB.FinancialReportRepository {
	return &financialReportRepository{db: db}
}

func (r *financialReportRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *financialReportRepository) FindGroup(ctx context.Context, groupID int64) (*entity.Group, error) {
	var group entity.Group
	err := r.getDB(ctx).Select("id", "status").First(&group, groupID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.ErrGroupNotFound
	}
	return &group, financialReportPersistenceError(err)
}

func (r *financialReportRepository) FindViewerMemberID(ctx context.Context, groupID, userID int64) (int64, error) {
	var member entity.GroupMember
	err := r.getDB(ctx).Select("id").Where("group_id = ? AND user_id = ? AND status = ? AND left_at IS NULL",
		groupID, userID, enum.GroupMemberStatusActive).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, exception.ErrFinancialReportForbidden
	}
	return member.ID, financialReportPersistenceError(err)
}

func (r *financialReportRepository) ListMembers(ctx context.Context, groupID int64, ids []int64) ([]entity.GroupMember, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var members []entity.GroupMember
	err := r.getDB(ctx).Where("group_id = ? AND id IN ?", groupID, ids).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped().Select("id", "name", "avatar_attachment_id")
		}).Order("id ASC").Find(&members).Error
	return members, financialReportPersistenceError(err)
}

func (r *financialReportRepository) ListExpenseShares(ctx context.Context, groupID, fromMemberID, toMemberID int64, limit, offset int) ([]domainDB.ExpenseReportShare, int64, error) {
	query := r.getDB(ctx).Table("expense_shares sh").
		Joins("JOIN expenses e ON e.id = sh.expense_id AND e.group_id = sh.group_id").
		Joins("JOIN group_members payer ON payer.id = e.created_by_member_id AND payer.group_id = sh.group_id").
		Where("sh.group_id = ? AND e.deleted_at IS NULL AND sh.deleted_at IS NULL", groupID).
		Where("sh.member_id <> e.created_by_member_id AND sh.share_amount_toman > 0")
	if fromMemberID != 0 {
		query = query.Where("sh.member_id = ?", fromMemberID)
	}
	if toMemberID != 0 {
		query = query.Where("e.created_by_member_id = ?", toMemberID)
	}
	query = query.Session(&gorm.Session{})
	var total int64
	if limit > 0 {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, financialReportPersistenceError(err)
		}
		query = query.Limit(limit).Offset(offset)
	}
	var shares []domainDB.ExpenseReportShare
	err := query.Select(`sh.id AS share_id, sh.expense_id, e.title AS expense_title, e.occurred_at,
		sh.member_id AS from_member_id, e.created_by_member_id AS to_member_id, sh.share_amount_toman AS amount_toman`).
		Order("e.occurred_at DESC NULLS LAST, e.id DESC, sh.id ASC").Scan(&shares).Error
	return shares, total, financialReportPersistenceError(err)
}

func (r *financialReportRepository) LoadBalanceData(ctx context.Context, groupID int64) (finance.BalanceData, error) {
	var data finance.BalanceData
	db := r.getDB(ctx)
	err := db.Model(&entity.GroupMember{}).Where("group_id = ? AND status = ? AND left_at IS NULL",
		groupID, enum.GroupMemberStatusActive).Pluck("id", &data.ActiveMembers).Error
	if err != nil {
		return data, financialReportPersistenceError(err)
	}

	err = db.Raw(`
		WITH active_members AS (
			SELECT id FROM group_members WHERE group_id = @group_id AND status = 'ACTIVE' AND left_at IS NULL
		), entries AS (
			SELECT sh.member_id AS from_id, e.created_by_member_id AS to_id, sh.share_amount_toman AS amount
			FROM expense_shares sh JOIN expenses e ON e.id = sh.expense_id AND e.group_id = sh.group_id
			WHERE sh.group_id = @group_id AND e.deleted_at IS NULL AND sh.deleted_at IS NULL
			  AND sh.share_amount_toman > 0 AND sh.member_id <> e.created_by_member_id
			UNION ALL
			SELECT leg.to_member_id, leg.from_member_id, leg.amount_toman
			FROM settlements p
			CROSS JOIN LATERAL jsonb_to_recordset(p.balance_legs)
				AS leg(from_member_id bigint, to_member_id bigint, amount_toman bigint)
			WHERE p.group_id = @group_id AND p.status = 'SUCCESS' AND p.deleted_at IS NULL
		), pairs AS (
			SELECT LEAST(from_id, to_id) AS low, GREATEST(from_id, to_id) AS high,
				SUM(CASE WHEN from_id < to_id THEN amount ELSE -amount END) AS net
			FROM entries
			JOIN active_members sender ON sender.id = entries.from_id
			JOIN active_members recipient ON recipient.id = entries.to_id
			GROUP BY 1, 2
		)
		SELECT CASE WHEN net > 0 THEN low ELSE high END AS from_member_id,
		       CASE WHEN net > 0 THEN high ELSE low END AS to_member_id,
		       ABS(net)::bigint AS amount_toman
		FROM pairs WHERE net <> 0
	`, sql.Named("group_id", groupID)).Scan(&data.OpenPairs).Error
	if err != nil {
		return data, financialReportPersistenceError(err)
	}

	var opening []byte
	err = db.Raw("SELECT state FROM financial_report_opening_states WHERE group_id = ?", groupID).Row().Scan(&opening)
	if errors.Is(err, sql.ErrNoRows) {
		return data, nil
	}
	if err != nil {
		return data, financialReportPersistenceError(err)
	}
	if err := json.Unmarshal(opening, &data.Opening); err != nil {
		return data, financialReportPersistenceError(err)
	}
	err = db.Model(&entity.Settlement{}).Select("from_member_id", "to_member_id", "amount_toman", "confirmed_at").
		Where("group_id = ? AND status = ? AND balance_legs IS NULL", groupID, enum.SettlementStatusSuccess).
		Order("confirmed_at ASC, id ASC").Scan(&data.Legacy).Error
	return data, financialReportPersistenceError(err)
}

func financialReportPersistenceError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "22003" {
		return exception.ErrFinancialAmountOutOfRange
	}
	return exception.Wrap(exception.ErrInternal, err)
}
