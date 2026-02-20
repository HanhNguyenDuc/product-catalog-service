package listproducts

import "time"

// ProductSummaryDTO is the lightweight read model returned by the ListProducts query.
// It intentionally omits heavy fields (e.g. Description) to keep list responses compact.
type ProductSummaryDTO struct {
	ID             string
	Name           string
	Category       string
	Status         string
	BasePrice      MoneyDTO
	EffectivePrice MoneyDTO
	IsDiscounted   bool
	DiscountEndsAt *time.Time // nil when no active discount
}

// MoneyDTO is a flat representation of a monetary amount.
type MoneyDTO struct {
	Amount   int64
	Currency string
}

// ListProductsRequest carries pagination and filter parameters.
type ListProductsRequest struct {
	Category *string // nil = all categories
	Limit    int     // max items per page; defaults to 20
	Offset   int     // 0-based offset for pagination
}

// ListProductsResponse wraps the result slice.
type ListProductsResponse struct {
	Items      []*ProductSummaryDTO
	TotalCount int // total matching rows (for pagination UI)
}
