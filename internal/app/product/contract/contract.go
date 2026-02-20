package contract

import (
	"context"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/app/product/domain"
)

// ProductRepository is the read/write contract for the Product aggregate.
type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	InsertMut(p *domain.Product) *spanner.Mutation
	UpdateMut(p *domain.Product) *spanner.Mutation
}

// EventRepository is the write-only contract for persisting domain events to the outbox.
type EventRepository interface {
	InsertMut(event domain.DomainEvent) *spanner.Mutation
}

// ListProductsFilter holds optional filter parameters for listing products.
type ListProductsFilter struct {
	Category *string // nil = no filter
}

// Page holds pagination parameters.
type Page struct {
	Limit  int
	Offset int
}

// QueryRepository is the read-only contract for product queries.
type QueryRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	ListActive(ctx context.Context, filter ListProductsFilter, page Page) ([]*domain.Product, error)
}
