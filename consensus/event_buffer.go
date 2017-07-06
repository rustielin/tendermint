package consensus

import (
	"github.com/tendermint/tendermint/types"
)

const (
	eventBufferCapacity = 1000
)

// Interface assertions
var _ types.EventsPublisher = (*EventBuffer)(nil)

// EventBuffer is a buffer of events.
type EventBuffer struct {
	next   types.EventsPublisher
	events []event
}

// NewEventBuffer returns a new buffer
func NewEventBuffer(next types.EventsPublisher) *EventBuffer {
	return &EventBuffer{
		next:   next,
		events: make([]event, 0, eventBufferCapacity),
	}
}

type event struct {
	msg  interface{}
	tags map[string]interface{}
}

// PublishWithTags buffers an event to be fired upon finality.
func (b *EventBuffer) PublishWithTags(msg interface{}, tags map[string]interface{}) error {
	b.events = append(b.events, event{msg, tags})
	return nil
}

// Flush fires events by running next.PublishWithTags on all cached events.
// Blocks. Clears cached events.
func (b *EventBuffer) Flush() error {
	for _, e := range b.events {
		err := b.next.PublishWithTags(e.msg, e.tags)
		if err != nil {
			return err
		}
	}
	b.events = make([]event, 0, eventBufferCapacity)
	return nil
}
