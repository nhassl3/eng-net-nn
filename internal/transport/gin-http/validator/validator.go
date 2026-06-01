package validator

import (
	"slices"

	"github.com/go-playground/validator/v10"
)

var (
	cities = []string{"дзержинск", "кстово", "красные баки", "нижний новгород"}
)

var (
	ValidCity validator.Func = func(fl validator.FieldLevel) bool {
		if city, ok := fl.Field().Interface().(string); ok {
			return slices.Contains(cities, city)
		}
		return false
	}
)
