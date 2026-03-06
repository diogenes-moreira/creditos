package handler

import (
	"io"
	"net/http"

	"github.com/diogenes-moreira/creditos/backend/internal/application/dto"
	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type VendorHandler struct {
	vendorService        *service.VendorService
	purchaseService      *service.PurchaseService
	vendorPaymentService *service.VendorPaymentService
	withdrawalService    *service.WithdrawalService
	pdfService           port.PDFService
	vendorRepo           port.VendorRepository
	vendorAccountRepo    port.VendorAccountRepository
	vendorMovementRepo   port.VendorMovementRepository
	clientService        *service.ClientService
	creditService        *service.CreditService
}

func NewVendorHandler(
	vendorService *service.VendorService,
	purchaseService *service.PurchaseService,
	vendorPaymentService *service.VendorPaymentService,
	withdrawalService *service.WithdrawalService,
	pdfService port.PDFService,
	vendorRepo port.VendorRepository,
	vendorAccountRepo port.VendorAccountRepository,
	vendorMovementRepo port.VendorMovementRepository,
	clientService *service.ClientService,
	creditService *service.CreditService,
) *VendorHandler {
	return &VendorHandler{
		vendorService:        vendorService,
		purchaseService:      purchaseService,
		vendorPaymentService: vendorPaymentService,
		withdrawalService:    withdrawalService,
		pdfService:           pdfService,
		vendorRepo:           vendorRepo,
		vendorAccountRepo:    vendorAccountRepo,
		vendorMovementRepo:   vendorMovementRepo,
		clientService:        clientService,
		creditService:        creditService,
	}
}

// GetProfile godoc
// @Summary Get vendor profile
// @Tags Vendor
// @Produce json
// @Success 200 {object} dto.VendorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/profile [get]
func (h *VendorHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	vendor, err := h.vendorService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	c.JSON(http.StatusOK, dto.ToVendorResponse(vendor, vendor.User.Email))
}

// UpdateProfile godoc
// @Summary Update vendor profile
// @Tags Vendor
// @Accept json
// @Produce json
// @Param request body dto.UpdateVendorProfileRequest true "Profile data"
// @Success 200 {object} dto.VendorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/profile [put]
func (h *VendorHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	var req dto.UpdateVendorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	vendor, err := h.vendorService.UpdateProfile(c.Request.Context(), userID, req.Phone, req.Address, req.City, req.Province)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToVendorResponse(vendor, vendor.User.Email))
}

// GetAccount godoc
// @Summary Get vendor account balance
// @Tags Vendor
// @Produce json
// @Success 200 {object} dto.VendorAccountResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/account [get]
func (h *VendorHandler) GetAccount(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	vendor, err := h.vendorRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	account, err := h.vendorAccountRepo.FindByVendorID(c.Request.Context(), vendor.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor account not found"})
		return
	}
	c.JSON(http.StatusOK, dto.ToVendorAccountResponse(account))
}

// GetMovements godoc
// @Summary Get vendor account movements
// @Tags Vendor
// @Produce json
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/account/movements [get]
func (h *VendorHandler) GetMovements(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	vendor, err := h.vendorRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	account, err := h.vendorAccountRepo.FindByVendorID(c.Request.Context(), vendor.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor account not found"})
		return
	}
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	movements, total, err := h.vendorMovementRepo.FindByAccountID(c.Request.Context(), account.ID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToVendorMovementResponses(movements), Total: total, Offset: req.Offset, Limit: req.Limit})
}

// GetPurchases godoc
// @Summary Get vendor's purchases
// @Tags Vendor
// @Produce json
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/purchases [get]
func (h *VendorHandler) GetPurchases(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	vendor, err := h.vendorRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	purchases, total, err := h.purchaseService.GetByVendorID(c.Request.Context(), vendor.ID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToPurchaseResponses(purchases), Total: total, Offset: req.Offset, Limit: req.Limit})
}

// SearchClients godoc
// @Summary Search clients (vendor)
// @Tags Vendor
// @Produce json
// @Param q query string false "Search query"
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Security BearerAuth
// @Router /me/vendor/clients/search [get]
func (h *VendorHandler) SearchClients(c *gin.Context) {
	var req dto.SearchClientsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	clients, total, err := h.clientService.Search(c.Request.Context(), req.Query, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	responses := make([]dto.ClientResponse, len(clients))
	for i, cl := range clients {
		responses[i] = dto.ToClientResponse(&cl, cl.User.Email)
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: responses, Total: total, Offset: req.Offset, Limit: req.Limit})
}

// GetClientCreditLines godoc
// @Summary Get client's credit lines (vendor)
// @Tags Vendor
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {array} dto.CreditLineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/clients/{id}/credit-lines [get]
func (h *VendorHandler) GetClientCreditLines(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid client ID"})
		return
	}
	creditLines, err := h.creditService.GetCreditLinesByClient(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToCreditLineResponses(creditLines))
}

// RegisterClient godoc
// @Summary Register a new client (vendor)
// @Tags Vendor
// @Accept json
// @Produce json
// @Param request body dto.RegisterClientByVendorRequest true "Client data"
// @Success 201 {object} dto.ClientResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/clients/register [post]
func (h *VendorHandler) RegisterClient(c *gin.Context) {
	var req dto.RegisterClientByVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	client, user, err := h.clientService.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName, req.DNI, req.CUIT, req.DateOfBirth, req.Phone, req.Address, req.City, req.Province, req.IsPEP)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToClientResponse(client, user.Email))
}

// RequestCreditLine godoc
// @Summary Request a credit line for client (vendor)
// @Tags Vendor
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body dto.RequestCreditLineByVendorRequest true "Credit line data"
// @Success 201 {object} dto.CreditLineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/clients/{id}/credit-lines [post]
func (h *VendorHandler) RequestCreditLine(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid client ID"})
		return
	}
	var req dto.RequestCreditLineByVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
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
	cl, err := h.creditService.CreateCreditLine(c.Request.Context(), userID, clientID, maxAmount, interestRate, req.MaxInstallments, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToCreditLineResponse(cl))
}

// RecordPurchase godoc
// @Summary Record a purchase
// @Tags Vendor
// @Accept json
// @Produce json
// @Param request body dto.RecordPurchaseRequest true "Purchase data"
// @Success 201 {object} dto.PurchaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /me/vendor/purchases [post]
func (h *VendorHandler) RecordPurchase(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	var req dto.RecordPurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid client ID"})
		return
	}
	creditLineID, err := uuid.Parse(req.CreditLineID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid credit line ID"})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}
	purchase, err := h.purchaseService.RecordPurchase(c.Request.Context(), userID, clientID, creditLineID, amount, req.Description, req.NumInstallments, model.AmortizationType(req.AmortizationType))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToPurchaseResponse(purchase))
}

// --- Admin endpoints ---

// AdminListVendors godoc
// @Summary List vendors
// @Tags Admin Vendors
// @Produce json
// @Param q query string false "Search query"
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Security BearerAuth
// @Router /admin/vendors [get]
func (h *VendorHandler) AdminListVendors(c *gin.Context) {
	var req dto.SearchClientsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	vendors, total, err := h.vendorService.Search(c.Request.Context(), req.Query, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	responses := make([]dto.VendorResponse, len(vendors))
	for i, v := range vendors {
		responses[i] = dto.ToVendorResponse(&v, v.User.Email)
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: responses, Total: total, Offset: req.Offset, Limit: req.Limit})
}

// AdminGetVendor godoc
// @Summary Get vendor detail
// @Tags Admin Vendors
// @Produce json
// @Param id path string true "Vendor ID"
// @Success 200 {object} dto.VendorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/vendors/{id} [get]
func (h *VendorHandler) AdminGetVendor(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	vendor, err := h.vendorService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	c.JSON(http.StatusOK, dto.ToVendorResponse(vendor, vendor.User.Email))
}

// AdminRegisterVendor godoc
// @Summary Register a new vendor
// @Tags Admin Vendors
// @Accept json
// @Produce json
// @Param request body dto.RegisterVendorRequest true "Vendor data"
// @Success 201 {object} dto.VendorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/vendors [post]
func (h *VendorHandler) AdminRegisterVendor(c *gin.Context) {
	var req dto.RegisterVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	vendor, user, err := h.vendorService.Register(c.Request.Context(), req.Email, req.Password, req.BusinessName, req.CUIT, req.Phone, req.Address, req.City, req.Province)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToVendorResponse(vendor, user.Email))
}

// AdminActivateVendor godoc
// @Summary Activate a vendor
// @Tags Admin Vendors
// @Produce json
// @Param id path string true "Vendor ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/vendors/{id}/activate [post]
func (h *VendorHandler) AdminActivateVendor(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	vendorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	if err := h.vendorService.Activate(c.Request.Context(), adminID, vendorID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vendor activated"})
}

// AdminDeactivateVendor godoc
// @Summary Deactivate a vendor
// @Tags Admin Vendors
// @Produce json
// @Param id path string true "Vendor ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/vendors/{id}/deactivate [post]
func (h *VendorHandler) AdminDeactivateVendor(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	vendorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	if err := h.vendorService.Deactivate(c.Request.Context(), adminID, vendorID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vendor deactivated"})
}

// AdminGetVendorPurchases godoc
// @Summary Get vendor's purchases (admin)
// @Tags Admin Vendors
// @Produce json
// @Param id path string true "Vendor ID"
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/vendors/{id}/purchases [get]
func (h *VendorHandler) AdminGetVendorPurchases(c *gin.Context) {
	vendorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	purchases, total, err := h.purchaseService.GetByVendorID(c.Request.Context(), vendorID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToPurchaseResponses(purchases), Total: total, Offset: req.Offset, Limit: req.Limit})
}

// AdminRecordPurchase godoc
// @Summary Record a purchase on behalf of a vendor
// @Tags Admin Vendors
// @Accept json
// @Produce json
// @Param id path string true "Vendor ID"
// @Param request body dto.RecordPurchaseRequest true "Purchase data"
// @Success 201 {object} dto.PurchaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/vendors/{id}/purchases [post]
func (h *VendorHandler) AdminRecordPurchase(c *gin.Context) {
	vendorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	var req dto.RecordPurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid client ID"})
		return
	}
	creditLineID, err := uuid.Parse(req.CreditLineID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid credit line ID"})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}
	vendor, err := h.vendorRepo.FindByID(c.Request.Context(), vendorID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	purchase, err := h.purchaseService.RecordPurchase(c.Request.Context(), vendor.UserID, clientID, creditLineID, amount, req.Description, req.NumInstallments, model.AmortizationType(req.AmortizationType))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToPurchaseResponse(purchase))
}

// AdminGetVendorPayments godoc
// @Summary Get vendor's payments (admin)
// @Tags Admin Vendors
// @Produce json
// @Param id path string true "Vendor ID"
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/vendors/{id}/payments [get]
func (h *VendorHandler) AdminGetVendorPayments(c *gin.Context) {
	vendorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	payments, total, err := h.vendorPaymentService.GetByVendorID(c.Request.Context(), vendorID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToVendorPaymentResponses(payments), Total: total, Offset: req.Offset, Limit: req.Limit})
}

// AdminRecordVendorPayment godoc
// @Summary Record payment to vendor (admin)
// @Tags Admin Vendors
// @Accept json
// @Produce json
// @Param id path string true "Vendor ID"
// @Param request body dto.RecordVendorPaymentRequest true "Payment data"
// @Success 201 {object} dto.VendorPaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/vendors/{id}/payments [post]
func (h *VendorHandler) AdminRecordVendorPayment(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	vendorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	var req dto.RecordVendorPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}
	payment, err := h.vendorPaymentService.RecordPayment(c.Request.Context(), adminID, vendorID, amount, model.VendorPaymentMethod(req.Method), req.Reference)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToVendorPaymentResponse(payment))
}

// --- Vendor Withdrawal endpoints ---

func (h *VendorHandler) RequestWithdrawal(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	var req dto.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid amount"})
		return
	}
	wr, err := h.withdrawalService.RequestWithdrawal(c.Request.Context(), userID, amount, model.VendorPaymentMethod(req.Method))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.ToWithdrawalRequestResponse(wr))
}

func (h *VendorHandler) GetMyWithdrawals(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	vendor, err := h.vendorRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	withdrawals, total, err := h.withdrawalService.GetByVendorID(c.Request.Context(), vendor.ID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToWithdrawalRequestResponses(withdrawals), Total: total, Offset: req.Offset, Limit: req.Limit})
}

func (h *VendorHandler) GetMyPaymentReceipt(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payment ID"})
		return
	}
	vendor, err := h.vendorRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	payment, err := h.vendorPaymentService.GetByID(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "payment not found"})
		return
	}
	if payment.VendorID != vendor.ID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "payment does not belong to this vendor"})
		return
	}
	reader, err := h.pdfService.GenerateVendorPaymentReceipt(payment, vendor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to generate receipt"})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=receipt-"+paymentID.String()[:8]+".pdf")
	io.Copy(c.Writer, reader)
}

// --- Admin Withdrawal endpoints ---

func (h *VendorHandler) AdminGetVendorWithdrawals(c *gin.Context) {
	vendorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	withdrawals, total, err := h.withdrawalService.GetByVendorID(c.Request.Context(), vendorID, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToWithdrawalRequestResponses(withdrawals), Total: total, Offset: req.Offset, Limit: req.Limit})
}

func (h *VendorHandler) AdminGetPendingWithdrawals(c *gin.Context) {
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 20
	}
	withdrawals, total, err := h.withdrawalService.GetPending(c.Request.Context(), req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToWithdrawalRequestResponses(withdrawals), Total: total, Offset: req.Offset, Limit: req.Limit})
}

func (h *VendorHandler) AdminApproveWithdrawal(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	withdrawalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid withdrawal ID"})
		return
	}
	var req dto.ApproveWithdrawalRequest
	_ = c.ShouldBindJSON(&req)
	wr, err := h.withdrawalService.ApproveWithdrawal(c.Request.Context(), adminID, withdrawalID, req.Reference)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToWithdrawalRequestResponse(wr))
}

func (h *VendorHandler) AdminRejectWithdrawal(c *gin.Context) {
	adminID := c.MustGet("userID").(uuid.UUID)
	withdrawalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid withdrawal ID"})
		return
	}
	var req dto.RejectWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	wr, err := h.withdrawalService.RejectWithdrawal(c.Request.Context(), adminID, withdrawalID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToWithdrawalRequestResponse(wr))
}

func (h *VendorHandler) AdminGetVendorPaymentReceipt(c *gin.Context) {
	vendorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid vendor ID"})
		return
	}
	paymentID, err := uuid.Parse(c.Param("paymentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payment ID"})
		return
	}
	payment, err := h.vendorPaymentService.GetByID(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "payment not found"})
		return
	}
	if payment.VendorID != vendorID {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "payment does not belong to this vendor"})
		return
	}
	vendor, err := h.vendorService.GetByID(c.Request.Context(), vendorID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "vendor not found"})
		return
	}
	reader, err := h.pdfService.GenerateVendorPaymentReceipt(payment, vendor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to generate receipt"})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=receipt-"+paymentID.String()[:8]+".pdf")
	io.Copy(c.Writer, reader)
}
