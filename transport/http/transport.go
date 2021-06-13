package http

import (
	"context"
	"net/http"

	"github.com/go-kratos/kratos/v2/transport"
)

var (
	_ transport.Transporter = &Transport{}
	_ transport.Metadata    = HeaderCarrier{}
)

// Transport is an HTTP transport.
type Transport struct {
	endpoint  string
	path      string
	method    string
	operation string
	metadata  HeaderCarrier
}

// Kind returns the transport kind.
func (tr *Transport) Kind() string {
	return "http"
}

// Endpoint returns the transport endpoint.
func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

// Operation returns the transport operation.
func (tr *Transport) Operation() string {
	return tr.operation
}

// SetOperation sets the transport operation.
func (tr *Transport) SetOperation(operation string) {
	tr.operation = operation
}

// Metadata returns the transport metadata.
func (tr *Transport) Metadata() transport.Metadata {
	return tr.metadata
}

// Path returns the Transport path from server context.
func Path(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			return tr.path
		}
	}
	return ""
}

// Method returns the Transport method from server context.
func Method(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			return tr.method
		}
	}
	return ""
}

type HeaderCarrier http.Header

// Get returns the value associated with the passed key.
func (hc HeaderCarrier) Get(key string) string {
	return http.Header(hc).Get(key)
}

// Set stores the key-value pair.
func (hc HeaderCarrier) Set(key string, value string) {
	http.Header(hc).Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (hc HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range http.Header(hc) {
		keys = append(keys, k)
	}
	return keys
}
