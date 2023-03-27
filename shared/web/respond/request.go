package respond

import (
	"context"
	"net/http"

	"github.com/darvaza-proxy/darvaza/shared/web/qlist"
)

// Request is a processed representation of [http.Request]
// to facilitate composing a [http.Response]
type Request struct {
	Context context.Context

	Supported []string
	Accepted  qlist.QualityList
}

// WithRequest creates a [Request] from a [http.Request]
func WithRequest(req http.Request) (*Request, error) {
	r := &Request{
		Context: req.Context(),
	}

	// Accept
	accepted, err := qlist.ParseQualityHeader(req.Header, "Accept")
	if err != nil {
		return r, err
	}
	r.Accepted = accepted

	return r, nil
}

// Supports specifies the supported Content-Types
func (r *Request) Supports(types ...string) *Request {
	r.Supported = append(r.Supported, types...)
	return r
}

// PreferredContentType decides what encoding should be used
func (r *Request) PreferredContentType() string {
	best, _, _ := qlist.BestQualityWithIdentity(r.Supported, r.Accepted, "identity")
	return best
}
