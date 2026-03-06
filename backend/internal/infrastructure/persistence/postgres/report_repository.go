package postgres

import (
	"context"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) FinancialReport(ctx context.Context, from, to time.Time) (*port.FinancialReport, error) {
	report := &port.FinancialReport{}

	activeStatuses := []string{
		string(model.LoanActive),
		string(model.LoanCompleted),
		string(model.LoanDefaulted),
	}

	// Accrued interest and IVA: installments with due_date in range, from active/completed/defaulted loans
	type accruedResult struct {
		InterestAccrued *decimal.Decimal
		IVAAccrued      *decimal.Decimal
	}
	var accrued accruedResult
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select("COALESCE(SUM(interest_amount), 0) as interest_accrued, COALESCE(SUM(iva_amount), 0) as iva_accrued").
		Joins("JOIN loans ON loans.id = installments.loan_id AND loans.deleted_at IS NULL").
		Where("installments.due_date BETWEEN ? AND ? AND loans.status IN ?", from, to, activeStatuses).
		Scan(&accrued)
	if accrued.InterestAccrued != nil {
		report.InterestAccrued = *accrued.InterestAccrued
	}
	if accrued.IVAAccrued != nil {
		report.IVAAccrued = *accrued.IVAAccrued
	}

	// Collected amounts: proportional based on paid_amount / total_amount
	type collectedResult struct {
		InterestCollected *decimal.Decimal
		IVACollected      *decimal.Decimal
		CapitalCollected  *decimal.Decimal
	}
	var collected collectedResult
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select(`COALESCE(SUM(CASE WHEN total_amount > 0 THEN interest_amount * paid_amount / total_amount ELSE 0 END), 0) as interest_collected,
				COALESCE(SUM(CASE WHEN total_amount > 0 THEN iva_amount * paid_amount / total_amount ELSE 0 END), 0) as iva_collected,
				COALESCE(SUM(CASE WHEN total_amount > 0 THEN capital_amount * paid_amount / total_amount ELSE 0 END), 0) as capital_collected`).
		Joins("JOIN loans ON loans.id = installments.loan_id AND loans.deleted_at IS NULL").
		Where("installments.due_date BETWEEN ? AND ? AND loans.status IN ?", from, to, activeStatuses).
		Scan(&collected)
	if collected.InterestCollected != nil {
		report.InterestCollected = *collected.InterestCollected
	}
	if collected.IVACollected != nil {
		report.IVACollected = *collected.IVACollected
	}
	if collected.CapitalCollected != nil {
		report.CapitalCollected = *collected.CapitalCollected
	}

	// Capital pending: remaining capital for unpaid installments in range
	type pendingResult struct {
		CapitalPending *decimal.Decimal
	}
	var pending pendingResult
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select(`COALESCE(SUM(CASE WHEN total_amount > 0 THEN capital_amount * remaining_amount / total_amount ELSE 0 END), 0) as capital_pending`).
		Joins("JOIN loans ON loans.id = installments.loan_id AND loans.deleted_at IS NULL").
		Where("installments.due_date BETWEEN ? AND ? AND loans.status IN ? AND installments.status != ?",
			from, to, activeStatuses, model.InstallmentPaid).
		Scan(&pending)
	if pending.CapitalPending != nil {
		report.CapitalPending = *pending.CapitalPending
	}

	return report, nil
}

func (r *ReportRepository) PortfolioPosition(ctx context.Context, from, to time.Time) ([]port.PortfolioPosition, error) {
	type positionResult struct {
		Status           string
		LoanCount        int64
		TotalPrincipal   decimal.Decimal
		TotalOutstanding decimal.Decimal
	}
	var results []positionResult

	subQuery := r.db.WithContext(ctx).Model(&model.Installment{}).
		Select("loan_id, COALESCE(SUM(remaining_amount), 0) as outstanding").
		Group("loan_id")

	r.db.WithContext(ctx).Model(&model.Loan{}).
		Select("loans.status as status, COUNT(*) as loan_count, COALESCE(SUM(loans.principal), 0) as total_principal, COALESCE(SUM(inst.outstanding), 0) as total_outstanding").
		Joins("LEFT JOIN (?) as inst ON inst.loan_id = loans.id", subQuery).
		Where("loans.deleted_at IS NULL AND loans.created_at BETWEEN ? AND ?", from, to).
		Group("loans.status").
		Scan(&results)

	positions := make([]port.PortfolioPosition, len(results))
	for i, r := range results {
		positions[i] = port.PortfolioPosition{
			Status:           r.Status,
			LoanCount:        r.LoanCount,
			TotalPrincipal:   r.TotalPrincipal,
			TotalOutstanding: r.TotalOutstanding,
		}
	}
	return positions, nil
}
