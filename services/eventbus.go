package services

import (
	"sync"
	"time"
)

type SSEEvent struct {
	ID         string
	OperatorID int32
	EventType  string
	Data       any
	CreatedAt  time.Time
}

type Subscriber chan SSEEvent

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[int32][]Subscriber
	buffer      map[int32][]SSEEvent
	bufferSize  int
}

func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		subscribers: make(map[int32][]Subscriber),
		buffer:      make(map[int32][]SSEEvent),
		bufferSize:  bufferSize,
	}
}

func (b *EventBus) Publish(evt SSEEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	buf := b.buffer[evt.OperatorID]
	buf = append(buf, evt)
	if len(buf) > b.bufferSize {
		buf = buf[len(buf)-b.bufferSize:]
	}
	b.buffer[evt.OperatorID] = buf

	for _, sub := range b.subscribers[evt.OperatorID] {
		select {
		case sub <- evt:
		default:
			// subscriber too slow → drop event (simple approach)
		}
	}
}

func (b *EventBus) Subscribe(operatorID int32) Subscriber {
	ch := make(Subscriber, 10)

	b.mu.Lock()
	b.subscribers[operatorID] = append(b.subscribers[operatorID], ch)
	b.mu.Unlock()

	return ch
}

func (b *EventBus) Unsubscribe(operatorID int32, ch Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subscribers[operatorID]
	out := subs[:0]

	for _, s := range subs {
		if s == ch {
			close(s)
			continue
		}
		out = append(out, s)
	}

	b.subscribers[operatorID] = out
}

func (b *EventBus) GetBufferedEvents(operatorID int32, lastID string) []SSEEvent {
	b.mu.RLock()
	defer b.mu.RUnlock()

	events := b.buffer[operatorID]
	if lastID == "" {
		return events
	}

	start := -1
	for i, e := range events {
		if e.ID == lastID {
			start = i + 1
			break
		}
	}

	if start == -1 || start >= len(events) {
		return nil
	}

	return events[start:]
}
