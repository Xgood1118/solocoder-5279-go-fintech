package money

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Money struct {
	Amount   decimal.Decimal
	Currency string
}

const DefaultCurrency = "CNY"
const precision = 2

func New(amount decimal.Decimal, currency string) Money {
	if currency == "" {
		currency = DefaultCurrency
	}
	return Money{
		Amount:   amount.Round(precision),
		Currency: currency,
	}
}

func NewFromInt(amount int64, currency string) Money {
	return New(decimal.NewFromInt(amount), currency)
}

func NewFromString(s string) (Money, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Money{}, fmt.Errorf("empty amount string")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Money{}, fmt.Errorf("invalid amount: %w", err)
	}
	return New(d, DefaultCurrency), nil
}

func Zero() Money {
	return New(decimal.Zero, DefaultCurrency)
}

func (m Money) Add(other Money) Money {
	m.assertSameCurrency(other)
	return New(m.Amount.Add(other.Amount), m.Currency)
}

func (m Money) Sub(other Money) Money {
	m.assertSameCurrency(other)
	return New(m.Amount.Sub(other.Amount), m.Currency)
}

func (m Money) Mul(n int64) Money {
	return New(m.Amount.Mul(decimal.NewFromInt(n)), m.Currency)
}

func (m Money) MulDecimal(d decimal.Decimal) Money {
	return New(m.Amount.Mul(d), m.Currency)
}

func (m Money) Div(n int64) Money {
	return New(m.Amount.Div(decimal.NewFromInt(n)), m.Currency)
}

func (m Money) Equal(other Money) bool {
	return m.Currency == other.Currency && m.Amount.Equal(other.Amount)
}

func (m Money) GreaterThan(other Money) bool {
	m.assertSameCurrency(other)
	return m.Amount.GreaterThan(other.Amount)
}

func (m Money) GreaterThanOrEqual(other Money) bool {
	m.assertSameCurrency(other)
	return m.Amount.GreaterThanOrEqual(other.Amount)
}

func (m Money) LessThan(other Money) bool {
	m.assertSameCurrency(other)
	return m.Amount.LessThan(other.Amount)
}

func (m Money) LessThanOrEqual(other Money) bool {
	m.assertSameCurrency(other)
	return m.Amount.LessThanOrEqual(other.Amount)
}

func (m Money) IsNegative() bool {
	return m.Amount.IsNegative()
}

func (m Money) IsPositive() bool {
	return m.Amount.IsPositive()
}

func (m Money) IsZero() bool {
	return m.Amount.IsZero()
}

func (m Money) Abs() Money {
	return New(m.Amount.Abs(), m.Currency)
}

func (m Money) Neg() Money {
	return New(m.Amount.Neg(), m.Currency)
}

func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.Amount.StringFixed(precision), m.Currency)
}

func (m Money) ToMinorUnits() int64 {
	multiplier := decimal.NewFromInt(100)
	return m.Amount.Mul(multiplier).IntPart()
}

func FromMinorUnits(amount int64, currency string) Money {
	d := decimal.NewFromInt(amount).Div(decimal.NewFromInt(100))
	return New(d, currency)
}

func (m Money) MarshalText() ([]byte, error) {
	return []byte(m.Amount.String()), nil
}

func (m *Money) UnmarshalText(data []byte) error {
	d, err := decimal.NewFromString(string(data))
	if err != nil {
		return err
	}
	m.Amount = d.Round(precision)
	m.Currency = DefaultCurrency
	return nil
}

func (m Money) assertSameCurrency(other Money) {
	if m.Currency != other.Currency {
		panic(fmt.Sprintf("currency mismatch: %s vs %s", m.Currency, other.Currency))
	}
}
