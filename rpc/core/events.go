package core

import (
	"github.com/pkg/errors"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmquery "github.com/tendermint/tmlibs/pubsub/query"
)

func Subscribe(wsCtx rpctypes.WSRPCContext, query string) (*ctypes.ResultSubscribe, error) {
	addr := wsCtx.GetRemoteAddr()

	logger.Info("Subscribe to query", "remote", addr, "query", query)
	q, err := tmquery.New(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse a query")
	}

	ch := make(chan interface{}, 1000)
	eventBus.Subscribe(addr, q, ch)

	go func() {
		for event := range ch {
			if wsCtx.IsRunning() {
				tmResult := &ctypes.ResultEvent{query, event.(tmtypes.TMEventData)}
				wsCtx.WriteRPCResponse(rpctypes.NewRPCResponse(wsCtx.Request.ID+"#event", tmResult, ""))
			} else {
				eventBus.Unsubscribe(addr, q)
			}
		}
	}()

	return &ctypes.ResultSubscribe{}, nil
}

func Unsubscribe(wsCtx rpctypes.WSRPCContext, query string) (*ctypes.ResultUnsubscribe, error) {
	addr := wsCtx.GetRemoteAddr()
	logger.Info("Unsubscribe from query", "remote", addr, "query", query)
	q, err := tmquery.New(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse a query")
	}
	eventBus.Unsubscribe(addr, q)
	return &ctypes.ResultUnsubscribe{}, nil
}

func UnsubscribeAll(wsCtx rpctypes.WSRPCContext) (*ctypes.ResultUnsubscribe, error) {
	addr := wsCtx.GetRemoteAddr()
	logger.Info("Unsubscribe from all", "remote", addr)
	eventBus.UnsubscribeAll(addr)
	return &ctypes.ResultUnsubscribe{}, nil
}
