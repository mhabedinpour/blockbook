package address

import (
	"blockbook/pkg/errors"
	"net/http"
)

var (
	ErrAddressAlreadySubscribed = errors.New("address already subscribed", errors.WithType("addressAlreadySubscribed"), errors.WithStatusCode(http.StatusConflict))
	ErrAddressNotSubscribed     = errors.New("address not subscribed", errors.WithType("addressNotSubscribed"), errors.WithStatusCode(http.StatusNotFound))
)
