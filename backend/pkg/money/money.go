package money

import "github.com/shopspring/decimal"

var Zero = decimal.NewFromInt(0)

func New(amount float64) decimal.Decimal {
	return decimal.NewFromFloat(amount)
}

func NewFromString(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

func NewFromInt(i int64) decimal.Decimal {
	return decimal.NewFromInt(i)
}

func IsPositive(d decimal.Decimal) bool {
	return d.GreaterThan(Zero)
}

func IsNegative(d decimal.Decimal) bool {
	return d.LessThan(Zero)
}

func IsZero(d decimal.Decimal) bool {
	return d.Equal(Zero)
}

func Round2(d decimal.Decimal) decimal.Decimal {
	return d.Round(2)
}

func Max(a, b decimal.Decimal) decimal.Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}

func Min(a, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}
