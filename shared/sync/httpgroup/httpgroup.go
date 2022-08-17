// Package httpgroup implements an errgroup for HTTP Servers
package httpgroup

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"syscall"

	"github.com/darvaza-proxy/core"
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
	cancelled int32
	count     int32

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
	if atomic.CompareAndSwapInt32(&heg.cancelled, 0, 1) {
		heg.cancel()
		return true
	}
	return false
}

// Cancelled tells if the Group has been cancelled
func (heg *Group) Cancelled() bool {
	return atomic.LoadInt32(&heg.cancelled) != 0
}

// SetContext initialises a Group with a given and externally
// cancellable context.
func (heg *Group) SetContext(ctx context.Context) {
	if heg.cancel != nil {
		panic(syscall.EBUSY)
	}

	heg.init(ctx)
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

	atomic.AddInt32(&heg.count, 1)

	heg.wg.Go(func() error {
		return w.Run()
	})

	heg.wg.Go(func() error {
		defer atomic.AddInt32(&heg.count, -1)

		<-heg.ctx.Done()
		return w.Shutdown(context.Background())
	})

	return nil
}

// Count returns how many servers are running in the Group
func (heg *Group) Count() uint {
	count := atomic.LoadInt32(&heg.count)
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
