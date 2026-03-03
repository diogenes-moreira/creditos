package postgres

import (
	"context"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VendorPaymentRepository struct {
	db *gorm.DB
}

func NewVendorPaymentRepository(db *gorm.DB) *VendorPaymentRepository {
	return &VendorPaymentRepository{db: db}
}

func (r *VendorPaymentRepository) Create(ctx context.Context, payment *model.VendorPayment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *VendorPaymentRepository) FindByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]model.VendorPayment, int64, error) {
	var payments []model.VendorPayment
	var total int64
	base := r.db.WithContext(ctx).Model(&model.VendorPayment{}).Where("vendor_id = ?", vendorID)
	base.Count(&total)
	if err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}
