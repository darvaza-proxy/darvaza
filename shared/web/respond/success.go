package respond

import (
	"net/http"
)

// OK prepares a 200 OK Response
func (r *Response) OK(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusOK)
}

// Created prepares a 201 Created Response
func (r *Response) Created(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusCreated)
}

// Accepted prepares a 202 Accepted Response
func (r *Response) Accepted(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusAccepted)
}

// NoContent sends a 204 No Content Response. Render isn't used.
func (r *Response) NoContent(rw http.ResponseWriter) {
	r.WithWriter(rw).WithStatus(http.StatusNoContent)

	r.hdrs.Del("Content-Type")

	r.writeHeaders()
	r.rw.WriteHeader(r.Code())
}
