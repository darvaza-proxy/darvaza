package simple

import (
	"errors"
	"strings"

	"darvaza.org/core"
)

var (
	_ error            = (*ErrInvalidCert)(nil)
	_ core.Unwrappable = (*ErrInvalidCert)(nil)
)

var (
	// ErrNotImplemented is returned when something hasn't been implemented yet
	ErrNotImplemented = errors.New("not implemented")
)

// ErrInvalidCert indicates the certificate can't be used
type ErrInvalidCert struct {
	Err    error
	Reason string
}

func (err ErrInvalidCert) Error() string {
	s := make([]string, 0, 3)
	s = append(s, "invalid certificate")

	if err.Reason != "" {
		s = append(s, err.Reason)
	}

	if err.Err != nil {
		s = append(s, err.Err.Error())
	}

	return strings.Join(s, ": ")
}

func (err ErrInvalidCert) Unwrap() error {
	return err.Err
}
