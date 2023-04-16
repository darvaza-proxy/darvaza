package sni

import (
	"context"
	"crypto/tls"
	"io/fs"
	"net"
	"sync"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/sync/httpgroup"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
)

var (
	_ net.Listener     = (*Dispatcher)(nil)
	_ httpgroup.Server = (*Dispatcher)(nil)
)

var (
	// ErrClosed is returned after Close() or Cancel()
	ErrClosed = fs.ErrClosed
	// ErrInvalid is returned when arguments aren't valid
	ErrInvalid = fs.ErrInvalid
	// ErrExists is returned when something is already created
	ErrExists = fs.ErrExist
)

// A Handler is a function that will take responsibility over a given
// connection. The Provided Context is used to indicate when a shut down
// has been initiated
type Handler func(context.Context, net.Conn) error

// The Dispatcher screens TCP connections and uses SNI to decide if
// they should be handled by a dedicated system or passed to
// the tls.Listener using it via Accept()
//
// dispatcher := &sni.Dispatcher{
// GetHandler: func() { ..... },
// }
//
// conf := &tls.Config{...}
// lsn, err := tls.NewListener(dispatcher, config)
type Dispatcher struct {
	mu sync.Mutex
	wg core.WaitGroup
	ch chan accept

	ctx       context.Context
	cancel    context.CancelFunc
	cancelled atomic.Bool
	ln        net.Listener
	log       slog.Logger
	err       error

	// Logger to report errors
	Logger slog.Logger
	// Context to be used as parent of the internal Canceller
	Context context.Context

	// GetHandler tells the Dispatcher if the connection associated with
	// a given ClientHelloInfo should be passed to a dedicated Handler
	// instead of passing it to the outer tls.Listener
	GetHandler func(*tls.ClientHelloInfo) Handler

	// OnAccept is optionally used to configure the inbound net.Conn
	OnAccept func(net.Conn) (net.Conn, error)

	// OnError let's the use decide if we shut down on critical errors or not
	// it also allows the user to act accordingly
	OnError func(err error) bool
}

type accept struct {
	conn net.Conn
	err  error
}

func (d *Dispatcher) init() {
	// Accept()
	d.ch = make(chan accept)

	// Cancel()
	ctx := d.Context
	if ctx == nil {
		ctx = context.Background()
	}
	d.ctx, d.cancel = context.WithCancel(ctx)

	// Logger
	log := d.Logger
	if log == nil {
		log = discard.New()
	}
	d.log = log

	// Callbacks
	d.wg.OnError(d.onError)
}

// Serve starts processing the underlying net.Listener
func (d *Dispatcher) Serve(ln net.Listener) error {
	if ln == nil {
		return ErrInvalid
	}

	d.mu.Lock()
	if d.ch == nil {
		d.init()
	}

	if d.ln != nil {
		d.mu.Unlock()
		return ErrExists
	}
	d.ln = ln
	d.mu.Unlock()

	return d.run()
}

func (d *Dispatcher) run() error {
	defer d.Close()

	for {
		conn, err := d.ln.Accept()
		if conn != nil {
			d.spawnHandler(conn)
			continue
		}

		if d.cancelled.Load() {
			// bye
			return nil
		}

		if err = d.catch(nil, err); err != nil {
			// oops
			return err
		}
	}
}

func (d *Dispatcher) spawnHandler(conn net.Conn) {
	d.wg.GoCatch(
		func() error {
			return d.handle(conn)
		},
		func(err error) error {
			return d.catch(conn.RemoteAddr(), err)
		})
}

func (d *Dispatcher) handle(conn net.Conn) error {
	if d.OnAccept != nil {
		conn2, err := d.OnAccept(conn)
		if err != nil {
			defer conn.Close()
			return err
		}
		conn = conn2
	}

	if d.GetHandler == nil {
		// no need to get the ClientHelloInfo here
		d.debug(conn.RemoteAddr()).
			Print("connected")
		return d.defaultHandler(d.ctx, conn)
	}

	// Get ClientHelloInfo
	chi, conn2, err := PeekClientHelloInfo(d.ctx, conn)
	if err != nil {
		defer conn.Close()
		return err
	}

	d.debug(conn.RemoteAddr()).
		WithField("sni", chi.ServerName).
		Print("connected")

	// Get alternative handler
	h := d.GetHandler(chi)
	if h == nil {
		h = d.defaultHandler
	}

	return h(d.ctx, conn2)
}

func (d *Dispatcher) defaultHandler(_ context.Context, conn net.Conn) error {
	d.ch <- accept{conn, nil}
	return nil
}

func (d *Dispatcher) catch(peer net.Addr, err error) error {
	if peer == nil {
		// Accept
		d.error(nil, err).Printf("accept: %s", err)
		return err
	}

	if err != nil {
		// don't propagate connection errors

		d.error(peer, err).Print(err)
		return nil
	}

	d.debug(peer).Print("done")
	return nil
}

func (d *Dispatcher) onError(err error) error {
	// catch considered this error to be fatal
	// initiate shutdown unless the user objects
	terminate := true

	if d.OnError != nil {
		terminate = d.OnError(err)
	}

	if terminate {
		d.Cancel()
		return err
	}

	// ignored
	return nil
}

// Shutdown initiates a shutdown and waits until the workers are done.
// The given Context isn't used, its only to allow implementing the
// httpgroup.Server interface
func (d *Dispatcher) Shutdown(context.Context) error {
	d.Cancel()
	return d.wg.Wait()
}

// Accept returns a connection that wasn't dispatched through
// the Handler provided by GetHandler
func (d *Dispatcher) Accept() (net.Conn, error) {
	d.mu.Lock()
	if d.ch == nil {
		d.init()
	}
	d.mu.Unlock()

	if msg := <-d.ch; msg.conn != nil {
		return msg.conn, msg.err
	}

	err := d.Err()
	if err == nil {
		if d.cancelled.Load() {
			return nil, ErrClosed
		}
		core.Panic("unreachable")
	}
	return nil, err
}

// Addr returns the address the underlying listener is using
func (d *Dispatcher) Addr() net.Addr {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.ln != nil {
		return d.ln.Addr()
	}
	return nil
}

// Close initiates a shut down but also returns
// the first fatal error if there was one
func (d *Dispatcher) Close() error {
	d.Cancel()
	return d.Err()
}

// Err tells the first fatal error
func (d *Dispatcher) Err() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.err
}

// Wait waits until all workers are done
func (d *Dispatcher) Wait() error {
	return d.wg.Wait()
}

// Cancel initiates a shut down. it will prevent
// new dispatchs and cancel existing workers, but
// the responsibility of closing the listener is on
// the tls.Listener
func (d *Dispatcher) Cancel() {
	if d.cancelled.CompareAndSwap(false, true) {
		d.cancel()
	}
}

// Cancelled tells if the Dispatcher has been shut down
func (d *Dispatcher) Cancelled() bool {
	return d.cancelled.Load()
}

func (d *Dispatcher) debug(peer net.Addr) slog.Logger {
	return d.loggerWithFields(slog.Debug, peer, nil)
}

func (d *Dispatcher) error(peer net.Addr, err error) slog.Logger {
	return d.loggerWithFields(slog.Error, peer, err)
}

func (d *Dispatcher) loggerWithFields(level slog.LogLevel, peer net.Addr, err error) slog.Logger {
	l := d.log.WithLevel(level).
		WithField("dispatcher", d.ln.Addr().String())

	if peer != nil {
		l = l.WithField("peer", peer.String())
	}

	if err != nil {
		l = l.WithField(slog.ErrorFieldName, err)
	}

	return l
}
