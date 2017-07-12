package client

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	data "github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcclient "github.com/tendermint/tendermint/rpc/lib/client"
	"github.com/tendermint/tendermint/types"
	tmpubsub "github.com/tendermint/tmlibs/pubsub"
)

/*
HTTP is a Client implementation that communicates
with a tendermint node over json rpc and websockets.

This is the main implementation you probably want to use in
production code.  There are other implementations when calling
the tendermint node in-process (local), or when you want to mock
out the server for test code (mock).
*/
type HTTP struct {
	remote string
	rpc    *rpcclient.JSONRPCClient
	*WSEvents
}

// New takes a remote endpoint in the form tcp://<host>:<port>
// and the websocket path (which always seems to be "/websocket")
func NewHTTP(remote, wsEndpoint string) *HTTP {
	return &HTTP{
		rpc:      rpcclient.NewJSONRPCClient(remote),
		remote:   remote,
		WSEvents: newWSEvents(remote, wsEndpoint),
	}
}

func (c *HTTP) _assertIsClient() Client {
	return c
}

func (c *HTTP) _assertIsNetworkClient() NetworkClient {
	return c
}

func (c *HTTP) Status() (*ctypes.ResultStatus, error) {
	result := new(ctypes.ResultStatus)
	_, err := c.rpc.Call("status", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Status")
	}
	return result, nil
}

func (c *HTTP) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	result := new(ctypes.ResultABCIInfo)
	_, err := c.rpc.Call("abci_info", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "ABCIInfo")
	}
	return result, nil
}

func (c *HTTP) ABCIQuery(path string, data data.Bytes, prove bool) (*ctypes.ResultABCIQuery, error) {
	result := new(ctypes.ResultABCIQuery)
	_, err := c.rpc.Call("abci_query",
		map[string]interface{}{"path": path, "data": data, "prove": prove},
		result)
	if err != nil {
		return nil, errors.Wrap(err, "ABCIQuery")
	}
	return result, nil
}

func (c *HTTP) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	result := new(ctypes.ResultBroadcastTxCommit)
	_, err := c.rpc.Call("broadcast_tx_commit", map[string]interface{}{"tx": tx}, result)
	if err != nil {
		return nil, errors.Wrap(err, "broadcast_tx_commit")
	}
	return result, nil
}

func (c *HTTP) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX("broadcast_tx_async", tx)
}

func (c *HTTP) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX("broadcast_tx_sync", tx)
}

func (c *HTTP) broadcastTX(route string, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	result := new(ctypes.ResultBroadcastTx)
	_, err := c.rpc.Call(route, map[string]interface{}{"tx": tx}, result)
	if err != nil {
		return nil, errors.Wrap(err, route)
	}
	return result, nil
}

func (c *HTTP) NetInfo() (*ctypes.ResultNetInfo, error) {
	result := new(ctypes.ResultNetInfo)
	_, err := c.rpc.Call("net_info", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "NetInfo")
	}
	return result, nil
}

func (c *HTTP) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	result := new(ctypes.ResultDumpConsensusState)
	_, err := c.rpc.Call("dump_consensus_state", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "DumpConsensusState")
	}
	return result, nil
}

func (c *HTTP) BlockchainInfo(minHeight, maxHeight int) (*ctypes.ResultBlockchainInfo, error) {
	result := new(ctypes.ResultBlockchainInfo)
	_, err := c.rpc.Call("blockchain",
		map[string]interface{}{"minHeight": minHeight, "maxHeight": maxHeight},
		result)
	if err != nil {
		return nil, errors.Wrap(err, "BlockchainInfo")
	}
	return result, nil
}

func (c *HTTP) Genesis() (*ctypes.ResultGenesis, error) {
	result := new(ctypes.ResultGenesis)
	_, err := c.rpc.Call("genesis", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Genesis")
	}
	return result, nil
}

func (c *HTTP) Block(height int) (*ctypes.ResultBlock, error) {
	result := new(ctypes.ResultBlock)
	_, err := c.rpc.Call("block", map[string]interface{}{"height": height}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Block")
	}
	return result, nil
}

func (c *HTTP) Commit(height int) (*ctypes.ResultCommit, error) {
	result := new(ctypes.ResultCommit)
	_, err := c.rpc.Call("commit", map[string]interface{}{"height": height}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Commit")
	}
	return result, nil
}

func (c *HTTP) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	result := new(ctypes.ResultTx)
	query := map[string]interface{}{
		"hash":  hash,
		"prove": prove,
	}
	_, err := c.rpc.Call("tx", query, result)
	if err != nil {
		return nil, errors.Wrap(err, "Tx")
	}
	return result, nil
}

func (c *HTTP) Validators() (*ctypes.ResultValidators, error) {
	result := new(ctypes.ResultValidators)
	_, err := c.rpc.Call("validators", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Validators")
	}
	return result, nil
}

/** websocket event stuff here... **/

type WSEvents struct {
	types.PubSub
	remote        string
	endpoint      string
	ws            *rpcclient.WSClient
	subscriptions map[string]chan<- types.TMEventData

	// used for signaling the goroutine that feeds ws -> EventSwitch
	quit chan bool
	done chan bool
}

func newWSEvents(remote, endpoint string) *WSEvents {
	return &WSEvents{
		PubSub:        tmpubsub.NewServer(1000), // should be configurable
		endpoint:      endpoint,
		remote:        remote,
		subscriptions: make(map[string]chan<- types.TMEventData),
		quit:          make(chan bool, 1),
		done:          make(chan bool, 1),
	}
}

// Start is the only way I could think the extend OnStart from
// events.eventSwitch.  If only it wasn't private...
// BaseService.Start -> eventSwitch.OnStart -> WSEvents.Start
func (w *WSEvents) Start() (bool, error) {
	st, err := w.PubSub.Start()
	// if we did start, then OnStart here...
	if st && err == nil {
		ws := rpcclient.NewWSClient(w.remote, w.endpoint)
		_, err = ws.Start()
		if err == nil {
			w.ws = ws
			go w.eventListener()
		}
	}
	return st, errors.Wrap(err, "StartWSEvent")
}

// Stop wraps the BaseService/eventSwitch actions as Start does
func (w *WSEvents) Stop() bool {
	stop := w.PubSub.Stop()
	if stop {
		// send a message to quit to stop the eventListener
		w.quit <- true
		<-w.done
		w.ws.Stop()
		w.ws = nil
	}
	return stop
}

/** TODO: more intelligent subscriptions! **/
func (w *WSEvents) Subscribe(query string, out chan<- types.TMEventData) error {
	eventType := extractEventType(query)
	if _, ok := w.subscriptions[eventType]; ok {
		return errors.New("already subscribed")
	}

	err := w.ws.Subscribe(query)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe")
	}
	w.subscriptions[eventType] = out

	return nil
}

func (w *WSEvents) Unsubscribe(query string) {
	err := w.ws.Unsubscribe(query)
	if err != nil {
		// FIXME: ignore?
	}
	eventType := extractEventType(query)
	delete(w.subscriptions, eventType)
}

func (w *WSEvents) UnsubscribeAll() {
	err := w.ws.UnsubscribeAll()
	if err != nil {
		// FIXME: ignore?
	}
	w.subscriptions = make(map[string]chan<- types.TMEventData)
}

// eventListener is an infinite loop pulling all websocket events
// and pushing them to the EventSwitch.
//
// the goroutine only stops by closing quit
func (w *WSEvents) eventListener() {
	for {
		select {
		case res := <-w.ws.ResultsCh:
			// res is json.RawMessage
			err := w.parseEvent(res)
			if err != nil {
				// FIXME: better logging/handling of errors??
				fmt.Printf("ws result: %+v\n", err)
			}
		case err := <-w.ws.ErrorsCh:
			// FIXME: better logging/handling of errors??
			fmt.Printf("ws err: %+v\n", err)
		case <-w.quit:
			// send a message so we can wait for the routine to exit
			// before cleaning up the w.ws stuff
			w.done <- true
			return
		}
	}
}

// parseEvent unmarshals the json message and converts it into
// some implementation of types.TMEventData, and sends it off
// on the merry way to the EventSwitch
func (w *WSEvents) parseEvent(data []byte) (err error) {
	result := new(ctypes.ResultEvent)
	err = json.Unmarshal(data, result)
	if err != nil {
		// ignore silently (eg. subscribe, unsubscribe and maybe other events)
		// TODO: ?
		return nil
	}
	// looks good!  let's fire this baby!
	if out, ok := w.subscriptions[result.Name]; ok {
		out <- result.Data
	}
	return nil
}

func extractEventType(query string) string {
	re := regexp.MustCompile(types.EventTypeKey + "=" + "([^\\s]+)")
	s := re.FindStringSubmatch(query)
	if len(s) > 0 {
		return s[0]
	}
	return ""
}
