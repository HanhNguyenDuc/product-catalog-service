// Package integration_test tests the application layer (use cases and queries)
// using in-memory mocks for infrastructure. No running service is required.
package integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"cloud.google.com/go/spanner"

	"github.com/product-catalog-service/common"
	"github.com/product-catalog-service/common/commitplanner"
	"github.com/product-catalog-service/internal/app/product/contract"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/app/product/domain/services"
	getproduct "github.com/product-catalog-service/internal/app/product/queries/get_product"
	listproducts "github.com/product-catalog-service/internal/app/product/queries/list_products"
	activateproduct "github.com/product-catalog-service/internal/app/product/usecases/activate_product"
	applydiscount "github.com/product-catalog-service/internal/app/product/usecases/apply_discount"
	createproduct "github.com/product-catalog-service/internal/app/product/usecases/create_product"
	deactivateproduct "github.com/product-catalog-service/internal/app/product/usecases/deactivate_product"
	removediscount "github.com/product-catalog-service/internal/app/product/usecases/remove_discount"
	updateproduct "github.com/product-catalog-service/internal/app/product/usecases/update_product"
)

// ────────────────────────────────────────────────────────────────────────────
// In-memory mocks
// ────────────────────────────────────────────────────────────────────────────

// fixedTicker returns a deterministic time for all Now() calls.
type fixedTicker struct{ t time.Time }

func (f fixedTicker) Now() time.Time { return f.t }

func newTicker(t time.Time) common.Ticker { return fixedTicker{t} }

// mockCommitter records whether Apply was called and can be made to return an error.
type mockCommitter struct {
	applied bool
	err     error
}

func (m *mockCommitter) Apply(_ context.Context, _ *commitplanner.Plan) error {
	if m.err != nil {
		return m.err
	}
	m.applied = true
	return nil
}

// inMemoryProductRepo is a simple map-backed implementation of both
// contract.ProductRepository and contract.QueryRepository.
type inMemoryProductRepo struct {
	store map[string]*domain.Product
}

func newInMemoryProductRepo() *inMemoryProductRepo {
	return &inMemoryProductRepo{store: make(map[string]*domain.Product)}
}

func (r *inMemoryProductRepo) GetByID(_ context.Context, id string) (*domain.Product, error) {
	p, ok := r.store[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}

func (r *inMemoryProductRepo) InsertMut(p *domain.Product) *spanner.Mutation {
	// In the e2e flow the committer calls Apply, but our mockCommitter doesn't
	// touch Spanner. We persist directly here so the query side can find the product.
	r.store[p.ID()] = p
	return nil // nil mutations are skipped by plan.Add callers
}

func (r *inMemoryProductRepo) UpdateMut(p *domain.Product) *spanner.Mutation {
	r.store[p.ID()] = p
	return nil
}

func (r *inMemoryProductRepo) ListActive(_ context.Context, filter contract.ListProductsFilter, page contract.Page) ([]*domain.Product, error) {
	var result []*domain.Product
	for _, p := range r.store {
		if p.Status() != domain.ProductStatusActive {
			continue
		}
		if filter.Category != nil && p.Category() != *filter.Category {
			continue
		}
		result = append(result, p)
	}
	// Apply offset + limit
	if page.Offset >= len(result) {
		return []*domain.Product{}, nil
	}
	result = result[page.Offset:]
	if page.Limit > 0 && len(result) > page.Limit {
		result = result[:page.Limit]
	}
	return result, nil
}

// inMemoryEventRepo just discards mutations (no Spanner in e2e).
type inMemoryEventRepo struct {
	events []domain.DomainEvent
}

func (r *inMemoryEventRepo) InsertMut(event domain.DomainEvent) *spanner.Mutation {
	r.events = append(r.events, event)
	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

var (
	baseTime = time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)
	pricing  = services.NewPricingCalculator()
)

func buildDeps(t *testing.T) (
	repo *inMemoryProductRepo,
	eventRepo *inMemoryEventRepo,
	committer *mockCommitter,
	ticker common.Ticker,
) {
	t.Helper()
	return newInMemoryProductRepo(), &inMemoryEventRepo{}, &mockCommitter{}, newTicker(baseTime)
}

// createOne is a test utility that runs CreateProduct and returns the new product ID.
func createOne(t *testing.T, repo *inMemoryProductRepo, eventRepo *inMemoryEventRepo, committer *mockCommitter, ticker common.Ticker, name, category string) string {
	t.Helper()
	it := createproduct.NewCreateProductInteractor(committer, repo, eventRepo, ticker)
	id, err := it.Execute(context.Background(), &createproduct.CreateProductRequest{
		Name:        name,
		Description: "a product",
		Category:    category,
	})
	if err != nil {
		t.Fatalf("createOne: %v", err)
	}
	return id
}

// ────────────────────────────────────────────────────────────────────────────
// CreateProduct
// ────────────────────────────────────────────────────────────────────────────

func TestCreateProduct_Success(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	it := createproduct.NewCreateProductInteractor(committer, repo, eventRepo, ticker)

	id, err := it.Execute(context.Background(), &createproduct.CreateProductRequest{
		Name:        "Laptop",
		Description: "High-end laptop",
		Category:    "electronics",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty product ID")
	}
	if _, ok := repo.store[id]; !ok {
		t.Fatal("product not persisted in repo")
	}
	if len(eventRepo.events) == 0 {
		t.Fatal("expected at least one domain event")
	}
}

func TestCreateProduct_EmptyName(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	it := createproduct.NewCreateProductInteractor(committer, repo, eventRepo, ticker)

	_, err := it.Execute(context.Background(), &createproduct.CreateProductRequest{
		Name:     "",
		Category: "electronics",
	})

	if !errors.Is(err, domain.ErrProductNameRequired) {
		t.Fatalf("expected ErrProductNameRequired, got %v", err)
	}
}

func TestCreateProduct_CommitterError(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	committer.err = errors.New("spanner unavailable")
	it := createproduct.NewCreateProductInteractor(committer, repo, eventRepo, ticker)

	_, err := it.Execute(context.Background(), &createproduct.CreateProductRequest{
		Name:     "Laptop",
		Category: "electronics",
	})

	if err == nil {
		t.Fatal("expected error from committer, got nil")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// UpdateProduct
// ────────────────────────────────────────────────────────────────────────────

func TestUpdateProduct_Success(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Old Name", "electronics")
	committer.applied = false // reset after create

	newName := "New Name"
	it := updateproduct.NewUpdateProductInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &updateproduct.UpdateProductRequest{
		ProductID: id,
		Name:      &newName,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.store[id].Name() != newName {
		t.Fatalf("expected name %q, got %q", newName, repo.store[id].Name())
	}
}

func TestUpdateProduct_ProductNotFound(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	it := updateproduct.NewUpdateProductInteractor(committer, repo, eventRepo, ticker)

	err := it.Execute(context.Background(), &updateproduct.UpdateProductRequest{
		ProductID: "non-existent-id",
	})

	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ApplyDiscount
// ────────────────────────────────────────────────────────────────────────────

func TestApplyDiscount_Success(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	startsAt := baseTime.Add(-time.Hour)
	endsAt := baseTime.Add(24 * time.Hour)

	it := applydiscount.NewApplyDiscountInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &applydiscount.ApplyDiscountRequest{
		ProductID:  id,
		Percentage: "10",
		StartsAt:   startsAt,
		EndsAt:     endsAt,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.store[id].Discount() == nil {
		t.Fatal("expected discount to be set")
	}
}

func TestApplyDiscount_InvalidPeriod(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	it := applydiscount.NewApplyDiscountInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &applydiscount.ApplyDiscountRequest{
		ProductID:  id,
		Percentage: "10",
		StartsAt:   baseTime.Add(time.Hour), // starts AFTER endsAt
		EndsAt:     baseTime.Add(-time.Hour),
	})

	if !errors.Is(err, domain.ErrDiscountInvalidPeriod) {
		t.Fatalf("expected ErrDiscountInvalidPeriod, got %v", err)
	}
}

func TestApplyDiscount_NotActive(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	// Deactivate the product first
	deactivateTicker := newTicker(baseTime)
	_ = deactivateTicker
	p := repo.store[id]
	_ = p.Deactivate(baseTime)

	it := applydiscount.NewApplyDiscountInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &applydiscount.ApplyDiscountRequest{
		ProductID:  id,
		Percentage: "10",
		StartsAt:   baseTime.Add(-time.Hour),
		EndsAt:     baseTime.Add(24 * time.Hour),
	})

	if !errors.Is(err, domain.ErrProductNotActive) {
		t.Fatalf("expected ErrProductNotActive, got %v", err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ActivateProduct
// ────────────────────────────────────────────────────────────────────────────

func TestActivateProduct_AlreadyActive(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	// Activate an already-active product → should be a no-op (idempotent)
	it := activateproduct.NewActivateProductInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &activateproduct.ActivateProductRequest{ProductID: id})

	if err != nil {
		t.Fatalf("expected no error for idempotent activate, got %v", err)
	}
}

func TestActivateProduct_FromInactive(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Lamp", "home")

	// Deactivate first
	_ = repo.store[id].Deactivate(baseTime)

	it := activateproduct.NewActivateProductInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &activateproduct.ActivateProductRequest{ProductID: id})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.store[id].Status() != domain.ProductStatusActive {
		t.Fatal("expected product to be active")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// GetProduct query
// ────────────────────────────────────────────────────────────────────────────

func TestGetProduct_WithNoDiscount(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Headphones", "electronics")

	q := getproduct.NewGetProductQuery(repo, pricing, ticker)
	dto, err := q.Execute(context.Background(), &getproduct.GetProductRequest{ProductID: id})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dto.ID != id {
		t.Fatalf("expected ID %q, got %q", id, dto.ID)
	}
	if dto.Discount != nil {
		t.Fatal("expected no discount")
	}
	if dto.EffectivePrice.Amount != dto.BasePrice.Amount {
		t.Fatal("effective price should equal base price when no discount")
	}
}

func TestGetProduct_WithActiveDiscount(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Keyboard", "electronics")

	// Apply 20% discount
	startsAt := baseTime.Add(-time.Hour)
	endsAt := baseTime.Add(24 * time.Hour)
	it := applydiscount.NewApplyDiscountInteractor(committer, repo, eventRepo, ticker)
	_ = it.Execute(context.Background(), &applydiscount.ApplyDiscountRequest{
		ProductID:  id,
		Percentage: "20",
		StartsAt:   startsAt,
		EndsAt:     endsAt,
	})

	q := getproduct.NewGetProductQuery(repo, pricing, ticker)
	dto, err := q.Execute(context.Background(), &getproduct.GetProductRequest{ProductID: id})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dto.Discount == nil || !dto.Discount.IsActive {
		t.Fatal("expected active discount on DTO")
	}
	if dto.EffectivePrice.Amount >= dto.BasePrice.Amount {
		t.Fatal("expected effective price to be lower than base price")
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	repo, _, _, ticker := buildDeps(t)
	q := getproduct.NewGetProductQuery(repo, pricing, ticker)

	_, err := q.Execute(context.Background(), &getproduct.GetProductRequest{ProductID: "ghost"})

	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ListProducts query
// ────────────────────────────────────────────────────────────────────────────

func TestListProducts_All(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")
	createOne(t, repo, eventRepo, committer, ticker, "Mouse", "electronics")
	createOne(t, repo, eventRepo, committer, ticker, "Desk", "furniture")

	q := listproducts.NewListProductsQuery(repo, pricing, ticker)
	resp, err := q.Execute(context.Background(), &listproducts.ListProductsRequest{Limit: 10})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(resp.Items))
	}
}

func TestListProducts_FilterByCategory(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")
	createOne(t, repo, eventRepo, committer, ticker, "Mouse", "electronics")
	createOne(t, repo, eventRepo, committer, ticker, "Desk", "furniture")

	cat := "electronics"
	q := listproducts.NewListProductsQuery(repo, pricing, ticker)
	resp, err := q.Execute(context.Background(), &listproducts.ListProductsRequest{
		Category: &cat,
		Limit:    10,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 electronics, got %d", len(resp.Items))
	}
	for _, item := range resp.Items {
		if item.Category != "electronics" {
			t.Fatalf("unexpected category %q", item.Category)
		}
	}
}

func TestListProducts_Pagination(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	for i := 0; i < 5; i++ {
		createOne(t, repo, eventRepo, committer, ticker, "Product", "misc")
	}

	q := listproducts.NewListProductsQuery(repo, pricing, ticker)

	page1, err := q.Execute(context.Background(), &listproducts.ListProductsRequest{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("page1 error: %v", err)
	}
	if len(page1.Items) != 2 {
		t.Fatalf("expected 2 items on page1, got %d", len(page1.Items))
	}

	page2, err := q.Execute(context.Background(), &listproducts.ListProductsRequest{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("page2 error: %v", err)
	}
	if len(page2.Items) != 2 {
		t.Fatalf("expected 2 items on page2, got %d", len(page2.Items))
	}
}

func TestListProducts_ExcludesInactive(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")
	_ = repo.store[id].Deactivate(baseTime) // make it inactive
	createOne(t, repo, eventRepo, committer, ticker, "Mouse", "electronics")

	q := listproducts.NewListProductsQuery(repo, pricing, ticker)
	resp, err := q.Execute(context.Background(), &listproducts.ListProductsRequest{Limit: 10})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 active item, got %d", len(resp.Items))
	}
}

// ────────────────────────────────────────────────────────────────────────────
// DeactivateProduct
// ────────────────────────────────────────────────────────────────────────────

func TestDeactivateProduct_Success(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	it := deactivateproduct.NewDeactivateProductInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &deactivateproduct.DeactivateProductRequest{ProductID: id})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.store[id].Status() != domain.ProductStatusInactive {
		t.Fatal("expected product to be inactive after deactivation")
	}
}

func TestDeactivateProduct_Idempotent(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")
	_ = repo.store[id].Deactivate(baseTime) // already inactive

	it := deactivateproduct.NewDeactivateProductInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &deactivateproduct.DeactivateProductRequest{ProductID: id})

	if err != nil {
		t.Fatalf("expected no error for idempotent deactivate, got %v", err)
	}
	if repo.store[id].Status() != domain.ProductStatusInactive {
		t.Fatal("expected product to remain inactive")
	}
}

func TestDeactivateProduct_RemovesDiscount(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	// Apply a discount first
	applyIt := applydiscount.NewApplyDiscountInteractor(committer, repo, eventRepo, ticker)
	_ = applyIt.Execute(context.Background(), &applydiscount.ApplyDiscountRequest{
		ProductID:  id,
		Percentage: "15",
		StartsAt:   baseTime.Add(-time.Hour),
		EndsAt:     baseTime.Add(24 * time.Hour),
	})

	it := deactivateproduct.NewDeactivateProductInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &deactivateproduct.DeactivateProductRequest{ProductID: id})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.store[id].Discount() != nil {
		t.Fatal("expected discount to be cleared on deactivation")
	}
}

func TestDeactivateProduct_NotFound(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	it := deactivateproduct.NewDeactivateProductInteractor(committer, repo, eventRepo, ticker)

	err := it.Execute(context.Background(), &deactivateproduct.DeactivateProductRequest{ProductID: "ghost"})

	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// RemoveDiscount
// ────────────────────────────────────────────────────────────────────────────

func TestRemoveDiscount_Success(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	// Apply discount first
	applyIt := applydiscount.NewApplyDiscountInteractor(committer, repo, eventRepo, ticker)
	_ = applyIt.Execute(context.Background(), &applydiscount.ApplyDiscountRequest{
		ProductID:  id,
		Percentage: "10",
		StartsAt:   baseTime.Add(-time.Hour),
		EndsAt:     baseTime.Add(24 * time.Hour),
	})

	it := removediscount.NewRemoveDiscountInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &removediscount.RemoveDiscountRequest{ProductID: id})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.store[id].Discount() != nil {
		t.Fatal("expected discount to be nil after removal")
	}
}

func TestRemoveDiscount_NoActiveDiscount(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	it := removediscount.NewRemoveDiscountInteractor(committer, repo, eventRepo, ticker)
	err := it.Execute(context.Background(), &removediscount.RemoveDiscountRequest{ProductID: id})

	if !errors.Is(err, domain.ErrNoActiveDiscount) {
		t.Fatalf("expected ErrNoActiveDiscount, got %v", err)
	}
}

func TestRemoveDiscount_NotFound(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	it := removediscount.NewRemoveDiscountInteractor(committer, repo, eventRepo, ticker)

	err := it.Execute(context.Background(), &removediscount.RemoveDiscountRequest{ProductID: "ghost"})

	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestRemoveDiscount_ThenGetProduct_PriceEqualsBase(t *testing.T) {
	repo, eventRepo, committer, ticker := buildDeps(t)
	id := createOne(t, repo, eventRepo, committer, ticker, "Laptop", "electronics")

	// Apply then remove discount
	applyIt := applydiscount.NewApplyDiscountInteractor(committer, repo, eventRepo, ticker)
	_ = applyIt.Execute(context.Background(), &applydiscount.ApplyDiscountRequest{
		ProductID:  id,
		Percentage: "30",
		StartsAt:   baseTime.Add(-time.Hour),
		EndsAt:     baseTime.Add(24 * time.Hour),
	})
	removeIt := removediscount.NewRemoveDiscountInteractor(committer, repo, eventRepo, ticker)
	_ = removeIt.Execute(context.Background(), &removediscount.RemoveDiscountRequest{ProductID: id})

	// After discount removal, effective price == base price
	q := getproduct.NewGetProductQuery(repo, pricing, ticker)
	dto, err := q.Execute(context.Background(), &getproduct.GetProductRequest{ProductID: id})
	if err != nil {
		t.Fatalf("expected no error on get, got %v", err)
	}
	if dto.EffectivePrice.Amount != dto.BasePrice.Amount {
		t.Fatalf("expected effective price %d == base price %d after discount removal",
			dto.EffectivePrice.Amount, dto.BasePrice.Amount)
	}
	if dto.Discount != nil {
		t.Fatal("expected no discount in DTO after removal")
	}
}
