package postgres

import (
	"context"
	"errors"

	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type attachmentRepository struct {
	db database.Database
}

func NewAttachmentRepository(db database.Database) domain_repo.AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) getDB(ctx context.Context) *gorm.DB {
	return database.ExtractTrxOrDB(ctx, r.db).GetGormDB().WithContext(ctx)
}

func (r *attachmentRepository) Create(ctx context.Context, attachment *entity.Attachment) error {
	db := r.getDB(ctx)
	if err := db.Create(attachment).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *attachmentRepository) CreateBatch(ctx context.Context, attachments []*entity.Attachment) error {
	if len(attachments) == 0 {
		return nil
	}
	db := r.getDB(ctx)
	if err := db.Create(&attachments).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *attachmentRepository) FindByID(ctx context.Context, id int64) (*entity.Attachment, error) {
	var attachment entity.Attachment
	db := r.getDB(ctx)

	err := db.Where("id = ?", id).First(&attachment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrAttachmentNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &attachment, nil
}

func (r *attachmentRepository) FindByIDWithRelations(ctx context.Context, id int64) (*entity.Attachment, error) {
	var attachment entity.Attachment
	db := r.getDB(ctx)

	err := db.Preload("UploadedByUser.AvatarAttachment").
		Preload("Group").
		Preload("Expense").
		Where("id = ?", id).
		First(&attachment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrAttachmentNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &attachment, nil
}

func (r *attachmentRepository) FindByIDs(ctx context.Context, ids []int64) ([]*entity.Attachment, error) {
	if len(ids) == 0 {
		return []*entity.Attachment{}, nil
	}
	var attachments []*entity.Attachment
	db := r.getDB(ctx)

	if err := db.Where("id IN (?)", ids).Find(&attachments).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return attachments, nil
}

func (r *attachmentRepository) FindByObjectKey(ctx context.Context, objectKey string) (*entity.Attachment, error) {
	var attachment entity.Attachment
	db := r.getDB(ctx)

	err := db.Where("object_key = ?", objectKey).First(&attachment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrAttachmentNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &attachment, nil
}

func (r *attachmentRepository) FindByHash(ctx context.Context, hash string) (*entity.Attachment, error) {
	var attachment entity.Attachment
	db := r.getDB(ctx)

	err := db.Where("file_hash = ?", hash).First(&attachment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrAttachmentNotFound
		}
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return &attachment, nil
}

func (r *attachmentRepository) FindByUploadedByUser(ctx context.Context, userID int64, limit, offset int) ([]*entity.Attachment, error) {
	var attachments []*entity.Attachment
	db := r.getDB(ctx).Where("uploaded_by_user_id = ?", userID).Order("id DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	if err := db.Find(&attachments).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return attachments, nil
}

func (r *attachmentRepository) FindByGroup(ctx context.Context, groupID int64, limit, offset int) ([]*entity.Attachment, error) {
	var attachments []*entity.Attachment
	db := r.getDB(ctx).Where("group_id = ?", groupID).Order("id DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	if err := db.Find(&attachments).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return attachments, nil
}

func (r *attachmentRepository) FindByGroupAndType(ctx context.Context, groupID int64, attachmentType enum.AttachmentType, limit, offset int) ([]*entity.Attachment, error) {
	var attachments []*entity.Attachment
	db := r.getDB(ctx).Where("group_id = ? AND type = ?", groupID, attachmentType).Order("id DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	if err := db.Find(&attachments).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return attachments, nil
}

func (r *attachmentRepository) FindByExpense(ctx context.Context, expenseID int64) ([]*entity.Attachment, error) {
	var attachments []*entity.Attachment
	db := r.getDB(ctx).Where("expense_id = ?", expenseID).Order("id DESC")

	if err := db.Find(&attachments).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return attachments, nil
}

func (r *attachmentRepository) FindByType(ctx context.Context, attachmentType enum.AttachmentType, limit, offset int) ([]*entity.Attachment, error) {
	var attachments []*entity.Attachment
	db := r.getDB(ctx).Where("type = ?", attachmentType).Order("id DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	if err := db.Find(&attachments).Error; err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	return attachments, nil
}

func (r *attachmentRepository) List(ctx context.Context, filter domain_repo.AttachmentFilter, pagination domain_repo.Pagination) ([]*entity.Attachment, int64, error) {
	db := r.getDB(ctx).Model(&entity.Attachment{})

	if filter.PreloadRelations {
		db = db.Preload("UploadedByUser.AvatarAttachment").Preload("Group").Preload("Expense")
	}

	if filter.UploadedByUserID != nil {
		db = db.Where("uploaded_by_user_id = ?", *filter.UploadedByUserID)
	}
	if filter.GroupID != nil {
		db = db.Where("group_id = ?", *filter.GroupID)
	}
	if filter.ExpenseID != nil {
		db = db.Where("expense_id = ?", *filter.ExpenseID)
	}
	if filter.Type != nil {
		db = db.Where("type = ?", *filter.Type)
	}
	if filter.IsPublic != nil {
		db = db.Where("is_public = ?", *filter.IsPublic)
	}
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		db = db.Where("original_filename ILIKE ?", searchTerm)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	pageSize := pagination.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := pagination.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var attachments []*entity.Attachment
	if err := db.Order("id DESC").Limit(pageSize).Offset(offset).Find(&attachments).Error; err != nil {
		return nil, 0, exception.Wrap(exception.ErrInternal, err)
	}

	return attachments, total, nil
}

func (r *attachmentRepository) Update(ctx context.Context, attachment *entity.Attachment) error {
	db := r.getDB(ctx)
	if err := db.Save(attachment).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *attachmentRepository) UpdateBatch(ctx context.Context, ids []int64, updates map[string]any) error {
	if len(ids) == 0 || len(updates) == 0 {
		return nil
	}
	db := r.getDB(ctx)
	if err := db.Model(&entity.Attachment{}).Where("id IN (?)", ids).Updates(updates).Error; err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return nil
}

func (r *attachmentRepository) DeleteByID(ctx context.Context, id int64) error {
	db := r.getDB(ctx)
	res := db.Where("id = ?", id).Delete(&entity.Attachment{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrAttachmentNotFound
	}
	return nil
}

func (r *attachmentRepository) DeleteByIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.getDB(ctx)
	res := db.Where("id IN (?)", ids).Delete(&entity.Attachment{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	return nil
}

func (r *attachmentRepository) HardDeleteByID(ctx context.Context, id int64) error {
	db := r.getDB(ctx)
	res := db.Unscoped().Where("id = ?", id).Delete(&entity.Attachment{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	if res.RowsAffected == 0 {
		return exception.ErrAttachmentNotFound
	}
	return nil
}

func (r *attachmentRepository) HardDeleteByIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.getDB(ctx)
	res := db.Unscoped().Where("id IN (?)", ids).Delete(&entity.Attachment{})
	if res.Error != nil {
		return exception.Wrap(exception.ErrInternal, res.Error)
	}
	return nil
}

func (r *attachmentRepository) ExistsByID(ctx context.Context, id int64) (bool, error) {
	var count int64
	db := r.getDB(ctx)
	if err := db.Model(&entity.Attachment{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, exception.Wrap(exception.ErrInternal, err)
	}
	return count > 0, nil
}

func (r *attachmentRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	db := r.getDB(ctx)
	if err := db.Model(&entity.Attachment{}).Count(&count).Error; err != nil {
		return 0, exception.Wrap(exception.ErrInternal, err)
	}
	return count, nil
}
