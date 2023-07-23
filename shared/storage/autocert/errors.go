package autocert

import (
	"io/fs"
	"sync/atomic"
)

var (
	_ error = (*errTimeout)(nil)
)

type errTimeout struct {
	Err string
}

func (e errTimeout) Error() string {
	return "timeout: " + e.Err
}

func (errTimeout) Timeout() bool { return true }

func mkcertErrUnknown(name string) *fs.PathError {
	return &fs.PathError{
		Path: name,
		Op:   "mkcert",
		Err:  errTimeout{"unexpected error"},
	}
}

func mkcertErrTimeout(name string) *fs.PathError {
	return &fs.PathError{
		Path: name,
		Op:   "mkcert",
		Err:  errTimeout{"failed to get certificate on time"},
	}
}

func mkcertErrTimeoutUnlessAtomic(name string, atomicErr *atomic.Value) error {
	if atomicErr != nil {
		if err, ok := atomicErr.Load().(error); ok {
			return err
		}
	}

	return mkcertErrTimeout(name)
}
