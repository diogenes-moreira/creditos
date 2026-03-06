package gin

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/auth"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/http/gin/handler"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/http/gin/middleware"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/persistence/postgres"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/pdf"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/storage"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

type Router struct {
	engine          *gin.Engine
	db              *gorm.DB
	defaultIVARate  float64
	latePenaltyRate decimal.Decimal
}

func NewRouter(db *gorm.DB, jwtSecret string, defaultIVARate float64, latePenaltyRate float64) *Router {
	r := &Router{
		engine:          gin.Default(),
		db:              db,
		defaultIVARate:  defaultIVARate,
		latePenaltyRate: decimal.NewFromFloat(latePenaltyRate),
	}
	r.engine.Use(middleware.CORS())
	r.engine.Use(middleware.AuditContext())
	r.setupRoutes(jwtSecret)
	return r
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) setupRoutes(jwtSecret string) {
	// Repositories
	userRepo := postgres.NewUserRepository(r.db)
	clientRepo := postgres.NewClientRepository(r.db)
	accountRepo := postgres.NewAccountRepository(r.db)
	movementRepo := postgres.NewMovementRepository(r.db)
	creditLineRepo := postgres.NewCreditLineRepository(r.db)
	loanRepo := postgres.NewLoanRepository(r.db)
	installmentRepo := postgres.NewInstallmentRepository(r.db)
	paymentRepo := postgres.NewPaymentRepository(r.db)
	auditLogRepo := postgres.NewAuditLogRepository(r.db)
	dashboardRepo := postgres.NewDashboardRepository(r.db)
	vendorRepo := postgres.NewVendorRepository(r.db)
	vendorAccountRepo := postgres.NewVendorAccountRepository(r.db)
	vendorMovementRepo := postgres.NewVendorMovementRepository(r.db)
	purchaseRepo := postgres.NewPurchaseRepository(r.db)
	vendorPaymentRepo := postgres.NewVendorPaymentRepository(r.db)

	// Adapters
	authService := auth.NewLocalAuthService(jwtSecret, 24)
	localStorage := storage.NewLocalStorage("./storage")
	pdfGenerator := pdf.NewGenerator()

	// Application services
	auditService := service.NewAuditService(auditLogRepo)
	clientService := service.NewClientService(userRepo, clientRepo, accountRepo, authService, auditService)
	creditService := service.NewCreditService(creditLineRepo, loanRepo, installmentRepo, accountRepo, movementRepo, auditService, clientRepo)
	paymentService := service.NewPaymentService(paymentRepo, loanRepo, installmentRepo, accountRepo, movementRepo, auditService, r.latePenaltyRate)
	accountService := service.NewAccountService(accountRepo, movementRepo)
	dashboardService := service.NewDashboardService(dashboardRepo)
	_ = service.NewPDFAppService(pdfGenerator, localStorage, loanRepo, clientRepo, paymentRepo)
	vendorService := service.NewVendorService(userRepo, vendorRepo, vendorAccountRepo, authService, auditService)
	purchaseService := service.NewPurchaseService(purchaseRepo, vendorRepo, vendorAccountRepo, vendorMovementRepo, clientRepo, creditService, auditService)
	vendorPaymentService := service.NewVendorPaymentService(vendorPaymentRepo, vendorAccountRepo, vendorMovementRepo, vendorRepo, auditService)
	withdrawalRepo := postgres.NewWithdrawalRequestRepository(r.db)
	withdrawalService := service.NewWithdrawalService(withdrawalRepo, vendorRepo, vendorAccountRepo, vendorPaymentRepo, vendorMovementRepo, auditService)

	// Handlers
	healthHandler := handler.NewHealthHandler(r.db)
	authHandler := handler.NewAuthHandler(clientService, userRepo, authService)
	clientHandler := handler.NewClientHandler(clientService, creditService, paymentService, purchaseRepo, accountRepo, movementRepo)
	accountHandler := handler.NewAccountHandler(accountService, clientRepo)
	creditHandler := handler.NewCreditHandler(creditService)
	loanHandler := handler.NewLoanHandler(creditService, paymentService, clientRepo, r.defaultIVARate)
	paymentHandler := handler.NewPaymentHandler(paymentService, clientRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	auditHandler := handler.NewAuditHandler(auditService)
	vendorHandler := handler.NewVendorHandler(vendorService, purchaseService, vendorPaymentService, withdrawalService, pdfGenerator, vendorRepo, vendorAccountRepo, vendorMovementRepo, clientService, creditService)

	// Health
	r.engine.GET("/health", healthHandler.Health)
	r.engine.GET("/health/ready", healthHandler.Ready)

	// Swagger
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.engine.Group("/api/v1")

	// Auth (public)
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// Authenticated routes
	authenticated := api.Group("")
	authenticated.Use(middleware.AuthMiddleware(authService))

	// Client routes
	clientRoutes := authenticated.Group("")
	clientRoutes.Use(middleware.RequireRole(userRepo, model.RoleClient, model.RoleAdmin))
	{
		clientRoutes.GET("/me/profile", clientHandler.GetProfile)
		clientRoutes.PUT("/me/profile", clientHandler.UpdateProfile)
		clientRoutes.PUT("/me/mercadopago", clientHandler.SetMercadoPagoLink)
		clientRoutes.GET("/me/account", accountHandler.GetAccount)
		clientRoutes.GET("/me/account/movements", accountHandler.GetMovements)
		clientRoutes.GET("/me/loans", loanHandler.GetMyLoans)
		clientRoutes.GET("/me/loans/:id", loanHandler.GetLoanDetail)
		clientRoutes.POST("/me/loans", loanHandler.RequestLoan)
		clientRoutes.POST("/me/loans/:loanId/payments", paymentHandler.RecordPayment)
		clientRoutes.GET("/me/payments", paymentHandler.GetMyPayments)
	}

	// Loan simulation (any authenticated user)
	authenticated.POST("/loans/simulate", loanHandler.Simulate)

	// Vendor routes
	vendorRoutes := authenticated.Group("/me/vendor")
	vendorRoutes.Use(middleware.RequireRole(userRepo, model.RoleVendor))
	{
		vendorRoutes.GET("/profile", vendorHandler.GetProfile)
		vendorRoutes.PUT("/profile", vendorHandler.UpdateProfile)
		vendorRoutes.GET("/account", vendorHandler.GetAccount)
		vendorRoutes.GET("/account/movements", vendorHandler.GetMovements)
		vendorRoutes.GET("/purchases", vendorHandler.GetPurchases)
		vendorRoutes.POST("/purchases", vendorHandler.RecordPurchase)
		vendorRoutes.GET("/clients/search", vendorHandler.SearchClients)
		vendorRoutes.POST("/clients/register", vendorHandler.RegisterClient)
		vendorRoutes.GET("/clients/:id/credit-lines", vendorHandler.GetClientCreditLines)
		vendorRoutes.POST("/clients/:id/credit-lines", vendorHandler.RequestCreditLine)
		vendorRoutes.POST("/withdrawals", vendorHandler.RequestWithdrawal)
		vendorRoutes.GET("/withdrawals", vendorHandler.GetMyWithdrawals)
		vendorRoutes.GET("/payments/:id/receipt", vendorHandler.GetMyPaymentReceipt)
	}

	// Admin routes
	adminRoutes := authenticated.Group("/admin")
	adminRoutes.Use(middleware.RequireRole(userRepo, model.RoleAdmin))
	{
		adminRoutes.GET("/clients", clientHandler.ListClients)
		adminRoutes.GET("/clients/search", clientHandler.ListClients)
		adminRoutes.GET("/clients/:id", clientHandler.GetClient)
		adminRoutes.PUT("/clients/:id/iva-rate", clientHandler.UpdateIVARate)
		adminRoutes.POST("/clients/:id/block", clientHandler.BlockClient)
		adminRoutes.POST("/clients/:id/unblock", clientHandler.UnblockClient)
		adminRoutes.GET("/clients/:id/loans", clientHandler.GetClientLoans)
		adminRoutes.GET("/clients/:id/credit-lines", clientHandler.GetClientCreditLines)
		adminRoutes.GET("/clients/:id/payments", clientHandler.GetClientPayments)
		adminRoutes.GET("/clients/:id/purchases", clientHandler.GetClientPurchases)
		adminRoutes.GET("/clients/:id/account", clientHandler.GetClientAccount)
		adminRoutes.GET("/clients/:id/account/movements", clientHandler.GetClientMovements)

		adminRoutes.POST("/credit-lines", creditHandler.CreateCreditLine)
		adminRoutes.GET("/credit-lines/pending", creditHandler.GetPendingCreditLines)
		adminRoutes.PUT("/credit-lines/:id", creditHandler.UpdateCreditLine)
		adminRoutes.POST("/credit-lines/:id/approve", creditHandler.ApproveCreditLine)
		adminRoutes.POST("/credit-lines/:id/reject", creditHandler.RejectCreditLine)

		adminRoutes.POST("/loans", loanHandler.AdminCreateLoan)
		adminRoutes.POST("/loans/withdrawal", loanHandler.AdminCreateWithdrawal)
		adminRoutes.GET("/loans/pending", loanHandler.GetPendingLoans)
		adminRoutes.POST("/loans/:id/approve", loanHandler.ApproveLoan)
		adminRoutes.POST("/loans/:id/disburse", loanHandler.DisburseLoan)
		adminRoutes.POST("/loans/:id/cancel", loanHandler.CancelLoan)
		adminRoutes.POST("/loans/:id/prepay", loanHandler.PrepayLoan)
		adminRoutes.POST("/loans/:id/payments", loanHandler.AdminRecordPayment)

		adminRoutes.PUT("/payments/:id/adjust", paymentHandler.AdjustPayment)

		adminRoutes.GET("/dashboard/portfolio", dashboardHandler.GetPortfolio)
		adminRoutes.GET("/dashboard/delinquency", dashboardHandler.GetDelinquency)
		adminRoutes.GET("/dashboard/kpis", dashboardHandler.GetKPIs)
		adminRoutes.GET("/dashboard/trends/disbursements", dashboardHandler.GetDisbursementTrend)
		adminRoutes.GET("/dashboard/trends/collections", dashboardHandler.GetCollectionTrend)

		adminRoutes.GET("/audit", auditHandler.GetAuditLogs)

		// Vendor management (admin)
		adminRoutes.GET("/vendors", vendorHandler.AdminListVendors)
		adminRoutes.GET("/vendors/:id", vendorHandler.AdminGetVendor)
		adminRoutes.POST("/vendors", vendorHandler.AdminRegisterVendor)
		adminRoutes.POST("/vendors/:id/activate", vendorHandler.AdminActivateVendor)
		adminRoutes.POST("/vendors/:id/deactivate", vendorHandler.AdminDeactivateVendor)
		adminRoutes.GET("/vendors/:id/purchases", vendorHandler.AdminGetVendorPurchases)
		adminRoutes.POST("/vendors/:id/purchases", vendorHandler.AdminRecordPurchase)
		adminRoutes.GET("/vendors/:id/payments", vendorHandler.AdminGetVendorPayments)
		adminRoutes.POST("/vendors/:id/payments", vendorHandler.AdminRecordVendorPayment)
		adminRoutes.GET("/vendors/:id/withdrawals", vendorHandler.AdminGetVendorWithdrawals)
		adminRoutes.GET("/vendors/:id/payments/:paymentId/receipt", vendorHandler.AdminGetVendorPaymentReceipt)
		adminRoutes.GET("/withdrawals/pending", vendorHandler.AdminGetPendingWithdrawals)
		adminRoutes.POST("/withdrawals/:id/approve", vendorHandler.AdminApproveWithdrawal)
		adminRoutes.POST("/withdrawals/:id/reject", vendorHandler.AdminRejectWithdrawal)
	}

	// Serve frontend SPA if FRONTEND_DIR is set
	if frontendDir := os.Getenv("FRONTEND_DIR"); frontendDir != "" {
		r.engine.Static("/assets", filepath.Join(frontendDir, "assets"))
		r.engine.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// Serve static files if they exist
			if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/swagger/") && !strings.HasPrefix(path, "/health") {
				filePath := filepath.Join(frontendDir, path)
				if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
					c.File(filePath)
					return
				}
				// SPA fallback: serve index.html
				c.File(filepath.Join(frontendDir, "index.html"))
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		})
	}
}
