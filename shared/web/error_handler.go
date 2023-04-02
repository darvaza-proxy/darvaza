package web

import (
	"context"
	"net/http"

	"darvaza.org/core"
)

// ErrorHandlerFunc is the signature of a function used as ErrorHandler
type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

// WithErrorHandler attaches an ErrorHandler function to a context
// for later retrieval
func WithErrorHandler(ctx context.Context, h ErrorHandlerFunc) context.Context {
	return errCtxKey.WithValue(ctx, h)
}

// ErrorHandler attempts to pull an ErrorHandler from the context.Context
func ErrorHandler(ctx context.Context) (ErrorHandlerFunc, bool) {
	return errCtxKey.Get(ctx)
}

var (
	errCtxKey = core.NewContextKey[ErrorHandlerFunc]("ErrorHandler")
)
