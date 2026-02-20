package repo

import (
	"context"
	"fmt"
	"math/big"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/app/product/contract"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/models/m_product"
)

type ProductRepo struct {
	db *spanner.Client
}

func NewProductRepo(db *spanner.Client) *ProductRepo {
	return &ProductRepo{db: db}
}

// GetByID loads a product from Spanner by its ID.
func (r *ProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	row, err := r.db.Single().ReadRow(ctx, m_product.Table,
		spanner.Key{id},
		[]string{
			m_product.ProductID,
			m_product.Name,
			m_product.Description,
			m_product.Category,
			m_product.BasePriceNumerator,
			m_product.BasePriceDenominator,
			m_product.DiscountPercent,
			m_product.DiscountStartDate,
			m_product.DiscountEndDate,
			m_product.Status,
			m_product.CreatedAt,
			m_product.UpdatedAt,
			m_product.ArchivedAt,
		},
	)
	if err != nil {
		if spanner.ErrCode(err) == 5 { // codes.NotFound
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}

	var pr m_product.ProductRow
	if err := row.ToStruct(&pr); err != nil {
		return nil, fmt.Errorf("GetByID decode: %w", err)
	}

	return pr.ToDomain()
}

// InsertMut returns a Spanner Mutation for a full INSERT of a new product.
func (r *ProductRepo) InsertMut(p *domain.Product) *spanner.Mutation {
	row := map[string]any{
		m_product.ProductID:            p.ID(),
		m_product.Name:                 p.Name(),
		m_product.Description:          p.Description(),
		m_product.Category:             p.Category(),
		m_product.BasePriceNumerator:   p.BasePrice().Amount(),
		m_product.BasePriceDenominator: int64(1),
		m_product.Status:               string(p.Status()),
		m_product.CreatedAt:            spanner.CommitTimestamp,
		m_product.UpdatedAt:            spanner.CommitTimestamp,
	}

	if d := p.Discount(); d != nil {
		var rat big.Rat
		rat.SetFloat64(d.PercentageFloat64())
		row[m_product.DiscountPercent] = rat
		row[m_product.DiscountStartDate] = d.StartsAt()
		row[m_product.DiscountEndDate] = d.EndsAt()
	}

	return spanner.InsertMap(m_product.Table, row)
}

// UpdateMut returns a Spanner Mutation containing only the dirty fields of a product.
// Returns nil when nothing has changed.
func (r *ProductRepo) UpdateMut(p *domain.Product) *spanner.Mutation {
	updates := map[string]any{
		m_product.ProductID: p.ID(),
		m_product.UpdatedAt: spanner.CommitTimestamp,
	}

	c := p.Changes()

	if c.Dirty(domain.FieldName) {
		updates[m_product.Name] = p.Name()
	}
	if c.Dirty(domain.FieldDescription) {
		updates[m_product.Description] = p.Description()
	}
	if c.Dirty(domain.FieldCategory) {
		updates[m_product.Category] = p.Category()
	}
	if c.Dirty(domain.FieldBasePrice) {
		updates[m_product.BasePriceNumerator] = p.BasePrice().Amount()
		updates[m_product.BasePriceDenominator] = int64(1)
	}
	if c.Dirty(domain.FieldStatus) {
		updates[m_product.Status] = string(p.Status())
	}
	if c.Dirty(domain.FieldDiscount) {
		if d := p.Discount(); d != nil {
			var rat big.Rat
			rat.SetFloat64(d.PercentageFloat64())
			updates[m_product.DiscountPercent] = rat
			updates[m_product.DiscountStartDate] = d.StartsAt()
			updates[m_product.DiscountEndDate] = d.EndsAt()
		} else {
			// Discount removed — clear all discount columns
			updates[m_product.DiscountPercent] = nil
			updates[m_product.DiscountStartDate] = nil
			updates[m_product.DiscountEndDate] = nil
		}
	}

	// Only ProductID + UpdatedAt → nothing actually changed
	if len(updates) == 2 {
		return nil
	}

	return spanner.UpdateMap(m_product.Table, updates)
}

// ListActive returns all active products, optionally filtered by category, with pagination.
func (r *ProductRepo) ListActive(ctx context.Context, filter contract.ListProductsFilter, page contract.Page) ([]*domain.Product, error) {
	stmt := spanner.Statement{
		SQL: `SELECT ` + allColumns + ` FROM ` + m_product.Table + `
		      WHERE ` + m_product.Status + ` = 'active'`,
	}

	if filter.Category != nil {
		stmt.SQL += " AND " + m_product.Category + " = @category"
		stmt.Params = map[string]any{"category": *filter.Category}
	}

	limit := page.Limit
	if limit <= 0 {
		limit = 20
	}
	stmt.SQL += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, page.Offset)

	var products []*domain.Product
	err := r.db.Single().Query(ctx, stmt).Do(func(row *spanner.Row) error {
		var pr m_product.ProductRow
		if err := row.ToStruct(&pr); err != nil {
			return fmt.Errorf("ListActive decode: %w", err)
		}
		p, err := pr.ToDomain()
		if err != nil {
			return err
		}
		products = append(products, p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ListActive: %w", err)
	}
	return products, nil
}

// allColumns is the full column list for SELECT queries.
const allColumns = `` +
	m_product.ProductID + `, ` +
	m_product.Name + `, ` +
	m_product.Description + `, ` +
	m_product.Category + `, ` +
	m_product.BasePriceNumerator + `, ` +
	m_product.BasePriceDenominator + `, ` +
	m_product.DiscountPercent + `, ` +
	m_product.DiscountStartDate + `, ` +
	m_product.DiscountEndDate + `, ` +
	m_product.Status + `, ` +
	m_product.CreatedAt + `, ` +
	m_product.UpdatedAt + `, ` +
	m_product.ArchivedAt
