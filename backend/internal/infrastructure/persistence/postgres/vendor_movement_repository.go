package postgres

import (
	"context"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VendorMovementRepository struct {
	db *gorm.DB
}

func NewVendorMovementRepository(db *gorm.DB) *VendorMovementRepository {
	return &VendorMovementRepository{db: db}
}

func (r *VendorMovementRepository) Create(ctx context.Context, movement *model.VendorMovement) error {
	return r.db.WithContext(ctx).Create(movement).Error
}

func (r *VendorMovementRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]model.VendorMovement, int64, error) {
	var movements []model.VendorMovement
	var total int64
	base := r.db.WithContext(ctx).Model(&model.VendorMovement{}).Where("vendor_account_id = ?", accountID)
	base.Count(&total)
	if err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, 0, err
	}
	return movements, total, nil
}
