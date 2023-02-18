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

	for _, v := range a {
		if !SliceContainsFn(b, v, eq) {
			out = append(out, v)
		}
	}

	return out
}

// SliceContains tells if a slice contains a given element
func SliceContains[T comparable](a []T, v T) bool {
	return SliceContainsFn(a, v, func(va, vb T) bool {
		return va == vb
	})
}

// SliceContainsFn tells if a slice contains a given element
// according to the callback eq
func SliceContainsFn[T any](a []T, v T, eq func(T, T) bool) bool {
	for _, va := range a {
		if eq(va, v) {
			return true
		}
	}
	return false
}
