package controller

import (
	"net/http"

	"blockbook/pkg/errors"
)

var ErrInvalidValidatorEngine = errors.New("could not get gin validator engine")
var ErrMalformedRequest = errors.New("could not unmarshal request payload", errors.WithStatusCode(http.StatusBadRequest), errors.WithType("MalformedRequest"))
