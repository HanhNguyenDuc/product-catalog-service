package domain

import (
	"strconv"
	"time"
)

// Discount is a value object belonging to the Product aggregate.
// It has no identity of its own — it exists only in the context of a Product.
type Discount struct {
	percentage string // stored as string to preserve exact decimal representation
	startsAt   time.Time
	endsAt     time.Time
}

// NewDiscount creates and validates a new Discount.
func NewDiscount(percentage string, startsAt, endsAt time.Time) (*Discount, error) {
	pct, err := strconv.ParseFloat(percentage, 64)
	if err != nil || pct < 0 || pct > 100 {
		return nil, ErrDiscountInvalidPercentage
	}
	if !endsAt.After(startsAt) {
		return nil, ErrDiscountInvalidPeriod
	}
	return &Discount{
		percentage: percentage,
		startsAt:   startsAt,
		endsAt:     endsAt,
	}, nil
}

// Accessors

func (d *Discount) Percentage() string  { return d.percentage }
func (d *Discount) StartsAt() time.Time { return d.startsAt }
func (d *Discount) EndsAt() time.Time   { return d.endsAt }

// IsValidAt returns true when now falls within [startsAt, endsAt).
func (d *Discount) IsValidAt(now time.Time) bool {
	return !now.Before(d.startsAt) && now.Before(d.endsAt)
}

// IsExpired returns true when now is at or after endsAt.
func (d *Discount) IsExpired(now time.Time) bool {
	return !now.Before(d.endsAt)
}

// IsUpcoming returns true when the discount has not started yet.
func (d *Discount) IsUpcoming(now time.Time) bool {
	return now.Before(d.startsAt)
}

// PercentageFloat64 returns the parsed percentage as float64.
func (d *Discount) PercentageFloat64() float64 {
	pct, _ := strconv.ParseFloat(d.percentage, 64)
	return pct
}
