package data

import (
	"container/list"
)

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
