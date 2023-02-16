// Package data provides helpers for generic data types
package data

// SliceMinus returns a new slice containing only the
// elements of one slice not present on the second
func SliceMinus[T comparable](A []T, B []T) []T {
	return SliceMinusFn(A, B, func(a, b T) bool {
		return a == b
	})
}

// SliceMinusFn returns a new slice containing only elements
// of slice A that aren't on slice B according to the callback
// eq
func SliceMinusFn[T any](A []T, B []T, eq func(T, T) bool) []T {
	out := make([]T, 0, len(A))

loop:
	for _, v := range A {
		for _, w := range B {
			if eq(v, w) {
				continue loop
			}
		}

		out = append(out, v)
	}

	return out
}
