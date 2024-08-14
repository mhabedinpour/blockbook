package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnknownErrorType = "unknown"
)

type Error struct {
	parentErr error

	Message string
	Type    string
	// StatusCode is used by the `controller` pkg to return the appropriate HTTP status code in response.
	StatusCode int
}

var _ error = Error{}

func (c Error) Error() string {
	result := ""
	if c.Type != "" {
		result += fmt.Sprintf("[TYPE:%s] ", c.Type)
	}
	result += c.Message
	if c.parentErr != nil {
		result += ":" + c.parentErr.Error()
	}

	return result
}

func (c Error) Unwrap() error {
	if c.parentErr != nil {
		return c.parentErr
	}

	return nil
}

type Options func(err error) error

func Wrap(err error, msg string, options ...Options) error {
	var cErr Error
	if errors.As(err, &cErr) {
		var result error = &Error{
			parentErr: cErr,
			Message:   msg,
			Type:      cErr.Type,
		}

		for _, option := range options {
			result = option(result)
		}

		return result
	}

	return errors.Wrap(err, msg)
}

func WithType(typ string) Options {
	return func(err error) error {
		var cErr Error
		if As(err, &cErr) {
			cErr.Type = typ

			return cErr
		}

		return Error{
			Message: err.Error(),
			Type:    typ,
		}
	}
}

func WithStatusCode(status int) Options {
	return func(err error) error {
		var cErr Error
		if As(err, &cErr) {
			cErr.StatusCode = status

			return cErr
		}

		return Error{
			Message:    err.Error(),
			StatusCode: status,
		}
	}
}

func New(msg string, options ...Options) error {
	if len(options) == 0 {
		return errors.New(msg)
	}

	var result error = Error{
		Message: msg,
		Type:    UnknownErrorType,
	}

	for _, option := range options {
		result = option(result)
	}

	return result
}

func Is(err, target error) bool { return errors.Is(err, target) }

func As(err error, target any) bool { return errors.As(err, target) }
