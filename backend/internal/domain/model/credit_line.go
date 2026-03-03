package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CreditLineStatus string

const (
	CreditLinePending  CreditLineStatus = "pending"
	CreditLineApproved CreditLineStatus = "approved"
	CreditLineRejected CreditLineStatus = "rejected"
)

type CreditLine struct {
	ID              uuid.UUID        `gorm:"type:uuid;primaryKey"`
	ClientID        uuid.UUID        `gorm:"type:uuid;index;not null"`
	Client          Client           `gorm:"foreignKey:ClientID"`
	MaxAmount       decimal.Decimal  `gorm:"type:decimal(18,2);not null"`
	UsedAmount      decimal.Decimal  `gorm:"type:decimal(18,2);not null;default:0"`
	InterestRate    decimal.Decimal  `gorm:"type:decimal(8,4);not null"`
	MaxInstallments int              `gorm:"not null"`
	Status          CreditLineStatus `gorm:"type:varchar(20);not null;default:'pending'"`
	ApprovedBy      *uuid.UUID       `gorm:"type:uuid"`
	ApprovedAt      *time.Time       `gorm:""`
	RejectedBy      *uuid.UUID       `gorm:"type:uuid"`
	RejectedAt      *time.Time       `gorm:""`
	RejectionReason string           `gorm:""`
	CreatedAt       time.Time        `gorm:"not null"`
	UpdatedAt       time.Time        `gorm:"not null"`
	DeletedAt       gorm.DeletedAt   `gorm:"index"`
}

func NewCreditLine(clientID uuid.UUID, maxAmount, interestRate decimal.Decimal, maxInstallments int) (*CreditLine, error) {
	if !maxAmount.IsPositive() {
		return nil, fmt.Errorf("max amount must be positive")
	}
	if interestRate.IsNegative() {
		return nil, fmt.Errorf("interest rate cannot be negative")
	}
	if maxInstallments < 1 || maxInstallments > 60 {
		return nil, fmt.Errorf("max installments must be between 1 and 60")
	}

	return &CreditLine{
		ID:              uuid.New(),
		ClientID:        clientID,
		MaxAmount:       maxAmount,
		UsedAmount:      decimal.NewFromInt(0),
		InterestRate:    interestRate,
		MaxInstallments: maxInstallments,
		Status:          CreditLinePending,
	}, nil
}

func (cl *CreditLine) Approve(approvedBy uuid.UUID) error {
	if cl.Status != CreditLinePending {
		return fmt.Errorf("can only approve pending credit lines, current status: %s", cl.Status)
	}
	cl.Status = CreditLineApproved
	cl.ApprovedBy = &approvedBy
	now := time.Now()
	cl.ApprovedAt = &now
	return nil
}

func (cl *CreditLine) Reject(rejectedBy uuid.UUID, reason string) error {
	if cl.Status != CreditLinePending {
		return fmt.Errorf("can only reject pending credit lines, current status: %s", cl.Status)
	}
	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}
	cl.Status = CreditLineRejected
	cl.RejectedBy = &rejectedBy
	now := time.Now()
	cl.RejectedAt = &now
	cl.RejectionReason = reason
	return nil
}

func (cl *CreditLine) AvailableAmount() decimal.Decimal {
	return cl.MaxAmount.Sub(cl.UsedAmount)
}

func (cl *CreditLine) CanDisburse(amount decimal.Decimal) error {
	if cl.Status != CreditLineApproved {
		return fmt.Errorf("credit line is not approved")
	}
	if amount.GreaterThan(cl.AvailableAmount()) {
		return fmt.Errorf("requested amount %s exceeds available %s", amount.StringFixed(2), cl.AvailableAmount().StringFixed(2))
	}
	return nil
}

func (cl *CreditLine) RecordDisbursement(amount decimal.Decimal) {
	cl.UsedAmount = cl.UsedAmount.Add(amount)
}

func (cl *CreditLine) ReleaseDisbursement(amount decimal.Decimal) {
	cl.UsedAmount = cl.UsedAmount.Sub(amount)
	if cl.UsedAmount.IsNegative() {
		cl.UsedAmount = decimal.NewFromInt(0)
	}
}
