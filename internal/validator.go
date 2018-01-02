package internal

import "gopkg.in/go-playground/validator.v9"

var v *validator.Validate

// returns a validator.Validate instance
// to be re-used internally
func validate() *validator.Validate {
	if v != nil {
		return v
	}
	v = validator.New()
	return v
}