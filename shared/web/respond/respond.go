// Package respond assists handlers to produce reponses
package respond

import (
	"net/http"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/web"
	"darvaza.org/darvaza/shared/web/qlist"
)

// A Responder is a narrowed set of supported Media Types
type Responder struct {
	registry *Registry

	supported []string
	ql        qlist.QualityList
	identity  string
}

// Supports sets additional supported media types on the
// [Responder], but they need to exist in the registry
func (res *Responder) Supports(types ...string) *Responder {
	for _, ct := range types {
		qv, ok := res.registry.GetParsed(ct)
		if !ok {
			core.Panicf("registry doesn't support %q", ct)
		}

		if len(res.supported) == 0 {
			// first
			res.identity = ct
		}

		res.supported = append(res.supported, ct)
		res.ql = append(res.ql, qv)
	}

	return res
}

// Response is the stage to build and send the response for a particular request
type Response struct {
	res  *Responder
	req  *http.Request
	rw   http.ResponseWriter
	hdrs http.Header
	code int
	h    Renderer
}

// ContentType is the preferred Media Type for this Response
func (r *Response) ContentType() string {
	return r.h.ContentType()
}

// Code returns the status set for this response
func (r *Response) Code() int {
	if r.code == 0 {
		return http.StatusOK
	}
	return r.code
}

// AddHeader appends a value to a header key
// for the response
func (r *Response) AddHeader(key, value string) {
	r.hdrs.Add(key, value)
}

// SetHeader sets the value of a header key
// for the response
func (r *Response) SetHeader(key, value string) {
	r.hdrs.Set(key, value)
}

// DeleteHeader removes a header key from the
// response
func (r *Response) DeleteHeader(key string) {
	r.hdrs.Del(key)
}

func (r *Response) writeHeaders() {
	var hasCT bool

	hdrs := r.rw.Header()
	for key, s := range r.hdrs {
		if key == "Content-Type" {
			hasCT = true
		}

		hdrs[key] = append(hdrs[key], s...)
	}

	if !hasCT {
		if ct := r.h.ContentType(); ct != "" {
			hdrs["Content-Type"] = []string{ct}
		}
	}
}

// WithWriter sets the writer directly, to be used in conjunction with [WithStatus].
func (r *Response) WithWriter(rw http.ResponseWriter) *Response {
	var err string

	// these panic because they indicate broken code
	switch {
	case rw == nil:
		err = "writer not provided"
	case r.rw != nil:
		err = "you can only set the writer once"
	default:
		r.rw = rw
	}

	if err != "" {
		core.Panic(err)
	}

	return r
}

// WithStatus sets the status code directly, to be used in conjunction with [WithWriter].
func (r *Response) WithStatus(code int) *Response {
	if code == 0 || r.code != 0 {
		core.Panic("invalid call to WithStatus")
	}
	r.code = code
	return r
}

// Render sends the rendered response to the user
func (r *Response) Render(v any) error {
	if r.rw == nil {
		core.Panic("call to Render without having provided the writer")
	}

	if err, ok := v.(error); ok {
		if h, ok := web.ErrorHandler(r.req.Context()); ok {
			// pass errors to the designated error handler
			wee := web.NewHTTPError(r.code, err, "")
			for k, s := range r.hdrs {
				// copy headers
				wee.Hdr[k] = append(wee.Hdr[k], s...)
			}

			h(r.rw, r.req, wee)
			return nil
		}
	}

	r.writeHeaders()
	r.rw.WriteHeader(r.Code())
	return r.h.Render(r.rw, v)
}

// WithRequest builds a new Response for the best accepted type
func (res *Responder) WithRequest(req *http.Request) (*Response, error) {
	accepted, err := qlist.ParseMediaRangeHeader(req.Header)
	if err != nil {
		return nil, web.NewHTTPError(http.StatusBadRequest, err, "Invalid Accept Header")
	}

	preferred, _, _ := qlist.BestQualityParsed(res.ql, accepted)
	if preferred == "" {
		preferred = res.identity
	}

	h, ok := res.registry.Get(preferred)
	if !ok {
		core.Panicf("%q renderer disappeared", preferred)
	}

	r := &Response{
		res: res,
		req: req,
		h:   h,
	}

	return r, nil
}
