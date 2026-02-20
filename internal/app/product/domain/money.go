package domain

import (
	"fmt"
	"math"
)

// Money is an immutable value object representing a monetary amount.
type Money struct {
	// amount stored in the smallest currency unit (e.g. cents for USD)
	amount   int64
	currency string
}

// NewMoney creates a new Money value-object.
// amount must be expressed in the smallest currency unit (e.g. 1000 = $10.00 for USD).
func NewMoney(amount int64, currency string) (*Money, error) {
	if amount < 0 {
		return nil, ErrNegativeAmount
	}
	if len(currency) != 3 {
		return nil, ErrInvalidCurrency
	}
	return &Money{amount: amount, currency: currency}, nil
}

// MustNewMoney is like NewMoney but panics on error. Useful in tests / constants.
func MustNewMoney(amount int64, currency string) *Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

// Amount returns the raw amount in the smallest currency unit.
func (m *Money) Amount() int64 {
	return m.amount
}

// Currency returns the ISO-4217 currency code.
func (m *Money) Currency() string {
	return m.currency
}

// String returns a human-readable representation, e.g. "10.00 USD".
func (m *Money) String() string {
	major := m.amount / 100
	minor := m.amount % 100
	return fmt.Sprintf("%d.%02d %s", major, minor, m.currency)
}

// Equals returns true when both amount and currency are equal.
func (m *Money) Equals(other *Money) bool {
	if other == nil {
		return false
	}
	return m.amount == other.amount && m.currency == other.currency
}

// Add returns a new Money that is the sum of m and other.
func (m *Money) Add(other *Money) (*Money, error) {
	if err := m.sameCurrency(other); err != nil {
		return nil, err
	}
	return &Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Subtract returns a new Money that is m minus other.
// Returns ErrNegativeAmount if the result would be negative.
func (m *Money) Subtract(other *Money) (*Money, error) {
	if err := m.sameCurrency(other); err != nil {
		return nil, err
	}
	result := m.amount - other.amount
	if result < 0 {
		return nil, ErrNegativeAmount
	}
	return &Money{amount: result, currency: m.currency}, nil
}

// Multiply returns a new Money scaled by factor (rounded to nearest cent).
func (m *Money) Multiply(factor float64) (*Money, error) {
	if factor < 0 {
		return nil, ErrNegativeAmount
	}
	result := int64(math.Round(float64(m.amount) * factor))
	return &Money{amount: result, currency: m.currency}, nil
}

// ApplyPercentageDiscount returns a new Money after applying a percentage discount.
// percentage must be between 0 and 100.
func (m *Money) ApplyPercentageDiscount(percentage float64) (*Money, error) {
	if percentage < 0 || percentage > 100 {
		return nil, ErrInvalidDiscountAmount
	}
	discounted := float64(m.amount) * (1 - percentage/100)
	return &Money{amount: int64(math.Round(discounted)), currency: m.currency}, nil
}

// IsGreaterThan returns true when m > other.
func (m *Money) IsGreaterThan(other *Money) (bool, error) {
	if err := m.sameCurrency(other); err != nil {
		return false, err
	}
	return m.amount > other.amount, nil
}

// IsLessThan returns true when m < other.
func (m *Money) IsLessThan(other *Money) (bool, error) {
	if err := m.sameCurrency(other); err != nil {
		return false, err
	}
	return m.amount < other.amount, nil
}

// IsZero returns true when the amount is zero.
func (m *Money) IsZero() bool {
	return m.amount == 0
}

// sameCurrency is a helper that guards cross-currency operations.
func (m *Money) sameCurrency(other *Money) error {
	if other == nil {
		return ErrCurrencyMismatch
	}
	if m.currency != other.currency {
		return fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.currency, other.currency)
	}
	return nil
}
