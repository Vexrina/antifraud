package outbox

type EventType int16

const (
	EventTypeUnknown     EventType = iota
	EventTypeTransaction EventType = iota
)

func (et EventType) String() string {
	switch et {
	case EventTypeUnknown:
		return "unknown"
	case EventTypeTransaction:
		return "transaction"
	default:
		return "unknown"
	}
}
