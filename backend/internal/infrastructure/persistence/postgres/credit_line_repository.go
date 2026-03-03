package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreditLineRepository struct {
	db *gorm.DB
}

func NewCreditLineRepository(db *gorm.DB) *CreditLineRepository {
	return &CreditLineRepository{db: db}
}

func (r *CreditLineRepository) Create(ctx context.Context, cl *model.CreditLine) error {
	return r.db.WithContext(ctx).Create(cl).Error
}

func (r *CreditLineRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.CreditLine, error) {
	var cl model.CreditLine
	if err := r.db.WithContext(ctx).Preload("Client").First(&cl, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("credit line not found: %w", err)
	}
	return &cl, nil
}

func (r *CreditLineRepository) FindByClientID(ctx context.Context, clientID uuid.UUID) ([]model.CreditLine, error) {
	var cls []model.CreditLine
	if err := r.db.WithContext(ctx).Where("client_id = ?", clientID).Order("created_at DESC").Find(&cls).Error; err != nil {
		return nil, err
	}
	return cls, nil
}

func (r *CreditLineRepository) FindByStatus(ctx context.Context, status model.CreditLineStatus, offset, limit int) ([]model.CreditLine, int64, error) {
	var cls []model.CreditLine
	var total int64
	base := r.db.WithContext(ctx).Model(&model.CreditLine{}).Where("status = ?", status)
	base.Count(&total)
	if err := base.Preload("Client").Offset(offset).Limit(limit).Order("created_at DESC").Find(&cls).Error; err != nil {
		return nil, 0, err
	}
	return cls, total, nil
}

func (r *CreditLineRepository) Update(ctx context.Context, cl *model.CreditLine) error {
	return r.db.WithContext(ctx).Save(cl).Error
}
