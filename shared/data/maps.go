package data

// MapContains tells if a given map contains a key.
// this helper is inteded for switch/case conditions
func MapContains[K comparable](m map[K]any, key K) bool {
	_, ok := m[key]
	return ok
}

// MappedValueOr returns a value of an entry or a default if
// not found
func MappedValueOr[K comparable, V any](m map[K]V, key K, def V) (V, bool) {
	if val, ok := m[key]; ok {
		return val, true
	}
	return def, false
}
