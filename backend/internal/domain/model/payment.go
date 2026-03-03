package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentMethod string

const (
	PaymentCash       PaymentMethod = "cash"
	PaymentTransfer   PaymentMethod = "transfer"
	PaymentMercadoPago PaymentMethod = "mercado_pago"
)

type Payment struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	LoanID        uuid.UUID       `gorm:"type:uuid;index;not null"`
	InstallmentID *uuid.UUID      `gorm:"type:uuid;index"`
	Amount        decimal.Decimal `gorm:"type:decimal(18,2);not null"`
	Method        PaymentMethod   `gorm:"type:varchar(20);not null"`
	Reference     string          `gorm:""`
	IsAdjustment  bool            `gorm:"not null;default:false"`
	AdjustedBy    *uuid.UUID      `gorm:"type:uuid"`
	AdjustmentNote string         `gorm:""`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`
}

func NewPayment(loanID uuid.UUID, amount decimal.Decimal, method PaymentMethod, reference string) (*Payment, error) {
	if !amount.IsPositive() {
		return nil, fmt.Errorf("payment amount must be positive")
	}
	if method != PaymentCash && method != PaymentTransfer && method != PaymentMercadoPago {
		return nil, fmt.Errorf("invalid payment method: %s", method)
	}

	return &Payment{
		ID:        uuid.New(),
		LoanID:    loanID,
		Amount:    amount,
		Method:    method,
		Reference: reference,
	}, nil
}

func (p *Payment) Adjust(adjustedBy uuid.UUID, note string) error {
	if p.IsAdjustment {
		return fmt.Errorf("payment is already an adjustment")
	}
	if note == "" {
		return fmt.Errorf("adjustment note is required")
	}
	p.IsAdjustment = true
	p.AdjustedBy = &adjustedBy
	p.AdjustmentNote = note
	return nil
}

func (p *Payment) LinkInstallment(installmentID uuid.UUID) {
	p.InstallmentID = &installmentID
}
