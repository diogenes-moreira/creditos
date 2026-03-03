package model_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVendorPayment(t *testing.T) {
	tests := []struct {
		name      string
		amount    decimal.Decimal
		method    model.VendorPaymentMethod
		reference string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid cash payment",
			amount:    decimal.NewFromFloat(5000),
			method:    model.VendorPaymentCash,
			reference: "recibo-001",
		},
		{
			name:      "valid transfer payment",
			amount:    decimal.NewFromFloat(10000),
			method:    model.VendorPaymentTransfer,
			reference: "TRF-123456",
		},
		{
			name:    "zero amount fails",
			amount:  decimal.NewFromInt(0),
			method:  model.VendorPaymentCash,
			wantErr: true,
			errMsg:  "payment amount must be positive",
		},
		{
			name:    "negative amount fails",
			amount:  decimal.NewFromFloat(-100),
			method:  model.VendorPaymentCash,
			wantErr: true,
			errMsg:  "payment amount must be positive",
		},
		{
			name:    "invalid method fails",
			amount:  decimal.NewFromInt(100),
			method:  model.VendorPaymentMethod("bitcoin"),
			wantErr: true,
			errMsg:  "invalid payment method",
		},
	}

	vendorID := uuid.New()
	paidBy := uuid.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := model.NewVendorPayment(vendorID, tt.amount, tt.method, tt.reference, paidBy)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, payment)
			} else {
				require.NoError(t, err)
				require.NotNil(t, payment)
				assert.NotEqual(t, uuid.Nil, payment.ID)
				assert.Equal(t, vendorID, payment.VendorID)
				assert.True(t, payment.Amount.Equal(tt.amount))
				assert.Equal(t, tt.method, payment.Method)
				assert.Equal(t, tt.reference, payment.Reference)
				assert.Equal(t, paidBy, payment.PaidBy)
			}
		})
	}
}
