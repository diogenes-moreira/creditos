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

func newTestLoan(t *testing.T) *model.Loan {
	t.Helper()
	loan, err := model.NewLoan(
		uuid.New(),
		uuid.New(),
		decimal.NewFromInt(100000),
		decimal.NewFromFloat(0.24),
		12,
		model.AmortizationFrench,
	)
	require.NoError(t, err)
	return loan
}

func TestNewLoan(t *testing.T) {
	clientID := uuid.New()
	creditLineID := uuid.New()

	tests := []struct {
		name            string
		principal       decimal.Decimal
		interestRate    decimal.Decimal
		numInstallments int
		amortType       model.AmortizationType
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "valid French loan",
			principal:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(0.24),
			numInstallments: 12,
			amortType:       model.AmortizationFrench,
		},
		{
			name:            "valid German loan",
			principal:       decimal.NewFromInt(50000),
			interestRate:    decimal.NewFromFloat(0.36),
			numInstallments: 6,
			amortType:       model.AmortizationGerman,
		},
		{
			name:            "zero principal fails",
			principal:       decimal.NewFromInt(0),
			interestRate:    decimal.NewFromFloat(0.24),
			numInstallments: 12,
			amortType:       model.AmortizationFrench,
			wantErr:         true,
			errMsg:          "principal must be positive",
		},
		{
			name:            "negative principal fails",
			principal:       decimal.NewFromInt(-1000),
			interestRate:    decimal.NewFromFloat(0.24),
			numInstallments: 12,
			amortType:       model.AmortizationFrench,
			wantErr:         true,
			errMsg:          "principal must be positive",
		},
		{
			name:            "zero installments fails",
			principal:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(0.24),
			numInstallments: 0,
			amortType:       model.AmortizationFrench,
			wantErr:         true,
			errMsg:          "number of installments must be at least 1",
		},
		{
			name:            "invalid amortization type fails",
			principal:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(0.24),
			numInstallments: 12,
			amortType:       model.AmortizationType("american"),
			wantErr:         true,
			errMsg:          "invalid amortization type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan, err := model.NewLoan(clientID, creditLineID, tt.principal, tt.interestRate, tt.numInstallments, tt.amortType)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, loan)
			} else {
				require.NoError(t, err)
				require.NotNil(t, loan)
				assert.Equal(t, clientID, loan.ClientID)
				assert.Equal(t, creditLineID, loan.CreditLineID)
				assert.True(t, tt.principal.Equal(loan.Principal))
				assert.Equal(t, model.LoanQuoted, loan.Status)
			}
		})
	}
}

func TestLoan_RequestApproval(t *testing.T) {
	t.Run("from quoted to pending", func(t *testing.T) {
		loan := newTestLoan(t)
		err := loan.RequestApproval()
		require.NoError(t, err)
		assert.Equal(t, model.LoanPending, loan.Status)
	})

	t.Run("fails from non-quoted status", func(t *testing.T) {
		loan := newTestLoan(t)
		_ = loan.RequestApproval()

		err := loan.RequestApproval()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only request approval for quoted loans")
	})
}

func TestLoan_Approve(t *testing.T) {
	approverID := uuid.New()

	t.Run("approve pending loan", func(t *testing.T) {
		loan := newTestLoan(t)
		_ = loan.RequestApproval()

		err := loan.Approve(approverID)
		require.NoError(t, err)
		assert.Equal(t, model.LoanApproved, loan.Status)
		assert.NotNil(t, loan.ApprovedBy)
		assert.Equal(t, approverID, *loan.ApprovedBy)
		assert.NotNil(t, loan.ApprovedAt)
	})

	t.Run("cannot approve quoted loan", func(t *testing.T) {
		loan := newTestLoan(t)
		err := loan.Approve(approverID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only approve pending loans")
	})
}

func TestLoan_Disburse(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("disburse approved loan creates installments", func(t *testing.T) {
		loan := newTestLoan(t)
		_ = loan.RequestApproval()
		_ = loan.Approve(uuid.New())

		installments, err := loan.Disburse(startDate, decimal.NewFromInt(21))
		require.NoError(t, err)
		assert.Equal(t, 12, len(installments))
		assert.Equal(t, model.LoanActive, loan.Status)
		assert.NotNil(t, loan.DisbursedAt)

		for i, inst := range installments {
			assert.Equal(t, i+1, inst.Number)
			assert.Equal(t, loan.ID, inst.LoanID)
			assert.Equal(t, model.InstallmentPending, inst.Status)
			assert.True(t, inst.PaidAmount.Equal(decimal.NewFromInt(0)))
			assert.True(t, inst.RemainingAmount.Equal(inst.TotalAmount))
		}
	})

	t.Run("cannot disburse non-approved loan", func(t *testing.T) {
		loan := newTestLoan(t)
		_, err := loan.Disburse(startDate, decimal.NewFromInt(21))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only disburse approved loans")
	})

	t.Run("disburse with German amortization", func(t *testing.T) {
		loan, _ := model.NewLoan(uuid.New(), uuid.New(),
			decimal.NewFromInt(60000), decimal.NewFromFloat(0.24), 6, model.AmortizationGerman)
		_ = loan.RequestApproval()
		_ = loan.Approve(uuid.New())

		installments, err := loan.Disburse(startDate, decimal.NewFromInt(21))
		require.NoError(t, err)
		assert.Equal(t, 6, len(installments))
	})
}

func TestLoan_Cancel(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("cancel active loan", func(t *testing.T) {
		loan := newTestLoan(t)
		_ = loan.RequestApproval()
		_ = loan.Approve(uuid.New())
		_, _ = loan.Disburse(startDate, decimal.NewFromInt(21))

		err := loan.Cancel()
		require.NoError(t, err)
		assert.Equal(t, model.LoanCancelled, loan.Status)
		assert.NotNil(t, loan.CancelledAt)
	})

	t.Run("cannot cancel non-active loan", func(t *testing.T) {
		loan := newTestLoan(t)
		err := loan.Cancel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only cancel active loans")
	})
}

func TestLoan_Complete(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("complete active loan", func(t *testing.T) {
		loan := newTestLoan(t)
		_ = loan.RequestApproval()
		_ = loan.Approve(uuid.New())
		_, _ = loan.Disburse(startDate, decimal.NewFromInt(21))

		err := loan.Complete()
		require.NoError(t, err)
		assert.Equal(t, model.LoanCompleted, loan.Status)
		assert.NotNil(t, loan.CompletedAt)
	})

	t.Run("cannot complete non-active loan", func(t *testing.T) {
		loan := newTestLoan(t)
		err := loan.Complete()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only complete active loans")
	})
}

func TestLoan_MarkDefaulted(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("default active loan", func(t *testing.T) {
		loan := newTestLoan(t)
		_ = loan.RequestApproval()
		_ = loan.Approve(uuid.New())
		_, _ = loan.Disburse(startDate, decimal.NewFromInt(21))

		err := loan.MarkDefaulted()
		require.NoError(t, err)
		assert.Equal(t, model.LoanDefaulted, loan.Status)
	})

	t.Run("cannot default non-active loan", func(t *testing.T) {
		loan := newTestLoan(t)
		err := loan.MarkDefaulted()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only default active loans")
	})
}

func TestLoan_StatusTransitionChain(t *testing.T) {
	// Test the full lifecycle: quoted -> pending -> approved -> active -> completed
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	loan := newTestLoan(t)

	assert.Equal(t, model.LoanQuoted, loan.Status)

	require.NoError(t, loan.RequestApproval())
	assert.Equal(t, model.LoanPending, loan.Status)

	require.NoError(t, loan.Approve(uuid.New()))
	assert.Equal(t, model.LoanApproved, loan.Status)

	_, err := loan.Disburse(startDate, decimal.NewFromInt(21))
	require.NoError(t, err)
	assert.Equal(t, model.LoanActive, loan.Status)

	require.NoError(t, loan.Complete())
	assert.Equal(t, model.LoanCompleted, loan.Status)
}

func TestLoan_CheckCompletion(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	loan := newTestLoan(t)
	_ = loan.RequestApproval()
	_ = loan.Approve(uuid.New())
	_, _ = loan.Disburse(startDate, decimal.NewFromInt(21))

	// Not all paid yet
	assert.False(t, loan.CheckCompletion())

	// Mark all as paid
	for i := range loan.Installments {
		loan.Installments[i].Status = model.InstallmentPaid
	}
	assert.True(t, loan.CheckCompletion())
}

func TestLoan_TotalPaidAndRemaining(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	loan := newTestLoan(t)
	_ = loan.RequestApproval()
	_ = loan.Approve(uuid.New())
	_, _ = loan.Disburse(startDate, decimal.NewFromInt(21))

	assert.True(t, loan.TotalPaid().Equal(decimal.NewFromInt(0)))
	assert.True(t, loan.TotalRemaining().IsPositive())

	totalExpected := decimal.NewFromInt(0)
	for _, inst := range loan.Installments {
		totalExpected = totalExpected.Add(inst.TotalAmount)
	}
	assert.True(t, loan.TotalRemaining().Equal(totalExpected))
}
