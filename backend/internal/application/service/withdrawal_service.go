package service

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WithdrawalService struct {
	withdrawalRepo     port.WithdrawalRequestRepository
	vendorRepo         port.VendorRepository
	vendorAccountRepo  port.VendorAccountRepository
	vendorPaymentRepo  port.VendorPaymentRepository
	vendorMovementRepo port.VendorMovementRepository
	audit              *AuditService
}

func NewWithdrawalService(
	withdrawalRepo port.WithdrawalRequestRepository,
	vendorRepo port.VendorRepository,
	vendorAccountRepo port.VendorAccountRepository,
	vendorPaymentRepo port.VendorPaymentRepository,
	vendorMovementRepo port.VendorMovementRepository,
	audit *AuditService,
) *WithdrawalService {
	return &WithdrawalService{
		withdrawalRepo:     withdrawalRepo,
		vendorRepo:         vendorRepo,
		vendorAccountRepo:  vendorAccountRepo,
		vendorPaymentRepo:  vendorPaymentRepo,
		vendorMovementRepo: vendorMovementRepo,
		audit:              audit,
	}
}

func (s *WithdrawalService) RequestWithdrawal(ctx context.Context, vendorUserID uuid.UUID, amount decimal.Decimal, method model.VendorPaymentMethod) (*model.WithdrawalRequest, error) {
	vendor, err := s.vendorRepo.FindByUserID(ctx, vendorUserID)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}
	if !vendor.IsActive {
		return nil, fmt.Errorf("vendor is not active")
	}

	account, err := s.vendorAccountRepo.FindByVendorID(ctx, vendor.ID)
	if err != nil {
		return nil, fmt.Errorf("vendor account not found: %w", err)
	}
	if account.Balance.LessThan(amount) {
		return nil, fmt.Errorf("insufficient balance: available %s, requested %s", account.Balance.StringFixed(2), amount.StringFixed(2))
	}

	wr, err := model.NewWithdrawalRequest(vendor.ID, amount, method)
	if err != nil {
		return nil, err
	}
	if err := s.withdrawalRepo.Create(ctx, wr); err != nil {
		return nil, fmt.Errorf("failed to create withdrawal request: %w", err)
	}

	s.audit.Record(ctx, &vendor.UserID, "request_withdrawal", "withdrawal_request", wr.ID.String(),
		fmt.Sprintf("Withdrawal request by vendor %s: %s via %s", vendor.BusinessName, amount.StringFixed(2), string(method)))

	return wr, nil
}

func (s *WithdrawalService) ApproveWithdrawal(ctx context.Context, adminUserID uuid.UUID, withdrawalID uuid.UUID, reference string) (*model.WithdrawalRequest, error) {
	wr, err := s.withdrawalRepo.FindByID(ctx, withdrawalID)
	if err != nil {
		return nil, fmt.Errorf("withdrawal request not found: %w", err)
	}

	if err := wr.Approve(adminUserID); err != nil {
		return nil, err
	}

	payment, err := model.NewVendorPayment(wr.VendorID, wr.Amount, wr.Method, reference, adminUserID)
	if err != nil {
		return nil, err
	}
	if err := s.vendorPaymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create vendor payment: %w", err)
	}

	vendorAccount, err := s.vendorAccountRepo.FindByVendorID(ctx, wr.VendorID)
	if err != nil {
		return nil, fmt.Errorf("vendor account not found: %w", err)
	}
	movement, err := vendorAccount.Debit(wr.Amount, fmt.Sprintf("Withdrawal approved: %s", string(wr.Method)), payment.ID.String())
	if err != nil {
		return nil, err
	}
	if err := s.vendorAccountRepo.Update(ctx, vendorAccount); err != nil {
		return nil, err
	}
	if err := s.vendorMovementRepo.Create(ctx, movement); err != nil {
		return nil, err
	}

	if err := wr.MarkPaid(payment.ID, reference); err != nil {
		return nil, err
	}
	if err := s.withdrawalRepo.Update(ctx, wr); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &adminUserID, "approve_withdrawal", "withdrawal_request", wr.ID.String(),
		fmt.Sprintf("Withdrawal approved for vendor %s: %s via %s", wr.Vendor.BusinessName, wr.Amount.StringFixed(2), string(wr.Method)))

	return wr, nil
}

func (s *WithdrawalService) RejectWithdrawal(ctx context.Context, adminUserID uuid.UUID, withdrawalID uuid.UUID, reason string) (*model.WithdrawalRequest, error) {
	wr, err := s.withdrawalRepo.FindByID(ctx, withdrawalID)
	if err != nil {
		return nil, fmt.Errorf("withdrawal request not found: %w", err)
	}

	if err := wr.Reject(adminUserID, reason); err != nil {
		return nil, err
	}
	if err := s.withdrawalRepo.Update(ctx, wr); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &adminUserID, "reject_withdrawal", "withdrawal_request", wr.ID.String(),
		fmt.Sprintf("Withdrawal rejected for vendor %s: %s", wr.Vendor.BusinessName, reason))

	return wr, nil
}

func (s *WithdrawalService) GetByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]model.WithdrawalRequest, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.withdrawalRepo.FindByVendorID(ctx, vendorID, offset, limit)
}

func (s *WithdrawalService) GetPending(ctx context.Context, offset, limit int) ([]model.WithdrawalRequest, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.withdrawalRepo.FindPending(ctx, offset, limit)
}

func (s *WithdrawalService) GetByID(ctx context.Context, id uuid.UUID) (*model.WithdrawalRequest, error) {
	return s.withdrawalRepo.FindByID(ctx, id)
}
