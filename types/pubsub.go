package types

import (
	"fmt"

	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/pubsub"
	"github.com/tendermint/tmlibs/pubsub/query"
)

const (
	EventKey = "tm.events.type"
)

var (
	EventQueryBond             = safeQueryForEvent("Bond")
	EventQueryUnbond           = safeQueryForEvent("Unbond")
	EventQueryRebond           = safeQueryForEvent("Rebond")
	EventQueryDupeout          = safeQueryForEvent("Dupeout")
	EventQueryFork             = safeQueryForEvent("Fork")
	EventQueryNewBlock         = safeQueryForEvent("NewBlock")
	EventQueryNewBlockHeader   = safeQueryForEvent("NewBlockHeader")
	EventQueryNewRound         = safeQueryForEvent("NewRound")
	EventQueryNewRoundStep     = safeQueryForEvent("NewRoundStep")
	EventQueryTimeoutPropose   = safeQueryForEvent("TimeoutPropose")
	EventQueryCompleteProposal = safeQueryForEvent("CompleteProposal")
	EventQueryPolka            = safeQueryForEvent("Polka")
	EventQueryUnlock           = safeQueryForEvent("Unlock")
	EventQueryLock             = safeQueryForEvent("Lock")
	EventQueryRelock           = safeQueryForEvent("Relock")
	EventQueryTimeoutWait      = safeQueryForEvent("TimeoutWait")
	EventQueryVote             = safeQueryForEvent("Vote")
)

func EventQueryTx(tx Tx) pubsub.Query {
	return safeQueryForEvent(fmt.Sprintf("Tx:%X", tx.Hash()))
}

func safeQueryForEvent(event string) pubsub.Query {
	queryStr := EventKey + "=" + event
	q, err := query.New(queryStr)
	if err != nil {
		panic(fmt.Sprintf("query %s has a syntax error in it", queryStr))
	}
	return q
}

type EventsSubscriber interface {
	Subscribe(pubsub.Query) chan interface{}
	Unsubscribe(chan interface{})
}

type EventsPublisher interface {
	PublishWithTags(interface{}, map[string]interface{})
}

type EventsPubsub interface {
	EventsPublisher
	EventsSubscriber
	cmn.Service
}

//--- block, tx, and vote events

func FireEventNewBlock(p EventsPublisher, block EventDataNewBlock) {
	fireEvent(p, EventStringNewBlock(), TMEventData{block})
}

func FireEventNewBlockHeader(p EventsPublisher, header EventDataNewBlockHeader) {
	fireEvent(p, EventStringNewBlockHeader(), TMEventData{header})
}

func FireEventVote(p EventsPublisher, vote EventDataVote) {
	fireEvent(p, EventStringVote(), TMEventData{vote})
}

func FireEventTx(p EventsPublisher, tx EventDataTx) {
	fireEvent(p, EventStringTx(tx.Tx), TMEventData{tx})
}

//--- EventDataRoundState events

func FireEventNewRoundStep(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringNewRoundStep(), TMEventData{rs})
}

func FireEventTimeoutPropose(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringTimeoutPropose(), TMEventData{rs})
}

func FireEventTimeoutWait(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringTimeoutWait(), TMEventData{rs})
}

func FireEventNewRound(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringNewRound(), TMEventData{rs})
}

func FireEventCompleteProposal(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringCompleteProposal(), TMEventData{rs})
}

func FireEventPolka(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringPolka(), TMEventData{rs})
}

func FireEventUnlock(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringUnlock(), TMEventData{rs})
}

func FireEventRelock(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringRelock(), TMEventData{rs})
}

func FireEventLock(p EventsPublisher, rs EventDataRoundState) {
	fireEvent(p, EventStringLock(), TMEventData{rs})
}

// All events should be based on this FireEvent to ensure they are TMEventData
func fireEvent(p EventsPublisher, event string, data TMEventData) {
	if p != nil {
		p.PublishWithTags(data, map[string]interface{}{EventKey: event})
	}
}
