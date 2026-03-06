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
	clientRepo      port.ClientRepository
	audit           *AuditService
}

func NewCreditService(
	creditLineRepo port.CreditLineRepository,
	loanRepo port.LoanRepository,
	installmentRepo port.InstallmentRepository,
	accountRepo port.AccountRepository,
	movementRepo port.MovementRepository,
	audit *AuditService,
	clientRepo port.ClientRepository,
) *CreditService {
	return &CreditService{
		creditLineRepo:  creditLineRepo,
		loanRepo:        loanRepo,
		installmentRepo: installmentRepo,
		accountRepo:     accountRepo,
		movementRepo:    movementRepo,
		clientRepo:      clientRepo,
		audit:           audit,
	}
}

func (s *CreditService) CreateCreditLine(ctx context.Context, adminID, clientID uuid.UUID, maxAmount, interestRate decimal.Decimal, maxInstallments int, recalculateOnPrepay bool) (*model.CreditLine, error) {
	existing, err := s.creditLineRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing credit lines: %w", err)
	}
	if len(existing) > 0 {
		return nil, fmt.Errorf("client already has a credit line")
	}

	cl, err := model.NewCreditLine(clientID, maxAmount, interestRate, maxInstallments, recalculateOnPrepay)
	if err != nil {
		return nil, err
	}
	// Admin-created credit lines are auto-approved
	if err := cl.Approve(adminID); err != nil {
		return nil, err
	}
	if err := s.creditLineRepo.Create(ctx, cl); err != nil {
		return nil, fmt.Errorf("failed to create credit line: %w", err)
	}
	s.audit.Record(ctx, &adminID, "create_credit_line", "credit_line", cl.ID.String(), fmt.Sprintf("Credit line created and approved: %s", maxAmount.StringFixed(2)))
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

func (s *CreditService) SimulateLoan(principal, interestRate decimal.Decimal, numInstallments int, amortType model.AmortizationType, ivaRate decimal.Decimal) model.AmortizationSchedule {
	startDate := time.Now()
	switch amortType {
	case model.AmortizationGerman:
		return model.CalculateGermanAmortization(principal, interestRate, numInstallments, startDate, ivaRate)
	default:
		return model.CalculateFrenchAmortization(principal, interestRate, numInstallments, startDate, ivaRate)
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

	client, err := s.clientRepo.FindByID(ctx, loan.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client for IVA rate: %w", err)
	}

	installments, err := loan.Disburse(time.Now(), client.IVARate)
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

func (s *CreditService) PrepayLoan(ctx context.Context, adminID, loanID uuid.UUID, amount decimal.Decimal, strategy model.PrepaymentStrategy) (*model.Loan, error) {
	loan, err := s.loanRepo.FindByIDWithInstallments(ctx, loanID)
	if err != nil {
		return nil, err
	}
	if loan.Status != model.LoanActive {
		return nil, fmt.Errorf("can only prepay active loans")
	}

	// Compute outstanding principal from unpaid installments
	outstandingPrincipal := loan.OutstandingPrincipal()
	if !amount.IsPositive() || amount.GreaterThan(outstandingPrincipal) {
		return nil, fmt.Errorf("prepayment amount must be between 0 and outstanding principal (%s)", outstandingPrincipal.StringFixed(2))
	}

	newOutstandingPrincipal := outstandingPrincipal.Sub(amount)

	// Get IVA rate for recalculation
	prepayClient, _ := s.clientRepo.FindByID(ctx, loan.ClientID)
	ivaRate := decimal.NewFromInt(21)
	if prepayClient != nil {
		ivaRate = prepayClient.IVARate
	}

	if newOutstandingPrincipal.IsZero() {
		// Prepay covers all remaining capital — mark all unpaid installments as paid
		for i := range loan.Installments {
			inst := &loan.Installments[i]
			if inst.Status == model.InstallmentPaid {
				continue
			}
			inst.PaidAmount = inst.TotalAmount
			inst.RemainingAmount = decimal.Zero
			inst.Status = model.InstallmentPaid
			if err := s.installmentRepo.Update(ctx, inst); err != nil {
				return nil, err
			}
		}
		_ = loan.Complete()
	} else {
		// Recalculate remaining installments with reduced principal
		switch strategy {
		case model.PrepaymentReduceTerm:
			removedIDs := loan.RecalculateReducingTerm(newOutstandingPrincipal, loan.InterestRate, ivaRate)
			if len(removedIDs) > 0 {
				if err := s.installmentRepo.DeleteBatch(ctx, removedIDs); err != nil {
					return nil, fmt.Errorf("failed to delete excess installments: %w", err)
				}
			}
			for i := range loan.Installments {
				inst := &loan.Installments[i]
				if inst.Status != model.InstallmentPaid {
					if err := s.installmentRepo.Update(ctx, inst); err != nil {
						return nil, err
					}
				}
			}
			s.audit.Record(ctx, &adminID, "recalculate_reduce_term", "loan", loan.ID.String(), fmt.Sprintf("Installments reduced after prepayment, %d removed", len(removedIDs)))
		default:
			loan.RecalculateRemainingInstallments(newOutstandingPrincipal, loan.InterestRate, ivaRate)
			for i := range loan.Installments {
				inst := &loan.Installments[i]
				if inst.Status != model.InstallmentPaid {
					if err := s.installmentRepo.Update(ctx, inst); err != nil {
						return nil, err
					}
				}
			}
			s.audit.Record(ctx, &adminID, "recalculate_installments", "loan", loan.ID.String(), "Remaining installments recalculated after prepayment")
		}

		if loan.CheckCompletion() {
			_ = loan.Complete()
		}
	}

	if err := s.loanRepo.Update(ctx, loan); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "prepay_loan", "loan", loan.ID.String(), fmt.Sprintf("Capital prepayment: %s", amount.StringFixed(2)))
	return loan, nil
}

func (s *CreditService) UpdateCreditLineMaxAmount(ctx context.Context, adminID, creditLineID uuid.UUID, newMaxAmount decimal.Decimal) (*model.CreditLine, error) {
	cl, err := s.creditLineRepo.FindByID(ctx, creditLineID)
	if err != nil {
		return nil, err
	}
	if err := cl.UpdateMaxAmount(newMaxAmount); err != nil {
		return nil, err
	}
	if err := s.creditLineRepo.Update(ctx, cl); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "update_credit_line", "credit_line", cl.ID.String(), fmt.Sprintf("Credit line max amount updated to %s", newMaxAmount.StringFixed(2)))
	return cl, nil
}

func (s *CreditService) UpdateCreditLineFields(ctx context.Context, cl *model.CreditLine) error {
	return s.creditLineRepo.Update(ctx, cl)
}

func (s *CreditService) CreateWithdrawal(ctx context.Context, adminID, clientID, creditLineID uuid.UUID, amount decimal.Decimal, numInstallments int, amortType model.AmortizationType) (*model.Loan, error) {
	loan, err := s.RequestLoan(ctx, clientID, creditLineID, amount, numInstallments, amortType)
	if err != nil {
		return nil, fmt.Errorf("failed to request loan for withdrawal: %w", err)
	}

	loan, err = s.ApproveLoan(ctx, adminID, loan.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to approve loan for withdrawal: %w", err)
	}

	loan, err = s.DisburseLoan(ctx, adminID, loan.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to disburse loan for withdrawal: %w", err)
	}

	s.audit.Record(ctx, &adminID, "create_withdrawal", "loan", loan.ID.String(), fmt.Sprintf("Cash withdrawal created and disbursed: %s", amount.StringFixed(2)))
	return loan, nil
}

func (s *CreditService) CreatePurchaseLoan(ctx context.Context, adminID, clientID, creditLineID uuid.UUID, amount decimal.Decimal, numInstallments int, amortType model.AmortizationType) (*model.Loan, error) {
	loan, err := s.RequestLoan(ctx, clientID, creditLineID, amount, numInstallments, amortType)
	if err != nil {
		return nil, fmt.Errorf("failed to request loan for purchase: %w", err)
	}

	loan, err = s.ApproveLoan(ctx, adminID, loan.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to approve loan for purchase: %w", err)
	}

	loan, err = s.loanRepo.FindByIDWithInstallments(ctx, loan.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch loan: %w", err)
	}

	cl, err := s.creditLineRepo.FindByID(ctx, loan.CreditLineID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credit line: %w", err)
	}
	if err := cl.CanDisburse(loan.Principal); err != nil {
		return nil, err
	}

	client, err := s.clientRepo.FindByID(ctx, loan.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client for IVA rate: %w", err)
	}

	installments, err := loan.Disburse(time.Now(), client.IVARate)
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

	if err := s.loanRepo.Update(ctx, loan); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &adminID, "create_purchase_loan", "loan", loan.ID.String(), fmt.Sprintf("Purchase loan created and disbursed: %s", amount.StringFixed(2)))
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
