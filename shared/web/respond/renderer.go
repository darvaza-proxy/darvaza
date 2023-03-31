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
