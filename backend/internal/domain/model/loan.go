package model

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PrepaymentStrategy string

const (
	PrepaymentReduceInstallment PrepaymentStrategy = "reduce_installment"
	PrepaymentReduceTerm        PrepaymentStrategy = "reduce_term"
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

func (l *Loan) Disburse(startDate time.Time, ivaRate decimal.Decimal) ([]Installment, error) {
	if l.Status != LoanApproved {
		return nil, fmt.Errorf("can only disburse approved loans, current status: %s", l.Status)
	}

	var schedule AmortizationSchedule
	switch l.AmortizationType {
	case AmortizationFrench:
		schedule = CalculateFrenchAmortization(l.Principal, l.InterestRate, l.NumInstallments, startDate, ivaRate)
	case AmortizationGerman:
		schedule = CalculateGermanAmortization(l.Principal, l.InterestRate, l.NumInstallments, startDate, ivaRate)
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
			IVAAmount:       calc.IVA,
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

// RecalculateRemainingInstallments recalculates interest for unpaid installments
// based on the given outstanding principal after a capital prepayment.
func (l *Loan) RecalculateRemainingInstallments(outstandingPrincipal, interestRate, ivaRate decimal.Decimal) {
	var pendingIndices []int
	for i, inst := range l.Installments {
		if inst.Status != InstallmentPaid {
			pendingIndices = append(pendingIndices, i)
		}
	}

	if len(pendingIndices) == 0 || outstandingPrincipal.IsZero() {
		return
	}

	numRemaining := len(pendingIndices)
	// Collect existing due dates
	dueDates := make([]time.Time, numRemaining)
	for j, idx := range pendingIndices {
		dueDates[j] = l.Installments[idx].DueDate
	}

	// Recalculate using the same amortization type
	var schedule AmortizationSchedule
	// Use the first pending due date minus 1 month as start date for calculation
	startDate := dueDates[0].AddDate(0, -1, 0)
	switch l.AmortizationType {
	case AmortizationFrench:
		schedule = CalculateFrenchAmortization(outstandingPrincipal, interestRate, numRemaining, startDate, ivaRate)
	case AmortizationGerman:
		schedule = CalculateGermanAmortization(outstandingPrincipal, interestRate, numRemaining, startDate, ivaRate)
	}

	// Update each pending installment with recalculated amounts
	for j, idx := range pendingIndices {
		calc := schedule.Installments[j]
		inst := &l.Installments[idx]
		inst.CapitalAmount = calc.Capital
		inst.InterestAmount = calc.Interest
		inst.IVAAmount = calc.IVA
		inst.TotalAmount = calc.Total
		inst.RemainingAmount = calc.Total
		inst.PaidAmount = decimal.NewFromInt(0)
	}
}

// RecalculateReducingTerm keeps the original installment amount and reduces the
// number of remaining installments after a prepayment. Returns the UUIDs of
// installments that were removed from the loan so the caller can delete them.
func (l *Loan) RecalculateReducingTerm(outstandingPrincipal, interestRate, ivaRate decimal.Decimal) []uuid.UUID {
	var pendingIndices []int
	for i, inst := range l.Installments {
		if inst.Status != InstallmentPaid {
			pendingIndices = append(pendingIndices, i)
		}
	}

	if len(pendingIndices) == 0 || outstandingPrincipal.IsZero() {
		return nil
	}

	pendingCount := len(pendingIndices)
	monthlyRate := interestRate.Div(decimal.NewFromInt(12))
	var newN int

	switch l.AmortizationType {
	case AmortizationFrench:
		if monthlyRate.IsZero() {
			originalCapital := l.Principal.Div(decimal.NewFromInt(int64(l.NumInstallments)))
			newNDec := outstandingPrincipal.Div(originalCapital)
			newN = int(math.Ceil(newNDec.InexactFloat64()))
		} else {
			// Original PMT = P * r * (1+r)^n / ((1+r)^n - 1)
			r, _ := monthlyRate.Float64()
			n := float64(l.NumInstallments)
			pow := math.Pow(1+r, n)
			pmt := l.Principal.Mul(monthlyRate).Mul(decimal.NewFromFloat(pow)).Div(decimal.NewFromFloat(pow - 1))

			// newN = ceil(-log(1 - outstanding*r/PMT) / log(1+r))
			oR, _ := outstandingPrincipal.Mul(monthlyRate).Div(pmt).Float64()
			inner := 1 - oR
			if inner <= 0 {
				return nil // cannot reduce term
			}
			newNFloat := math.Ceil(-math.Log(inner) / math.Log(1+r))
			newN = int(newNFloat)
		}
	case AmortizationGerman:
		originalCapital := l.Principal.Div(decimal.NewFromInt(int64(l.NumInstallments)))
		newNDec := outstandingPrincipal.Div(originalCapital)
		newN = int(math.Ceil(newNDec.InexactFloat64()))
	}

	if newN < 1 {
		newN = 1
	}
	if newN >= pendingCount {
		return nil // no reduction possible
	}

	// Collect due dates for the installments we'll keep
	dueDates := make([]time.Time, pendingCount)
	for j, idx := range pendingIndices {
		dueDates[j] = l.Installments[idx].DueDate
	}

	// Recalculate the first newN pending installments
	startDate := dueDates[0].AddDate(0, -1, 0)
	var schedule AmortizationSchedule
	switch l.AmortizationType {
	case AmortizationFrench:
		schedule = CalculateFrenchAmortization(outstandingPrincipal, interestRate, newN, startDate, ivaRate)
	case AmortizationGerman:
		schedule = CalculateGermanAmortization(outstandingPrincipal, interestRate, newN, startDate, ivaRate)
	}

	// Update the first newN pending installments
	for j := 0; j < newN; j++ {
		idx := pendingIndices[j]
		calc := schedule.Installments[j]
		inst := &l.Installments[idx]
		inst.CapitalAmount = calc.Capital
		inst.InterestAmount = calc.Interest
		inst.IVAAmount = calc.IVA
		inst.TotalAmount = calc.Total
		inst.RemainingAmount = calc.Total
		inst.PaidAmount = decimal.NewFromInt(0)
	}

	// Collect IDs of excess installments to delete
	var removedIDs []uuid.UUID
	for j := newN; j < pendingCount; j++ {
		idx := pendingIndices[j]
		removedIDs = append(removedIDs, l.Installments[idx].ID)
	}

	// Remove excess installments from the slice
	removeSet := make(map[uuid.UUID]bool, len(removedIDs))
	for _, rid := range removedIDs {
		removeSet[rid] = true
	}
	kept := make([]Installment, 0, len(l.Installments)-len(removedIDs))
	for _, inst := range l.Installments {
		if !removeSet[inst.ID] {
			kept = append(kept, inst)
		}
	}
	l.Installments = kept

	return removedIDs
}

// OutstandingPrincipal returns the sum of capital amounts for all unpaid installments.
func (l *Loan) OutstandingPrincipal() decimal.Decimal {
	total := decimal.Zero
	for _, inst := range l.Installments {
		if inst.Status != InstallmentPaid {
			total = total.Add(inst.CapitalAmount)
		}
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

// CancellationSettlement computes the early cancellation settlement breakdown.
// Formula: outstanding capital + accrued interest from disbursement to now + IVA on that interest.
// Accrued interest = outstanding capital * monthly rate * months elapsed (pro-rata by days).
func (l *Loan) CancellationSettlement(ivaRate decimal.Decimal) (pendingCapital, accruedInterest, accruedIVA, total decimal.Decimal) {
	pendingCapital = decimal.Zero
	accruedInterest = decimal.Zero
	accruedIVA = decimal.Zero

	for _, inst := range l.Installments {
		if inst.Status != InstallmentPaid {
			pendingCapital = pendingCapital.Add(inst.CapitalAmount)
		}
	}

	// Calculate months elapsed from disbursement (pro-rata by days)
	if l.DisbursedAt != nil && pendingCapital.IsPositive() {
		daysElapsed := decimal.NewFromFloat(time.Since(*l.DisbursedAt).Hours() / 24)
		monthsElapsed := daysElapsed.Div(decimal.NewFromInt(30))
		monthlyRate := l.InterestRate.Div(decimal.NewFromInt(100))
		accruedInterest = pendingCapital.Mul(monthlyRate).Mul(monthsElapsed).Round(2)
		accruedIVA = accruedInterest.Mul(ivaRate).Div(decimal.NewFromInt(100)).Round(2)
	}

	total = pendingCapital.Add(accruedInterest).Add(accruedIVA)
	return
}
