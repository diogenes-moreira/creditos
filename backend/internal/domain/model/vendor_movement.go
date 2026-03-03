package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type VendorMovement struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey"`
	VendorAccountID uuid.UUID       `gorm:"type:uuid;index;not null"`
	Type            MovementType    `gorm:"type:varchar(10);not null"`
	Amount          decimal.Decimal `gorm:"type:decimal(18,2);not null"`
	BalanceAfter    decimal.Decimal `gorm:"type:decimal(18,2);not null"`
	Description     string          `gorm:"not null"`
	Reference       string          `gorm:""`
	CreatedAt       time.Time       `gorm:"not null"`
}

func NewVendorMovement(accountID uuid.UUID, mType MovementType, amount, balanceAfter decimal.Decimal, description, reference string) *VendorMovement {
	return &VendorMovement{
		ID:              uuid.New(),
		VendorAccountID: accountID,
		Type:            mType,
		Amount:          amount,
		BalanceAfter:    balanceAfter,
		Description:     description,
		Reference:       reference,
		CreatedAt:       time.Now(),
	}
}
