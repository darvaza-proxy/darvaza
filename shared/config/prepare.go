package config

import (
	"github.com/amery/defaults"

	"darvaza.org/core"
)

// SetDefaults applies `defaults` structtags and SetDefaults()
// recursively
func SetDefaults(v any) error {
	return defaults.Set(v)
}

// Prepare runs SetDefaults and Validate
func Prepare(v any) error {
	if err := SetDefaults(v); err != nil {
		return core.Wrap(err, "SetDefaults")
	}

	if err := Validate(v); err != nil {
		return core.Wrap(err, "Validate")
	}

	return nil
}
