package httpserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"github.com/darvaza-proxy/darvaza/acme"
	"github.com/darvaza-proxy/darvaza/shared/tls/sni"
)

// Config describes how Server needs to be set up
type Config struct {
	// Logger is an optional slog.Logger used to debug the Server
	Logger slog.Logger
	// Context is the parent context.Context of the cancellable we use
	// with the workers
	Context context.Context

	// Bind defines the ports and addresses we listen
	Bind BindingConfig

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, http.DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int

	// MaxRecvBufferSize is the buffer size we will attempt to set to
	// UDP listeners
	// If zero, bind.DefaultMaxRecvBufferSize is used.
	MaxRecvBufferSize int

	// TLSConfig optionally serves as starting point allowing the
	// use to specify different constraints
	TLSConfig *tls.Config

	// TLS Callbacks
	GetHandlerForClient func(*tls.ClientHelloInfo) sni.Handler
	GetConfigForClient  func(*tls.ClientHelloInfo) (*tls.Config, error)
	GetCertificate      func(*tls.ClientHelloInfo) (*tls.Certificate, error)
	GetRootCAs          func() *x509.CertPool
	GetClientCAs        func() *x509.CertPool

	// Optional resolver for the ACME-HTTP-01 challenge
	AcmeHTTP01 acme.HTTP01Resolver

	// Handler is the HTTPS application we serve on Bind.Port via H1/H2/H3 mapped
	// to `/` on our internal router. This internal router also takes
	// responsibility of handling the ACME-HTTP-01 challenge and other special cases
	// defined using the Handle() and HandleFunc() methods on the created Server
	Handler http.Handler
	// HandleInsecure tells us to use the same handler via H1/H2C on Bind.PortInsecure.
	// Otherwise the insecure port, if allowed, will only redirect to https and optionally
	// handle the ACME-HTTP-01 challenge using the resolver specified on the AcmeHTTP01
	// field
	HandleInsecure bool
}

// SetDefaults attempts to fill any configuration gap
func (cfg *Config) SetDefaults() error {
	if cfg.Logger == nil {
		cfg.Logger = discard.New()
	}

	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	return nil
}

// Update alters the config based on what's found on the given server
func (cfg *Config) Update(hs *Server) error {
	if hs == nil {
		return os.ErrInvalid
	}

	*cfg = hs.cfg
	return nil
}

// BindingConfig includes the information needed to listen TCP/UDP ports
type BindingConfig struct {
	Interfaces []string
	Addresses  []string

	Port         uint16
	PortStrict   bool
	PortAttempts int
	// PortInsecure specifies the port used to listen plain HTTP
	// if AllowInsecure is enabled
	PortInsecure uint16
	// AllowInsecure tells if plain HTTP 1.1 or H2C is allowed
	AllowInsecure bool

	// KeepAlive specifies the KeepAlive value to use on net.ListenConfig
	// when calling Listen()
	KeepAlive time.Duration
}

func (srv *Server) getReadHeaderTimeout() time.Duration {
	if t := srv.cfg.ReadHeaderTimeout; t > 0 {
		return t
	}
	return srv.cfg.ReadTimeout
}

func getTimeout(d time.Duration) time.Time {
	if d > 0 {
		return time.Now().Add(d)
	}
	return time.Time{}
}
