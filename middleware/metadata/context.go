package metadata

import (
	"context"

	"github.com/go-kratos/kratos/v2/metadata"
)

type clientMetadataKey struct{}

func NewClientContext(ctx context.Context, md metadata.Metadata) context.Context {
	return context.WithValue(ctx, clientMetadataKey{}, md)
}

func FromClientContext(ctx context.Context) (metadata.Metadata, bool) {
	md, ok := ctx.Value(clientMetadataKey{}).(metadata.Metadata)
	return md, ok
}

type serverMetadataKey struct{}

func NewServerContext(ctx context.Context, md metadata.Metadata) context.Context {
	return context.WithValue(ctx, serverMetadataKey{}, md)
}

func FromServerContext(ctx context.Context) (metadata.Metadata, bool) {
	md, ok := ctx.Value(serverMetadataKey{}).(metadata.Metadata)
	return md, ok
}
