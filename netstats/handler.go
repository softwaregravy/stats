package netstats

import (
	"context"
	"net"

	"github.com/segmentio/netx"
	"github.com/segmentio/stats"
)

// NewHandler returns a netx.Handler object that warps handler and produces
// metrics on the default engine.
func NewHandler(handler netx.Handler) netx.Handler {
	return NewHandlerWith(nil, handler)
}

// NewHandlerWith returns a netx.Handler object that warps handler and produces
// metrics on eng.
//
// If eng is nil, the default engine is used instead.
func NewHandlerWith(eng *stats.Engine, handler netx.Handler) netx.Handler {
	if eng == nil {
		eng = stats.DefaultEngine
	}
	return netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		handler.ServeConn(ctx, NewConnWith(eng, conn))
	})
}
