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
	creditLineRepo     port.CreditLineRepository
	accountRepo        port.AccountRepository
	movementRepo       port.MovementRepository
	vendorRepo         port.VendorRepository
	vendorAccountRepo  port.VendorAccountRepository
	vendorMovementRepo port.VendorMovementRepository
	clientRepo         port.ClientRepository
	audit              *AuditService
}

func NewPurchaseService(
	purchaseRepo port.PurchaseRepository,
	creditLineRepo port.CreditLineRepository,
	accountRepo port.AccountRepository,
	movementRepo port.MovementRepository,
	vendorRepo port.VendorRepository,
	vendorAccountRepo port.VendorAccountRepository,
	vendorMovementRepo port.VendorMovementRepository,
	clientRepo port.ClientRepository,
	audit *AuditService,
) *PurchaseService {
	return &PurchaseService{
		purchaseRepo:       purchaseRepo,
		creditLineRepo:     creditLineRepo,
		accountRepo:        accountRepo,
		movementRepo:       movementRepo,
		vendorRepo:         vendorRepo,
		vendorAccountRepo:  vendorAccountRepo,
		vendorMovementRepo: vendorMovementRepo,
		clientRepo:         clientRepo,
		audit:              audit,
	}
}

func (s *PurchaseService) RecordPurchase(ctx context.Context, vendorUserID, clientID, creditLineID uuid.UUID, amount decimal.Decimal, description string) (*model.Purchase, error) {
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

	cl.RecordDisbursement(amount)
	if err := s.creditLineRepo.Update(ctx, cl); err != nil {
		return nil, fmt.Errorf("failed to update credit line: %w", err)
	}

	purchase, err := model.NewPurchase(vendor.ID, clientID, creditLineID, amount, description)
	if err != nil {
		return nil, err
	}
	if err := s.purchaseRepo.Create(ctx, purchase); err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	account, err := s.accountRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client account not found: %w", err)
	}
	clientMovement, err := account.Debit(amount, fmt.Sprintf("Purchase: %s", description), purchase.ID.String())
	if err != nil {
		return nil, err
	}
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}
	if err := s.movementRepo.Create(ctx, clientMovement); err != nil {
		return nil, err
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
		fmt.Sprintf("Purchase %s for client %s: %s", amount.StringFixed(2), client.FullName(), description))

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
