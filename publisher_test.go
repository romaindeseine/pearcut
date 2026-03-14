package pearcut

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestStdoutPublisher(t *testing.T) {
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name  string
		event AssignmentEvent
	}{
		{
			name: "basic event",
			event: AssignmentEvent{
				UserID:     "user-1",
				Experiment: "exp-1",
				Variant:    "control",
				Timestamp:  ts,
			},
		},
		{
			name: "different variant",
			event: AssignmentEvent{
				UserID:     "user-2",
				Experiment: "exp-2",
				Variant:    "treatment",
				Timestamp:  ts,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			pub := NewStdoutPublisher(&buf)
			pub.Publish(context.Background(), tt.event)

			var got map[string]string
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON output: %v", err)
			}

			if got["user_id"] != tt.event.UserID {
				t.Errorf("user_id = %q, want %q", got["user_id"], tt.event.UserID)
			}
			if got["experiment"] != tt.event.Experiment {
				t.Errorf("experiment = %q, want %q", got["experiment"], tt.event.Experiment)
			}
			if got["variant"] != tt.event.Variant {
				t.Errorf("variant = %q, want %q", got["variant"], tt.event.Variant)
			}
			if got["timestamp"] == "" {
				t.Error("timestamp is empty")
			}
		})
	}
}
