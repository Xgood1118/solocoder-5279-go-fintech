package money

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	m := New(decimal.NewFromInt(100), "CNY")
	assert.Equal(t, "100.00", m.Amount.StringFixed(2))
	assert.Equal(t, "CNY", m.Currency)
}

func TestNewFromInt(t *testing.T) {
	m := NewFromInt(100, "CNY")
	assert.Equal(t, "100.00", m.Amount.StringFixed(2))
}

func TestNewFromString(t *testing.T) {
	m, err := NewFromString("123.45")
	assert.NoError(t, err)
	assert.Equal(t, "123.45", m.Amount.StringFixed(2))
	assert.Equal(t, "CNY", m.Currency)

	_, err = NewFromString("invalid")
	assert.Error(t, err)
}

func TestZero(t *testing.T) {
	m := Zero()
	assert.True(t, m.IsZero())
}

func TestAdd(t *testing.T) {
	a := NewFromInt(100, "CNY")
	b := NewFromInt(50, "CNY")
	c := a.Add(b)
	assert.Equal(t, "150.00", c.Amount.StringFixed(2))
}

func TestSub(t *testing.T) {
	a := NewFromInt(100, "CNY")
	b := NewFromInt(30, "CNY")
	c := a.Sub(b)
	assert.Equal(t, "70.00", c.Amount.StringFixed(2))
}

func TestMul(t *testing.T) {
	a := NewFromInt(10, "CNY")
	b := a.Mul(5)
	assert.Equal(t, "50.00", b.Amount.StringFixed(2))
}

func TestComparison(t *testing.T) {
	a := NewFromInt(100, "CNY")
	b := NewFromInt(50, "CNY")
	c := NewFromInt(100, "CNY")

	assert.True(t, a.GreaterThan(b))
	assert.True(t, b.LessThan(a))
	assert.True(t, a.Equal(c))
	assert.True(t, a.GreaterThanOrEqual(c))
	assert.True(t, a.LessThanOrEqual(c))
}

func TestIsNegative(t *testing.T) {
	pos := NewFromInt(100, "CNY")
	neg := NewFromInt(-50, "CNY")
	zero := Zero()

	assert.False(t, pos.IsNegative())
	assert.True(t, neg.IsNegative())
	assert.False(t, zero.IsNegative())
	assert.True(t, zero.IsZero())
}

func TestAbs(t *testing.T) {
	m := NewFromInt(-100, "CNY")
	abs := m.Abs()
	assert.Equal(t, "100.00", abs.Amount.StringFixed(2))
}

func TestNeg(t *testing.T) {
	m := NewFromInt(100, "CNY")
	neg := m.Neg()
	assert.Equal(t, "-100.00", neg.Amount.StringFixed(2))
}

func TestString(t *testing.T) {
	m := NewFromInt(1234, "CNY")
	assert.Equal(t, "1234.00 CNY", m.String())
}

func TestToMinorUnits(t *testing.T) {
	m, _ := NewFromString("123.45")
	assert.Equal(t, int64(12345), m.ToMinorUnits())
}

func TestFromMinorUnits(t *testing.T) {
	m := FromMinorUnits(12345, "CNY")
	assert.Equal(t, "123.45", m.Amount.StringFixed(2))
}
