package handler

import (
	"net/http"

	"github.com/diogenes-moreira/creditos/backend/internal/application/dto"
	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LoanHandler struct {
	creditService  *service.CreditService
	paymentService *service.PaymentService
	clientRepo     port.ClientRepository
	defaultIVARate decimal.Decimal
}

func NewLoanHandler(creditService *service.CreditService, paymentService *service.PaymentService, clientRepo port.ClientRepository, defaultIVARate float64) *LoanHandler {
	return &LoanHandler{
		creditService:  creditService,
		paymentService: paymentService,
		clientRepo:     clientRepo,
		defaultIVARate: decimal.NewFromFloat(defaultIVARate),
	}
}

// Simulate godoc
// @Summary Simulate a loan
// @Description Calculates an amortization schedule for a given loan amount, installments, and type (French or German)
// @Tags Loans
// @Accept json
// @Produce json
// @Param request body dto.SimulateLoanRequest true "Simulation parameters"
// @Success 200 {object} dto.SimulationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /loans/simulate [post]
func (h *LoanHandler) Simulate(c *gin.Context) {
	var req dto.SimulateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}

	// Use a placeholder rate for simulation
	rate := decimal.NewFromFloat(0.45) // 45% annual as default
	schedule := h.creditService.SimulateLoan(amount, rate, req.NumInstallments, model.AmortizationType(req.AmortizationType), h.defaultIVARate)

	installments := make([]dto.InstallmentResponse, len(schedule.Installments))
	for i, inst := range schedule.Installments {
		installments[i] = dto.InstallmentResponse{
			Number:         inst.Number,
			DueDate:        inst.DueDate.Format("2006-01-02"),
			CapitalAmount:  inst.Capital.StringFixed(2),
			InterestAmount: inst.Interest.StringFixed(2),
			IVAAmount:      inst.IVA.StringFixed(2),
			TotalAmount:    inst.Total.StringFixed(2),
		}
	}

	c.JSON(http.StatusOK, dto.SimulationResponse{
		Principal:     amount.StringFixed(2),
		InterestRate:  rate.StringFixed(4),
		TotalInterest: schedule.TotalInterest.StringFixed(2),
		TotalPayment:  schedule.TotalPayment.StringFixed(2),
		Installments:  installments,
	})
}

// RequestLoan godoc
// @Summary Request a new loan
// @Description Submits a loan request against an approved credit line for the authenticated client
// @Tags Loans
// @Accept json
// @Produce json
// @Param request body dto.RequestLoanRequest true "Loan request data"
// @Success 201 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/loans [post]
func (h *LoanHandler) RequestLoan(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	client, err := h.clientRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "client not found"})
		return
	}

	var req dto.RequestLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	clID, _ := uuid.Parse(req.CreditLineID)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}

	loan, err := h.creditService.RequestLoan(c.Request.Context(), client.ID, clID, amount, req.NumInstallments, model.AmortizationType(req.AmortizationType))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToLoanResponse(loan))
}

// GetMyLoans godoc
// @Summary List my loans
// @Description Returns a paginated list of loans belonging to the authenticated client
// @Tags Loans
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/loans [get]
func (h *LoanHandler) GetMyLoans(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	client, err := h.clientRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "client not found"})
		return
	}

	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	loans, total, err := h.creditService.GetLoansByClient(c.Request.Context(), client.ID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToLoanResponses(loans), Total: total, Offset: req.Offset, Limit: req.Limit})
}

// GetLoanDetail godoc
// @Summary Get loan details
// @Description Returns the full details of a specific loan including its installment schedule
// @Tags Loans
// @Produce json
// @Param id path string true "Loan UUID"
// @Success 200 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/loans/{id} [get]
func (h *LoanHandler) GetLoanDetail(c *gin.Context) {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid loan ID"})
		return
	}
	loan, err := h.creditService.GetLoanDetail(c.Request.Context(), loanID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "loan not found"})
		return
	}
	c.JSON(http.StatusOK, dto.ToLoanResponse(loan))
}

// ApproveLoan godoc
// @Summary Approve a loan
// @Description Approves a pending loan request
// @Tags Admin Loans
// @Produce json
// @Param id path string true "Loan UUID"
// @Success 200 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/loans/{id}/approve [post]
func (h *LoanHandler) ApproveLoan(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid loan ID"})
		return
	}
	loan, err := h.creditService.ApproveLoan(c.Request.Context(), adminID, loanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToLoanResponse(loan))
}

// DisburseLoan godoc
// @Summary Disburse a loan
// @Description Disburses an approved loan, crediting the funds to the client's account
// @Tags Admin Loans
// @Produce json
// @Param id path string true "Loan UUID"
// @Success 200 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/loans/{id}/disburse [post]
func (h *LoanHandler) DisburseLoan(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid loan ID"})
		return
	}
	loan, err := h.creditService.DisburseLoan(c.Request.Context(), adminID, loanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToLoanResponse(loan))
}

// CancelLoan godoc
// @Summary Cancel a loan
// @Description Cancels a loan that has not yet been fully disbursed
// @Tags Admin Loans
// @Produce json
// @Param id path string true "Loan UUID"
// @Success 200 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/loans/{id}/cancel [post]
func (h *LoanHandler) CancelLoan(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid loan ID"})
		return
	}
	loan, err := h.creditService.CancelLoan(c.Request.Context(), adminID, loanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToLoanResponse(loan))
}

// PrepayLoan godoc
// @Summary Early loan prepayment
// @Description Processes an early cancellation or partial prepayment of a loan
// @Tags Admin Loans
// @Accept json
// @Produce json
// @Param id path string true "Loan UUID"
// @Param request body dto.PrepayLoanRequest true "Prepayment amount"
// @Success 200 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/loans/{id}/prepay [post]
func (h *LoanHandler) PrepayLoan(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid loan ID"})
		return
	}
	var req dto.PrepayLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}
	loan, err := h.creditService.PrepayLoan(c.Request.Context(), adminID, loanID, amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToLoanResponse(loan))
}

// GetPendingLoans godoc
// @Summary List pending loans
// @Description Returns a paginated list of loans awaiting approval
// @Tags Admin Loans
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/loans/pending [get]
func (h *LoanHandler) GetPendingLoans(c *gin.Context) {
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	loans, total, err := h.creditService.GetLoansByStatus(c.Request.Context(), model.LoanPending, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToLoanResponses(loans), Total: total, Offset: req.Offset, Limit: req.Limit})
}

// AdminCreateLoan godoc
// @Summary Create a loan on behalf of a client
// @Description Admin creates a loan for a given client and credit line
// @Tags Loans
// @Accept json
// @Produce json
// @Param request body dto.AdminCreateLoanRequest true "Loan data"
// @Success 201 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/loans [post]
func (h *LoanHandler) AdminCreateLoan(c *gin.Context) {
	var req dto.AdminCreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	clientID, _ := uuid.Parse(req.ClientID)
	clID, _ := uuid.Parse(req.CreditLineID)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}

	loan, err := h.creditService.RequestLoan(c.Request.Context(), clientID, clID, amount, req.NumInstallments, model.AmortizationType(req.AmortizationType))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToLoanResponse(loan))
}

// AdminCreateWithdrawal godoc
// @Summary Create a cash withdrawal for a client
// @Description Admin creates a loan that is immediately approved and disbursed (cash withdrawal)
// @Tags Admin Loans
// @Accept json
// @Produce json
// @Param request body dto.AdminCreateLoanRequest true "Withdrawal data"
// @Success 201 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/loans/withdrawal [post]
func (h *LoanHandler) AdminCreateWithdrawal(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	var req dto.AdminCreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	clientID, _ := uuid.Parse(req.ClientID)
	clID, _ := uuid.Parse(req.CreditLineID)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}

	loan, err := h.creditService.CreateWithdrawal(c.Request.Context(), adminID, clientID, clID, amount, req.NumInstallments, model.AmortizationType(req.AmortizationType))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToLoanResponse(loan))
}

// AdminRecordPayment godoc
// @Summary Record a payment on behalf of a client
// @Description Admin records a payment for a client's active loan
// @Tags Admin Loans
// @Accept json
// @Produce json
// @Param id path string true "Loan UUID"
// @Param request body dto.RecordPaymentRequest true "Payment data"
// @Success 201 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/loans/{id}/payments [post]
func (h *LoanHandler) AdminRecordPayment(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid loan ID"})
		return
	}
	var req dto.RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}
	var instID *uuid.UUID
	if req.InstallmentID != "" {
		parsed, parseErr := uuid.Parse(req.InstallmentID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid installment ID"})
			return
		}
		instID = &parsed
	}

	payment, err := h.paymentService.RecordPayment(c.Request.Context(), adminID, loanID, amount, model.PaymentMethod(req.Method), req.Reference, instID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToPaymentResponse(payment))
}
