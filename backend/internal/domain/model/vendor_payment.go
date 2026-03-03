package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type VendorPaymentMethod string

const (
	VendorPaymentCash     VendorPaymentMethod = "cash"
	VendorPaymentTransfer VendorPaymentMethod = "transfer"
)

type VendorPayment struct {
	ID        uuid.UUID           `gorm:"type:uuid;primaryKey"`
	VendorID  uuid.UUID           `gorm:"type:uuid;index;not null"`
	Vendor    Vendor              `gorm:"foreignKey:VendorID"`
	Amount    decimal.Decimal     `gorm:"type:decimal(18,2);not null"`
	Method    VendorPaymentMethod `gorm:"type:varchar(20);not null"`
	Reference string              `gorm:""`
	PaidBy    uuid.UUID           `gorm:"type:uuid;not null"`
	CreatedAt time.Time           `gorm:"not null"`
}

func NewVendorPayment(vendorID uuid.UUID, amount decimal.Decimal, method VendorPaymentMethod, reference string, paidBy uuid.UUID) (*VendorPayment, error) {
	if !amount.IsPositive() {
		return nil, fmt.Errorf("payment amount must be positive")
	}
	if method != VendorPaymentCash && method != VendorPaymentTransfer {
		return nil, fmt.Errorf("invalid payment method: %s", method)
	}

	return &VendorPayment{
		ID:        uuid.New(),
		VendorID:  vendorID,
		Amount:    amount,
		Method:    method,
		Reference: reference,
		PaidBy:    paidBy,
		CreatedAt: time.Now(),
	}, nil
}
