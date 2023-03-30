package respond

import "net/http"

// BadRequest prepares a 400 Bad Request Response
func (r *Response) BadRequest(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusBadRequest)
}

// Unauthorized prepares a 401 Unauthorized Response
func (r *Response) Unauthorized(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusUnauthorized)
}

// Forbidden prepares a 403 Forbidden Response
func (r *Response) Forbidden(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusForbidden)
}

// NotFound prepares a 404 Not Found Response
func (r *Response) NotFound(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusNotFound)
}

// MethodNotAllowed prepares a 405 Method Not Allowed Response
func (r *Response) MethodNotAllowed(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusMethodNotAllowed)
}

// NotAcceptable prepares a 406 Not Acceptable Response
func (r *Response) NotAcceptable(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusNotAcceptable)
}

// Conflict prepares a 409 Conflict Response
func (r *Response) Conflict(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusConflict)
}

// Gone prepares a 410 Gone Response
func (r *Response) Gone(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusGone)
}

// LengthRequired prepares a 411 Length Required Response
func (r *Response) LengthRequired(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusLengthRequired)
}

// PreconditionFailed prepares a 412 Precondition Failed Response
func (r *Response) PreconditionFailed(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusPreconditionFailed)
}

// RequestEntityTooLarge prepares a 413 Request Entity Too Large Response
func (r *Response) RequestEntityTooLarge(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusRequestEntityTooLarge)
}

// UnsupportedMediaType prepares a 415 Unsupported Media Type Response
func (r *Response) UnsupportedMediaType(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusUnsupportedMediaType)
}

// UnprocessableEntity prepares a 422 Unprocessable Entity Response
func (r *Response) UnprocessableEntity(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusUnprocessableEntity)
}
