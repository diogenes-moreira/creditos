package model_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVendorAccount(t *testing.T) {
	vendorID := uuid.New()
	account := model.NewVendorAccount(vendorID)

	require.NotNil(t, account)
	assert.Equal(t, vendorID, account.VendorID)
	assert.True(t, account.Balance.Equal(decimal.NewFromInt(0)))
	assert.NotEqual(t, uuid.Nil, account.ID)
}

func TestVendorAccount_Credit(t *testing.T) {
	tests := []struct {
		name        string
		amount      decimal.Decimal
		wantErr     bool
		errMsg      string
		wantBalance string
	}{
		{
			name:        "credit positive amount",
			amount:      decimal.NewFromFloat(100.50),
			wantBalance: "100.5",
		},
		{
			name:    "credit zero amount fails",
			amount:  decimal.NewFromInt(0),
			wantErr: true,
			errMsg:  "credit amount must be positive",
		},
		{
			name:    "credit negative amount fails",
			amount:  decimal.NewFromFloat(-50),
			wantErr: true,
			errMsg:  "credit amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := model.NewVendorAccount(uuid.New())
			movement, err := account.Credit(tt.amount, "test credit", "ref-001")

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, movement)
			} else {
				require.NoError(t, err)
				require.NotNil(t, movement)
				assert.Equal(t, tt.wantBalance, account.Balance.String())
				assert.Equal(t, "test credit", movement.Description)
				assert.Equal(t, "ref-001", movement.Reference)
			}
		})
	}
}

func TestVendorAccount_Debit(t *testing.T) {
	tests := []struct {
		name        string
		amount      decimal.Decimal
		wantErr     bool
		errMsg      string
		wantBalance string
	}{
		{
			name:        "debit positive amount",
			amount:      decimal.NewFromFloat(50),
			wantBalance: "-50",
		},
		{
			name:    "debit zero amount fails",
			amount:  decimal.NewFromInt(0),
			wantErr: true,
			errMsg:  "debit amount must be positive",
		},
		{
			name:    "debit negative amount fails",
			amount:  decimal.NewFromFloat(-50),
			wantErr: true,
			errMsg:  "debit amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := model.NewVendorAccount(uuid.New())
			movement, err := account.Debit(tt.amount, "test debit", "ref-002")

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, movement)
			} else {
				require.NoError(t, err)
				require.NotNil(t, movement)
				assert.Equal(t, tt.wantBalance, account.Balance.String())
			}
		})
	}
}

func TestVendorAccount_BalanceTracking(t *testing.T) {
	account := model.NewVendorAccount(uuid.New())

	m1, err := account.Credit(decimal.NewFromInt(1000), "purchase sale", "pur-001")
	require.NoError(t, err)
	assert.True(t, account.Balance.Equal(decimal.NewFromInt(1000)))
	assert.True(t, m1.BalanceAfter.Equal(decimal.NewFromInt(1000)))

	m2, err := account.Debit(decimal.NewFromInt(300), "payment received", "pay-001")
	require.NoError(t, err)
	assert.True(t, account.Balance.Equal(decimal.NewFromInt(700)))
	assert.True(t, m2.BalanceAfter.Equal(decimal.NewFromInt(700)))

	m3, err := account.Credit(decimal.NewFromInt(200), "another sale", "pur-002")
	require.NoError(t, err)
	assert.True(t, account.Balance.Equal(decimal.NewFromInt(900)))
	assert.True(t, m3.BalanceAfter.Equal(decimal.NewFromInt(900)))
}

func TestVendorAccount_MovementTypes(t *testing.T) {
	account := model.NewVendorAccount(uuid.New())

	creditMov, err := account.Credit(decimal.NewFromInt(100), "credit", "")
	require.NoError(t, err)
	assert.Equal(t, model.MovementTypeCredit, creditMov.Type)
	assert.True(t, creditMov.Amount.Equal(decimal.NewFromInt(100)))

	debitMov, err := account.Debit(decimal.NewFromInt(50), "debit", "")
	require.NoError(t, err)
	assert.Equal(t, model.MovementTypeDebit, debitMov.Type)
	assert.True(t, debitMov.Amount.Equal(decimal.NewFromInt(50)))
}
