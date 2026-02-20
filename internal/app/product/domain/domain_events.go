package domain

import "time"

// DomainEvent is the marker interface for all domain events.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// ────────────────────────────────────────────────────────────────────────────
// Product events
// ────────────────────────────────────────────────────────────────────────────

// ProductCreatedEvent is raised when a new product is created.
type ProductCreatedEvent struct {
	productID string
	name      string
	category  string
	basePrice *Money
	status    ProductStatus
	at        time.Time
}

func NewProductCreatedEvent(productID, name, category string, basePrice *Money, status ProductStatus, at time.Time) *ProductCreatedEvent {
	return &ProductCreatedEvent{productID: productID, name: name, category: category, basePrice: basePrice, status: status, at: at}
}

func (e *ProductCreatedEvent) EventName() string     { return "product.created" }
func (e *ProductCreatedEvent) OccurredAt() time.Time { return e.at }
func (e *ProductCreatedEvent) ProductID() string     { return e.productID }
func (e *ProductCreatedEvent) Name() string          { return e.name }
func (e *ProductCreatedEvent) Category() string      { return e.category }
func (e *ProductCreatedEvent) BasePrice() *Money     { return e.basePrice }
func (e *ProductCreatedEvent) Status() ProductStatus { return e.status }

// ProductUpdatedEvent is raised when mutable fields of a product change.
type ProductUpdatedEvent struct {
	productID     string
	changedFields []Field
	at            time.Time
}

func NewProductUpdatedEvent(productID string, changedFields []Field, at time.Time) *ProductUpdatedEvent {
	return &ProductUpdatedEvent{productID: productID, changedFields: changedFields, at: at}
}

func (e *ProductUpdatedEvent) EventName() string      { return "product.updated" }
func (e *ProductUpdatedEvent) OccurredAt() time.Time  { return e.at }
func (e *ProductUpdatedEvent) ProductID() string      { return e.productID }
func (e *ProductUpdatedEvent) ChangedFields() []Field { return e.changedFields }

// ProductActivatedEvent is raised when a product transitions to active status.
type ProductActivatedEvent struct {
	productID string
	at        time.Time
}

func NewProductActivatedEvent(productID string, at time.Time) *ProductActivatedEvent {
	return &ProductActivatedEvent{productID: productID, at: at}
}

func (e *ProductActivatedEvent) EventName() string     { return "product.activated" }
func (e *ProductActivatedEvent) OccurredAt() time.Time { return e.at }
func (e *ProductActivatedEvent) ProductID() string     { return e.productID }

// ProductDeactivatedEvent is raised when a product transitions to inactive status.
type ProductDeactivatedEvent struct {
	productID string
	at        time.Time
}

func NewProductDeactivatedEvent(productID string, at time.Time) *ProductDeactivatedEvent {
	return &ProductDeactivatedEvent{productID: productID, at: at}
}

func (e *ProductDeactivatedEvent) EventName() string     { return "product.deactivated" }
func (e *ProductDeactivatedEvent) OccurredAt() time.Time { return e.at }
func (e *ProductDeactivatedEvent) ProductID() string     { return e.productID }

// ────────────────────────────────────────────────────────────────────────────
// Discount events
// ────────────────────────────────────────────────────────────────────────────

// DiscountAppliedEvent is raised when a discount is successfully applied to a product.
type DiscountAppliedEvent struct {
	productID  string
	percentage string
	startsAt   time.Time
	endsAt     time.Time
	at         time.Time
}

func NewDiscountAppliedEvent(productID, percentage string, startsAt, endsAt, at time.Time) *DiscountAppliedEvent {
	return &DiscountAppliedEvent{productID: productID, percentage: percentage, startsAt: startsAt, endsAt: endsAt, at: at}
}

func (e *DiscountAppliedEvent) EventName() string     { return "product.discount_applied" }
func (e *DiscountAppliedEvent) OccurredAt() time.Time { return e.at }
func (e *DiscountAppliedEvent) ProductID() string     { return e.productID }
func (e *DiscountAppliedEvent) Percentage() string    { return e.percentage }
func (e *DiscountAppliedEvent) StartsAt() time.Time   { return e.startsAt }
func (e *DiscountAppliedEvent) EndsAt() time.Time     { return e.endsAt }

// DiscountRemovedEvent is raised when an active discount is removed from a product.
type DiscountRemovedEvent struct {
	productID string
	at        time.Time
}

func NewDiscountRemovedEvent(productID string, at time.Time) *DiscountRemovedEvent {
	return &DiscountRemovedEvent{productID: productID, at: at}
}

func (e *DiscountRemovedEvent) EventName() string     { return "product.discount_removed" }
func (e *DiscountRemovedEvent) OccurredAt() time.Time { return e.at }
func (e *DiscountRemovedEvent) ProductID() string     { return e.productID }
