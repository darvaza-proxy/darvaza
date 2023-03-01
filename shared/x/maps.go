package x

// MappedValueOr returns a value of an entry or a default if
// not found
func MappedValueOr[K comparable, V any](m map[K]V, key K, def V) (V, bool) {
	if val, ok := m[key]; ok {
		return val, true
	}
	return def, false
}
