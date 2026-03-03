package service

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentService struct {
	paymentRepo     port.PaymentRepository
	loanRepo        port.LoanRepository
	installmentRepo port.InstallmentRepository
	accountRepo     port.AccountRepository
	movementRepo    port.MovementRepository
	audit           *AuditService
}

func NewPaymentService(
	paymentRepo port.PaymentRepository,
	loanRepo port.LoanRepository,
	installmentRepo port.InstallmentRepository,
	accountRepo port.AccountRepository,
	movementRepo port.MovementRepository,
	audit *AuditService,
) *PaymentService {
	return &PaymentService{
		paymentRepo:     paymentRepo,
		loanRepo:        loanRepo,
		installmentRepo: installmentRepo,
		accountRepo:     accountRepo,
		movementRepo:    movementRepo,
		audit:           audit,
	}
}

func (s *PaymentService) RecordPayment(ctx context.Context, userID uuid.UUID, loanID uuid.UUID, amount decimal.Decimal, method model.PaymentMethod, reference string) (*model.Payment, error) {
	loan, err := s.loanRepo.FindByIDWithInstallments(ctx, loanID)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}
	if loan.Status != model.LoanActive {
		return nil, fmt.Errorf("can only pay active loans")
	}

	payment, err := model.NewPayment(loanID, amount, method, reference)
	if err != nil {
		return nil, err
	}

	// Apply payment to oldest unpaid installments
	remaining := amount
	for i := range loan.Installments {
		if remaining.IsZero() || !remaining.IsPositive() {
			break
		}
		inst := &loan.Installments[i]
		if inst.Status == model.InstallmentPaid {
			continue
		}
		applied, surplus, applyErr := inst.ApplyPayment(remaining)
		if applyErr != nil {
			return nil, applyErr
		}
		if applied.IsPositive() {
			payment.LinkInstallment(inst.ID)
			if err := s.installmentRepo.Update(ctx, inst); err != nil {
				return nil, err
			}
		}
		remaining = surplus
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Debit account
	account, err := s.accountRepo.FindByClientID(ctx, loan.ClientID)
	if err != nil {
		return nil, err
	}
	movement, err := account.Debit(amount, "Loan payment", payment.ID.String())
	if err != nil {
		return nil, err
	}
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	if err := s.movementRepo.Create(ctx, movement); err != nil {
		return nil, err
	}

	// Check if loan is fully paid
	if loan.CheckCompletion() {
		_ = loan.Complete()
	}
	if err := s.loanRepo.Update(ctx, loan); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &userID, "record_payment", "payment", payment.ID.String(), fmt.Sprintf("Payment recorded: %s via %s", amount.StringFixed(2), method))
	return payment, nil
}

func (s *PaymentService) AdjustPayment(ctx context.Context, adminID, paymentID uuid.UUID, note string) (*model.Payment, error) {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if err := payment.Adjust(adminID, note); err != nil {
		return nil, err
	}
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &adminID, "adjust_payment", "payment", payment.ID.String(), fmt.Sprintf("Payment adjusted: %s", note))
	return payment, nil
}

func (s *PaymentService) GetPaymentsByLoan(ctx context.Context, loanID uuid.UUID, offset, limit int) ([]model.Payment, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.paymentRepo.FindByLoanID(ctx, loanID, offset, limit)
}

func (s *PaymentService) GetPaymentsByClient(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]model.Payment, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.paymentRepo.FindByClientLoans(ctx, clientID, offset, limit)
}
