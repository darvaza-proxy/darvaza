package x

import (
	"container/list"

	"darvaza.org/core"
)

// CloneMapList duplicates a map containing a list.List
func CloneMapList[T comparable](src map[T]*list.List) map[T]*list.List {
	fn := func(v any) (any, bool) { return v, true }
	return CloneMapListFn(src, fn)
}

// CloneMapListFn duplicates a map containing a list.List but
// allows the element's values to be cloned by a helper function
func CloneMapListFn[K comparable, V any](src map[K]*list.List,
	fn func(v V) (V, bool)) map[K]*list.List {
	out := make(map[K]*list.List, len(src))
	for k, l := range src {
		out[k] = cloneList(l, fn)
	}
	return out
}

func cloneList[T any](src *list.List, fn func(v T) (T, bool)) *list.List {
	if fn == nil {
		fn = func(v T) (T, bool) {
			return v, true
		}
	}

	out := list.New()
	core.ListForEach(src, func(v0 T) bool {
		if v1, ok := fn(v0); ok {
			out.PushBack(v1)
		}
		return false
	})
	return out
}
