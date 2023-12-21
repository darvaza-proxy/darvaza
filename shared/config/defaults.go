package config

import "github.com/amery/defaults"

// SetDefaults applies `defaults` structtags and SetDefaults()
// recursively
func SetDefaults(v any) error {
	if t, ok := v.(defaults.SetterWithError); ok {
		// defaults.Set will omit this setter on
		// the entry function to prevent loops.
		return t.SetDefaults()
	}

	return defaults.Set(v)
}

// CanUpdate returns true when the given value is an initial value of its type
func CanUpdate(v any) bool {
	return defaults.CanUpdate(v)
}
