package m_product

import (
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/app/product/domain"
)

// Table name for Spanner.
const Table = "products"

// ProductRow is the Spanner row representation of a product.
// It mirrors the products table schema 1-to-1.
type ProductRow struct {
	ProductID            string              `spanner:"product_id"`
	Name                 string              `spanner:"name"`
	Description          string              `spanner:"description"`
	Category             string              `spanner:"category"`
	BasePriceNumerator   int64               `spanner:"base_price_numerator"`
	BasePriceDenominator int64               `spanner:"base_price_denominator"`
	DiscountPercent      spanner.NullNumeric `spanner:"discount_percent"` // nullable → zero value when absent
	DiscountStartDate    spanner.NullTime    `spanner:"discount_start_date"`
	DiscountEndDate      spanner.NullTime    `spanner:"discount_end_date"`
	Status               string              `spanner:"status"`
	CreatedAt            time.Time           `spanner:"created_at"`
	UpdatedAt            time.Time           `spanner:"updated_at"`
	ArchivedAt           spanner.NullTime    `spanner:"archived_at"`
}

// ToDomain converts a ProductRow (from Spanner) to a domain.Product aggregate.
func (r *ProductRow) ToDomain() (*domain.Product, error) {
	basePrice, err := domain.NewMoney(r.BasePriceNumerator, "VND") // currency stored implicitly
	if err != nil {
		return nil, err
	}

	var discount *domain.Discount
	if r.DiscountPercent.Valid && r.DiscountStartDate.Valid && r.DiscountEndDate.Valid {
		f, _ := r.DiscountPercent.Numeric.Float64()
		pct := formatDecimal(f)
		discount, err = domain.NewDiscount(pct, r.DiscountStartDate.Time, r.DiscountEndDate.Time)
		if err != nil {
			return nil, err
		}
	}

	return domain.Reconstitute(
		r.ProductID,
		r.Name,
		r.Description,
		r.Category,
		basePrice,
		discount,
		domain.ProductStatus(r.Status),
	)
}

// formatDecimal converts a float64 percentage to its string representation.
func formatDecimal(f float64) string {
	return fmt.Sprintf("%g", f)
}
