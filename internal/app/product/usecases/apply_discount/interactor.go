package applydiscount

import (
	"context"
	"time"

	"github.com/product-catalog-service/common"
	"github.com/product-catalog-service/common/commitplanner"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/app/product/contract"
)

type ApplyDiscountInteractor struct {
	committer commitplanner.Applier
	repo      contract.ProductRepository
	eventRepo contract.EventRepository
	ticker    common.Ticker
}

func NewApplyDiscountInteractor(committer commitplanner.Applier, repo contract.ProductRepository, eventRepo contract.EventRepository, ticker common.Ticker) *ApplyDiscountInteractor {
	return &ApplyDiscountInteractor{committer: committer, repo: repo, eventRepo: eventRepo, ticker: ticker}
}

type ApplyDiscountRequest struct {
	ProductID  string
	Percentage string
	StartsAt   time.Time
	EndsAt     time.Time
}

func (it *ApplyDiscountInteractor) Execute(ctx context.Context, req *ApplyDiscountRequest) error {
	product, err := it.repo.GetByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	discount, err := domain.NewDiscount(req.Percentage, req.StartsAt, req.EndsAt)
	if err != nil {
		return err
	}

	if err := product.ApplyDiscount(discount, it.ticker.Now()); err != nil {
		return err
	}

	plan := commitplanner.NewPlan()

	if mut := it.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.Events() {
		if mut := it.eventRepo.InsertMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	return it.committer.Apply(ctx, plan)
}
