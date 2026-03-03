package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VendorAccountRepository struct {
	db *gorm.DB
}

func NewVendorAccountRepository(db *gorm.DB) *VendorAccountRepository {
	return &VendorAccountRepository{db: db}
}

func (r *VendorAccountRepository) Create(ctx context.Context, account *model.VendorAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *VendorAccountRepository) FindByVendorID(ctx context.Context, vendorID uuid.UUID) (*model.VendorAccount, error) {
	var account model.VendorAccount
	if err := r.db.WithContext(ctx).First(&account, "vendor_id = ?", vendorID).Error; err != nil {
		return nil, fmt.Errorf("vendor account not found: %w", err)
	}
	return &account, nil
}

func (r *VendorAccountRepository) Update(ctx context.Context, account *model.VendorAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}
