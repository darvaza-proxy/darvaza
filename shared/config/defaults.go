package config

import "github.com/amery/defaults"

// SetDefaults applies `defaults` structtags and SetDefaults()
// recursively
func SetDefaults(v any) error {
	return defaults.Set(v)
}
