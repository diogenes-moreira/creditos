package model_test

import (
	"testing"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestInstallment(total decimal.Decimal) model.Installment {
	return model.Installment{
		ID:              uuid.New(),
		LoanID:          uuid.New(),
		Number:          1,
		DueDate:         time.Now().AddDate(0, 1, 0),
		CapitalAmount:   total.Mul(decimal.NewFromFloat(0.8)),
		InterestAmount:  total.Mul(decimal.NewFromFloat(0.2)),
		TotalAmount:     total,
		PaidAmount:      decimal.NewFromInt(0),
		RemainingAmount: total,
		Status:          model.InstallmentPending,
	}
}

func TestInstallment_ApplyPayment_Full(t *testing.T) {
	inst := newTestInstallment(decimal.NewFromInt(10000))

	applied, surplus, err := inst.ApplyPayment(decimal.NewFromInt(10000))
	require.NoError(t, err)

	assert.True(t, applied.Equal(decimal.NewFromInt(10000)))
	assert.True(t, surplus.Equal(decimal.NewFromInt(0)))
	assert.Equal(t, model.InstallmentPaid, inst.Status)
	assert.True(t, inst.PaidAmount.Equal(decimal.NewFromInt(10000)))
	assert.True(t, inst.RemainingAmount.Equal(decimal.NewFromInt(0)))
	assert.NotNil(t, inst.PaidAt)
}

func TestInstallment_ApplyPayment_Partial(t *testing.T) {
	inst := newTestInstallment(decimal.NewFromInt(10000))

	applied, surplus, err := inst.ApplyPayment(decimal.NewFromInt(3000))
	require.NoError(t, err)

	assert.True(t, applied.Equal(decimal.NewFromInt(3000)))
	assert.True(t, surplus.Equal(decimal.NewFromInt(0)))
	assert.Equal(t, model.InstallmentPartial, inst.Status)
	assert.True(t, inst.PaidAmount.Equal(decimal.NewFromInt(3000)))
	assert.True(t, inst.RemainingAmount.Equal(decimal.NewFromInt(7000)))
	assert.Nil(t, inst.PaidAt)
}

func TestInstallment_ApplyPayment_Overpayment(t *testing.T) {
	inst := newTestInstallment(decimal.NewFromInt(10000))

	applied, surplus, err := inst.ApplyPayment(decimal.NewFromInt(15000))
	require.NoError(t, err)

	assert.True(t, applied.Equal(decimal.NewFromInt(10000)))
	assert.True(t, surplus.Equal(decimal.NewFromInt(5000)))
	assert.Equal(t, model.InstallmentPaid, inst.Status)
	assert.True(t, inst.PaidAmount.Equal(decimal.NewFromInt(10000)))
	assert.True(t, inst.RemainingAmount.Equal(decimal.NewFromInt(0)))
	assert.NotNil(t, inst.PaidAt)
}

func TestInstallment_ApplyPayment_MultiplePartials(t *testing.T) {
	inst := newTestInstallment(decimal.NewFromInt(10000))

	// First partial payment
	applied1, surplus1, err := inst.ApplyPayment(decimal.NewFromInt(3000))
	require.NoError(t, err)
	assert.True(t, applied1.Equal(decimal.NewFromInt(3000)))
	assert.True(t, surplus1.Equal(decimal.NewFromInt(0)))
	assert.Equal(t, model.InstallmentPartial, inst.Status)

	// Second partial payment
	applied2, surplus2, err := inst.ApplyPayment(decimal.NewFromInt(4000))
	require.NoError(t, err)
	assert.True(t, applied2.Equal(decimal.NewFromInt(4000)))
	assert.True(t, surplus2.Equal(decimal.NewFromInt(0)))
	assert.Equal(t, model.InstallmentPartial, inst.Status)
	assert.True(t, inst.PaidAmount.Equal(decimal.NewFromInt(7000)))

	// Final payment completes
	applied3, surplus3, err := inst.ApplyPayment(decimal.NewFromInt(3000))
	require.NoError(t, err)
	assert.True(t, applied3.Equal(decimal.NewFromInt(3000)))
	assert.True(t, surplus3.Equal(decimal.NewFromInt(0)))
	assert.Equal(t, model.InstallmentPaid, inst.Status)
}

func TestInstallment_ApplyPayment_ZeroAmount(t *testing.T) {
	inst := newTestInstallment(decimal.NewFromInt(10000))

	_, _, err := inst.ApplyPayment(decimal.NewFromInt(0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment amount must be positive")
}

func TestInstallment_ApplyPayment_NegativeAmount(t *testing.T) {
	inst := newTestInstallment(decimal.NewFromInt(10000))

	_, _, err := inst.ApplyPayment(decimal.NewFromInt(-100))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment amount must be positive")
}

func TestInstallment_ApplyPayment_AlreadyPaid(t *testing.T) {
	inst := newTestInstallment(decimal.NewFromInt(10000))
	_, _, _ = inst.ApplyPayment(decimal.NewFromInt(10000))
	assert.Equal(t, model.InstallmentPaid, inst.Status)

	// Paying again returns full surplus
	applied, surplus, err := inst.ApplyPayment(decimal.NewFromInt(5000))
	require.NoError(t, err)
	assert.True(t, applied.Equal(decimal.NewFromInt(0)))
	assert.True(t, surplus.Equal(decimal.NewFromInt(5000)))
}

func TestInstallment_IsOverdue(t *testing.T) {
	t.Run("future due date is not overdue", func(t *testing.T) {
		inst := newTestInstallment(decimal.NewFromInt(10000))
		inst.DueDate = time.Now().AddDate(0, 1, 0)
		assert.False(t, inst.IsOverdue())
	})

	t.Run("past due date is overdue", func(t *testing.T) {
		inst := newTestInstallment(decimal.NewFromInt(10000))
		inst.DueDate = time.Now().AddDate(0, -1, 0)
		assert.True(t, inst.IsOverdue())
	})

	t.Run("paid installment is not overdue even if past due", func(t *testing.T) {
		inst := newTestInstallment(decimal.NewFromInt(10000))
		inst.DueDate = time.Now().AddDate(0, -1, 0)
		inst.Status = model.InstallmentPaid
		assert.False(t, inst.IsOverdue())
	})
}

func TestInstallment_MarkOverdue(t *testing.T) {
	t.Run("marks pending as overdue", func(t *testing.T) {
		inst := newTestInstallment(decimal.NewFromInt(10000))
		inst.MarkOverdue()
		assert.Equal(t, model.InstallmentOverdue, inst.Status)
	})

	t.Run("marks partial as overdue", func(t *testing.T) {
		inst := newTestInstallment(decimal.NewFromInt(10000))
		inst.Status = model.InstallmentPartial
		inst.MarkOverdue()
		assert.Equal(t, model.InstallmentOverdue, inst.Status)
	})

	t.Run("does not mark paid as overdue", func(t *testing.T) {
		inst := newTestInstallment(decimal.NewFromInt(10000))
		inst.Status = model.InstallmentPaid
		inst.MarkOverdue()
		assert.Equal(t, model.InstallmentPaid, inst.Status)
	})
}

func TestInstallment_DaysOverdue(t *testing.T) {
	t.Run("not overdue returns 0", func(t *testing.T) {
		inst := newTestInstallment(decimal.NewFromInt(10000))
		inst.DueDate = time.Now().AddDate(0, 1, 0)
		assert.Equal(t, 0, inst.DaysOverdue())
	})

	t.Run("overdue by 30 days", func(t *testing.T) {
		inst := newTestInstallment(decimal.NewFromInt(10000))
		inst.DueDate = time.Now().AddDate(0, 0, -30)
		days := inst.DaysOverdue()
		assert.True(t, days >= 29 && days <= 31, "expected around 30 days overdue, got %d", days)
	})
}
