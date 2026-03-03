package port

import (
	"context"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByFirebaseUID(ctx context.Context, uid string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

type ClientRepository interface {
	Create(ctx context.Context, client *model.Client) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Client, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Client, error)
	FindByDNI(ctx context.Context, dni string) (*model.Client, error)
	FindByCUIT(ctx context.Context, cuit string) (*model.Client, error)
	Search(ctx context.Context, query string, offset, limit int) ([]model.Client, int64, error)
	FindAll(ctx context.Context, offset, limit int) ([]model.Client, int64, error)
	Update(ctx context.Context, client *model.Client) error
}

type AccountRepository interface {
	Create(ctx context.Context, account *model.CurrentAccount) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.CurrentAccount, error)
	FindByClientID(ctx context.Context, clientID uuid.UUID) (*model.CurrentAccount, error)
	Update(ctx context.Context, account *model.CurrentAccount) error
}

type MovementRepository interface {
	Create(ctx context.Context, movement *model.Movement) error
	FindByAccountID(ctx context.Context, accountID uuid.UUID, offset, limit int) ([]model.Movement, int64, error)
}

type CreditLineRepository interface {
	Create(ctx context.Context, cl *model.CreditLine) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.CreditLine, error)
	FindByClientID(ctx context.Context, clientID uuid.UUID) ([]model.CreditLine, error)
	FindByStatus(ctx context.Context, status model.CreditLineStatus, offset, limit int) ([]model.CreditLine, int64, error)
	Update(ctx context.Context, cl *model.CreditLine) error
}

type LoanRepository interface {
	Create(ctx context.Context, loan *model.Loan) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Loan, error)
	FindByIDWithInstallments(ctx context.Context, id uuid.UUID) (*model.Loan, error)
	FindByClientID(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]model.Loan, int64, error)
	FindByStatus(ctx context.Context, status model.LoanStatus, offset, limit int) ([]model.Loan, int64, error)
	FindActive(ctx context.Context, offset, limit int) ([]model.Loan, int64, error)
	Update(ctx context.Context, loan *model.Loan) error
}

type InstallmentRepository interface {
	CreateBatch(ctx context.Context, installments []model.Installment) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Installment, error)
	FindByLoanID(ctx context.Context, loanID uuid.UUID) ([]model.Installment, error)
	FindUnpaidByLoanID(ctx context.Context, loanID uuid.UUID) ([]model.Installment, error)
	FindOverdue(ctx context.Context) ([]model.Installment, error)
	Update(ctx context.Context, installment *model.Installment) error
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	FindByLoanID(ctx context.Context, loanID uuid.UUID, offset, limit int) ([]model.Payment, int64, error)
	FindByClientLoans(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]model.Payment, int64, error)
	Update(ctx context.Context, payment *model.Payment) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	FindByEntity(ctx context.Context, entityType, entityID string, offset, limit int) ([]model.AuditLog, int64, error)
	FindByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.AuditLog, int64, error)
	FindAll(ctx context.Context, offset, limit int) ([]model.AuditLog, int64, error)
}

type DashboardRepository interface {
	PortfolioSummary(ctx context.Context) (*PortfolioSummary, error)
	DelinquencyRates(ctx context.Context) (*DelinquencyStats, error)
	DisbursementTrend(ctx context.Context, from, to time.Time) ([]TrendPoint, error)
	CollectionTrend(ctx context.Context, from, to time.Time) ([]TrendPoint, error)
}

type PortfolioSummary struct {
	TotalClients      int64
	ActiveLoans       int64
	TotalDisbursed    decimal.Decimal
	TotalOutstanding  decimal.Decimal
	TotalCollected    decimal.Decimal
	PendingApprovals  int64
}

type DelinquencyStats struct {
	PAR30  decimal.Decimal
	PAR60  decimal.Decimal
	PAR90  decimal.Decimal
	TotalOverdue     decimal.Decimal
	OverdueCount     int64
	DelinquencyRate  decimal.Decimal
}

type TrendPoint struct {
	Date   time.Time
	Amount decimal.Decimal
	Count  int64
}
