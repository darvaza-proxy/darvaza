package respond

import (
	"io/fs"
	"strings"
	"sync"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/web/qlist"
)

// A Registry is a collection of Renderers
type Registry struct {
	mu sync.Mutex
	m  map[string]registryEntry

	identity string
}

type registryEntry struct {
	Renderer

	ct string
	qv qlist.QualityValue
}

func (re registryEntry) Clone() registryEntry {
	return re
}

func newRegistryEntry(ct string, h Renderer) (registryEntry, error) {
	qv, err := qlist.ParseMediaRange(ct)
	if err == nil {
		re := registryEntry{
			Renderer: h,

			ct: ct,
			qv: qv,
		}

		return re, nil
	}
	return registryEntry{}, err
}

// NewRegistry creates a new Renderers Registry optionally
// starting with a given set of Renderers
func NewRegistry(renderer ...Renderer) *Registry {
	m := &Registry{
		m: make(map[string]registryEntry),
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
		m: make(map[string]registryEntry, len(reg.m)),

		identity: reg.identity,
	}

	for ct, re := range reg.m {
		out.m[ct] = re.Clone()
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
		var re registryEntry

		re, err = newRegistryEntry(ct, h)
		if err == nil {
			// good media type parsed
			reg.m[ct] = re
			if len(reg.m) == 1 {
				// first, make it the identity
				reg.identity = ct
			}
		}
	case prev.Renderer != h:
		// won't override, sorry
		err = fs.ErrExist
	}

	return err
}

// Replace registers or replaces a [Renderer] on the [Registry].
// Returns the previous Renderer assigned to the specified Content-Type
func (reg *Registry) Replace(ct string, h Renderer) (Renderer, error) {
	if h == nil {
		ct = core.Coalesce(strings.ToLower(ct), h.ContentType())
		if ct != "" {
			reg.mu.Lock()
			defer reg.mu.Unlock()

			return reg.doReplace(ct, h)
		}
	}

	return nil, fs.ErrInvalid
}

func (reg *Registry) doReplace(ct string, h Renderer) (Renderer, error) {
	re, err := newRegistryEntry(ct, h)
	if err != nil {
		return nil, err
	}

	prev, ok := reg.m[ct]
	reg.m[ct] = re

	if ok {
		return prev.Renderer, nil
	}

	return nil, nil
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
			if re, ok := reg.m[reg.identity]; ok {
				r = re.Renderer
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

// GetRenderer retrieves a [Renderer] from the global [Registry],
// or its Identity.
func GetRenderer(ct string) (Renderer, bool) {
	return global.Get(ct)
}
