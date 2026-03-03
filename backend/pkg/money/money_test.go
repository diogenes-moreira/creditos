package money_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/pkg/money"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{name: "positive amount", input: 100.50, expected: "100.5"},
		{name: "zero", input: 0, expected: "0"},
		{name: "negative amount", input: -50.25, expected: "-50.25"},
		{name: "very small amount", input: 0.01, expected: "0.01"},
		{name: "large amount", input: 1000000.99, expected: "1000000.99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := money.New(tt.input)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestNewFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid decimal", input: "100.50", want: "100.5"},
		{name: "valid integer", input: "100", want: "100"},
		{name: "valid negative", input: "-50.25", want: "-50.25"},
		{name: "valid zero", input: "0", want: "0"},
		{name: "invalid string", input: "abc", wantErr: true},
		{name: "empty string", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := money.NewFromString(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result.String())
			}
		})
	}
}

func TestNewFromInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{name: "positive int", input: 100, expected: "100"},
		{name: "zero", input: 0, expected: "0"},
		{name: "negative int", input: -50, expected: "-50"},
		{name: "large int", input: 1000000, expected: "1000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := money.NewFromInt(tt.input)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestIsPositive(t *testing.T) {
	tests := []struct {
		name     string
		value    decimal.Decimal
		expected bool
	}{
		{name: "positive value", value: money.New(100.50), expected: true},
		{name: "zero", value: money.Zero, expected: false},
		{name: "negative value", value: money.New(-1), expected: false},
		{name: "small positive", value: money.New(0.01), expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, money.IsPositive(tt.value))
		})
	}
}

func TestIsNegative(t *testing.T) {
	tests := []struct {
		name     string
		value    decimal.Decimal
		expected bool
	}{
		{name: "negative value", value: money.New(-100.50), expected: true},
		{name: "zero", value: money.Zero, expected: false},
		{name: "positive value", value: money.New(1), expected: false},
		{name: "small negative", value: money.New(-0.01), expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, money.IsNegative(tt.value))
		})
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name     string
		value    decimal.Decimal
		expected bool
	}{
		{name: "zero constant", value: money.Zero, expected: true},
		{name: "new zero from int", value: money.NewFromInt(0), expected: true},
		{name: "new zero from float", value: money.New(0), expected: true},
		{name: "positive", value: money.New(1), expected: false},
		{name: "negative", value: money.New(-1), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, money.IsZero(tt.value))
		})
	}
}

func TestRound2(t *testing.T) {
	tests := []struct {
		name     string
		input    decimal.Decimal
		expected string
	}{
		{name: "rounds down third decimal", input: money.New(100.554), expected: "100.55"},
		{name: "rounds up third decimal", input: money.New(100.555), expected: "100.56"},
		{name: "already two decimals", input: money.New(100.55), expected: "100.55"},
		{name: "integer value", input: money.NewFromInt(100), expected: "100.00"},
		{name: "many decimals", input: money.New(1.23456789), expected: "1.23"},
		{name: "negative rounds", input: money.New(-100.555), expected: "-100.56"},
		{name: "zero", input: money.Zero, expected: "0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := money.Round2(tt.input)
			assert.Equal(t, tt.expected, result.StringFixed(2))
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a        decimal.Decimal
		b        decimal.Decimal
		expected decimal.Decimal
	}{
		{name: "a greater than b", a: money.New(100), b: money.New(50), expected: money.New(100)},
		{name: "b greater than a", a: money.New(50), b: money.New(100), expected: money.New(100)},
		{name: "equal values", a: money.New(100), b: money.New(100), expected: money.New(100)},
		{name: "negative values", a: money.New(-10), b: money.New(-20), expected: money.New(-10)},
		{name: "zero and positive", a: money.Zero, b: money.New(1), expected: money.New(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := money.Max(tt.a, tt.b)
			assert.True(t, tt.expected.Equal(result), "expected %s, got %s", tt.expected, result)
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        decimal.Decimal
		b        decimal.Decimal
		expected decimal.Decimal
	}{
		{name: "a less than b", a: money.New(50), b: money.New(100), expected: money.New(50)},
		{name: "b less than a", a: money.New(100), b: money.New(50), expected: money.New(50)},
		{name: "equal values", a: money.New(100), b: money.New(100), expected: money.New(100)},
		{name: "negative values", a: money.New(-10), b: money.New(-20), expected: money.New(-20)},
		{name: "zero and positive", a: money.Zero, b: money.New(1), expected: money.Zero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := money.Min(tt.a, tt.b)
			assert.True(t, tt.expected.Equal(result), "expected %s, got %s", tt.expected, result)
		})
	}
}

func TestZeroConstant(t *testing.T) {
	assert.True(t, money.Zero.Equal(decimal.NewFromInt(0)))
	assert.True(t, money.IsZero(money.Zero))
	assert.False(t, money.IsPositive(money.Zero))
	assert.False(t, money.IsNegative(money.Zero))
}
