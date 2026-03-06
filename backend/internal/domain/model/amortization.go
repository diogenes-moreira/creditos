package model

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

type AmortizationType string

const (
	AmortizationFrench AmortizationType = "french"
	AmortizationGerman AmortizationType = "german"
)

type AmortizationSchedule struct {
	Installments  []InstallmentCalc
	TotalInterest decimal.Decimal
	TotalIVA      decimal.Decimal
	TotalPayment  decimal.Decimal
}

type InstallmentCalc struct {
	Number    int
	DueDate   time.Time
	Capital   decimal.Decimal
	Interest  decimal.Decimal
	IVA       decimal.Decimal
	Total     decimal.Decimal
	Remaining decimal.Decimal
}

// CalculateFrenchAmortization computes a fixed-payment (French) amortization schedule.
// Monthly interest rate = annual rate / 12.
// ivaRate is the IVA percentage applied on interest (e.g. 21 for 21%). Use zero to skip IVA.
func CalculateFrenchAmortization(principal, annualRate decimal.Decimal, installments int, startDate time.Time, ivaRate decimal.Decimal) AmortizationSchedule {
	monthlyRate := annualRate.Div(decimal.NewFromInt(12))
	hundred := decimal.NewFromInt(100)
	schedule := AmortizationSchedule{}
	remaining := principal

	if monthlyRate.IsZero() {
		capitalPerMonth := principal.Div(decimal.NewFromInt(int64(installments))).Round(2)
		for i := 1; i <= installments; i++ {
			capital := capitalPerMonth
			if i == installments {
				capital = remaining
			}
			remaining = remaining.Sub(capital)
			dueDate := startDate.AddDate(0, i, 0)
			schedule.Installments = append(schedule.Installments, InstallmentCalc{
				Number:    i,
				DueDate:   dueDate,
				Capital:   capital,
				Interest:  decimal.NewFromInt(0),
				IVA:       decimal.NewFromInt(0),
				Total:     capital,
				Remaining: remaining,
			})
		}
		schedule.TotalInterest = decimal.NewFromInt(0)
		schedule.TotalIVA = decimal.NewFromInt(0)
		schedule.TotalPayment = principal
		return schedule
	}

	// Fixed monthly payment: P * r * (1+r)^n / ((1+r)^n - 1)
	r, _ := monthlyRate.Float64()
	n := float64(installments)
	pow := math.Pow(1+r, n)
	pmt := principal.Mul(monthlyRate).Mul(decimal.NewFromFloat(pow)).Div(decimal.NewFromFloat(pow - 1))
	pmt = pmt.Round(2)

	totalInterest := decimal.NewFromInt(0)
	totalIVA := decimal.NewFromInt(0)
	for i := 1; i <= installments; i++ {
		interest := remaining.Mul(monthlyRate).Round(2)
		capital := pmt.Sub(interest)
		if i == installments {
			capital = remaining
			pmt = capital.Add(interest)
		}
		remaining = remaining.Sub(capital)
		if remaining.IsNegative() {
			remaining = decimal.NewFromInt(0)
		}
		iva := interest.Mul(ivaRate).Div(hundred).Round(2)
		totalInterest = totalInterest.Add(interest)
		totalIVA = totalIVA.Add(iva)
		dueDate := startDate.AddDate(0, i, 0)

		schedule.Installments = append(schedule.Installments, InstallmentCalc{
			Number:    i,
			DueDate:   dueDate,
			Capital:   capital.Round(2),
			Interest:  interest,
			IVA:       iva,
			Total:     capital.Add(interest).Add(iva).Round(2),
			Remaining: remaining.Round(2),
		})
	}

	schedule.TotalInterest = totalInterest.Round(2)
	schedule.TotalIVA = totalIVA.Round(2)
	schedule.TotalPayment = principal.Add(totalInterest).Add(totalIVA).Round(2)
	return schedule
}

// CalculateGermanAmortization computes a fixed-capital (German) amortization schedule.
// Capital portion is equal each month, interest decreases as remaining principal decreases.
// ivaRate is the IVA percentage applied on interest (e.g. 21 for 21%). Use zero to skip IVA.
func CalculateGermanAmortization(principal, annualRate decimal.Decimal, installments int, startDate time.Time, ivaRate decimal.Decimal) AmortizationSchedule {
	monthlyRate := annualRate.Div(decimal.NewFromInt(12))
	hundred := decimal.NewFromInt(100)
	capitalPerMonth := principal.Div(decimal.NewFromInt(int64(installments))).Round(2)
	remaining := principal
	schedule := AmortizationSchedule{}
	totalInterest := decimal.NewFromInt(0)
	totalIVA := decimal.NewFromInt(0)

	for i := 1; i <= installments; i++ {
		interest := remaining.Mul(monthlyRate).Round(2)
		capital := capitalPerMonth
		if i == installments {
			capital = remaining
		}
		remaining = remaining.Sub(capital)
		if remaining.IsNegative() {
			remaining = decimal.NewFromInt(0)
		}
		iva := interest.Mul(ivaRate).Div(hundred).Round(2)
		total := capital.Add(interest).Add(iva).Round(2)
		totalInterest = totalInterest.Add(interest)
		totalIVA = totalIVA.Add(iva)
		dueDate := startDate.AddDate(0, i, 0)

		schedule.Installments = append(schedule.Installments, InstallmentCalc{
			Number:    i,
			DueDate:   dueDate,
			Capital:   capital.Round(2),
			Interest:  interest,
			IVA:       iva,
			Total:     total,
			Remaining: remaining.Round(2),
		})
	}

	schedule.TotalInterest = totalInterest.Round(2)
	schedule.TotalIVA = totalIVA.Round(2)
	schedule.TotalPayment = principal.Add(totalInterest).Add(totalIVA).Round(2)
	return schedule
}

// CalculateEarlyCancellation computes the remaining balance for early cancellation.
// Returns the sum of remaining capital of all unpaid installments.
func CalculateEarlyCancellation(unpaidInstallments []InstallmentCalc) decimal.Decimal {
	total := decimal.NewFromInt(0)
	for _, inst := range unpaidInstallments {
		total = total.Add(inst.Capital)
	}
	return total.Round(2)
}
