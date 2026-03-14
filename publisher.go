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

// AsyncPublisher wraps an EventPublisher and publishes events asynchronously
// via a buffered channel and a background goroutine.
type AsyncPublisher struct {
	inner EventPublisher
	ch    chan AssignmentEvent
	done  chan struct{}
}

func NewAsyncPublisher(inner EventPublisher) *AsyncPublisher {
	a := &AsyncPublisher{
		inner: inner,
		ch:    make(chan AssignmentEvent, 4096),
		done:  make(chan struct{}),
	}
	go func() {
		defer close(a.done)
		for event := range a.ch {
			a.inner.Publish(context.Background(), event)
		}
	}()
	return a
}

func (a *AsyncPublisher) Publish(_ context.Context, event AssignmentEvent) {
	select {
	case a.ch <- event:
	default:
		slog.Warn("event buffer full, dropping event",
			"experiment", event.Experiment,
			"user_id", event.UserID,
		)
	}
}

func (a *AsyncPublisher) Close() error {
	close(a.ch)
	<-a.done
	return a.inner.Close()
}
