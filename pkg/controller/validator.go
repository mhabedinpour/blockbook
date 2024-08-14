package controller

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const (
	SplitNParam = 2
)

func setupValidator() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return ErrInvalidValidatorEngine
	}
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", SplitNParam)[0]
		formName := strings.SplitN(field.Tag.Get("form"), ",", SplitNParam)[0]
		if name == "-" || formName == "-" {
			return ""
		} else if formName != "" {
			return formName
		}

		return name
	})

	return nil
}
