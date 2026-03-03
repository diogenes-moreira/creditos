package handler

import (
	"net/http"

	"github.com/diogenes-moreira/creditos/backend/internal/application/dto"
	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccountHandler struct {
	accountService *service.AccountService
	clientRepo     port.ClientRepository
}

func NewAccountHandler(accountService *service.AccountService, clientRepo port.ClientRepository) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		clientRepo:     clientRepo,
	}
}

// GetAccount godoc
// @Summary Get current account
// @Description Returns the current account details for the authenticated client
// @Tags Account
// @Produce json
// @Success 200 {object} dto.AccountResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/account [get]
func (h *AccountHandler) GetAccount(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	client, err := h.clientRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "client not found"})
		return
	}
	account, err := h.accountService.GetByClientID(c.Request.Context(), client.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "account not found"})
		return
	}
	c.JSON(http.StatusOK, dto.ToAccountResponse(account))
}

// GetMovements godoc
// @Summary Get account movements
// @Description Returns a paginated list of movements for the authenticated client's account
// @Tags Account
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/account/movements [get]
func (h *AccountHandler) GetMovements(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	client, err := h.clientRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "client not found"})
		return
	}
	account, err := h.accountService.GetByClientID(c.Request.Context(), client.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "account not found"})
		return
	}

	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	movements, total, err := h.accountService.GetMovements(c.Request.Context(), account.ID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToMovementResponses(movements), Total: total, Offset: req.Offset, Limit: req.Limit})
}
