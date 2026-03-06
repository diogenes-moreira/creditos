package model_test

import (
	"testing"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateFrenchAmortization(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("basic 12 installments at 24 percent annual", func(t *testing.T) {
		principal := decimal.NewFromInt(120000)
		annualRate := decimal.NewFromFloat(0.24)
		installments := 12

		schedule := model.CalculateFrenchAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

		assert.Equal(t, installments, len(schedule.Installments))
		assert.True(t, schedule.TotalInterest.IsPositive(), "total interest should be positive")
		assert.True(t, schedule.TotalPayment.GreaterThan(principal), "total payment should exceed principal")

		// French: all installments should have approximately the same base payment (capital + interest)
		// Note: Total includes IVA which varies proportionally to interest, so we check the base payment
		firstBase := schedule.Installments[0].Capital.Add(schedule.Installments[0].Interest)
		for i := 0; i < len(schedule.Installments)-1; i++ {
			base := schedule.Installments[i].Capital.Add(schedule.Installments[i].Interest)
			diff := base.Sub(firstBase).Abs()
			assert.True(t, diff.LessThanOrEqual(decimal.NewFromFloat(0.02)),
				"installment %d base payment %s differs from first %s", i+1,
				base.StringFixed(2), firstBase.StringFixed(2))
		}

		// Verify IVA is computed on each installment
		for _, inst := range schedule.Installments {
			expectedIVA := inst.Interest.Mul(decimal.NewFromInt(21)).Div(decimal.NewFromInt(100)).Round(2)
			assert.True(t, inst.IVA.Equal(expectedIVA),
				"installment %d IVA should be %s, got %s", inst.Number, expectedIVA.StringFixed(2), inst.IVA.StringFixed(2))
		}

		// Last installment remaining should be zero
		lastInst := schedule.Installments[len(schedule.Installments)-1]
		assert.True(t, lastInst.Remaining.Equal(decimal.NewFromInt(0)),
			"last installment remaining should be 0, got %s", lastInst.Remaining.StringFixed(2))

		// Verify due dates are monthly
		for i, inst := range schedule.Installments {
			expectedDate := startDate.AddDate(0, i+1, 0)
			assert.Equal(t, expectedDate, inst.DueDate, "installment %d due date mismatch", i+1)
		}

		// Verify installment numbers
		for i, inst := range schedule.Installments {
			assert.Equal(t, i+1, inst.Number)
		}
	})

	t.Run("zero interest rate", func(t *testing.T) {
		principal := decimal.NewFromInt(12000)
		annualRate := decimal.NewFromInt(0)
		installments := 12

		schedule := model.CalculateFrenchAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

		assert.Equal(t, installments, len(schedule.Installments))
		assert.True(t, schedule.TotalInterest.Equal(decimal.NewFromInt(0)),
			"total interest should be zero, got %s", schedule.TotalInterest.StringFixed(2))
		assert.True(t, schedule.TotalPayment.Equal(principal),
			"total payment should equal principal")

		// Each installment should be 1000
		for _, inst := range schedule.Installments {
			assert.True(t, inst.Interest.Equal(decimal.NewFromInt(0)))
		}

		// Sum of capital should equal principal
		totalCapital := decimal.NewFromInt(0)
		for _, inst := range schedule.Installments {
			totalCapital = totalCapital.Add(inst.Capital)
		}
		assert.True(t, totalCapital.Equal(principal),
			"sum of capital %s should equal principal %s", totalCapital.StringFixed(2), principal.StringFixed(2))
	})

	t.Run("single installment", func(t *testing.T) {
		principal := decimal.NewFromInt(10000)
		annualRate := decimal.NewFromFloat(0.12)
		installments := 1

		schedule := model.CalculateFrenchAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

		assert.Equal(t, 1, len(schedule.Installments))
		inst := schedule.Installments[0]
		assert.True(t, inst.Capital.Equal(principal))
		// Interest = 10000 * 0.01 = 100
		assert.True(t, inst.Interest.Equal(decimal.NewFromInt(100)))
		assert.True(t, inst.Remaining.Equal(decimal.NewFromInt(0)))
	})

	t.Run("interest decreases and capital increases over time", func(t *testing.T) {
		principal := decimal.NewFromInt(100000)
		annualRate := decimal.NewFromFloat(0.36)
		installments := 24

		schedule := model.CalculateFrenchAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

		// In French amortization, interest portion decreases and capital increases
		for i := 1; i < len(schedule.Installments)-1; i++ {
			assert.True(t, schedule.Installments[i].Interest.LessThanOrEqual(schedule.Installments[i-1].Interest),
				"interest should decrease: installment %d (%s) > installment %d (%s)",
				i+1, schedule.Installments[i].Interest.StringFixed(2),
				i, schedule.Installments[i-1].Interest.StringFixed(2))
		}
	})
}

func TestCalculateGermanAmortization(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("basic 12 installments at 24 percent annual", func(t *testing.T) {
		principal := decimal.NewFromInt(120000)
		annualRate := decimal.NewFromFloat(0.24)
		installments := 12

		schedule := model.CalculateGermanAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

		assert.Equal(t, installments, len(schedule.Installments))
		assert.True(t, schedule.TotalInterest.IsPositive())
		assert.True(t, schedule.TotalPayment.GreaterThan(principal))

		// German: capital portion is constant (equal each month)
		expectedCapital := principal.Div(decimal.NewFromInt(int64(installments))).Round(2)
		for i := 0; i < len(schedule.Installments)-1; i++ {
			assert.True(t, schedule.Installments[i].Capital.Equal(expectedCapital),
				"installment %d capital should be %s, got %s",
				i+1, expectedCapital.StringFixed(2), schedule.Installments[i].Capital.StringFixed(2))
		}

		// German: total payment decreases over time (decreasing installments)
		for i := 1; i < len(schedule.Installments); i++ {
			assert.True(t, schedule.Installments[i].Total.LessThanOrEqual(schedule.Installments[i-1].Total),
				"installment %d total %s should be <= installment %d total %s",
				i+1, schedule.Installments[i].Total.StringFixed(2),
				i, schedule.Installments[i-1].Total.StringFixed(2))
		}

		// Last installment remaining should be zero
		lastInst := schedule.Installments[len(schedule.Installments)-1]
		assert.True(t, lastInst.Remaining.Equal(decimal.NewFromInt(0)),
			"last remaining should be 0, got %s", lastInst.Remaining.StringFixed(2))
	})

	t.Run("zero interest rate", func(t *testing.T) {
		principal := decimal.NewFromInt(12000)
		annualRate := decimal.NewFromInt(0)
		installments := 12

		schedule := model.CalculateGermanAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

		assert.True(t, schedule.TotalInterest.Equal(decimal.NewFromInt(0)))
		assert.True(t, schedule.TotalPayment.Equal(principal))

		for _, inst := range schedule.Installments {
			assert.True(t, inst.Interest.Equal(decimal.NewFromInt(0)))
		}
	})

	t.Run("single installment", func(t *testing.T) {
		principal := decimal.NewFromInt(10000)
		annualRate := decimal.NewFromFloat(0.12)

		schedule := model.CalculateGermanAmortization(principal, annualRate, 1, startDate, decimal.NewFromInt(21))

		require.Equal(t, 1, len(schedule.Installments))
		inst := schedule.Installments[0]
		assert.True(t, inst.Capital.Equal(principal))
		assert.True(t, inst.Interest.Equal(decimal.NewFromInt(100)))
	})

	t.Run("interest decreases over time", func(t *testing.T) {
		principal := decimal.NewFromInt(100000)
		annualRate := decimal.NewFromFloat(0.36)
		installments := 24

		schedule := model.CalculateGermanAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

		for i := 1; i < len(schedule.Installments); i++ {
			assert.True(t, schedule.Installments[i].Interest.LessThan(schedule.Installments[i-1].Interest),
				"interest should strictly decrease")
		}
	})

	t.Run("first installment is largest", func(t *testing.T) {
		principal := decimal.NewFromInt(60000)
		annualRate := decimal.NewFromFloat(0.24)
		installments := 6

		schedule := model.CalculateGermanAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

		firstTotal := schedule.Installments[0].Total
		lastTotal := schedule.Installments[len(schedule.Installments)-1].Total
		assert.True(t, firstTotal.GreaterThan(lastTotal),
			"first installment %s should be > last %s", firstTotal.StringFixed(2), lastTotal.StringFixed(2))
	})
}

func TestCalculateEarlyCancellation(t *testing.T) {
	t.Run("sum of remaining capital", func(t *testing.T) {
		unpaid := []model.InstallmentCalc{
			{Number: 3, Capital: decimal.NewFromInt(5000), Interest: decimal.NewFromInt(500)},
			{Number: 4, Capital: decimal.NewFromInt(5000), Interest: decimal.NewFromInt(400)},
			{Number: 5, Capital: decimal.NewFromInt(5000), Interest: decimal.NewFromInt(300)},
		}

		result := model.CalculateEarlyCancellation(unpaid)
		assert.True(t, result.Equal(decimal.NewFromInt(15000)),
			"early cancellation should be 15000, got %s", result.StringFixed(2))
	})

	t.Run("single unpaid installment", func(t *testing.T) {
		unpaid := []model.InstallmentCalc{
			{Number: 12, Capital: decimal.NewFromFloat(8333.33), Interest: decimal.NewFromInt(100)},
		}

		result := model.CalculateEarlyCancellation(unpaid)
		assert.True(t, result.Equal(decimal.NewFromFloat(8333.33)),
			"got %s", result.StringFixed(2))
	})

	t.Run("no unpaid installments", func(t *testing.T) {
		unpaid := []model.InstallmentCalc{}
		result := model.CalculateEarlyCancellation(unpaid)
		assert.True(t, result.Equal(decimal.NewFromInt(0)))
	})

	t.Run("nil unpaid installments", func(t *testing.T) {
		result := model.CalculateEarlyCancellation(nil)
		assert.True(t, result.Equal(decimal.NewFromInt(0)))
	})
}

func TestFrenchVsGerman_TotalInterestComparison(t *testing.T) {
	// With the same parameters, German amortization should have less total interest
	// because the principal is paid down faster (equal capital portions).
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	principal := decimal.NewFromInt(100000)
	annualRate := decimal.NewFromFloat(0.24)
	installments := 12

	french := model.CalculateFrenchAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))
	german := model.CalculateGermanAmortization(principal, annualRate, installments, startDate, decimal.NewFromInt(21))

	assert.True(t, german.TotalInterest.LessThan(french.TotalInterest),
		"German interest %s should be less than French interest %s",
		german.TotalInterest.StringFixed(2), french.TotalInterest.StringFixed(2))
}
