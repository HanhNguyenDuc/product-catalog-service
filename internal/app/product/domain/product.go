package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
)

const (
	FieldName        Field = "name"
	FieldDiscount    Field = "discount"
	FieldDescription Field = "description"
	FieldCategory    Field = "category"
	FieldBasePrice   Field = "base_price"
	FieldStatus      Field = "status"
)

// Product is the aggregate root of the product domain.
// All state mutations go through its methods, which enforce invariants and record domain events.
type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	changes     *Changes
	events      []DomainEvent
}

// ────────────────────────────────────────────────────────────────────────────
// Constructors
// ────────────────────────────────────────────────────────────────────────────

// NewProduct creates a brand-new product aggregate.
// It validates required fields and raises a ProductCreatedEvent.
func NewProduct(name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	if name == "" {
		return nil, ErrProductNameRequired
	}
	if basePrice == nil {
		return nil, ErrProductBasePriceRequired
	}

	id := uuid.NewString()
	p := &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		status:      ProductStatusActive,
		changes:     NewChanges(),
	}

	p.events = append(p.events, NewProductCreatedEvent(id, name, category, basePrice, ProductStatusActive, now))

	return p, nil
}

// Reconstitute rebuilds a Product from persisted state without raising events.
// Use this in repository implementations when loading from storage.
func Reconstitute(
	id, name, description, category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
) (*Product, error) {
	if id == "" {
		return nil, ErrProductIDRequired
	}
	if name == "" {
		return nil, ErrProductNameRequired
	}
	if status != ProductStatusActive && status != ProductStatusInactive {
		return nil, ErrInvalidStatus
	}
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		changes:     NewChanges(),
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Accessors (read-only)
// ────────────────────────────────────────────────────────────────────────────

func (p *Product) Changes() *Changes     { return p.changes }
func (p *Product) ID() string            { return p.id }
func (p *Product) Name() string          { return p.name }
func (p *Product) Description() string   { return p.description }
func (p *Product) Category() string      { return p.category }
func (p *Product) BasePrice() *Money     { return p.basePrice }
func (p *Product) Discount() *Discount   { return p.discount }
func (p *Product) Status() ProductStatus { return p.status }
func (p *Product) Events() []DomainEvent { return p.events }
func (p *Product) IsActive() bool        { return p.status == ProductStatusActive }

// ClearEvents resets the in-memory event slice after they have been dispatched.
func (p *Product) ClearEvents() {
	p.events = nil
}

// ────────────────────────────────────────────────────────────────────────────
// Business methods (mutating, enforce invariants)
// ────────────────────────────────────────────────────────────────────────────

// SetName updates the product name and marks the field dirty.
func (p *Product) SetName(name string) error {
	if name == "" {
		return ErrProductNameRequired
	}
	if p.name == name {
		return nil
	}
	p.name = name
	p.changes.MarkDirty(FieldName)
	return nil
}

// SetDescription updates the product description and marks the field dirty.
func (p *Product) SetDescription(description string) {
	if p.description == description {
		return
	}
	p.description = description
	p.changes.MarkDirty(FieldDescription)
}

// SetCategory updates the product category and marks the field dirty.
func (p *Product) SetCategory(category string) {
	if p.category == category {
		return
	}
	p.category = category
	p.changes.MarkDirty(FieldCategory)
}

// SetBasePrice updates the product base price and marks the field dirty.
func (p *Product) SetBasePrice(price *Money) error {
	if price == nil {
		return ErrProductBasePriceRequired
	}
	p.basePrice = price
	p.changes.MarkDirty(FieldBasePrice)
	return nil
}

// Activate transitions the product to active status and raises ProductActivatedEvent.
func (p *Product) Activate(now time.Time) error {
	if p.status == ProductStatusActive {
		return nil
	}
	p.status = ProductStatusActive
	p.changes.MarkDirty(FieldStatus)
	p.events = append(p.events, NewProductActivatedEvent(p.id, now))
	return nil
}

// Deactivate transitions the product to inactive status and raises ProductDeactivatedEvent.
// Any active discount is also removed.
func (p *Product) Deactivate(now time.Time) error {
	if p.status == ProductStatusInactive {
		return nil
	}
	p.status = ProductStatusInactive
	p.changes.MarkDirty(FieldStatus)
	if p.discount != nil {
		p.events = append(p.events, NewDiscountRemovedEvent(p.id, now))
		p.discount = nil
		p.changes.MarkDirty(FieldDiscount)
	}
	p.events = append(p.events, NewProductDeactivatedEvent(p.id, now))
	return nil
}

// ApplyDiscount applies a discount to the product.
// Only active products can receive discounts and the discount period must be valid.
func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}
	if discount == nil {
		return errors.New("discount must not be nil")
	}
	if !discount.IsValidAt(now) {
		return ErrInvalidDiscountPeriod
	}

	p.discount = discount
	p.changes.MarkDirty(FieldDiscount)
	p.events = append(p.events, NewDiscountAppliedEvent(p.id, discount.Percentage(), discount.StartsAt(), discount.EndsAt(), now))
	return nil
}

// RemoveDiscount removes any active discount from the product and raises DiscountRemovedEvent.
func (p *Product) RemoveDiscount(now time.Time) error {
	if p.discount == nil {
		return ErrNoActiveDiscount
	}
	p.discount = nil
	p.changes.MarkDirty(FieldDiscount)
	p.events = append(p.events, NewDiscountRemovedEvent(p.id, now))
	return nil
}

// RecordUpdate raises a ProductUpdatedEvent listing all currently dirty fields.
// Call this once before persisting if you want a single "updated" event for all field changes.
func (p *Product) RecordUpdate(now time.Time) {
	var dirty []Field
	for f := range p.changes.dirty {
		dirty = append(dirty, f)
	}
	if len(dirty) == 0 {
		return
	}
	p.events = append(p.events, NewProductUpdatedEvent(p.id, dirty, now))
}
