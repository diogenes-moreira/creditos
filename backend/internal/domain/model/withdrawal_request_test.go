package model_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithdrawalRequest(t *testing.T) {
	vendorID := uuid.New()

	tests := []struct {
		name    string
		amount  decimal.Decimal
		method  model.VendorPaymentMethod
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid cash withdrawal",
			amount: decimal.NewFromInt(5000),
			method: model.VendorPaymentCash,
		},
		{
			name:   "valid transfer withdrawal",
			amount: decimal.NewFromFloat(10000.50),
			method: model.VendorPaymentTransfer,
		},
		{
			name:    "zero amount fails",
			amount:  decimal.NewFromInt(0),
			method:  model.VendorPaymentCash,
			wantErr: true,
			errMsg:  "withdrawal amount must be positive",
		},
		{
			name:    "negative amount fails",
			amount:  decimal.NewFromInt(-100),
			method:  model.VendorPaymentCash,
			wantErr: true,
			errMsg:  "withdrawal amount must be positive",
		},
		{
			name:    "invalid method fails",
			amount:  decimal.NewFromInt(5000),
			method:  model.VendorPaymentMethod("bitcoin"),
			wantErr: true,
			errMsg:  "invalid withdrawal method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr, err := model.NewWithdrawalRequest(vendorID, tt.amount, tt.method)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, wr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, wr)
				assert.Equal(t, vendorID, wr.VendorID)
				assert.True(t, tt.amount.Equal(wr.Amount))
				assert.Equal(t, tt.method, wr.Method)
				assert.Equal(t, model.WithdrawalStatusPending, wr.Status)
				assert.Nil(t, wr.ProcessedAt)
				assert.Nil(t, wr.ProcessedBy)
				assert.Nil(t, wr.PaymentID)
				assert.Empty(t, wr.RejectionReason)
			}
		})
	}
}

func TestWithdrawalRequest_Approve(t *testing.T) {
	adminID := uuid.New()

	t.Run("approve pending request", func(t *testing.T) {
		wr, _ := model.NewWithdrawalRequest(uuid.New(), decimal.NewFromInt(5000), model.VendorPaymentCash)
		err := wr.Approve(adminID)

		require.NoError(t, err)
		assert.Equal(t, model.WithdrawalStatusApproved, wr.Status)
		assert.NotNil(t, wr.ProcessedAt)
		assert.NotNil(t, wr.ProcessedBy)
		assert.Equal(t, adminID, *wr.ProcessedBy)
	})

	t.Run("cannot approve non-pending request", func(t *testing.T) {
		wr, _ := model.NewWithdrawalRequest(uuid.New(), decimal.NewFromInt(5000), model.VendorPaymentCash)
		_ = wr.Approve(adminID)

		err := wr.Approve(adminID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only approve pending")
	})
}

func TestWithdrawalRequest_MarkPaid(t *testing.T) {
	adminID := uuid.New()
	paymentID := uuid.New()

	t.Run("mark approved request as paid", func(t *testing.T) {
		wr, _ := model.NewWithdrawalRequest(uuid.New(), decimal.NewFromInt(5000), model.VendorPaymentTransfer)
		_ = wr.Approve(adminID)

		err := wr.MarkPaid(paymentID, "TRF-12345")
		require.NoError(t, err)
		assert.Equal(t, model.WithdrawalStatusPaid, wr.Status)
		assert.NotNil(t, wr.PaymentID)
		assert.Equal(t, paymentID, *wr.PaymentID)
		assert.Equal(t, "TRF-12345", wr.Reference)
	})

	t.Run("cannot mark pending request as paid", func(t *testing.T) {
		wr, _ := model.NewWithdrawalRequest(uuid.New(), decimal.NewFromInt(5000), model.VendorPaymentCash)

		err := wr.MarkPaid(paymentID, "ref")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only mark approved")
	})
}

func TestWithdrawalRequest_Reject(t *testing.T) {
	adminID := uuid.New()

	t.Run("reject pending request", func(t *testing.T) {
		wr, _ := model.NewWithdrawalRequest(uuid.New(), decimal.NewFromInt(5000), model.VendorPaymentCash)
		err := wr.Reject(adminID, "insufficient funds")

		require.NoError(t, err)
		assert.Equal(t, model.WithdrawalStatusRejected, wr.Status)
		assert.Equal(t, "insufficient funds", wr.RejectionReason)
		assert.NotNil(t, wr.ProcessedAt)
		assert.NotNil(t, wr.ProcessedBy)
		assert.Equal(t, adminID, *wr.ProcessedBy)
	})

	t.Run("cannot reject non-pending request", func(t *testing.T) {
		wr, _ := model.NewWithdrawalRequest(uuid.New(), decimal.NewFromInt(5000), model.VendorPaymentCash)
		_ = wr.Approve(adminID)

		err := wr.Reject(adminID, "reason")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only reject pending")
	})

	t.Run("empty reason fails", func(t *testing.T) {
		wr, _ := model.NewWithdrawalRequest(uuid.New(), decimal.NewFromInt(5000), model.VendorPaymentCash)
		err := wr.Reject(adminID, "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "rejection reason is required")
	})
}
