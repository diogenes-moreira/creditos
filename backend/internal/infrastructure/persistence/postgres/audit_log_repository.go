package postgres

import (
	"context"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditLogRepository) FindByEntity(ctx context.Context, entityType, entityID string, offset, limit int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	base := r.db.WithContext(ctx).Model(&model.AuditLog{}).Where("entity_type = ? AND entity_id = ?", entityType, entityID)
	base.Count(&total)
	if err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *AuditLogRepository) FindByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	base := r.db.WithContext(ctx).Model(&model.AuditLog{}).Where("user_id = ?", userID)
	base.Count(&total)
	if err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *AuditLogRepository) FindAll(ctx context.Context, offset, limit int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	r.db.WithContext(ctx).Model(&model.AuditLog{}).Count(&total)
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
