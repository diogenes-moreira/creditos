package model_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayment(t *testing.T) {
	loanID := uuid.New()

	tests := []struct {
		name      string
		amount    decimal.Decimal
		method    model.PaymentMethod
		reference string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid cash payment",
			amount:    decimal.NewFromInt(5000),
			method:    model.PaymentCash,
			reference: "receipt-001",
		},
		{
			name:      "valid transfer payment",
			amount:    decimal.NewFromInt(10000),
			method:    model.PaymentTransfer,
			reference: "TRF-12345",
		},
		{
			name:      "valid mercado pago payment",
			amount:    decimal.NewFromFloat(7500.50),
			method:    model.PaymentMercadoPago,
			reference: "MP-98765",
		},
		{
			name:    "zero amount fails",
			amount:  decimal.NewFromInt(0),
			method:  model.PaymentCash,
			wantErr: true,
			errMsg:  "payment amount must be positive",
		},
		{
			name:    "negative amount fails",
			amount:  decimal.NewFromInt(-100),
			method:  model.PaymentCash,
			wantErr: true,
			errMsg:  "payment amount must be positive",
		},
		{
			name:    "invalid payment method fails",
			amount:  decimal.NewFromInt(5000),
			method:  model.PaymentMethod("bitcoin"),
			wantErr: true,
			errMsg:  "invalid payment method",
		},
		{
			name:    "empty payment method fails",
			amount:  decimal.NewFromInt(5000),
			method:  model.PaymentMethod(""),
			wantErr: true,
			errMsg:  "invalid payment method",
		},
		{
			name:      "valid payment with empty reference",
			amount:    decimal.NewFromInt(1000),
			method:    model.PaymentCash,
			reference: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := model.NewPayment(loanID, tt.amount, tt.method, tt.reference)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, payment)
			} else {
				require.NoError(t, err)
				require.NotNil(t, payment)
				assert.Equal(t, loanID, payment.LoanID)
				assert.True(t, tt.amount.Equal(payment.Amount))
				assert.Equal(t, tt.method, payment.Method)
				assert.Equal(t, tt.reference, payment.Reference)
				assert.False(t, payment.IsAdjustment)
				assert.Nil(t, payment.AdjustedBy)
				assert.Empty(t, payment.AdjustmentNote)
				assert.Nil(t, payment.InstallmentID)
			}
		})
	}
}

func TestPayment_Adjust(t *testing.T) {
	adjusterID := uuid.New()

	t.Run("adjust normal payment", func(t *testing.T) {
		payment, _ := model.NewPayment(uuid.New(), decimal.NewFromInt(5000), model.PaymentCash, "ref")
		err := payment.Adjust(adjusterID, "incorrect amount applied")

		require.NoError(t, err)
		assert.True(t, payment.IsAdjustment)
		assert.NotNil(t, payment.AdjustedBy)
		assert.Equal(t, adjusterID, *payment.AdjustedBy)
		assert.Equal(t, "incorrect amount applied", payment.AdjustmentNote)
	})

	t.Run("cannot adjust already adjusted payment", func(t *testing.T) {
		payment, _ := model.NewPayment(uuid.New(), decimal.NewFromInt(5000), model.PaymentCash, "ref")
		_ = payment.Adjust(adjusterID, "first adjustment")

		err := payment.Adjust(adjusterID, "second adjustment")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "payment is already an adjustment")
	})

	t.Run("empty note fails", func(t *testing.T) {
		payment, _ := model.NewPayment(uuid.New(), decimal.NewFromInt(5000), model.PaymentCash, "ref")
		err := payment.Adjust(adjusterID, "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "adjustment note is required")
	})
}

func TestPayment_LinkInstallment(t *testing.T) {
	payment, _ := model.NewPayment(uuid.New(), decimal.NewFromInt(5000), model.PaymentCash, "ref")
	assert.Nil(t, payment.InstallmentID)

	instID := uuid.New()
	payment.LinkInstallment(instID)

	require.NotNil(t, payment.InstallmentID)
	assert.Equal(t, instID, *payment.InstallmentID)
}
