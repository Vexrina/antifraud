package outbox

const (
	// OutboxEventType_Unknown .
	OutboxEventType_Unknown = iota
	// OutboxEventType_Transaction - кладем транзакцию в аутбокс
	OutboxEventType_Transaction
)
