package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InstallmentRepository struct {
	db *gorm.DB
}

func NewInstallmentRepository(db *gorm.DB) *InstallmentRepository {
	return &InstallmentRepository{db: db}
}

func (r *InstallmentRepository) CreateBatch(ctx context.Context, installments []model.Installment) error {
	return r.db.WithContext(ctx).Create(&installments).Error
}

func (r *InstallmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Installment, error) {
	var inst model.Installment
	if err := r.db.WithContext(ctx).First(&inst, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("installment not found: %w", err)
	}
	return &inst, nil
}

func (r *InstallmentRepository) FindByLoanID(ctx context.Context, loanID uuid.UUID) ([]model.Installment, error) {
	var installments []model.Installment
	if err := r.db.WithContext(ctx).Where("loan_id = ?", loanID).Order("number ASC").Find(&installments).Error; err != nil {
		return nil, err
	}
	return installments, nil
}

func (r *InstallmentRepository) FindUnpaidByLoanID(ctx context.Context, loanID uuid.UUID) ([]model.Installment, error) {
	var installments []model.Installment
	if err := r.db.WithContext(ctx).Where("loan_id = ? AND status != ?", loanID, model.InstallmentPaid).Order("number ASC").Find(&installments).Error; err != nil {
		return nil, err
	}
	return installments, nil
}

func (r *InstallmentRepository) FindOverdue(ctx context.Context) ([]model.Installment, error) {
	var installments []model.Installment
	if err := r.db.WithContext(ctx).Where("status IN ? AND due_date < NOW()", []string{string(model.InstallmentPending), string(model.InstallmentPartial), string(model.InstallmentOverdue)}).Find(&installments).Error; err != nil {
		return nil, err
	}
	return installments, nil
}

func (r *InstallmentRepository) Update(ctx context.Context, installment *model.Installment) error {
	return r.db.WithContext(ctx).Save(installment).Error
}

func (r *InstallmentRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&model.Installment{}, "id IN ?", ids).Error
}
