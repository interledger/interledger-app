package graph

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

type GraphqlHttpHandlerOpts struct {
	WebSocketKeepAlivePingInterval time.Duration
}

func MakeHandler(gs *handler.Server, opts GraphqlHttpHandlerOpts) http.Handler {
	webSocketKeepAlivePingInterval := 10 * time.Second
	if opts.WebSocketKeepAlivePingInterval != 0 {
		webSocketKeepAlivePingInterval = opts.WebSocketKeepAlivePingInterval
	}

	gs.AddTransport(transport.Websocket{
		KeepAlivePingInterval: webSocketKeepAlivePingInterval,
	})
	gs.AddTransport(transport.Options{})
	gs.AddTransport(transport.GET{})
	gs.AddTransport(transport.POST{})
	gs.AddTransport(transport.MultipartForm{})

	return gs
}
