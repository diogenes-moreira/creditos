package postgres

import (
	"context"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WithdrawalRequestRepository struct {
	db *gorm.DB
}

func NewWithdrawalRequestRepository(db *gorm.DB) *WithdrawalRequestRepository {
	return &WithdrawalRequestRepository{db: db}
}

func (r *WithdrawalRequestRepository) Create(ctx context.Context, req *model.WithdrawalRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *WithdrawalRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.WithdrawalRequest, error) {
	var req model.WithdrawalRequest
	if err := r.db.WithContext(ctx).Preload("Vendor").First(&req, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *WithdrawalRequestRepository) FindByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]model.WithdrawalRequest, int64, error) {
	var requests []model.WithdrawalRequest
	var total int64
	base := r.db.WithContext(ctx).Model(&model.WithdrawalRequest{}).Where("vendor_id = ?", vendorID)
	base.Count(&total)
	if err := base.Offset(offset).Limit(limit).Order("requested_at DESC").Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

func (r *WithdrawalRequestRepository) FindPending(ctx context.Context, offset, limit int) ([]model.WithdrawalRequest, int64, error) {
	var requests []model.WithdrawalRequest
	var total int64
	base := r.db.WithContext(ctx).Model(&model.WithdrawalRequest{}).Where("status = ?", model.WithdrawalStatusPending)
	base.Count(&total)
	if err := base.Preload("Vendor").Offset(offset).Limit(limit).Order("requested_at ASC").Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

func (r *WithdrawalRequestRepository) Update(ctx context.Context, req *model.WithdrawalRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}
