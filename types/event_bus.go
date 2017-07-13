package types

import (
	cmn "github.com/tendermint/tmlibs/common"
	tmpubsub "github.com/tendermint/tmlibs/pubsub"
)

// EventBus is a common bus for all events going through the system.
type EventBus struct {
	cmn.BaseService
	pubsub *tmpubsub.Server
}

// NewEventBus returns new event bus wrapping
func NewEventBus(pubsub *tmpubsub.Server) *EventBus {
	b := &EventBus{pubsub: pubsub}
	b.BaseService = *cmn.NewBaseService(nil, "EventBus", b)
	return b
}

func (b *EventBus) OnStart() error {
	return b.pubsub.OnStart()
}

func (b *EventBus) OnStop() {
	b.pubsub.OnStop()
}

func (b *EventBus) Subscribe(subscriber string, query tmpubsub.Query, out chan<- interface{}) {
	b.pubsub.Subscribe(subscriber, query, out)
}

func (b *EventBus) Unsubscribe(subscriber string, query tmpubsub.Query) {
	b.pubsub.Unsubscribe(subscriber, query)
}

func (b *EventBus) UnsubscribeAll(subscriber string) {
	b.pubsub.UnsubscribeAll(subscriber)
}

func (b *EventBus) publish(eventType string, eventData TMEventData) error {
	if b.pubsub != nil {
		return b.pubsub.PublishWithTags(eventData, map[string]interface{}{EventTypeKey: eventType})
	}
	return nil
}

//--- block, tx, and vote events

func (b *EventBus) PublishEventNewBlock(block EventDataNewBlock) error {
	return b.publish(EventNewBlock, TMEventData{block})
}

func (b *EventBus) PublishEventNewBlockHeader(header EventDataNewBlockHeader) error {
	return b.publish(EventNewBlockHeader, TMEventData{header})
}

func (b *EventBus) PublishEventVote(vote EventDataVote) error {
	return b.publish(EventVote, TMEventData{vote})
}

func (b *EventBus) PublishEventTx(tx EventDataTx) error {
	return b.publish(EventTx(tx.Tx), TMEventData{tx})
}

//--- EventDataRoundState events

func (b *EventBus) PublishEventNewRoundStep(rs EventDataRoundState) error {
	return b.publish(EventNewRoundStep, TMEventData{rs})
}

func (b *EventBus) PublishEventTimeoutPropose(rs EventDataRoundState) error {
	return b.publish(EventTimeoutPropose, TMEventData{rs})
}

func (b *EventBus) PublishEventTimeoutWait(rs EventDataRoundState) error {
	return b.publish(EventTimeoutWait, TMEventData{rs})
}

func (b *EventBus) PublishEventNewRound(rs EventDataRoundState) error {
	return b.publish(EventNewRound, TMEventData{rs})
}

func (b *EventBus) PublishEventCompleteProposal(rs EventDataRoundState) error {
	return b.publish(EventCompleteProposal, TMEventData{rs})
}

func (b *EventBus) PublishEventPolka(rs EventDataRoundState) error {
	return b.publish(EventPolka, TMEventData{rs})
}

func (b *EventBus) PublishEventUnlock(rs EventDataRoundState) error {
	return b.publish(EventUnlock, TMEventData{rs})
}

func (b *EventBus) PublishEventRelock(rs EventDataRoundState) error {
	return b.publish(EventRelock, TMEventData{rs})
}

func (b *EventBus) PublishEventLock(rs EventDataRoundState) error {
	return b.publish(EventLock, TMEventData{rs})
}
