package core

import (
	"github.com/pkg/errors"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/pubsub/query"
)

func Subscribe(wsCtx rpctypes.WSRPCContext, queryStr string) (*ctypes.ResultSubscribe, error) {
	logger.Info("Subscribe to query", "remote", wsCtx.GetRemoteAddr(), "query", queryStr)
	q, err := query.New(queryStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse a query")
	}
	ch := eventsSub.Subscribe(q)
	if err = wsCtx.AddSubscription(queryStr, ch); err != nil {
		eventsSub.Unsubscribe(ch)
		return nil, err
	}
	go func() {
		for event := range ch {
			if wsCtx.IsRunning() {
				tmResult := &ctypes.ResultEvent{queryStr, event.(tmtypes.TMEventData)}
				wsCtx.TryWriteRPCResponse(rpctypes.NewRPCResponse(wsCtx.Request.ID+"#event", tmResult, ""))
			} else {
				eventsSub.Unsubscribe(ch)
			}
		}
	}()
	return &ctypes.ResultSubscribe{}, nil
}

func Unsubscribe(wsCtx rpctypes.WSRPCContext, queryStr string) (*ctypes.ResultUnsubscribe, error) {
	logger.Info("Unsubscribe from query", "remote", wsCtx.GetRemoteAddr(), "query", queryStr)
	ch := wsCtx.DeleteSubscription(queryStr)
	if ch != nil {
		eventsSub.Unsubscribe(ch)
	}
	return &ctypes.ResultUnsubscribe{}, nil
}

func UnsubscribeAll(wsCtx rpctypes.WSRPCContext) (*ctypes.ResultUnsubscribe, error) {
	logger.Info("Unsubscribe from all", "remote", wsCtx.GetRemoteAddr())
	channels := wsCtx.DeleteAllSubscriptions()
	for _, ch := range channels {
		eventsSub.Unsubscribe(ch)
	}
	return &ctypes.ResultUnsubscribe{}, nil
}
