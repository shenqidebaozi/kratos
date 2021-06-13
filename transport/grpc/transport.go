package grpc

import (
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/metadata"
)

var (
	_ transport.Transporter = &Transport{}
)

// Transport is a gRPC transport.
type Transport struct {
	endpoint  string
	operation string
	metadata  MetadataCarrier
}

// Kind returns the transport kind.
func (tr *Transport) Kind() string {
	return "grpc"
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

type MetadataCarrier metadata.MD

// Get returns the value associated with the passed key.
func (mc MetadataCarrier) Get(key string) string {
	vals := metadata.MD(mc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// Set stores the key-value pair.
func (mc MetadataCarrier) Set(key string, value string) {
	metadata.MD(mc).Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (mc MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range metadata.MD(mc) {
		keys = append(keys, k)
	}
	return keys
}
