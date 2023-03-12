package x

import "darvaza.org/core"

// Intersect returns a slice containing the items on both
// provided slices
func Intersect[T comparable](a, b []T) []T {
	return IntersectFn(a, b, func(va, vb T) bool {
		return va == vb
	})
}

// IntersectFn returns a slice containing the items on both
// provided slices using a helper function to compare them
func IntersectFn[T any](a, b []T, eq func(va, vb T) bool) []T {
	out := make([]T, 0, len(a))
	for _, va := range a {
		if core.SliceContainsFn(b, va, eq) {
			out = append(out, va)
		}
	}
	return out
}
