package service

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type VendorPaymentService struct {
	vendorPaymentRepo  port.VendorPaymentRepository
	vendorAccountRepo  port.VendorAccountRepository
	vendorMovementRepo port.VendorMovementRepository
	vendorRepo         port.VendorRepository
	audit              *AuditService
}

func NewVendorPaymentService(
	vendorPaymentRepo port.VendorPaymentRepository,
	vendorAccountRepo port.VendorAccountRepository,
	vendorMovementRepo port.VendorMovementRepository,
	vendorRepo port.VendorRepository,
	audit *AuditService,
) *VendorPaymentService {
	return &VendorPaymentService{
		vendorPaymentRepo:  vendorPaymentRepo,
		vendorAccountRepo:  vendorAccountRepo,
		vendorMovementRepo: vendorMovementRepo,
		vendorRepo:         vendorRepo,
		audit:              audit,
	}
}

func (s *VendorPaymentService) RecordPayment(ctx context.Context, adminUserID, vendorID uuid.UUID, amount decimal.Decimal, method model.VendorPaymentMethod, reference string) (*model.VendorPayment, error) {
	vendor, err := s.vendorRepo.FindByID(ctx, vendorID)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}
	if !vendor.IsActive {
		return nil, fmt.Errorf("vendor is not active")
	}

	payment, err := model.NewVendorPayment(vendorID, amount, method, reference, adminUserID)
	if err != nil {
		return nil, err
	}
	if err := s.vendorPaymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create vendor payment: %w", err)
	}

	vendorAccount, err := s.vendorAccountRepo.FindByVendorID(ctx, vendorID)
	if err != nil {
		return nil, fmt.Errorf("vendor account not found: %w", err)
	}
	movement, err := vendorAccount.Debit(amount, fmt.Sprintf("Payment received: %s", string(method)), payment.ID.String())
	if err != nil {
		return nil, err
	}
	if err := s.vendorAccountRepo.Update(ctx, vendorAccount); err != nil {
		return nil, err
	}
	if err := s.vendorMovementRepo.Create(ctx, movement); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &adminUserID, "record_vendor_payment", "vendor_payment", payment.ID.String(),
		fmt.Sprintf("Payment to vendor %s: %s via %s", vendor.BusinessName, amount.StringFixed(2), string(method)))

	return payment, nil
}

func (s *VendorPaymentService) GetByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]model.VendorPayment, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.vendorPaymentRepo.FindByVendorID(ctx, vendorID, offset, limit)
}
