// Package data provides helpers for generic data types
package data

// SliceMinus returns a new slice containing only the
// elements of one slice not present on the second
func SliceMinus[T comparable](a []T, b []T) []T {
	return SliceMinusFn(a, b, func(va, vb T) bool {
		return va == vb
	})
}

// SliceMinusFn returns a new slice containing only elements
// of slice A that aren't on slice B according to the callback
// eq
func SliceMinusFn[T any](a []T, b []T, eq func(T, T) bool) []T {
	out := make([]T, 0, len(a))

loop:
	for _, v := range a {
		for _, w := range b {
			if eq(v, w) {
				continue loop
			}
		}

		out = append(out, v)
	}

	return out
}
