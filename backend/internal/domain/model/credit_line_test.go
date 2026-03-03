package model_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCreditLine(t *testing.T) {
	clientID := uuid.New()

	tests := []struct {
		name            string
		maxAmount       decimal.Decimal
		interestRate    decimal.Decimal
		maxInstallments int
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "valid credit line",
			maxAmount:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(0.25),
			maxInstallments: 12,
		},
		{
			name:            "valid with zero interest",
			maxAmount:       decimal.NewFromInt(50000),
			interestRate:    decimal.NewFromInt(0),
			maxInstallments: 6,
		},
		{
			name:            "max 60 installments",
			maxAmount:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(0.30),
			maxInstallments: 60,
		},
		{
			name:            "min 1 installment",
			maxAmount:       decimal.NewFromInt(10000),
			interestRate:    decimal.NewFromFloat(0.10),
			maxInstallments: 1,
		},
		{
			name:            "zero max amount fails",
			maxAmount:       decimal.NewFromInt(0),
			interestRate:    decimal.NewFromFloat(0.25),
			maxInstallments: 12,
			wantErr:         true,
			errMsg:          "max amount must be positive",
		},
		{
			name:            "negative max amount fails",
			maxAmount:       decimal.NewFromInt(-1000),
			interestRate:    decimal.NewFromFloat(0.25),
			maxInstallments: 12,
			wantErr:         true,
			errMsg:          "max amount must be positive",
		},
		{
			name:            "negative interest rate fails",
			maxAmount:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(-0.01),
			maxInstallments: 12,
			wantErr:         true,
			errMsg:          "interest rate cannot be negative",
		},
		{
			name:            "zero installments fails",
			maxAmount:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(0.25),
			maxInstallments: 0,
			wantErr:         true,
			errMsg:          "max installments must be between 1 and 60",
		},
		{
			name:            "over 60 installments fails",
			maxAmount:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(0.25),
			maxInstallments: 61,
			wantErr:         true,
			errMsg:          "max installments must be between 1 and 60",
		},
		{
			name:            "negative installments fails",
			maxAmount:       decimal.NewFromInt(100000),
			interestRate:    decimal.NewFromFloat(0.25),
			maxInstallments: -1,
			wantErr:         true,
			errMsg:          "max installments must be between 1 and 60",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, err := model.NewCreditLine(clientID, tt.maxAmount, tt.interestRate, tt.maxInstallments)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, cl)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cl)
				assert.Equal(t, clientID, cl.ClientID)
				assert.True(t, tt.maxAmount.Equal(cl.MaxAmount))
				assert.True(t, tt.interestRate.Equal(cl.InterestRate))
				assert.Equal(t, tt.maxInstallments, cl.MaxInstallments)
				assert.Equal(t, model.CreditLinePending, cl.Status)
				assert.True(t, cl.UsedAmount.Equal(decimal.NewFromInt(0)))
			}
		})
	}
}

func TestCreditLine_Approve(t *testing.T) {
	approverID := uuid.New()

	t.Run("approve pending credit line", func(t *testing.T) {
		cl, err := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		require.NoError(t, err)

		err = cl.Approve(approverID)
		require.NoError(t, err)
		assert.Equal(t, model.CreditLineApproved, cl.Status)
		assert.NotNil(t, cl.ApprovedBy)
		assert.Equal(t, approverID, *cl.ApprovedBy)
		assert.NotNil(t, cl.ApprovedAt)
	})

	t.Run("cannot approve already approved", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		_ = cl.Approve(approverID)

		err := cl.Approve(approverID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only approve pending credit lines")
	})

	t.Run("cannot approve rejected", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		_ = cl.Reject(uuid.New(), "bad credit")

		err := cl.Approve(approverID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only approve pending credit lines")
	})
}

func TestCreditLine_Reject(t *testing.T) {
	rejectorID := uuid.New()

	t.Run("reject pending credit line", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		err := cl.Reject(rejectorID, "insufficient income")

		require.NoError(t, err)
		assert.Equal(t, model.CreditLineRejected, cl.Status)
		assert.NotNil(t, cl.RejectedBy)
		assert.Equal(t, rejectorID, *cl.RejectedBy)
		assert.NotNil(t, cl.RejectedAt)
		assert.Equal(t, "insufficient income", cl.RejectionReason)
	})

	t.Run("reject without reason fails", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		err := cl.Reject(rejectorID, "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "rejection reason is required")
	})

	t.Run("cannot reject approved", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		_ = cl.Approve(uuid.New())

		err := cl.Reject(rejectorID, "changed mind")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only reject pending credit lines")
	})
}

func TestCreditLine_AvailableAmount(t *testing.T) {
	cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)

	assert.True(t, cl.AvailableAmount().Equal(decimal.NewFromInt(100000)))

	cl.RecordDisbursement(decimal.NewFromInt(30000))
	assert.True(t, cl.AvailableAmount().Equal(decimal.NewFromInt(70000)))

	cl.RecordDisbursement(decimal.NewFromInt(70000))
	assert.True(t, cl.AvailableAmount().Equal(decimal.NewFromInt(0)))
}

func TestCreditLine_CanDisburse(t *testing.T) {
	t.Run("approved line with enough balance", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		_ = cl.Approve(uuid.New())

		err := cl.CanDisburse(decimal.NewFromInt(50000))
		assert.NoError(t, err)
	})

	t.Run("approved line exact balance", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		_ = cl.Approve(uuid.New())

		err := cl.CanDisburse(decimal.NewFromInt(100000))
		assert.NoError(t, err)
	})

	t.Run("exceeds available amount", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		_ = cl.Approve(uuid.New())

		err := cl.CanDisburse(decimal.NewFromInt(100001))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds available")
	})

	t.Run("not approved fails", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)

		err := cl.CanDisburse(decimal.NewFromInt(50000))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "credit line is not approved")
	})

	t.Run("partially used line", func(t *testing.T) {
		cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
		_ = cl.Approve(uuid.New())
		cl.RecordDisbursement(decimal.NewFromInt(60000))

		err := cl.CanDisburse(decimal.NewFromInt(40000))
		assert.NoError(t, err)

		err = cl.CanDisburse(decimal.NewFromInt(40001))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds available")
	})
}

func TestCreditLine_ReleaseDisbursement(t *testing.T) {
	cl, _ := model.NewCreditLine(uuid.New(), decimal.NewFromInt(100000), decimal.NewFromFloat(0.25), 12)
	cl.RecordDisbursement(decimal.NewFromInt(50000))

	cl.ReleaseDisbursement(decimal.NewFromInt(30000))
	assert.True(t, cl.UsedAmount.Equal(decimal.NewFromInt(20000)))

	// Release more than used should floor at zero
	cl.ReleaseDisbursement(decimal.NewFromInt(50000))
	assert.True(t, cl.UsedAmount.Equal(decimal.NewFromInt(0)))
}
