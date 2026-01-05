package outbox

import "testing"

func TestEventType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		et   EventType
		want string
	}{
		{
			name: "Unknown",
			et:   EventTypeUnknown,
			want: "unknown",
		},
		{
			name: "Transaction",
			et:   EventTypeTransaction,
			want: "transaction",
		},
		{
			name: "Unknown iota",
			et:   EventType(100),
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.et.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
