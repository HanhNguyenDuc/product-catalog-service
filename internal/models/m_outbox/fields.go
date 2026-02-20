package m_outbox

const Table = "outbox_events"
const (
	EventID     string = "event_id"
	EventType   string = "event_type"
	AggregateID string = "aggregate_id"
	Payload     string = "payload"
	Status      string = "status"
	CreatedAt   string = "created_at"
	ProcessedAt string = "processed_at"
)

const StatusPending = "pending"
