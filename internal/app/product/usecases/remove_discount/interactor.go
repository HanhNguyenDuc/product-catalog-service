package removediscount

import (
	"context"

	"github.com/product-catalog-service/common"
	"github.com/product-catalog-service/common/commitplanner"
	"github.com/product-catalog-service/internal/app/product/contract"
)

type RemoveDiscountInteractor struct {
	committer commitplanner.Applier
	repo      contract.ProductRepository
	eventRepo contract.EventRepository
	ticker    common.Ticker
}

func NewRemoveDiscountInteractor(committer commitplanner.Applier, repo contract.ProductRepository, eventRepo contract.EventRepository, ticker common.Ticker) *RemoveDiscountInteractor {
	return &RemoveDiscountInteractor{committer: committer, repo: repo, eventRepo: eventRepo, ticker: ticker}
}

type RemoveDiscountRequest struct {
	ProductID string
}

func (it *RemoveDiscountInteractor) Execute(ctx context.Context, req *RemoveDiscountRequest) error {
	product, err := it.repo.GetByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	if err := product.RemoveDiscount(it.ticker.Now()); err != nil {
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
