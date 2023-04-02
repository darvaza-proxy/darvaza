package respond

import "net/http"

// InternalServerError prepares a 500 Internal Server Error Response
func (r *Response) InternalServerError(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusInternalServerError)
}

// NotImplemented prepares a 501 Not Implemented Response
func (r *Response) NotImplemented(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusNotImplemented)
}

// BadGateway prepares a 502 Bad Gateway Response
func (r *Response) BadGateway(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusBadGateway)
}

// ServiceUnavailable prepares a 503 Service Unavailable Response
func (r *Response) ServiceUnavailable(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusServiceUnavailable)
}

// GatewayTimeout prepares a 504 Gateway Timeout Response
func (r *Response) GatewayTimeout(rw http.ResponseWriter) *Response {
	return r.WithWriter(rw).WithStatus(http.StatusGatewayTimeout)
}
