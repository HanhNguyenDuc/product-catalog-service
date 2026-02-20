package listproducts

import (
	"context"

	"github.com/product-catalog-service/common"
	"github.com/product-catalog-service/internal/app/product/domain/services"
	"github.com/product-catalog-service/internal/app/product/contract"
)

const defaultLimit = 20

// ListProductsQuery lists active products with optional category filter and pagination.
// It uses the PricingCalculator to compute the effective price for each product.
type ListProductsQuery struct {
	queryRepo contract.QueryRepository
	pricing   *services.PricingCalculator
	ticker    common.Ticker
}

func NewListProductsQuery(queryRepo contract.QueryRepository, pricing *services.PricingCalculator, ticker common.Ticker) *ListProductsQuery {
	return &ListProductsQuery{queryRepo: queryRepo, pricing: pricing, ticker: ticker}
}

func (q *ListProductsQuery) Execute(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLimit
	}

	products, err := q.queryRepo.ListActive(ctx,
		contract.ListProductsFilter{Category: req.Category},
		contract.Page{Limit: limit, Offset: req.Offset},
	)
	if err != nil {
		return nil, err
	}

	now := q.ticker.Now()
	items := make([]*ProductSummaryDTO, 0, len(products))

	for _, p := range products {
		effective, err := q.pricing.EffectivePrice(p.BasePrice(), p.Discount(), now)
		if err != nil {
			return nil, err
		}

		summary := &ProductSummaryDTO{
			ID:       p.ID(),
			Name:     p.Name(),
			Category: p.Category(),
			Status:   string(p.Status()),
			BasePrice: MoneyDTO{
				Amount:   p.BasePrice().Amount(),
				Currency: p.BasePrice().Currency(),
			},
			EffectivePrice: MoneyDTO{
				Amount:   effective.Amount(),
				Currency: effective.Currency(),
			},
			IsDiscounted: q.pricing.IsDiscounted(p.Discount(), now),
		}

		if d := p.Discount(); d != nil && d.IsValidAt(now) {
			endsAt := d.EndsAt()
			summary.DiscountEndsAt = &endsAt
		}

		items = append(items, summary)
	}

	return &ListProductsResponse{
		Items:      items,
		TotalCount: len(items), // Note: replace with a real COUNT query when DB is wired
	}, nil
}
