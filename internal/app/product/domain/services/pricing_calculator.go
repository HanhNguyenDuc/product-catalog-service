package services

import (
	"strconv"
	"time"

	"github.com/product-catalog-service/internal/app/product/domain"
)

// PricingCalculator is a domain service that handles price computation logic.
// It is stateless and depends only on domain value objects (Money, Discount).
type PricingCalculator struct{}

// NewPricingCalculator returns a new PricingCalculator.
func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

// EffectivePrice returns the price a customer would pay for a product at a given point in time.
// If the product has a valid discount at 'now', the discounted price is returned.
// Otherwise, the base price is returned unchanged.
func (pc *PricingCalculator) EffectivePrice(basePrice *domain.Money, discount *domain.Discount, now time.Time) (*domain.Money, error) {
	if basePrice == nil {
		return nil, domain.ErrProductBasePriceRequired
	}

	if discount == nil || !discount.IsValidAt(now) {
		return basePrice, nil
	}

	pct, err := strconv.ParseFloat(discount.Percentage(), 64)
	if err != nil {
		return nil, domain.ErrDiscountInvalidPercentage
	}

	return basePrice.ApplyPercentageDiscount(pct)
}

// DiscountAmount returns the absolute monetary value saved by the discount at a given time.
// Returns zero money (same currency as basePrice) when there is no active discount.
func (pc *PricingCalculator) DiscountAmount(basePrice *domain.Money, discount *domain.Discount, now time.Time) (*domain.Money, error) {
	if basePrice == nil {
		return nil, domain.ErrProductBasePriceRequired
	}

	zero, err := domain.NewMoney(0, basePrice.Currency())
	if err != nil {
		return nil, err
	}

	if discount == nil || !discount.IsValidAt(now) {
		return zero, nil
	}

	effective, err := pc.EffectivePrice(basePrice, discount, now)
	if err != nil {
		return nil, err
	}

	return basePrice.Subtract(effective)
}

// IsDiscounted returns true when the product has a valid discount at the given time.
func (pc *PricingCalculator) IsDiscounted(discount *domain.Discount, now time.Time) bool {
	return discount != nil && discount.IsValidAt(now)
}
