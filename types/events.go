package types

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire/data"
	cmn "github.com/tendermint/tmlibs/common"
	tmpubsub "github.com/tendermint/tmlibs/pubsub"
	tmquery "github.com/tendermint/tmlibs/pubsub/query"
)

// Reserved event types
const (
	EventBond    = "Bond"
	EventUnbond  = "Unbond"
	EventRebond  = "Rebond"
	EventDupeout = "Dupeout"
	EventFork    = "Fork"

	EventNewBlock         = "NewBlock"
	EventNewBlockHeader   = "NewBlockHeader"
	EventNewRound         = "NewRound"
	EventNewRoundStep     = "NewRoundStep"
	EventTimeoutPropose   = "TimeoutPropose"
	EventCompleteProposal = "CompleteProposal"
	EventPolka            = "Polka"
	EventUnlock           = "Unlock"
	EventLock             = "Lock"
	EventRelock           = "Relock"
	EventTimeoutWait      = "TimeoutWait"
	EventVote             = "Vote"
)

func EventTx(tx Tx) string { return fmt.Sprintf("Tx:%X", tx.Hash()) }

///////////////////////////////////////////////////////////////////////////////
// ENCODING / DECODING
///////////////////////////////////////////////////////////////////////////////

var (
	EventDataNameNewBlock       = "new_block"
	EventDataNameNewBlockHeader = "new_block_header"
	EventDataNameTx             = "tx"
	EventDataNameRoundState     = "round_state"
	EventDataNameVote           = "vote"
)

// implements events.EventData
type TMEventDataInner interface {
	// empty interface
}

type TMEventData struct {
	TMEventDataInner `json:"unwrap"`
}

func (tmr TMEventData) MarshalJSON() ([]byte, error) {
	return tmEventDataMapper.ToJSON(tmr.TMEventDataInner)
}

func (tmr *TMEventData) UnmarshalJSON(data []byte) (err error) {
	parsed, err := tmEventDataMapper.FromJSON(data)
	if err == nil && parsed != nil {
		tmr.TMEventDataInner = parsed.(TMEventDataInner)
	}
	return
}

func (tmr TMEventData) Unwrap() TMEventDataInner {
	tmrI := tmr.TMEventDataInner
	for wrap, ok := tmrI.(TMEventData); ok; wrap, ok = tmrI.(TMEventData) {
		tmrI = wrap.TMEventDataInner
	}
	return tmrI
}

func (tmr TMEventData) Empty() bool {
	return tmr.TMEventDataInner == nil
}

const (
	EventDataTypeNewBlock       = byte(0x01)
	EventDataTypeFork           = byte(0x02)
	EventDataTypeTx             = byte(0x03)
	EventDataTypeNewBlockHeader = byte(0x04)

	EventDataTypeRoundState = byte(0x11)
	EventDataTypeVote       = byte(0x12)
)

var tmEventDataMapper = data.NewMapper(TMEventData{}).
	RegisterImplementation(EventDataNewBlock{}, EventDataNameNewBlock, EventDataTypeNewBlock).
	RegisterImplementation(EventDataNewBlockHeader{}, EventDataNameNewBlockHeader, EventDataTypeNewBlockHeader).
	RegisterImplementation(EventDataTx{}, EventDataNameTx, EventDataTypeTx).
	RegisterImplementation(EventDataRoundState{}, EventDataNameRoundState, EventDataTypeRoundState).
	RegisterImplementation(EventDataVote{}, EventDataNameVote, EventDataTypeVote)

// Most event messages are basic types (a block, a transaction)
// but some (an input to a call tx or a receive) are more exotic

type EventDataNewBlock struct {
	Block *Block `json:"block"`
}

// light weight event for benchmarking
type EventDataNewBlockHeader struct {
	Header *Header `json:"header"`
}

// All txs fire EventDataTx
type EventDataTx struct {
	Height int           `json:"height"`
	Tx     Tx            `json:"tx"`
	Data   data.Bytes    `json:"data"`
	Log    string        `json:"log"`
	Code   abci.CodeType `json:"code"`
	Error  string        `json:"error"` // this is redundant information for now
}

// NOTE: This goes into the replay WAL
type EventDataRoundState struct {
	Height int    `json:"height"`
	Round  int    `json:"round"`
	Step   string `json:"step"`

	// private, not exposed to websockets
	RoundState interface{} `json:"-"`
}

type EventDataVote struct {
	Vote *Vote
}

func (_ EventDataNewBlock) AssertIsTMEventData()       {}
func (_ EventDataNewBlockHeader) AssertIsTMEventData() {}
func (_ EventDataTx) AssertIsTMEventData()             {}
func (_ EventDataRoundState) AssertIsTMEventData()     {}
func (_ EventDataVote) AssertIsTMEventData()           {}

///////////////////////////////////////////////////////////////////////////////
// PUBSUB
///////////////////////////////////////////////////////////////////////////////

const (
	// EventTypeKey is a reserved key, used to specify event type in tags.
	EventTypeKey = "tm.events.type"
)

// EventsPublisher is an interface for somebody who wants to publish events.
type EventsPublisher interface {
	PublishWithTags(interface{}, map[string]interface{}) error
}

// EventsSubscriber is an interface for somebody who wants to listen for events.
type EventsSubscriber interface {
	Subscribe(string, tmpubsub.Query, chan<- interface{})
	Unsubscribe(string, tmpubsub.Query)
	UnsubscribeAll(string)
}

// PubSub is a common interface unifying publisher and subscriber.
type PubSub interface {
	EventsPublisher
	EventsSubscriber
	cmn.Service
}

var (
	EventQueryBond             = safeQueryFor(EventBond)
	EventQueryUnbond           = safeQueryFor(EventUnbond)
	EventQueryRebond           = safeQueryFor(EventRebond)
	EventQueryDupeout          = safeQueryFor(EventDupeout)
	EventQueryFork             = safeQueryFor(EventFork)
	EventQueryNewBlock         = safeQueryFor(EventNewBlock)
	EventQueryNewBlockHeader   = safeQueryFor(EventNewBlockHeader)
	EventQueryNewRound         = safeQueryFor(EventNewRound)
	EventQueryNewRoundStep     = safeQueryFor(EventNewRoundStep)
	EventQueryTimeoutPropose   = safeQueryFor(EventTimeoutPropose)
	EventQueryCompleteProposal = safeQueryFor(EventCompleteProposal)
	EventQueryPolka            = safeQueryFor(EventPolka)
	EventQueryUnlock           = safeQueryFor(EventUnlock)
	EventQueryLock             = safeQueryFor(EventLock)
	EventQueryRelock           = safeQueryFor(EventRelock)
	EventQueryTimeoutWait      = safeQueryFor(EventTimeoutWait)
	EventQueryVote             = safeQueryFor(EventVote)
)

func EventQueryTx(tx Tx) tmpubsub.Query {
	return safeQueryFor(EventTx(tx))
}

func safeQueryFor(eventType string) tmpubsub.Query {
	q, err := tmquery.New(EventTypeKey + "=" + eventType)
	if err != nil {
		panic(fmt.Sprintf("query %v has a syntax error in it", q))
	}
	return q
}

//--- block, tx, and vote events

func FireEventNewBlock(p EventsPublisher, block EventDataNewBlock) error {
	return fireEvent(p, EventNewBlock, TMEventData{block})
}

func FireEventNewBlockHeader(p EventsPublisher, header EventDataNewBlockHeader) error {
	return fireEvent(p, EventNewBlockHeader, TMEventData{header})
}

func FireEventVote(p EventsPublisher, vote EventDataVote) error {
	return fireEvent(p, EventVote, TMEventData{vote})
}

func FireEventTx(p EventsPublisher, tx EventDataTx) error {
	return fireEvent(p, EventTx(tx.Tx), TMEventData{tx})
}

//--- EventDataRoundState events

func FireEventNewRoundStep(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventNewRoundStep, TMEventData{rs})
}

func FireEventTimeoutPropose(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventTimeoutPropose, TMEventData{rs})
}

func FireEventTimeoutWait(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventTimeoutWait, TMEventData{rs})
}

func FireEventNewRound(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventNewRound, TMEventData{rs})
}

func FireEventCompleteProposal(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventCompleteProposal, TMEventData{rs})
}

func FireEventPolka(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventPolka, TMEventData{rs})
}

func FireEventUnlock(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventUnlock, TMEventData{rs})
}

func FireEventRelock(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventRelock, TMEventData{rs})
}

func FireEventLock(p EventsPublisher, rs EventDataRoundState) error {
	return fireEvent(p, EventLock, TMEventData{rs})
}

// All events should be based on this FireEvent to ensure they are TMEventData.
func fireEvent(p EventsPublisher, eventType string, eventData TMEventData) error {
	if p != nil {
		return p.PublishWithTags(eventData, map[string]interface{}{EventTypeKey: eventType})
	}
	return nil
}
