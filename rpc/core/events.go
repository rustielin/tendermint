package core

import (
	"github.com/pkg/errors"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmquery "github.com/tendermint/tmlibs/pubsub/query"
)

func Subscribe(wsCtx rpctypes.WSRPCContext, query string) (*ctypes.ResultSubscribe, error) {
	logger.Info("Subscribe to query", "remote", wsCtx.GetRemoteAddr(), "query", query)
	q, err := tmquery.New(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse a query")
	}
	ch := pubsub.Subscribe(q)
	if err = wsCtx.AddSubscription(query, ch); err != nil {
		pubsub.Unsubscribe(ch)
		return nil, err
	}
	go func() {
		for event := range ch {
			if wsCtx.IsRunning() {
				tmResult := &ctypes.ResultEvent{query, event.(tmtypes.TMEventData)}
				wsCtx.TryWriteRPCResponse(rpctypes.NewRPCResponse(wsCtx.Request.ID+"#event", tmResult, ""))
			} else {
				pubsub.Unsubscribe(ch)
			}
		}
	}()
	return &ctypes.ResultSubscribe{}, nil
}

func Unsubscribe(wsCtx rpctypes.WSRPCContext, query string) (*ctypes.ResultUnsubscribe, error) {
	logger.Info("Unsubscribe from query", "remote", wsCtx.GetRemoteAddr(), "query", query)
	ch := wsCtx.DeleteSubscription(query)
	if ch != nil {
		pubsub.Unsubscribe(ch)
	}
	return &ctypes.ResultUnsubscribe{}, nil
}

func UnsubscribeAll(wsCtx rpctypes.WSRPCContext) (*ctypes.ResultUnsubscribe, error) {
	logger.Info("Unsubscribe from all", "remote", wsCtx.GetRemoteAddr())
	channels := wsCtx.DeleteAllSubscriptions()
	for _, ch := range channels {
		pubsub.Unsubscribe(ch)
	}
	return &ctypes.ResultUnsubscribe{}, nil
}
