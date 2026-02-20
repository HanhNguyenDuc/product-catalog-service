package m_outbox

import (
	"time"

	"cloud.google.com/go/spanner"
)

// OutboxEventRow is the Spanner row representation of an outbox event.
// It mirrors the outbox_events table schema 1-to-1.
type OutboxEventRow struct {
	EventID     string           `spanner:"event_id"`
	EventType   string           `spanner:"event_type"`
	AggregateID string           `spanner:"aggregate_id"`
	Payload     string           `spanner:"payload"`
	Status      string           `spanner:"status"`
	CreatedAt   time.Time        `spanner:"created_at"`
	ProcessedAt spanner.NullTime `spanner:"processed_at"`
}
