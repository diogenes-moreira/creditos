package model_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPurchase(t *testing.T) {
	tests := []struct {
		name        string
		amount      decimal.Decimal
		description string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid purchase",
			amount:      decimal.NewFromFloat(1500.50),
			description: "Electrodomestico",
		},
		{
			name:        "zero amount fails",
			amount:      decimal.NewFromInt(0),
			description: "item",
			wantErr:     true,
			errMsg:      "purchase amount must be positive",
		},
		{
			name:        "negative amount fails",
			amount:      decimal.NewFromFloat(-100),
			description: "item",
			wantErr:     true,
			errMsg:      "purchase amount must be positive",
		},
		{
			name:        "empty description fails",
			amount:      decimal.NewFromInt(100),
			description: "",
			wantErr:     true,
			errMsg:      "purchase description is required",
		},
	}

	vendorID := uuid.New()
	clientID := uuid.New()
	creditLineID := uuid.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			purchase, err := model.NewPurchase(vendorID, clientID, creditLineID, tt.amount, tt.description)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, purchase)
			} else {
				require.NoError(t, err)
				require.NotNil(t, purchase)
				assert.NotEqual(t, uuid.Nil, purchase.ID)
				assert.Equal(t, vendorID, purchase.VendorID)
				assert.Equal(t, clientID, purchase.ClientID)
				assert.Equal(t, creditLineID, purchase.CreditLineID)
				assert.True(t, purchase.Amount.Equal(tt.amount))
				assert.Equal(t, tt.description, purchase.Description)
			}
		})
	}
}
