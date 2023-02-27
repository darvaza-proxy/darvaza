package data

import (
	"container/list"
)

// ListContains checks if a container/list contains an element
func ListContains[T comparable](l *list.List, val T) bool {
	return ListContainsFn(l, func(v T) bool {
		return v == val
	})
}

// ListContainsFn checks if a container/list contains an element
// that satisfies a given function
func ListContainsFn[T comparable](l *list.List, match func(T) bool) bool {
	var found bool

	if l != nil && match != nil {
		ListForEach(l, func(v T) bool {
			found = match(v)
			return found
		})
	}
	return found
}

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
	ListForEach(src, func(v0 T) bool {
		if v1, ok := fn(v0); ok {
			out.PushBack(v1)
		}
		return false
	})
	return out
}

// ListForEach calls a function for each value until told to stop
func ListForEach[T any](l *list.List, fn func(v T) bool) {
	if l == nil || fn == nil {
		return
	}

	ListForEachElement(l, func(e *list.Element) bool {
		if v, ok := e.Value.(T); ok {
			return fn(v)
		}
		return false
	})
}

// ListForEachElement calls a function for each element until told to stop
func ListForEachElement(l *list.List, fn func(*list.Element) bool) {
	if l == nil || fn == nil {
		return
	}

	for e := l.Front(); e != nil; e = e.Next() {
		if fn(e) {
			break
		}
	}
}