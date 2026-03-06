package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Purchase struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey"`
	VendorID     uuid.UUID       `gorm:"type:uuid;index;not null"`
	Vendor       Vendor          `gorm:"foreignKey:VendorID"`
	ClientID     uuid.UUID       `gorm:"type:uuid;index;not null"`
	Client       Client          `gorm:"foreignKey:ClientID"`
	CreditLineID uuid.UUID       `gorm:"type:uuid;not null"`
	CreditLine   CreditLine      `gorm:"foreignKey:CreditLineID"`
	LoanID       uuid.UUID       `gorm:"type:uuid;not null"`
	Loan         Loan            `gorm:"foreignKey:LoanID"`
	Amount       decimal.Decimal `gorm:"type:decimal(18,2);not null"`
	Description  string          `gorm:"not null"`
	CreatedAt    time.Time       `gorm:"not null"`
}

func NewPurchase(vendorID, clientID, creditLineID, loanID uuid.UUID, amount decimal.Decimal, description string) (*Purchase, error) {
	if !amount.IsPositive() {
		return nil, fmt.Errorf("purchase amount must be positive")
	}
	if description == "" {
		return nil, fmt.Errorf("purchase description is required")
	}

	return &Purchase{
		ID:           uuid.New(),
		VendorID:     vendorID,
		ClientID:     clientID,
		CreditLineID: creditLineID,
		LoanID:       loanID,
		Amount:       amount,
		Description:  description,
		CreatedAt:    time.Now(),
	}, nil
}
