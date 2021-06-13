package metadata

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Option is metadata option.
type Option func(*options)

type options struct {
	globalPrefix []string
	md           metadata.Metadata
}

// WithConstants is option with constant metadata key value.
func WithConstants(md metadata.Metadata) Option {
	return func(o *options) {
		o.md = md
	}
}

// WithGlobalPropagatedPrefix is option with global propagated key prefix.
func WithGlobalPropagatedPrefix(prefix ...string) Option {
	return func(o *options) {
		o.globalPrefix = append(o.globalPrefix, prefix...)
	}
}

// Client is middleware client-side metadata.
func Client(opts ...Option) middleware.Middleware {
	options := options{
		globalPrefix: []string{"x-md-g-"},
	}
	for _, o := range opts {
		o(&options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			md := options.md.Clone()
			// passing through the global propagated metadata
			if tr, ok := transport.FromServerContext(ctx); ok {
				for k, v := range tr.Metadata() {
					if strings.HasPrefix(k, options.prefix) {
						md.Set(k, v)
					}
				}
			}
			// passing through the client outgoing metadata
			if cmd, ok := FromClientContext(ctx); ok {
				for k, v := range cmd {
					md.Set(k, v)
				}
			}
			if tr, ok := transport.FromClientContext(ctx); ok {
				tr.WithMetadata(md)
			}
			return handler(ctx, req)
		}
	}
}

// Server is middleware server-side metadata.
func Server(opts ...Option) middleware.Middleware {
	options := options{
		globalPrefix: []string{"x-md-g-"},
	}
	for _, o := range opts {
		o(&options)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// passing through the global propagated metadata
			if tr, ok := transport.FromServerContext(ctx); ok {
				md := metadata.New()
				for _, k := range tr.Metadata().Keys() {
					md[k] = tr.Metadata().Get(k)
				}
				ctx = metadata.NewServerContext(ctx, md)
			}
			return handler(ctx, req)
		}
	}
}
