package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CurrentAccount struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ClientID  uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null"`
	Client    Client         `gorm:"foreignKey:ClientID"`
	Balance   decimal.Decimal `gorm:"type:decimal(18,2);not null;default:0"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Movements []Movement     `gorm:"foreignKey:AccountID"`
}

func NewCurrentAccount(clientID uuid.UUID) *CurrentAccount {
	return &CurrentAccount{
		ID:       uuid.New(),
		ClientID: clientID,
		Balance:  decimal.NewFromInt(0),
	}
}

func (a *CurrentAccount) Credit(amount decimal.Decimal, description, reference string) (*Movement, error) {
	if !amount.IsPositive() {
		return nil, fmt.Errorf("credit amount must be positive")
	}

	a.Balance = a.Balance.Add(amount)

	return NewMovement(a.ID, MovementTypeCredit, amount, a.Balance, description, reference), nil
}

func (a *CurrentAccount) Debit(amount decimal.Decimal, description, reference string) (*Movement, error) {
	if !amount.IsPositive() {
		return nil, fmt.Errorf("debit amount must be positive")
	}

	a.Balance = a.Balance.Sub(amount)

	return NewMovement(a.ID, MovementTypeDebit, amount, a.Balance, description, reference), nil
}
