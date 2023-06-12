package config

import (
	"darvaza.org/core"
)

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
