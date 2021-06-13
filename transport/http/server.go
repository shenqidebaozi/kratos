package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	ic "github.com/go-kratos/kratos/v2/internal/context"
	"github.com/go-kratos/kratos/v2/internal/host"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/gorilla/mux"
)

var _ transport.Server = (*Server)(nil)
var _ transport.Endpointer = (*Server)(nil)

// ServerOption is an HTTP server option.
type ServerOption func(*Server)

// Network with server network.
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Timeout with server timeout.
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Logger with server logger.
func Logger(logger log.Logger) ServerOption {
	return func(s *Server) {
		s.log = log.NewHelper(logger)
	}
}

// Middleware with service middleware option.
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(o *Server) {
		o.ms = m
	}
}

// Filter with HTTP middleware option.
func Filter(filters ...FilterFunc) ServerOption {
	return func(o *Server) {
		o.filters = filters
	}
}

// RequestDecoder with request decoder.
func RequestDecoder(dec DecodeRequestFunc) ServerOption {
	return func(o *Server) {
		o.dec = dec
	}
}

// ResponseEncoder with response encoder.
func ResponseEncoder(en EncodeResponseFunc) ServerOption {
	return func(o *Server) {
		o.enc = en
	}
}

// ErrorEncoder with error encoder.
func ErrorEncoder(en EncodeErrorFunc) ServerOption {
	return func(o *Server) {
		o.ene = en
	}
}

// Server is an HTTP server wrapper.
type Server struct {
	*http.Server
	ctx         context.Context
	lis         net.Listener
	once        sync.Once
	endpoint    *url.URL
	err         error
	network     string
	address     string
	timeout     time.Duration
	filters     []FilterFunc
	ms          []middleware.Middleware
	dec         DecodeRequestFunc
	enc         EncodeResponseFunc
	ene         EncodeErrorFunc
	router      *mux.Router
	log         *log.Helper
	mdKeyPrefix string
}

// NewServer creates an HTTP server by options.
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: ":0",
		timeout: 1 * time.Second,
		dec:     DefaultRequestDecoder,
		enc:     DefaultResponseEncoder,
		ene:     DefaultErrorEncoder,
		log:     log.NewHelper(log.DefaultLogger),
	}
	for _, o := range opts {
		o(srv)
	}
	srv.Server = &http.Server{Handler: srv}
	srv.router = mux.NewRouter()
	srv.router.Use(srv.filter())
	return srv
}

// Route registers an HTTP route.
func (s *Server) Route(prefix string, filters ...FilterFunc) *Route {
	return newRoute(prefix, s, filters...)
}

// Handle registers a new route with a matcher for the URL path.
func (s *Server) Handle(path string, h http.Handler) {
	s.router.Handle(path, h)
}

// HandlePrefix registers a new route with a matcher for the URL path prefix.
func (s *Server) HandlePrefix(prefix string, h http.Handler) {
	s.router.PathPrefix(prefix).Handler(h)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (s *Server) HandleFunc(path string, h http.HandlerFunc) {
	s.router.HandleFunc(path, h)
}

// ServeHTTP should write reply headers and data to the ResponseWriter and then return.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(res, req)
}

func (s *Server) filter() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		next = FilterChain(s.filters...)(next)
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := ic.Merge(req.Context(), s.ctx)
			defer cancel()
			if s.timeout > 0 {
				ctx, cancel = context.WithTimeout(ctx, s.timeout)
				defer cancel()
			}
			tr := &Transport{
				endpoint:  s.endpoint.String(),
				path:      req.RequestURI,
				method:    req.Method,
				operation: req.RequestURI,
				metadata:  HeaderCarrier(req.Header),
			}
			if r := mux.CurrentRoute(req); r != nil {
				if path, err := r.GetPathTemplate(); err == nil {
					tr.operation = path
				}
			}
			ctx = transport.NewServerContext(ctx, tr)
			md := metadata.New()
			for k, v := range req.Header {
				if strings.HasPrefix(k, s.mdKeyPrefix) && len(v) > 0 {
					md[strings.TrimLeft(k, s.mdKeyPrefix)] = v[0]
				}
			}
			ctx = metadata.NewServerContext(ctx, md)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// Endpoint return a real address to registry endpoint.
// examples:
//   http://127.0.0.1:8000?isSecure=false
func (s *Server) Endpoint() (*url.URL, error) {
	s.once.Do(func() {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return
		}
		addr, err := host.Extract(s.address, s.lis)
		if err != nil {
			lis.Close()
			s.err = err
			return
		}
		s.lis = lis
		s.endpoint = &url.URL{Scheme: "http", Host: addr}
	})
	if s.err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}

// Start start the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if _, err := s.Endpoint(); err != nil {
		return err
	}
	s.ctx = ctx
	s.log.Infof("[HTTP] server listening on: %s", s.lis.Addr().String())
	if err := s.Serve(s.lis); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop stop the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("[HTTP] server stopping")
	return s.Shutdown(context.Background())
}
