package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type VendorAccount struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey"`
	VendorID  uuid.UUID        `gorm:"type:uuid;uniqueIndex;not null"`
	Vendor    Vendor           `gorm:"foreignKey:VendorID"`
	Balance   decimal.Decimal  `gorm:"type:decimal(18,2);not null;default:0"`
	CreatedAt time.Time        `gorm:"not null"`
	UpdatedAt time.Time        `gorm:"not null"`
	DeletedAt gorm.DeletedAt   `gorm:"index"`
	Movements []VendorMovement `gorm:"foreignKey:VendorAccountID"`
}

func NewVendorAccount(vendorID uuid.UUID) *VendorAccount {
	return &VendorAccount{
		ID:       uuid.New(),
		VendorID: vendorID,
		Balance:  decimal.NewFromInt(0),
	}
}

func (a *VendorAccount) Credit(amount decimal.Decimal, description, reference string) (*VendorMovement, error) {
	if !amount.IsPositive() {
		return nil, fmt.Errorf("credit amount must be positive")
	}

	a.Balance = a.Balance.Add(amount)

	return NewVendorMovement(a.ID, MovementTypeCredit, amount, a.Balance, description, reference), nil
}

func (a *VendorAccount) Debit(amount decimal.Decimal, description, reference string) (*VendorMovement, error) {
	if !amount.IsPositive() {
		return nil, fmt.Errorf("debit amount must be positive")
	}

	a.Balance = a.Balance.Sub(amount)

	return NewVendorMovement(a.ID, MovementTypeDebit, amount, a.Balance, description, reference), nil
}
