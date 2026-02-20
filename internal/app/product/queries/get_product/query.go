package getproduct

import (
	"context"

	"github.com/product-catalog-service/common"
	"github.com/product-catalog-service/internal/app/product/contract"
	"github.com/product-catalog-service/internal/app/product/domain/services"
)

// GetProductQuery fetches a single product by ID and enriches it with the
// current effective price computed by the PricingCalculator domain service.
type GetProductQuery struct {
	queryRepo contract.QueryRepository
	pricing   *services.PricingCalculator
	ticker    common.Ticker
}

func NewGetProductQuery(queryRepo contract.QueryRepository, pricing *services.PricingCalculator, ticker common.Ticker) *GetProductQuery {
	return &GetProductQuery{queryRepo: queryRepo, pricing: pricing, ticker: ticker}
}

type GetProductRequest struct {
	ProductID string
}

func (q *GetProductQuery) Execute(ctx context.Context, req *GetProductRequest) (*ProductDTO, error) {
	product, err := q.queryRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	now := q.ticker.Now()

	effective, err := q.pricing.EffectivePrice(product.BasePrice(), product.Discount(), now)
	if err != nil {
		return nil, err
	}

	dto := &ProductDTO{
		ID:          product.ID(),
		Name:        product.Name(),
		Description: product.Description(),
		Category:    product.Category(),
		Status:      string(product.Status()),
		BasePrice: MoneyDTO{
			Amount:   product.BasePrice().Amount(),
			Currency: product.BasePrice().Currency(),
		},
		EffectivePrice: MoneyDTO{
			Amount:   effective.Amount(),
			Currency: effective.Currency(),
		},
	}

	if d := product.Discount(); d != nil {
		dto.Discount = &DiscountDTO{
			Percentage: d.Percentage(),
			StartsAt:   d.StartsAt(),
			EndsAt:     d.EndsAt(),
			IsActive:   d.IsValidAt(now),
		}
	}

	return dto, nil
}
