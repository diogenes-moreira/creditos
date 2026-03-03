package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VendorRepository struct {
	db *gorm.DB
}

func NewVendorRepository(db *gorm.DB) *VendorRepository {
	return &VendorRepository{db: db}
}

func (r *VendorRepository) Create(ctx context.Context, vendor *model.Vendor) error {
	return r.db.WithContext(ctx).Create(vendor).Error
}

func (r *VendorRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Vendor, error) {
	var vendor model.Vendor
	if err := r.db.WithContext(ctx).Preload("User").First(&vendor, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}
	return &vendor, nil
}

func (r *VendorRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Vendor, error) {
	var vendor model.Vendor
	if err := r.db.WithContext(ctx).Preload("User").First(&vendor, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}
	return &vendor, nil
}

func (r *VendorRepository) FindByCUIT(ctx context.Context, cuit string) (*model.Vendor, error) {
	var vendor model.Vendor
	if err := r.db.WithContext(ctx).First(&vendor, "cuit = ?", cuit).Error; err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}
	return &vendor, nil
}

func (r *VendorRepository) FindAll(ctx context.Context, offset, limit int) ([]model.Vendor, int64, error) {
	var vendors []model.Vendor
	var total int64
	r.db.WithContext(ctx).Model(&model.Vendor{}).Count(&total)
	if err := r.db.WithContext(ctx).Preload("User").Offset(offset).Limit(limit).Order("created_at DESC").Find(&vendors).Error; err != nil {
		return nil, 0, err
	}
	return vendors, total, nil
}

func (r *VendorRepository) Search(ctx context.Context, query string, offset, limit int) ([]model.Vendor, int64, error) {
	var vendors []model.Vendor
	var total int64
	q := "%" + query + "%"
	base := r.db.WithContext(ctx).Model(&model.Vendor{}).
		Where("business_name ILIKE ? OR cuit ILIKE ?", q, q)
	base.Count(&total)
	if err := base.Preload("User").Offset(offset).Limit(limit).Order("created_at DESC").Find(&vendors).Error; err != nil {
		return nil, 0, err
	}
	return vendors, total, nil
}

func (r *VendorRepository) Update(ctx context.Context, vendor *model.Vendor) error {
	return r.db.WithContext(ctx).Save(vendor).Error
}
