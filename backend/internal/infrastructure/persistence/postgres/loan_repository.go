package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) Create(ctx context.Context, loan *model.Loan) error {
	return r.db.WithContext(ctx).Create(loan).Error
}

func (r *LoanRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Loan, error) {
	var loan model.Loan
	if err := r.db.WithContext(ctx).First(&loan, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}
	return &loan, nil
}

func (r *LoanRepository) FindByIDWithInstallments(ctx context.Context, id uuid.UUID) (*model.Loan, error) {
	var loan model.Loan
	if err := r.db.WithContext(ctx).Preload("Installments", func(db *gorm.DB) *gorm.DB {
		return db.Order("number ASC")
	}).First(&loan, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}
	return &loan, nil
}

func (r *LoanRepository) FindByClientID(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]model.Loan, int64, error) {
	var loans []model.Loan
	var total int64
	base := r.db.WithContext(ctx).Model(&model.Loan{}).Where("client_id = ?", clientID)
	base.Count(&total)
	if err := base.Preload("Client").Preload("Installments", func(db *gorm.DB) *gorm.DB {
		return db.Order("number ASC")
	}).Offset(offset).Limit(limit).Order("created_at DESC").Find(&loans).Error; err != nil {
		return nil, 0, err
	}
	return loans, total, nil
}

func (r *LoanRepository) FindByStatus(ctx context.Context, status model.LoanStatus, offset, limit int) ([]model.Loan, int64, error) {
	var loans []model.Loan
	var total int64
	base := r.db.WithContext(ctx).Model(&model.Loan{}).Where("status = ?", status)
	base.Count(&total)
	if err := base.Preload("Client").Preload("Installments").Offset(offset).Limit(limit).Order("created_at DESC").Find(&loans).Error; err != nil {
		return nil, 0, err
	}
	return loans, total, nil
}

func (r *LoanRepository) FindActive(ctx context.Context, offset, limit int) ([]model.Loan, int64, error) {
	return r.FindByStatus(ctx, model.LoanActive, offset, limit)
}

func (r *LoanRepository) Update(ctx context.Context, loan *model.Loan) error {
	return r.db.WithContext(ctx).Save(loan).Error
}
