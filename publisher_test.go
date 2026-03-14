package pearcut

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
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

type spyPublisher struct {
	mu     sync.Mutex
	events []AssignmentEvent
	closed bool
}

func (s *spyPublisher) Publish(_ context.Context, e AssignmentEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
}

func (s *spyPublisher) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func (s *spyPublisher) Events() []AssignmentEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]AssignmentEvent, len(s.events))
	copy(cp, s.events)
	return cp
}

func TestAsyncPublisher(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	makeEvent := func(experiment, userID, variant string) AssignmentEvent {
		return AssignmentEvent{
			Experiment: experiment,
			UserID:     userID,
			Variant:    variant,
			Timestamp:  now,
		}
	}

	tests := []struct {
		name       string
		events     []AssignmentEvent
		wantCount  int
		wantEvents []AssignmentEvent
	}{
		{
			name:       "single event is delivered",
			events:     []AssignmentEvent{makeEvent("exp-1", "user-1", "control")},
			wantCount:  1,
			wantEvents: []AssignmentEvent{makeEvent("exp-1", "user-1", "control")},
		},
		{
			name: "multiple events are delivered in order",
			events: []AssignmentEvent{
				makeEvent("exp-1", "user-1", "control"),
				makeEvent("exp-2", "user-2", "treatment"),
				makeEvent("exp-3", "user-3", "control"),
			},
			wantCount: 3,
			wantEvents: []AssignmentEvent{
				makeEvent("exp-1", "user-1", "control"),
				makeEvent("exp-2", "user-2", "treatment"),
				makeEvent("exp-3", "user-3", "control"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &spyPublisher{}
			async := NewAsyncPublisher(spy)

			for _, e := range tt.events {
				async.Publish(context.Background(), e)
			}

			if err := async.Close(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := spy.Events()
			if len(got) != tt.wantCount {
				t.Fatalf("got %d events, want %d", len(got), tt.wantCount)
			}

			for i, want := range tt.wantEvents {
				if got[i] != want {
					t.Errorf("event[%d] = %+v, want %+v", i, got[i], want)
				}
			}
		})
	}
}

func TestAsyncPublisherCloseDrainsBuffer(t *testing.T) {
	spy := &spyPublisher{}
	async := NewAsyncPublisher(spy)

	n := 100
	for i := 0; i < n; i++ {
		async.Publish(context.Background(), AssignmentEvent{
			Experiment: "exp",
			UserID:     "user",
			Variant:    "v",
			Timestamp:  time.Now(),
		})
	}

	if err := async.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := spy.Events()
	if len(got) != n {
		t.Fatalf("got %d events, want %d", len(got), n)
	}
}

func TestAsyncPublisherBufferFullDropsEvent(t *testing.T) {
	// Use a blocking spy to fill the buffer: the spy blocks until released,
	// so the single buffer slot stays occupied.
	block := make(chan struct{})
	blocking := &blockingPublisher{block: block}
	async := &AsyncPublisher{
		inner: blocking,
		ch:    make(chan AssignmentEvent, 1),
		done:  make(chan struct{}),
	}
	go func() {
		defer close(async.done)
		for event := range async.ch {
			async.inner.Publish(context.Background(), event)
		}
	}()

	// First event: picked up by goroutine, which blocks in Publish.
	async.Publish(context.Background(), AssignmentEvent{Experiment: "e1"})
	// Wait for goroutine to pick it up and block.
	blocking.waitBlocked()

	// Second event: fills the buffer (size 1).
	async.Publish(context.Background(), AssignmentEvent{Experiment: "e2"})

	// Third event: buffer full, should be dropped.
	async.Publish(context.Background(), AssignmentEvent{Experiment: "e3"})

	// Unblock and drain.
	close(block)
	if err := async.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := blocking.Events()
	if len(got) != 2 {
		t.Fatalf("got %d events, want 2 (third should be dropped)", len(got))
	}
}

func TestAsyncPublisherCloseCallsInnerClose(t *testing.T) {
	spy := &spyPublisher{}
	async := NewAsyncPublisher(spy)

	if err := async.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spy.mu.Lock()
	defer spy.mu.Unlock()
	if !spy.closed {
		t.Error("inner publisher Close was not called")
	}
}

// blockingPublisher blocks in Publish until the block channel is closed.
type blockingPublisher struct {
	mu      sync.Mutex
	events  []AssignmentEvent
	block   chan struct{}
	blocked chan struct{}
	once    sync.Once
}

func (b *blockingPublisher) Publish(_ context.Context, e AssignmentEvent) {
	b.once.Do(func() {
		b.blocked = make(chan struct{})
		close(b.blocked)
	})
	<-b.block
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, e)
}

func (b *blockingPublisher) Close() error { return nil }

func (b *blockingPublisher) waitBlocked() {
	// Spin until blocked channel exists and is closed.
	for {
		b.mu.Lock()
		ch := b.blocked
		b.mu.Unlock()
		if ch != nil {
			<-ch
			return
		}
		time.Sleep(time.Millisecond)
	}
}

func (b *blockingPublisher) Events() []AssignmentEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	cp := make([]AssignmentEvent, len(b.events))
	copy(cp, b.events)
	return cp
}
