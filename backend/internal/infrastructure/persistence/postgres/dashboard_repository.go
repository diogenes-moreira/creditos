package postgres

import (
	"context"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) PortfolioSummary(ctx context.Context) (*port.PortfolioSummary, error) {
	summary := &port.PortfolioSummary{}

	r.db.WithContext(ctx).Model(&model.Client{}).Count(&summary.TotalClients)
	r.db.WithContext(ctx).Model(&model.Loan{}).Where("status = ?", model.LoanActive).Count(&summary.ActiveLoans)
	r.db.WithContext(ctx).Model(&model.CreditLine{}).Where("status = ?", model.CreditLinePending).Count(&summary.PendingApprovals)

	var disbursed struct{ Total *decimal.Decimal }
	r.db.WithContext(ctx).Model(&model.Loan{}).
		Select("COALESCE(SUM(principal), 0) as total").
		Where("status IN ?", []string{string(model.LoanActive), string(model.LoanCompleted)}).
		Scan(&disbursed)
	if disbursed.Total != nil {
		summary.TotalDisbursed = *disbursed.Total
	}

	var outstanding struct{ Total *decimal.Decimal }
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select("COALESCE(SUM(remaining_amount), 0) as total").
		Where("status != ?", model.InstallmentPaid).
		Scan(&outstanding)
	if outstanding.Total != nil {
		summary.TotalOutstanding = *outstanding.Total
	}

	var collected struct{ Total *decimal.Decimal }
	r.db.WithContext(ctx).Model(&model.Payment{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Scan(&collected)
	if collected.Total != nil {
		summary.TotalCollected = *collected.Total
	}

	return summary, nil
}

func (r *DashboardRepository) DelinquencyRates(ctx context.Context) (*port.DelinquencyStats, error) {
	stats := &port.DelinquencyStats{}
	now := time.Now()

	var totalOutstanding struct{ Total *decimal.Decimal }
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select("COALESCE(SUM(remaining_amount), 0) as total").
		Where("status != ?", model.InstallmentPaid).
		Scan(&totalOutstanding)

	par30Date := now.AddDate(0, 0, -30)
	par60Date := now.AddDate(0, 0, -60)
	par90Date := now.AddDate(0, 0, -90)

	var par30 struct{ Total *decimal.Decimal }
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select("COALESCE(SUM(remaining_amount), 0) as total").
		Where("status != ? AND due_date < ?", model.InstallmentPaid, par30Date).
		Scan(&par30)
	if par30.Total != nil {
		stats.PAR30 = *par30.Total
	}

	var par60 struct{ Total *decimal.Decimal }
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select("COALESCE(SUM(remaining_amount), 0) as total").
		Where("status != ? AND due_date < ?", model.InstallmentPaid, par60Date).
		Scan(&par60)
	if par60.Total != nil {
		stats.PAR60 = *par60.Total
	}

	var par90 struct{ Total *decimal.Decimal }
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select("COALESCE(SUM(remaining_amount), 0) as total").
		Where("status != ? AND due_date < ?", model.InstallmentPaid, par90Date).
		Scan(&par90)
	if par90.Total != nil {
		stats.PAR90 = *par90.Total
	}

	var overdue struct {
		Total *decimal.Decimal
		Count int64
	}
	r.db.WithContext(ctx).Model(&model.Installment{}).
		Select("COALESCE(SUM(remaining_amount), 0) as total, COUNT(*) as count").
		Where("status != ? AND due_date < ?", model.InstallmentPaid, now).
		Scan(&overdue)
	if overdue.Total != nil {
		stats.TotalOverdue = *overdue.Total
	}
	stats.OverdueCount = overdue.Count

	if totalOutstanding.Total != nil && !totalOutstanding.Total.IsZero() {
		stats.DelinquencyRate = stats.TotalOverdue.Div(*totalOutstanding.Total).Mul(decimal.NewFromInt(100)).Round(2)
	}

	return stats, nil
}

func (r *DashboardRepository) DisbursementTrend(ctx context.Context, from, to time.Time) ([]port.TrendPoint, error) {
	type result struct {
		Date   time.Time
		Total  decimal.Decimal
		Count  int64
	}
	var results []result
	r.db.WithContext(ctx).Model(&model.Loan{}).
		Select("DATE(disbursed_at) as date, SUM(principal) as total, COUNT(*) as count").
		Where("disbursed_at BETWEEN ? AND ? AND status IN ?", from, to, []string{string(model.LoanActive), string(model.LoanCompleted)}).
		Group("DATE(disbursed_at)").
		Order("date").
		Scan(&results)

	points := make([]port.TrendPoint, len(results))
	for i, r := range results {
		points[i] = port.TrendPoint{Date: r.Date, Amount: r.Total, Count: r.Count}
	}
	return points, nil
}

func (r *DashboardRepository) CollectionTrend(ctx context.Context, from, to time.Time) ([]port.TrendPoint, error) {
	type result struct {
		Date  time.Time
		Total decimal.Decimal
		Count int64
	}
	var results []result
	r.db.WithContext(ctx).Model(&model.Payment{}).
		Select("DATE(created_at) as date, SUM(amount) as total, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", from, to).
		Group("DATE(created_at)").
		Order("date").
		Scan(&results)

	points := make([]port.TrendPoint, len(results))
	for i, r := range results {
		points[i] = port.TrendPoint{Date: r.Date, Amount: r.Total, Count: r.Count}
	}
	return points, nil
}
