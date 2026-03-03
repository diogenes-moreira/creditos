package service

import (
	"context"
	"fmt"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreditService struct {
	creditLineRepo  port.CreditLineRepository
	loanRepo        port.LoanRepository
	installmentRepo port.InstallmentRepository
	accountRepo     port.AccountRepository
	movementRepo    port.MovementRepository
	audit           *AuditService
}

func NewCreditService(
	creditLineRepo port.CreditLineRepository,
	loanRepo port.LoanRepository,
	installmentRepo port.InstallmentRepository,
	accountRepo port.AccountRepository,
	movementRepo port.MovementRepository,
	audit *AuditService,
) *CreditService {
	return &CreditService{
		creditLineRepo:  creditLineRepo,
		loanRepo:        loanRepo,
		installmentRepo: installmentRepo,
		accountRepo:     accountRepo,
		movementRepo:    movementRepo,
		audit:           audit,
	}
}

func (s *CreditService) CreateCreditLine(ctx context.Context, adminID, clientID uuid.UUID, maxAmount, interestRate decimal.Decimal, maxInstallments int) (*model.CreditLine, error) {
	cl, err := model.NewCreditLine(clientID, maxAmount, interestRate, maxInstallments)
	if err != nil {
		return nil, err
	}
	if err := s.creditLineRepo.Create(ctx, cl); err != nil {
		return nil, fmt.Errorf("failed to create credit line: %w", err)
	}
	s.audit.Record(ctx, &adminID, "create_credit_line", "credit_line", cl.ID.String(), fmt.Sprintf("Credit line created: %s", maxAmount.StringFixed(2)))
	return cl, nil
}

func (s *CreditService) ApproveCreditLine(ctx context.Context, adminID, creditLineID uuid.UUID) (*model.CreditLine, error) {
	cl, err := s.creditLineRepo.FindByID(ctx, creditLineID)
	if err != nil {
		return nil, err
	}
	if err := cl.Approve(adminID); err != nil {
		return nil, err
	}
	if err := s.creditLineRepo.Update(ctx, cl); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "approve_credit_line", "credit_line", cl.ID.String(), "Credit line approved")
	return cl, nil
}

func (s *CreditService) RejectCreditLine(ctx context.Context, adminID, creditLineID uuid.UUID, reason string) (*model.CreditLine, error) {
	cl, err := s.creditLineRepo.FindByID(ctx, creditLineID)
	if err != nil {
		return nil, err
	}
	if err := cl.Reject(adminID, reason); err != nil {
		return nil, err
	}
	if err := s.creditLineRepo.Update(ctx, cl); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "reject_credit_line", "credit_line", cl.ID.String(), fmt.Sprintf("Credit line rejected: %s", reason))
	return cl, nil
}

func (s *CreditService) GetCreditLinesByClient(ctx context.Context, clientID uuid.UUID) ([]model.CreditLine, error) {
	return s.creditLineRepo.FindByClientID(ctx, clientID)
}

func (s *CreditService) GetPendingCreditLines(ctx context.Context, offset, limit int) ([]model.CreditLine, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.creditLineRepo.FindByStatus(ctx, model.CreditLinePending, offset, limit)
}

func (s *CreditService) SimulateLoan(principal, interestRate decimal.Decimal, numInstallments int, amortType model.AmortizationType) model.AmortizationSchedule {
	startDate := time.Now()
	switch amortType {
	case model.AmortizationGerman:
		return model.CalculateGermanAmortization(principal, interestRate, numInstallments, startDate)
	default:
		return model.CalculateFrenchAmortization(principal, interestRate, numInstallments, startDate)
	}
}

func (s *CreditService) RequestLoan(ctx context.Context, clientID, creditLineID uuid.UUID, amount decimal.Decimal, numInstallments int, amortType model.AmortizationType) (*model.Loan, error) {
	cl, err := s.creditLineRepo.FindByID(ctx, creditLineID)
	if err != nil {
		return nil, fmt.Errorf("credit line not found: %w", err)
	}
	if cl.ClientID != clientID {
		return nil, fmt.Errorf("credit line does not belong to client")
	}
	if err := cl.CanDisburse(amount); err != nil {
		return nil, err
	}
	if numInstallments > cl.MaxInstallments {
		return nil, fmt.Errorf("requested installments %d exceeds max %d", numInstallments, cl.MaxInstallments)
	}

	loan, err := model.NewLoan(clientID, creditLineID, amount, cl.InterestRate, numInstallments, amortType)
	if err != nil {
		return nil, err
	}
	if err := loan.RequestApproval(); err != nil {
		return nil, err
	}
	if err := s.loanRepo.Create(ctx, loan); err != nil {
		return nil, fmt.Errorf("failed to create loan: %w", err)
	}
	s.audit.Record(ctx, &clientID, "request_loan", "loan", loan.ID.String(), fmt.Sprintf("Loan requested: %s", amount.StringFixed(2)))
	return loan, nil
}

func (s *CreditService) ApproveLoan(ctx context.Context, adminID, loanID uuid.UUID) (*model.Loan, error) {
	loan, err := s.loanRepo.FindByID(ctx, loanID)
	if err != nil {
		return nil, err
	}
	if err := loan.Approve(adminID); err != nil {
		return nil, err
	}
	if err := s.loanRepo.Update(ctx, loan); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "approve_loan", "loan", loan.ID.String(), "Loan approved")
	return loan, nil
}

func (s *CreditService) DisburseLoan(ctx context.Context, adminID, loanID uuid.UUID) (*model.Loan, error) {
	loan, err := s.loanRepo.FindByIDWithInstallments(ctx, loanID)
	if err != nil {
		return nil, err
	}

	cl, err := s.creditLineRepo.FindByID(ctx, loan.CreditLineID)
	if err != nil {
		return nil, err
	}
	if err := cl.CanDisburse(loan.Principal); err != nil {
		return nil, err
	}

	installments, err := loan.Disburse(time.Now())
	if err != nil {
		return nil, err
	}

	cl.RecordDisbursement(loan.Principal)
	if err := s.creditLineRepo.Update(ctx, cl); err != nil {
		return nil, err
	}

	if err := s.installmentRepo.CreateBatch(ctx, installments); err != nil {
		return nil, fmt.Errorf("failed to create installments: %w", err)
	}

	account, err := s.accountRepo.FindByClientID(ctx, loan.ClientID)
	if err != nil {
		return nil, err
	}
	movement, err := account.Credit(loan.Principal, "Loan disbursement", loan.ID.String())
	if err != nil {
		return nil, err
	}
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	if err := s.movementRepo.Create(ctx, movement); err != nil {
		return nil, err
	}

	if err := s.loanRepo.Update(ctx, loan); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "disburse_loan", "loan", loan.ID.String(), fmt.Sprintf("Loan disbursed: %s", loan.Principal.StringFixed(2)))
	return loan, nil
}

func (s *CreditService) CancelLoan(ctx context.Context, adminID, loanID uuid.UUID) (*model.Loan, error) {
	loan, err := s.loanRepo.FindByIDWithInstallments(ctx, loanID)
	if err != nil {
		return nil, err
	}
	if err := loan.Cancel(); err != nil {
		return nil, err
	}

	cl, err := s.creditLineRepo.FindByID(ctx, loan.CreditLineID)
	if err != nil {
		return nil, err
	}
	cl.ReleaseDisbursement(loan.Principal)
	if err := s.creditLineRepo.Update(ctx, cl); err != nil {
		return nil, err
	}

	if err := s.loanRepo.Update(ctx, loan); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "cancel_loan", "loan", loan.ID.String(), "Loan cancelled (early cancellation)")
	return loan, nil
}

func (s *CreditService) PrepayLoan(ctx context.Context, adminID, loanID uuid.UUID, amount decimal.Decimal) (*model.Loan, error) {
	loan, err := s.loanRepo.FindByIDWithInstallments(ctx, loanID)
	if err != nil {
		return nil, err
	}
	if loan.Status != model.LoanActive {
		return nil, fmt.Errorf("can only prepay active loans")
	}

	remaining := amount
	for i := range loan.Installments {
		if remaining.IsZero() || !remaining.IsPositive() {
			break
		}
		inst := &loan.Installments[i]
		if inst.Status == model.InstallmentPaid {
			continue
		}
		applied, surplus, err := inst.ApplyPayment(remaining)
		if err != nil {
			return nil, err
		}
		remaining = surplus
		if err := s.installmentRepo.Update(ctx, inst); err != nil {
			return nil, err
		}
		_ = applied
	}

	if loan.CheckCompletion() {
		_ = loan.Complete()
	}

	if err := s.loanRepo.Update(ctx, loan); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "prepay_loan", "loan", loan.ID.String(), fmt.Sprintf("Capital prepayment: %s", amount.StringFixed(2)))
	return loan, nil
}

func (s *CreditService) GetLoansByClient(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]model.Loan, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.loanRepo.FindByClientID(ctx, clientID, offset, limit)
}

func (s *CreditService) GetLoanDetail(ctx context.Context, loanID uuid.UUID) (*model.Loan, error) {
	return s.loanRepo.FindByIDWithInstallments(ctx, loanID)
}

func (s *CreditService) GetLoansByStatus(ctx context.Context, status model.LoanStatus, offset, limit int) ([]model.Loan, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.loanRepo.FindByStatus(ctx, status, offset, limit)
}
