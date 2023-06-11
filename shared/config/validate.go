package config

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate validates exposed fields including nested structs
func Validate(v any) error {
	return validate.Struct(v)
}

// AsValidationErrors gives access to a slice of [validator.FieldError]
func AsValidationErrors(err error) (validator.ValidationErrors, bool) {
	p, ok := err.(validator.ValidationErrors)
	return p, ok
}
