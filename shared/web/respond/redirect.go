package respond

import (
	"fmt"
	"net/http"

	"darvaza.org/core"
)

func (r *Response) redirect(rw http.ResponseWriter, loc string, code int) error {
	var err error

	r.WithWriter(rw).WithStatus(code)
	if loc != "" {
		r.SetHeader("Location", loc)
	}
	r.SetHeader("Content-Type", "text/plain; charset=UTF-8")

	r.writeHeaders()
	r.rw.WriteHeader(code)

	if loc == "" {
		// did they set the location manually?
		loc = r.hdrs.Get("Location")
	}

	if loc == "" {
		core.Panic("Redirect without Location")
	}

	_, err = fmt.Fprintln(rw, http.StatusText(code), "Redirecting to", loc)
	return err
}

// MovedPermanently sends a 301 Moved Permanently Response
func (r *Response) MovedPermanently(rw http.ResponseWriter, loc string) error {
	return r.redirect(rw, loc, http.StatusMovedPermanently)
}

// Found sends a 302 Found Response
func (r *Response) Found(rw http.ResponseWriter, loc string) error {
	return r.redirect(rw, loc, http.StatusFound)
}

// SeeOther sends a 303 See Other Response
func (r *Response) SeeOther(rw http.ResponseWriter, loc string) error {
	return r.redirect(rw, loc, http.StatusSeeOther)
}

// TemporaryRedirect sends a 307 Temporary Redirect Response
func (r *Response) TemporaryRedirect(rw http.ResponseWriter, loc string) error {
	return r.redirect(rw, loc, http.StatusTemporaryRedirect)
}

// PermanentRedirect sends a 308 Permanent Redirect Response
func (r *Response) PermanentRedirect(rw http.ResponseWriter, loc string) error {
	return r.redirect(rw, loc, http.StatusPermanentRedirect)
}
