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

type PaymentHandler struct {
	paymentService *service.PaymentService
	clientRepo     port.ClientRepository
}

func NewPaymentHandler(paymentService *service.PaymentService, clientRepo port.ClientRepository) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		clientRepo:     clientRepo,
	}
}

// RecordPayment godoc
// @Summary Record a payment
// @Description Records a payment against a specific loan for the authenticated client
// @Tags Payments
// @Accept json
// @Produce json
// @Param loanId path string true "Loan UUID"
// @Param request body dto.RecordPaymentRequest true "Payment data"
// @Success 201 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/loans/{loanId}/payments [post]
func (h *PaymentHandler) RecordPayment(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	loanID, err := uuid.Parse(c.Param("loanId"))
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

	payment, err := h.paymentService.RecordPayment(c.Request.Context(), userID, loanID, amount, model.PaymentMethod(req.Method), req.Reference, instID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToPaymentResponse(payment))
}

// GetMyPayments godoc
// @Summary List my payments
// @Description Returns a paginated list of all payments made by the authenticated client
// @Tags Payments
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/payments [get]
func (h *PaymentHandler) GetMyPayments(c *gin.Context) {
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
	payments, total, err := h.paymentService.GetPaymentsByClient(c.Request.Context(), client.ID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToPaymentResponses(payments), Total: total, Offset: req.Offset, Limit: req.Limit})
}

// AdjustPayment godoc
// @Summary Adjust a payment
// @Description Applies an administrative adjustment to an existing payment with a note
// @Tags Admin Payments
// @Accept json
// @Produce json
// @Param id path string true "Payment UUID"
// @Param request body dto.AdjustPaymentRequest true "Adjustment note"
// @Success 200 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/payments/{id}/adjust [put]
func (h *PaymentHandler) AdjustPayment(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payment ID"})
		return
	}

	var req dto.AdjustPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	payment, err := h.paymentService.AdjustPayment(c.Request.Context(), adminID, paymentID, req.Note)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToPaymentResponse(payment))
}
