package service

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PurchaseService struct {
	purchaseRepo       port.PurchaseRepository
	vendorRepo         port.VendorRepository
	vendorAccountRepo  port.VendorAccountRepository
	vendorMovementRepo port.VendorMovementRepository
	clientRepo         port.ClientRepository
	creditService      *CreditService
	audit              *AuditService
}

func NewPurchaseService(
	purchaseRepo port.PurchaseRepository,
	vendorRepo port.VendorRepository,
	vendorAccountRepo port.VendorAccountRepository,
	vendorMovementRepo port.VendorMovementRepository,
	clientRepo port.ClientRepository,
	creditService *CreditService,
	audit *AuditService,
) *PurchaseService {
	return &PurchaseService{
		purchaseRepo:       purchaseRepo,
		vendorRepo:         vendorRepo,
		vendorAccountRepo:  vendorAccountRepo,
		vendorMovementRepo: vendorMovementRepo,
		clientRepo:         clientRepo,
		creditService:      creditService,
		audit:              audit,
	}
}

func (s *PurchaseService) RecordPurchase(ctx context.Context, vendorUserID, clientID, creditLineID uuid.UUID, amount decimal.Decimal, description string, numInstallments int, amortType model.AmortizationType) (*model.Purchase, error) {
	vendor, err := s.vendorRepo.FindByUserID(ctx, vendorUserID)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}
	if !vendor.IsActive {
		return nil, fmt.Errorf("vendor is not active")
	}

	client, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}
	if client.IsBlocked {
		return nil, fmt.Errorf("client is blocked")
	}

	loan, err := s.creditService.CreatePurchaseLoan(ctx, vendorUserID, clientID, creditLineID, amount, numInstallments, amortType)
	if err != nil {
		return nil, fmt.Errorf("failed to create purchase loan: %w", err)
	}

	purchase, err := model.NewPurchase(vendor.ID, clientID, creditLineID, loan.ID, amount, description)
	if err != nil {
		return nil, err
	}
	if err := s.purchaseRepo.Create(ctx, purchase); err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	vendorAccount, err := s.vendorAccountRepo.FindByVendorID(ctx, vendor.ID)
	if err != nil {
		return nil, fmt.Errorf("vendor account not found: %w", err)
	}
	vendorMovement, err := vendorAccount.Credit(amount, fmt.Sprintf("Sale: %s", description), purchase.ID.String())
	if err != nil {
		return nil, err
	}
	if err := s.vendorAccountRepo.Update(ctx, vendorAccount); err != nil {
		return nil, err
	}
	if err := s.vendorMovementRepo.Create(ctx, vendorMovement); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &vendorUserID, "record_purchase", "purchase", purchase.ID.String(),
		fmt.Sprintf("Purchase %s for client %s: %s (%d installments, %s)", amount.StringFixed(2), client.FullName(), description, numInstallments, amortType))

	return purchase, nil
}

func (s *PurchaseService) GetByVendorID(ctx context.Context, vendorID uuid.UUID, offset, limit int) ([]model.Purchase, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.purchaseRepo.FindByVendorID(ctx, vendorID, offset, limit)
}

func (s *PurchaseService) GetByClientID(ctx context.Context, clientID uuid.UUID, offset, limit int) ([]model.Purchase, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.purchaseRepo.FindByClientID(ctx, clientID, offset, limit)
}
