package gin

import (
	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/auth"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/http/gin/handler"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/http/gin/middleware"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/persistence/postgres"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/pdf"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/storage"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

type Router struct {
	engine *gin.Engine
	db     *gorm.DB
}

func NewRouter(db *gorm.DB, jwtSecret string) *Router {
	r := &Router{
		engine: gin.Default(),
		db:     db,
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

	// Adapters
	authService := auth.NewLocalAuthService(jwtSecret, 24)
	localStorage := storage.NewLocalStorage("./storage")
	pdfGenerator := pdf.NewGenerator()

	// Application services
	auditService := service.NewAuditService(auditLogRepo)
	clientService := service.NewClientService(userRepo, clientRepo, accountRepo, authService, auditService)
	creditService := service.NewCreditService(creditLineRepo, loanRepo, installmentRepo, accountRepo, movementRepo, auditService)
	paymentService := service.NewPaymentService(paymentRepo, loanRepo, installmentRepo, accountRepo, movementRepo, auditService)
	accountService := service.NewAccountService(accountRepo, movementRepo)
	dashboardService := service.NewDashboardService(dashboardRepo)
	_ = service.NewPDFAppService(pdfGenerator, localStorage, loanRepo, clientRepo, paymentRepo)

	// Handlers
	healthHandler := handler.NewHealthHandler(r.db)
	authHandler := handler.NewAuthHandler(clientService, userRepo, authService)
	clientHandler := handler.NewClientHandler(clientService)
	accountHandler := handler.NewAccountHandler(accountService, clientRepo)
	creditHandler := handler.NewCreditHandler(creditService)
	loanHandler := handler.NewLoanHandler(creditService, clientRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService, clientRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	auditHandler := handler.NewAuditHandler(auditService)

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

	// Admin routes
	adminRoutes := authenticated.Group("/admin")
	adminRoutes.Use(middleware.RequireRole(userRepo, model.RoleAdmin))
	{
		adminRoutes.GET("/clients", clientHandler.ListClients)
		adminRoutes.GET("/clients/:id", clientHandler.GetClient)
		adminRoutes.GET("/clients/search", clientHandler.ListClients)
		adminRoutes.POST("/clients/:id/block", clientHandler.BlockClient)
		adminRoutes.POST("/clients/:id/unblock", clientHandler.UnblockClient)

		adminRoutes.POST("/credit-lines", creditHandler.CreateCreditLine)
		adminRoutes.GET("/credit-lines/pending", creditHandler.GetPendingCreditLines)
		adminRoutes.POST("/credit-lines/:id/approve", creditHandler.ApproveCreditLine)
		adminRoutes.POST("/credit-lines/:id/reject", creditHandler.RejectCreditLine)

		adminRoutes.GET("/loans/pending", loanHandler.GetPendingLoans)
		adminRoutes.POST("/loans/:id/approve", loanHandler.ApproveLoan)
		adminRoutes.POST("/loans/:id/disburse", loanHandler.DisburseLoan)
		adminRoutes.POST("/loans/:id/cancel", loanHandler.CancelLoan)
		adminRoutes.POST("/loans/:id/prepay", loanHandler.PrepayLoan)

		adminRoutes.PUT("/payments/:id/adjust", paymentHandler.AdjustPayment)

		adminRoutes.GET("/dashboard/portfolio", dashboardHandler.GetPortfolio)
		adminRoutes.GET("/dashboard/delinquency", dashboardHandler.GetDelinquency)
		adminRoutes.GET("/dashboard/kpis", dashboardHandler.GetKPIs)
		adminRoutes.GET("/dashboard/trends/disbursements", dashboardHandler.GetDisbursementTrend)
		adminRoutes.GET("/dashboard/trends/collections", dashboardHandler.GetCollectionTrend)

		adminRoutes.GET("/audit", auditHandler.GetAuditLogs)
	}
}
