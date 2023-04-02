package respond

import "io"

// A RenderFunc writes the v object encoded
type RenderFunc func(w io.Writer, v any) error

// Renderer is an object that renders objects with a
// predetermined encoding
type Renderer interface {
	Render(w io.Writer, v any) error
	ContentType() string
}

var (
	_ Renderer = (*renderer)(nil)
)

type renderer struct {
	ct string
	h  RenderFunc
}

func (r renderer) ContentType() string             { return r.ct }
func (r renderer) Render(w io.Writer, v any) error { return r.h(w, v) }

// NewRenderer creates a Renderer implementation using
// the given content type string and render functions
func NewRenderer(ct string, h RenderFunc) Renderer {
	return &renderer{
		ct: ct,
		h:  h,
	}
}
