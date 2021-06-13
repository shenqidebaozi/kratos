package transport

import (
	"context"
	"net/url"

	// init encoding
	_ "github.com/go-kratos/kratos/v2/encoding/json"
	_ "github.com/go-kratos/kratos/v2/encoding/proto"
	_ "github.com/go-kratos/kratos/v2/encoding/xml"
	_ "github.com/go-kratos/kratos/v2/encoding/yaml"
)

// Metadata is the storage medium used by a Transporter.
type Metadata interface {
	// Get returns the value associated with the passed key.
	Get(key string) string
	// Set stores the key-value pair.
	Set(key string, value string)
	// Keys lists the keys stored in this carrier.
	Keys() []string
}

// Server is transport server.
type Server interface {
	Start(context.Context) error
	Stop(context.Context) error
}

// Endpointer is registry endpoint.
type Endpointer interface {
	Endpoint() (*url.URL, error)
}

// Transporter is transport context value interface.
type Transporter interface {
	Kind() string
	Endpoint() string

	Operation() string
	SetOperation(string)

	// metadata must not be nil
	Metadata() Metadata
}

type serverTransportKey struct{}
type clientTransportKey struct{}

// NewServerContext returns a new Context that carries value.
func NewServerContext(ctx context.Context, tr Transporter) context.Context {
	return context.WithValue(ctx, serverTransportKey{}, tr)
}

// FromServerContext returns the Transport value stored in ctx, if any.
func FromServerContext(ctx context.Context) (tr Transporter, ok bool) {
	tr, ok = ctx.Value(serverTransportKey{}).(Transporter)
	return
}

// NewClientContext returns a new Context that carries value.
func NewClientContext(ctx context.Context, tr Transporter) context.Context {
	return context.WithValue(ctx, clientTransportKey{}, tr)
}

// FromClientContext returns the Transport value stored in ctx, if any.
func FromClientContext(ctx context.Context) (tr Transporter, ok bool) {
	tr, ok = ctx.Value(clientTransportKey{}).(Transporter)
	return
}

// SetOperation set operation into context transport.
func SetOperation(ctx context.Context, method string) {
	if tr, ok := FromServerContext(ctx); ok {
		tr.SetOperation(method)
	}
}

// Operation returns the Transport operation from server context.
func Operation(ctx context.Context) string {
	if tr, ok := FromServerContext(ctx); ok {
		return tr.Operation()
	}
	return ""
}
