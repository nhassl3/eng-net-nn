package validator

import (
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	cities = []string{"казань", "челябинск", "нижний новгород"}
)

var (
	ValidCity validator.Func = func(fl validator.FieldLevel) bool {
		if city, ok := fl.Field().Interface().(string); ok {
			return slices.Contains(cities, strings.TrimSpace(strings.ToLower(city)))
		}
		return false
	}
)
