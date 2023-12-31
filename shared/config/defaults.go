package config

import "github.com/amery/defaults"

// SetDefaults applies `default` struct-tags and SetDefaults()
// recursively. If the given object has a `SetDefaults() error`
// method, it will be invoked instead.
func SetDefaults(v any) error {
	if t, ok := v.(defaults.SetterWithError); ok {
		// defaults.Set will omit this setter on
		// the entry function to prevent loops.
		return t.SetDefaults()
	}

	return defaults.Set(v)
}

// Set applies `default` struct-tags and SetDefaults()
// recursively. If the given object has a `SetDefaults() error`
// method, it will be ignored. Any `SetDefaults() error` deeper
// in the struct will be called.
func Set(v any) error {
	return defaults.Set(v)
}

// CanUpdate returns true when the given value is an initial value of its type
func CanUpdate(v any) bool {
	return defaults.CanUpdate(v)
}
