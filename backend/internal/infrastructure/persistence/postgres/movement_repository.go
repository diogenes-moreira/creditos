package postgres

import (
	"context"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MovementRepository struct {
	db *gorm.DB
}

func NewMovementRepository(db *gorm.DB) *MovementRepository {
	return &MovementRepository{db: db}
}

func (r *MovementRepository) Create(ctx context.Context, movement *model.Movement) error {
	return r.db.WithContext(ctx).Create(movement).Error
}

func (r *MovementRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]model.Movement, int64, error) {
	var movements []model.Movement
	var total int64
	base := r.db.WithContext(ctx).Model(&model.Movement{}).Where("account_id = ?", accountID)
	base.Count(&total)
	if err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, 0, err
	}
	return movements, total, nil
}
