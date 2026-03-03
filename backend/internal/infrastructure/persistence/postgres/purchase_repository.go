package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) Create(ctx context.Context, purchase *model.Purchase) error {
	return r.db.WithContext(ctx).Create(purchase).Error
}

func (r *PurchaseRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Purchase, error) {
	var purchase model.Purchase
	if err := r.db.WithContext(ctx).Preload("Vendor").Preload("Client").First(&purchase, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("purchase not found: %w", err)
	}
	return &purchase, nil
}

func (r *PurchaseRepository) FindByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]model.Purchase, int64, error) {
	var purchases []model.Purchase
	var total int64
	base := r.db.WithContext(ctx).Model(&model.Purchase{}).Where("vendor_id = ?", vendorID)
	base.Count(&total)
	if err := base.Preload("Client").Offset(offset).Limit(limit).Order("created_at DESC").Find(&purchases).Error; err != nil {
		return nil, 0, err
	}
	return purchases, total, nil
}

func (r *PurchaseRepository) FindByClientID(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]model.Purchase, int64, error) {
	var purchases []model.Purchase
	var total int64
	base := r.db.WithContext(ctx).Model(&model.Purchase{}).Where("client_id = ?", clientID)
	base.Count(&total)
	if err := base.Preload("Vendor").Offset(offset).Limit(limit).Order("created_at DESC").Find(&purchases).Error; err != nil {
		return nil, 0, err
	}
	return purchases, total, nil
}
