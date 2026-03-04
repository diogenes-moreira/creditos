package handler

import (
	"net/http"

	"github.com/diogenes-moreira/creditos/backend/internal/application/dto"
	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreditHandler struct {
	creditService *service.CreditService
}

func NewCreditHandler(creditService *service.CreditService) *CreditHandler {
	return &CreditHandler{creditService: creditService}
}

// CreateCreditLine godoc
// @Summary Create a credit line
// @Description Creates a new credit line for a client with specified limits and interest rate
// @Tags Credit Lines
// @Accept json
// @Produce json
// @Param request body dto.CreateCreditLineRequest true "Credit line data"
// @Success 201 {object} dto.CreditLineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/credit-lines [post]
func (h *CreditHandler) CreateCreditLine(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	var req dto.CreateCreditLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	clientID, _ := uuid.Parse(req.ClientID)
	maxAmount, err := decimal.NewFromString(req.MaxAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid max amount"})
		return
	}
	interestRate, err := decimal.NewFromString(req.InterestRate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid interest rate"})
		return
	}
	cl, err := h.creditService.CreateCreditLine(c.Request.Context(), adminID, clientID, maxAmount, interestRate, req.MaxInstallments)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToCreditLineResponse(cl))
}

// ApproveCreditLine godoc
// @Summary Approve a credit line
// @Description Approves a pending credit line, making it available for the client to request loans
// @Tags Credit Lines
// @Produce json
// @Param id path string true "Credit Line UUID"
// @Success 200 {object} dto.CreditLineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/credit-lines/{id}/approve [post]
func (h *CreditHandler) ApproveCreditLine(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	clID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid credit line ID"})
		return
	}
	cl, err := h.creditService.ApproveCreditLine(c.Request.Context(), adminID, clID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToCreditLineResponse(cl))
}

// RejectCreditLine godoc
// @Summary Reject a credit line
// @Description Rejects a pending credit line with a reason
// @Tags Credit Lines
// @Accept json
// @Produce json
// @Param id path string true "Credit Line UUID"
// @Param request body dto.RejectCreditLineRequest true "Rejection reason"
// @Success 200 {object} dto.CreditLineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/credit-lines/{id}/reject [post]
func (h *CreditHandler) RejectCreditLine(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	clID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid credit line ID"})
		return
	}
	var req dto.RejectCreditLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	cl, err := h.creditService.RejectCreditLine(c.Request.Context(), adminID, clID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToCreditLineResponse(cl))
}

// UpdateCreditLine godoc
// @Summary Update a credit line's max amount
// @Description Updates the maximum amount of an existing credit line
// @Tags Credit Lines
// @Accept json
// @Produce json
// @Param id path string true "Credit Line UUID"
// @Param request body dto.UpdateCreditLineRequest true "New max amount"
// @Success 200 {object} dto.CreditLineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/credit-lines/{id} [put]
func (h *CreditHandler) UpdateCreditLine(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	clID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid credit line ID"})
		return
	}
	var req dto.UpdateCreditLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	newMaxAmount, err := decimal.NewFromString(req.MaxAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid max amount"})
		return
	}
	cl, err := h.creditService.UpdateCreditLineMaxAmount(c.Request.Context(), adminID, clID, newMaxAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToCreditLineResponse(cl))
}

// GetPendingCreditLines godoc
// @Summary List pending credit lines
// @Description Returns a paginated list of credit lines awaiting approval
// @Tags Credit Lines
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/credit-lines/pending [get]
func (h *CreditHandler) GetPendingCreditLines(c *gin.Context) {
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	cls, total, err := h.creditService.GetPendingCreditLines(c.Request.Context(), req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToCreditLineResponses(cls), Total: total, Offset: req.Offset, Limit: req.Limit})
}
