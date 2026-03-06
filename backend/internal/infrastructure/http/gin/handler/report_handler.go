package handler

import (
	"net/http"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/application/dto"
	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) parseDateRange(c *gin.Context) (time.Time, time.Time) {
	to := time.Now()
	from := to.AddDate(-1, 0, 0)

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = parsed
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			to = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}
	return from, to
}

// GetFinancialReport godoc
// @Summary Get financial report
// @Description Returns accrued/collected interest, IVA, capital collected and pending for a date range
// @Tags Reports
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD), defaults to 12 months ago"
// @Param to query string false "End date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} dto.FinancialReportResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/reports/financial [get]
func (h *ReportHandler) GetFinancialReport(c *gin.Context) {
	from, to := h.parseDateRange(c)

	report, err := h.reportService.GetFinancialReport(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.FinancialReportResponse{
		InterestAccrued:   report.InterestAccrued.StringFixed(2),
		InterestCollected: report.InterestCollected.StringFixed(2),
		IVAAccrued:        report.IVAAccrued.StringFixed(2),
		IVACollected:      report.IVACollected.StringFixed(2),
		CapitalCollected:  report.CapitalCollected.StringFixed(2),
		CapitalPending:    report.CapitalPending.StringFixed(2),
	})
}

// GetPortfolioPosition godoc
// @Summary Get portfolio position
// @Description Returns loan count, principal, and outstanding balance grouped by loan status
// @Tags Reports
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD), defaults to 12 months ago"
// @Param to query string false "End date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} dto.PortfolioPositionResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/reports/portfolio [get]
func (h *ReportHandler) GetPortfolioPosition(c *gin.Context) {
	from, to := h.parseDateRange(c)

	positions, err := h.reportService.GetPortfolioPosition(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	items := make([]dto.PortfolioPositionItemResponse, len(positions))
	for i, p := range positions {
		items[i] = dto.PortfolioPositionItemResponse{
			Status:           p.Status,
			LoanCount:        p.LoanCount,
			TotalPrincipal:   p.TotalPrincipal.StringFixed(2),
			TotalOutstanding: p.TotalOutstanding.StringFixed(2),
		}
	}

	c.JSON(http.StatusOK, dto.PortfolioPositionResponse{Items: items})
}
