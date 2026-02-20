package getproduct

import "time"

// ProductDTO is the read model returned by the GetProduct query.
type ProductDTO struct {
	ID             string
	Name           string
	Description    string
	Category       string
	Status         string
	BasePrice      MoneyDTO
	EffectivePrice MoneyDTO
	Discount       *DiscountDTO // nil when no active discount
}

// MoneyDTO is a flat representation of a monetary amount.
type MoneyDTO struct {
	Amount   int64
	Currency string
}

// DiscountDTO contains the discount details for a product.
type DiscountDTO struct {
	Percentage string
	StartsAt   time.Time
	EndsAt     time.Time
	IsActive   bool
}
