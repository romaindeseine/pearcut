package pearcut

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
)

// NoopPublisher discards all events. Used as the default when no publisher is configured.
type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, AssignmentEvent) {}
func (NoopPublisher) Close() error                             { return nil }

// StdoutPublisher writes assignment events as JSON lines to the given writer.
type StdoutPublisher struct {
	enc *json.Encoder
}

func NewStdoutPublisher(w io.Writer) *StdoutPublisher {
	return &StdoutPublisher{enc: json.NewEncoder(w)}
}

func (p *StdoutPublisher) Publish(_ context.Context, event AssignmentEvent) {
	if err := p.enc.Encode(event); err != nil {
		slog.Error("failed to write event", "error", err)
	}
}

func (p *StdoutPublisher) Close() error { return nil }
