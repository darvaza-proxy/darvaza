package web

import (
	"errors"
	"fmt"
	"net/http"

	"darvaza.org/core"
)

var (
	_ Error            = (*HTTPError)(nil)
	_ core.Unwrappable = (*HTTPError)(nil)
	_ http.Handler     = (*HTTPError)(nil)
)

// Error is an error that knows its HTTP Status Code
type Error interface {
	Error() string
	Status() int
}

// HTTPError extends [core.WrappedError] with HTTP Status Code
type HTTPError struct {
	Err  error
	Code int
	Hdr  http.Header
}

// Status returns the StatusCode associated with the Error
func (err *HTTPError) Status() int {
	switch {
	case err.Code == 0:
		return http.StatusOK
	case err.Code < 0:
		return http.StatusInternalServerError
	default:
		return err.Code
	}
}

// Header returns a [http.Header] attached to this error for custom fields
func (err *HTTPError) Header() http.Header {
	if err.Hdr == nil {
		err.Hdr = make(http.Header)
	}
	return err.Hdr
}

// AddHeader appends a value to an HTTP header entry of the HTTPError
func (err *HTTPError) AddHeader(key, value string) {
	err.Header().Add(key, value)
}

// SetHeader sets the value of a header key of the HTTPError
func (err *HTTPError) SetHeader(key, value string) {
	err.Header().Set(key, value)
}

// DeleteHeader removes a header key from the HTTPError if present
func (err *HTTPError) DeleteHeader(key string) {
	err.Header().Del(key)
}

// ServeHTTP is a very primitive handler that will try to pass the error
// to a [middleware.ErrorHandlerFunc] provided via the request's context.Context
func (err *HTTPError) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h, ok := ErrorHandler(req.Context()); ok {
		// pass over to the error handler
		h(rw, req, err)
		return
	}

	code := err.Status()

	switch {
	case code == http.StatusOK:
		rw.WriteHeader(http.StatusNoContent)
	case code < http.StatusBadRequest:
		rw.WriteHeader(code)
	default:
		err.writeError(rw)
	}
}

func (err *HTTPError) writeError(rw http.ResponseWriter) {
	hdr := rw.Header()
	for k, s := range err.Hdr {
		// apply headers
		hdr[k] = append(hdr[k], s...)
	}
	// override media type
	hdr["Content-Type"] = []string{"text/plain; charset=UTF-8"}

	code := err.Status()

	rw.WriteHeader(code)
	fmt.Fprintln(rw, ErrorText(code))

	if err.Err != nil {
		if msg := err.Err.Error(); msg != "" {
			fmt.Fprint(rw, "\n", msg)
		}
	}
}

func (err *HTTPError) Error() string {
	var msg string

	text := ErrorText(err.Status())
	if err.Err != nil {
		msg = err.Err.Error()
	}
	if msg == "" {
		return text
	}

	return fmt.Sprintf("%s: %s", text, msg)
}

func (err *HTTPError) Unwrap() error {
	return err.Err
}

// NewHTTPError creates a new HTTPError with a given StatusCode
// and optional cause and annotation
func NewHTTPError(code int, err error, note string) *HTTPError {
	switch {
	case err != nil:
		err = core.Wrap(err, note)
	case note != "":
		err = errors.New(note)
	}

	return &HTTPError{Err: err, Code: code}
}

// NewHTTPErrorf creates a new HTTPError with a given StatusCode
// and optional cause and formatted annotation
func NewHTTPErrorf(code int, err error, format string, args ...any) *HTTPError {
	switch {
	case err != nil:
		err = core.Wrap(err, format, args...)
	default:
		err = fmt.Errorf(format, args...)
		if err.Error() == "" {
			err = nil
		}
	}

	return &HTTPError{Err: err, Code: code}
}

// ErrorText returns the title corresponding to
// a given HTTP Status code
func ErrorText(code int) string {
	text := http.StatusText(code)

	switch {
	case text == "":
		text = fmt.Sprintf("Unknown Error %d", code)
	case code >= 400:
		text = fmt.Sprintf("%s (Error %d)", text, code)
	}

	return text
}
