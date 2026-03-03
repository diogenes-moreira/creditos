package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MovementType string

const (
	MovementTypeCredit MovementType = "credit"
	MovementTypeDebit  MovementType = "debit"
)

type Movement struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey"`
	AccountID    uuid.UUID       `gorm:"type:uuid;index;not null"`
	Type         MovementType    `gorm:"type:varchar(10);not null"`
	Amount       decimal.Decimal `gorm:"type:decimal(18,2);not null"`
	BalanceAfter decimal.Decimal `gorm:"type:decimal(18,2);not null"`
	Description  string          `gorm:"not null"`
	Reference    string          `gorm:""`
	CreatedAt    time.Time       `gorm:"not null"`
}

func NewMovement(accountID uuid.UUID, mType MovementType, amount, balanceAfter decimal.Decimal, description, reference string) *Movement {
	return &Movement{
		ID:           uuid.New(),
		AccountID:    accountID,
		Type:         mType,
		Amount:       amount,
		BalanceAfter: balanceAfter,
		Description:  description,
		Reference:    reference,
		CreatedAt:    time.Now(),
	}
}
