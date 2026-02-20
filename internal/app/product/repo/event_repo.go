package repo

import (
	"encoding/json"
	"fmt"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"github.com/product-catalog-service/internal/app/product/domain"
	"github.com/product-catalog-service/internal/models/m_outbox"
)

// EventRepo handles persistence of domain events into the outbox_events table.
// Mutations are atomic — they are added to the same commit plan as the business mutation.
type EventRepo struct{}

func NewEventRepo() *EventRepo {
	return &EventRepo{}
}

// InsertMut converts a DomainEvent into an outbox_events INSERT mutation.
// Returns nil when the event payload cannot be serialised (should never happen in practice).
func (r *EventRepo) InsertMut(event domain.DomainEvent) *spanner.Mutation {
	aggregateID := aggregateIDOf(event)

	payload, err := marshalPayload(event)
	if err != nil {
		// Payload serialisation failure is a programming error; surface as nil
		// so callers can decide whether to skip or fail hard.
		return nil
	}

	row := map[string]any{
		m_outbox.EventID:     uuid.NewString(),
		m_outbox.EventType:   event.EventName(),
		m_outbox.AggregateID: aggregateID,
		m_outbox.Payload:     payload,
		m_outbox.Status:      m_outbox.StatusPending,
		m_outbox.CreatedAt:   spanner.CommitTimestamp,
	}

	return spanner.InsertMap(m_outbox.Table, row)
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

// aggregateIDOf extracts the product_id from every known DomainEvent type.
func aggregateIDOf(event domain.DomainEvent) string {
	type hasProductID interface {
		ProductID() string
	}
	if e, ok := event.(hasProductID); ok {
		return e.ProductID()
	}
	return ""
}

// marshalPayload serialises a DomainEvent to a JSON string accepted by Spanner's JSON column.
// Each event type is serialised via an anonymous struct so that field names are stable
// and independent of future rename refactors.
func marshalPayload(event domain.DomainEvent) (string, error) {
	var data any

	switch e := event.(type) {
	case *domain.ProductCreatedEvent:
		data = struct {
			ProductID string `json:"product_id"`
			Name      string `json:"name"`
			Category  string `json:"category"`
			Status    string `json:"status"`
		}{
			ProductID: e.ProductID(),
			Name:      e.Name(),
			Category:  e.Category(),
			Status:    string(e.Status()),
		}

	case *domain.ProductUpdatedEvent:
		fields := make([]string, 0)
		for _, f := range e.ChangedFields() {
			fields = append(fields, string(f))
		}
		data = struct {
			ProductID     string   `json:"product_id"`
			ChangedFields []string `json:"changed_fields"`
		}{
			ProductID:     e.ProductID(),
			ChangedFields: fields,
		}

	case *domain.ProductActivatedEvent:
		data = struct {
			ProductID string `json:"product_id"`
		}{ProductID: e.ProductID()}

	case *domain.ProductDeactivatedEvent:
		data = struct {
			ProductID string `json:"product_id"`
		}{ProductID: e.ProductID()}

	case *domain.DiscountAppliedEvent:
		data = struct {
			ProductID  string `json:"product_id"`
			Percentage string `json:"percentage"`
			StartsAt   string `json:"starts_at"`
			EndsAt     string `json:"ends_at"`
		}{
			ProductID:  e.ProductID(),
			Percentage: e.Percentage(),
			StartsAt:   e.StartsAt().Format("2006-01-02T15:04:05Z07:00"),
			EndsAt:     e.EndsAt().Format("2006-01-02T15:04:05Z07:00"),
		}

	case *domain.DiscountRemovedEvent:
		data = struct {
			ProductID string `json:"product_id"`
		}{ProductID: e.ProductID()}

	default:
		return "", fmt.Errorf("unknown event type: %T", event)
	}

	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
