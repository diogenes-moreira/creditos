package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InstallmentStatus string

const (
	InstallmentPending  InstallmentStatus = "pending"
	InstallmentPartial  InstallmentStatus = "partial"
	InstallmentPaid     InstallmentStatus = "paid"
	InstallmentOverdue  InstallmentStatus = "overdue"
)

type Installment struct {
	ID              uuid.UUID         `gorm:"type:uuid;primaryKey"`
	LoanID          uuid.UUID         `gorm:"type:uuid;index;not null"`
	Number          int               `gorm:"not null"`
	DueDate         time.Time         `gorm:"not null"`
	CapitalAmount   decimal.Decimal   `gorm:"type:decimal(18,2);not null"`
	InterestAmount  decimal.Decimal   `gorm:"type:decimal(18,2);not null"`
	IVAAmount       decimal.Decimal   `gorm:"type:decimal(18,2);not null;default:0"`
	TotalAmount     decimal.Decimal   `gorm:"type:decimal(18,2);not null"`
	PaidAmount      decimal.Decimal   `gorm:"type:decimal(18,2);not null;default:0"`
	RemainingAmount decimal.Decimal   `gorm:"type:decimal(18,2);not null"`
	PenaltyApplied  bool              `gorm:"not null;default:false"`
	Status          InstallmentStatus `gorm:"type:varchar(20);not null;default:'pending'"`
	PaidAt          *time.Time        `gorm:""`
	CreatedAt       time.Time         `gorm:"not null"`
	UpdatedAt       time.Time         `gorm:"not null"`
}

// ApplyPayment applies a payment amount to this installment.
// Returns the amount actually applied and any surplus.
func (i *Installment) ApplyPayment(amount decimal.Decimal) (applied, surplus decimal.Decimal, err error) {
	if !amount.IsPositive() {
		return decimal.Zero, decimal.Zero, fmt.Errorf("payment amount must be positive")
	}
	if i.Status == InstallmentPaid {
		return decimal.Zero, amount, nil
	}

	if amount.GreaterThanOrEqual(i.RemainingAmount) {
		applied = i.RemainingAmount
		surplus = amount.Sub(i.RemainingAmount)
		i.PaidAmount = i.TotalAmount
		i.RemainingAmount = decimal.NewFromInt(0)
		i.Status = InstallmentPaid
		now := time.Now()
		i.PaidAt = &now
	} else {
		applied = amount
		surplus = decimal.NewFromInt(0)
		i.PaidAmount = i.PaidAmount.Add(amount)
		i.RemainingAmount = i.TotalAmount.Sub(i.PaidAmount)
		i.Status = InstallmentPartial
	}

	return applied, surplus, nil
}

func (i *Installment) IsOverdue() bool {
	return i.Status != InstallmentPaid && time.Now().After(i.DueDate)
}

func (i *Installment) MarkOverdue() {
	if i.Status != InstallmentPaid {
		i.Status = InstallmentOverdue
	}
}

func (i *Installment) DaysOverdue() int {
	if !i.IsOverdue() {
		return 0
	}
	return int(time.Since(i.DueDate).Hours() / 24)
}

// ApplyLatePenalty recalculates interest with a penalty surcharge for overdue installments.
// penaltyRate is a percentage (e.g. 10 means 10%).
// Only applies once per installment (guarded by PenaltyApplied flag).
func (i *Installment) ApplyLatePenalty(penaltyRate decimal.Decimal) {
	if i.Status == InstallmentPaid || i.PenaltyApplied || !i.IsOverdue() {
		return
	}

	multiplier := decimal.NewFromInt(1).Add(penaltyRate.Div(decimal.NewFromInt(100)))
	newInterest := i.InterestAmount.Mul(multiplier).Round(2)

	// Preserve IVA proportion: if original interest > 0, scale IVA the same way
	newIVA := i.IVAAmount
	if i.InterestAmount.IsPositive() {
		newIVA = i.IVAAmount.Mul(newInterest).Div(i.InterestAmount).Round(2)
	}

	i.InterestAmount = newInterest
	i.IVAAmount = newIVA
	i.TotalAmount = i.CapitalAmount.Add(newInterest).Add(newIVA)
	i.RemainingAmount = i.TotalAmount.Sub(i.PaidAmount)
	i.PenaltyApplied = true
}
