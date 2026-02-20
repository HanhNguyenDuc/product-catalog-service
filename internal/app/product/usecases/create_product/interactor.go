package createproduct

import (
	"context"

	"github.com/product-catalog-service/common"
	"github.com/product-catalog-service/common/commitplanner"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/app/product/contract"
)

const basePrice = 100

type CreateProductInteractor struct {
	committer commitplanner.Applier
	repo      contract.ProductRepository
	eventRepo contract.EventRepository
	ticker    common.Ticker
}

func NewCreateProductInteractor(committer commitplanner.Applier, repo contract.ProductRepository, eventRepo contract.EventRepository, ticker common.Ticker) *CreateProductInteractor {
	return &CreateProductInteractor{committer: committer, repo: repo, eventRepo: eventRepo, ticker: ticker}
}

type CreateProductRequest struct {
	Name        string
	Description string
	Category    string
}

func (it *CreateProductInteractor) Execute(ctx context.Context, req *CreateProductRequest) (string, error) {
	money, err := domain.NewMoney(basePrice, "USD")
	if err != nil {
		return "", err
	}
	product, err := domain.NewProduct(req.Name, req.Description, req.Category, money, it.ticker.Now())
	if err != nil {
		return "", err
	}

	plan := commitplanner.NewPlan()

	if mut := it.repo.InsertMut(product); mut != nil {
		plan.Add(mut)
	}

	for _, event := range product.Events() {
		if mut := it.eventRepo.InsertMut(event); mut != nil {
			plan.Add(mut)
		}
	}

	if err := it.committer.Apply(ctx, plan); err != nil {
		return "", err
	}

	return product.ID(), nil
}
