package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *PaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).First(&payment, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}
	return &payment, nil
}

func (r *PaymentRepository) FindByLoanID(ctx context.Context, loanID uuid.UUID, offset, limit int) ([]model.Payment, int64, error) {
	var payments []model.Payment
	var total int64
	base := r.db.WithContext(ctx).Model(&model.Payment{}).Where("loan_id = ?", loanID)
	base.Count(&total)
	if err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}

func (r *PaymentRepository) FindByClientLoans(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]model.Payment, int64, error) {
	var payments []model.Payment
	var total int64
	subQuery := r.db.Model(&model.Loan{}).Select("id").Where("client_id = ?", clientID)
	base := r.db.WithContext(ctx).Model(&model.Payment{}).Where("loan_id IN (?)", subQuery)
	base.Count(&total)
	if err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}
