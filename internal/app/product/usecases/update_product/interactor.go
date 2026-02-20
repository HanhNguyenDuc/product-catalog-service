package updateproduct

import (
	"context"

	"github.com/product-catalog-service/common"
	"github.com/product-catalog-service/common/commitplanner"
	"github.com/product-catalog-service/internal/app/product/contract"
)

type UpdateProductInteractor struct {
	committer commitplanner.Applier
	repo      contract.ProductRepository
	eventRepo contract.EventRepository
	ticker    common.Ticker
}

func NewUpdateProductInteractor(committer commitplanner.Applier, repo contract.ProductRepository, eventRepo contract.EventRepository, ticker common.Ticker) *UpdateProductInteractor {
	return &UpdateProductInteractor{committer: committer, repo: repo, eventRepo: eventRepo, ticker: ticker}
}

type UpdateProductRequest struct {
	ProductID   string
	Name        *string
	Description *string
	Category    *string
}

func (it *UpdateProductInteractor) Execute(ctx context.Context, req *UpdateProductRequest) error {
	product, err := it.repo.GetByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	if req.Name != nil {
		if err := product.SetName(*req.Name); err != nil {
			return err
		}
	}
	if req.Description != nil {
		product.SetDescription(*req.Description)
	}
	if req.Category != nil {
		product.SetCategory(*req.Category)
	}

	product.RecordUpdate(it.ticker.Now())

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
