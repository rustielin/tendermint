package types

import tmpubsub "github.com/tendermint/tmlibs/pubsub"

type NopEventBus struct{}

func (NopEventBus) Subscribe(subscriber string, query tmpubsub.Query, out chan<- interface{}) {
}

func (NopEventBus) Unsubscribe(subscriber string, query tmpubsub.Query) {
}

func (NopEventBus) UnsubscribeAll(subscriber string) {
}

//--- block, tx, and vote events

func (NopEventBus) PublishEventNewBlock(block EventDataNewBlock) error {
	return nil
}

func (NopEventBus) PublishEventNewBlockHeader(header EventDataNewBlockHeader) error {
	return nil
}

func (NopEventBus) PublishEventVote(vote EventDataVote) error {
	return nil
}

func (NopEventBus) PublishEventTx(tx EventDataTx) error {
	return nil
}

//--- EventDataRoundState events

func (NopEventBus) PublishEventNewRoundStep(rs EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventTimeoutPropose(rs EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventTimeoutWait(rs EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventNewRound(rs EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventCompleteProposal(rs EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventPolka(rs EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventUnlock(rs EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventRelock(rs EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventLock(rs EventDataRoundState) error {
	return nil
}
