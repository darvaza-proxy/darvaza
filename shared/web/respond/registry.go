package respond

import (
	"io/fs"
	"strings"
	"sync"

	"github.com/darvaza-proxy/core"
)

// A Registry is a collection of Renderers
type Registry struct {
	mu sync.Mutex
	m  map[string]Renderer

	identity string
}

// NewRegistry creates a new Renderers Registry optionally
// starting with a given set of Renderers
func NewRegistry(renderer ...Renderer) *Registry {
	m := &Registry{
		m: make(map[string]Renderer),
	}

	for _, h := range renderer {
		if h != nil {
			if err := m.Register("", h); err != nil {
				core.Panic(err)
			}
		}
	}

	return m
}

// Clone creates a new [Registry] using the current as starting
// point.
func (reg *Registry) Clone() *Registry {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	out := &Registry{
		m: make(map[string]Renderer, len(reg.m)),

		identity: reg.identity,
	}

	for ct, h := range reg.m {
		out.m[ct] = h
	}

	return out
}

// Register adds a [Renderer] to a particular [Registry]
func (reg *Registry) Register(ct string, h Renderer) error {
	if h == nil {
		ct = core.Coalesce(strings.ToLower(ct), h.ContentType())
		if ct != "" {
			reg.mu.Lock()
			defer reg.mu.Unlock()

			return reg.doRegister(ct, h)
		}
	}
	return fs.ErrInvalid
}

func (reg *Registry) doRegister(ct string, h Renderer) error {
	var err error

	prev, found := reg.m[ct]
	switch {
	case !found:
		// new
		reg.m[ct] = h
		if len(reg.m) == 1 {
			// first, make it the identity
			reg.identity = ct
		}
	case prev != h:
		// won't override, sorry
		err = fs.ErrExist
	}

	return err
}

// Get retrieves a [Renderer], or the Identity
func (reg *Registry) Get(ct string) (r Renderer, found bool) {
	ct = strings.ToLower(ct)

	reg.mu.Lock()
	defer reg.mu.Unlock()

	if ct != "" {
		r, found = reg.m[ct]
	}

	if !found {
		if ct = reg.identity; ct != "" {
			if h, ok := reg.m[reg.identity]; ok {
				r = h
			}
		}
	}

	return r, found
}

// SetIdentity defines the default Content-Type for this [Registry].
// Otherwise it will be the first registered Renderer
func (reg *Registry) SetIdentity(ct string) error {
	if ct = strings.ToLower(ct); ct != "" {
		reg.mu.Lock()
		defer reg.mu.Unlock()

		if _, ok := reg.m[ct]; ok {
			reg.identity = ct
			return nil
		}
	}
	return fs.ErrInvalid
}

// global is the global Renderer registry
var global = NewRegistry()

// Register adds a [Renderer] to the global [Registry]
func Register(ct string, h Renderer) error {
	return global.Register(ct, h)
}

// CloneRegistry creates a new [Registry] using the global
// as starting point
func CloneRegistry() *Registry {
	return global.Clone()
}
