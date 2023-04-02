// Package httpgroup implements an errgroup for HTTP Servers
package httpgroup

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"syscall"

	"darvaza.org/core"
	"darvaza.org/slog"
)

var (
	_ Server = (*http.Server)(nil)
)

// Server is a subset of the standard *http.Server including what httpgroup uses
type Server interface {
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

// Worker is an abstraction of a running Server
type Worker struct {
	Listener net.Listener
	Server   Server
}

// IsError filters out errors that can stop the Group
func (*Worker) IsError(err error) bool {
	switch err {
	case nil, http.ErrServerClosed:
		return false
	default:
		return true
	}
}

// Run is the blocking call that runs the Server
func (w *Worker) Run() error {
	var err error

	if w.Listener == nil || w.Server == nil {
		err = syscall.EINVAL
	} else if e := w.Server.Serve(w.Listener); w.IsError(e) {
		err = e
	}

	return err
}

// Shutdown is the blocking call that stops a Server
func (w *Worker) Shutdown(ctx context.Context) error {
	return w.Server.Shutdown(ctx)
}

// Group is a variant of errgroup.Group on which workers
// are *http.Server/net.Listener instances
type Group struct {
	ctx       context.Context
	cancel    context.CancelFunc
	cancelled atomic.Bool
	count     atomic.Int32
	logger    atomic.Value

	wg core.WaitGroup
}

// init initialises the Group when needed
func (heg *Group) init(ctx context.Context) context.Context {
	if heg.cancel == nil {
		switch ctx {
		case nil, context.TODO():
			ctx = context.Background()
		}

		heg.wg.OnError(heg.onError)

		ctx1, cancel := context.WithCancel(ctx)
		heg.ctx = ctx1
		heg.cancel = cancel
	}

	return heg.ctx
}

func (heg *Group) onError(err error) error {
	if err != nil {
		heg.tryCancel()
	}
	return err
}

func (heg *Group) tryCancel() bool {
	if heg.cancelled.CompareAndSwap(false, true) {
		heg.cancel()
		return true
	}
	return false
}

// Cancelled tells if the Group has been cancelled
func (heg *Group) Cancelled() bool {
	return heg.cancelled.Load()
}

// SetContext initialises a Group with a given and externally
// cancellable context.
func (heg *Group) SetContext(ctx context.Context) {
	if heg.cancel != nil {
		panic(syscall.EBUSY)
	}

	heg.init(ctx)
}

// SetLogger sets the slog.Logger to be used when supervising
// workers
func (heg *Group) SetLogger(logger slog.Logger) {
	heg.logger.Store(logger)
}

// Cancel initiates a shutdown of all *http.Server{}s
func (heg *Group) Cancel() error {
	heg.init(context.TODO())

	heg.tryCancel()
	return nil
}

// Go spawns a new Server controlled by the Group
func (heg *Group) Go(srv Server, lsn net.Listener) error {
	if srv == nil || lsn == nil {
		return syscall.EINVAL
	} else if heg.Cancelled() {
		return syscall.ECANCELED
	}

	heg.init(context.TODO())

	w := &Worker{
		Server:   srv,
		Listener: lsn,
	}

	// make a copy of the Listener's Address
	// in case something happens to it
	addr, _ := core.AddrPort(lsn)
	name := fmt.Sprintf("http://%s", addr.String())

	heg.count.Add(1)

	heg.wg.GoCatch(func() error {
		return heg.runWorker(w, name)
	}, func(err error) error {
		return heg.catchWorker(name, err)
	})

	heg.wg.GoCatch(func() error {
		defer heg.count.Add(-1)

		return heg.runSupervisor(w, name)
	}, func(err error) error {
		return heg.catchSupervisor(name, err)
	})

	return nil
}

func (heg *Group) runWorker(w *Worker, name string) error {
	if log, ok := heg.debug(); ok {
		log.Println(name, "started")
	}

	return w.Run()
}

func (heg *Group) catchWorker(name string, err error) error {
	if err != nil {
		if log, ok := heg.error(err); ok {
			log.Println(name, "crashed")
		}
		return err
	}

	if log, ok := heg.debug(); ok {
		log.Println(name, "ended")
	}

	return nil
}

func (heg *Group) runSupervisor(w *Worker, name string) error {
	if log, ok := heg.debug(); ok {
		log.Println(name, "supervisor started")
	}

	// wait for cancellation
	<-heg.ctx.Done()

	if log, ok := heg.debug(); ok {
		log.Println(name, "shutting down")
	}

	return w.Shutdown(context.Background())
}

func (heg *Group) catchSupervisor(name string, err error) error {
	if err != nil {
		if log, ok := heg.error(err); ok {
			log.Println(name, "shutdown error")
		}
		return err
	}

	if log, ok := heg.debug(); ok {
		log.Println(name, "shutdown ended")
	}

	return nil
}

// Count returns how many servers are running in the Group
func (heg *Group) Count() uint {
	count := heg.count.Load()
	if count > 0 {
		return uint(count)
	}
	return 0
}

// Wait blocks until all servers have shut down
func (heg *Group) Wait() error {
	heg.init(context.TODO())

	return heg.wg.Wait()
}
