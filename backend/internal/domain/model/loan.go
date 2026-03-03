package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LoanStatus string

const (
	LoanQuoted    LoanStatus = "quoted"
	LoanPending   LoanStatus = "pending"
	LoanApproved  LoanStatus = "approved"
	LoanActive    LoanStatus = "active"
	LoanCompleted LoanStatus = "completed"
	LoanCancelled LoanStatus = "cancelled"
	LoanDefaulted LoanStatus = "defaulted"
)

type Loan struct {
	ID               uuid.UUID        `gorm:"type:uuid;primaryKey"`
	ClientID         uuid.UUID        `gorm:"type:uuid;index;not null"`
	Client           Client           `gorm:"foreignKey:ClientID"`
	CreditLineID     uuid.UUID        `gorm:"type:uuid;index;not null"`
	CreditLine       CreditLine       `gorm:"foreignKey:CreditLineID"`
	Principal        decimal.Decimal  `gorm:"type:decimal(18,2);not null"`
	InterestRate     decimal.Decimal  `gorm:"type:decimal(8,4);not null"`
	NumInstallments  int              `gorm:"not null"`
	AmortizationType AmortizationType `gorm:"type:varchar(20);not null"`
	Status           LoanStatus       `gorm:"type:varchar(20);not null;default:'quoted'"`
	DisbursedAt      *time.Time       `gorm:""`
	ApprovedBy       *uuid.UUID       `gorm:"type:uuid"`
	ApprovedAt       *time.Time       `gorm:""`
	CompletedAt      *time.Time       `gorm:""`
	CancelledAt      *time.Time       `gorm:""`
	Installments     []Installment    `gorm:"foreignKey:LoanID"`
	CreatedAt        time.Time        `gorm:"not null"`
	UpdatedAt        time.Time        `gorm:"not null"`
	DeletedAt        gorm.DeletedAt   `gorm:"index"`
}

func NewLoan(clientID, creditLineID uuid.UUID, principal, interestRate decimal.Decimal, numInstallments int, amortType AmortizationType) (*Loan, error) {
	if !principal.IsPositive() {
		return nil, fmt.Errorf("principal must be positive")
	}
	if numInstallments < 1 {
		return nil, fmt.Errorf("number of installments must be at least 1")
	}
	if amortType != AmortizationFrench && amortType != AmortizationGerman {
		return nil, fmt.Errorf("invalid amortization type: %s", amortType)
	}

	return &Loan{
		ID:               uuid.New(),
		ClientID:         clientID,
		CreditLineID:     creditLineID,
		Principal:        principal,
		InterestRate:     interestRate,
		NumInstallments:  numInstallments,
		AmortizationType: amortType,
		Status:           LoanQuoted,
	}, nil
}

func (l *Loan) RequestApproval() error {
	if l.Status != LoanQuoted {
		return fmt.Errorf("can only request approval for quoted loans, current status: %s", l.Status)
	}
	l.Status = LoanPending
	return nil
}

func (l *Loan) Approve(approvedBy uuid.UUID) error {
	if l.Status != LoanPending {
		return fmt.Errorf("can only approve pending loans, current status: %s", l.Status)
	}
	l.Status = LoanApproved
	l.ApprovedBy = &approvedBy
	now := time.Now()
	l.ApprovedAt = &now
	return nil
}

func (l *Loan) Disburse(startDate time.Time) ([]Installment, error) {
	if l.Status != LoanApproved {
		return nil, fmt.Errorf("can only disburse approved loans, current status: %s", l.Status)
	}

	var schedule AmortizationSchedule
	switch l.AmortizationType {
	case AmortizationFrench:
		schedule = CalculateFrenchAmortization(l.Principal, l.InterestRate, l.NumInstallments, startDate)
	case AmortizationGerman:
		schedule = CalculateGermanAmortization(l.Principal, l.InterestRate, l.NumInstallments, startDate)
	}

	installments := make([]Installment, len(schedule.Installments))
	for i, calc := range schedule.Installments {
		installments[i] = Installment{
			ID:              uuid.New(),
			LoanID:          l.ID,
			Number:          calc.Number,
			DueDate:         calc.DueDate,
			CapitalAmount:   calc.Capital,
			InterestAmount:  calc.Interest,
			TotalAmount:     calc.Total,
			PaidAmount:      decimal.NewFromInt(0),
			RemainingAmount: calc.Total,
			Status:          InstallmentPending,
		}
	}

	l.Status = LoanActive
	now := time.Now()
	l.DisbursedAt = &now
	l.Installments = installments

	return installments, nil
}

func (l *Loan) Cancel() error {
	if l.Status != LoanActive {
		return fmt.Errorf("can only cancel active loans, current status: %s", l.Status)
	}
	l.Status = LoanCancelled
	now := time.Now()
	l.CancelledAt = &now
	return nil
}

func (l *Loan) Complete() error {
	if l.Status != LoanActive {
		return fmt.Errorf("can only complete active loans, current status: %s", l.Status)
	}
	l.Status = LoanCompleted
	now := time.Now()
	l.CompletedAt = &now
	return nil
}

func (l *Loan) MarkDefaulted() error {
	if l.Status != LoanActive {
		return fmt.Errorf("can only default active loans, current status: %s", l.Status)
	}
	l.Status = LoanDefaulted
	return nil
}

func (l *Loan) CheckCompletion() bool {
	for _, inst := range l.Installments {
		if inst.Status != InstallmentPaid {
			return false
		}
	}
	return true
}

func (l *Loan) OverdueInstallments() []Installment {
	var overdue []Installment
	now := time.Now()
	for _, inst := range l.Installments {
		if inst.Status == InstallmentPending && inst.DueDate.Before(now) {
			overdue = append(overdue, inst)
		}
	}
	return overdue
}

func (l *Loan) TotalPaid() decimal.Decimal {
	total := decimal.NewFromInt(0)
	for _, inst := range l.Installments {
		total = total.Add(inst.PaidAmount)
	}
	return total
}

func (l *Loan) TotalRemaining() decimal.Decimal {
	total := decimal.NewFromInt(0)
	for _, inst := range l.Installments {
		total = total.Add(inst.RemainingAmount)
	}
	return total
}
