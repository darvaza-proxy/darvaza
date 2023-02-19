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
