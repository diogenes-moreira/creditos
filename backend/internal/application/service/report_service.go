package service

import (
	"context"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
)

type ReportService struct {
	reportRepo port.ReportRepository
}

func NewReportService(reportRepo port.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

func (s *ReportService) GetFinancialReport(ctx context.Context, from, to time.Time) (*port.FinancialReport, error) {
	return s.reportRepo.FinancialReport(ctx, from, to)
}

func (s *ReportService) GetPortfolioPosition(ctx context.Context, from, to time.Time) ([]port.PortfolioPosition, error) {
	return s.reportRepo.PortfolioPosition(ctx, from, to)
}
