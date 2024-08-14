package errors

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

const (
	RequiredConstraint      = "required"
	MaxConstraint           = "max"
	OneOfConstraint         = "oneof"
	EqualConstraint         = "eq"
	LessThanFieldConstraint = "ltfield"
	MoreThanFieldConstraint = "gtfield"
	PositiveConstraint      = "positive"
)

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	result := ""
	for i, validationError := range v {
		result += fmt.Sprintf("%d. %s;", i, validationError.Error())
	}

	return result
}

func (v *ValidationErrors) AddError(validationError ValidationError) {
	*v = append(*v, validationError)
}

// ValidationError is used to construct user validation errors in controller pkg.
type ValidationError struct {
	baseErr         error
	field           string
	constraint      string
	constraintParam string
	Message         string `json:"message"`
}

var _ error = ValidationError{}

func (v ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Error string `json:"error"`
	}{
		Error: v.Message,
	})
}
func (v ValidationError) Error() string {
	return v.Message
}

func (v ValidationError) getMessage() string {
	switch v.constraint {
	case RequiredConstraint:
		return fmt.Sprintf("`%s` field is required", v.field)
	case MaxConstraint:
		return fmt.Sprintf("maximum for `%s` field is %s", v.field, v.constraintParam)
	case OneOfConstraint:
		return fmt.Sprintf("the value of `%s` field should be one of: %s", v.field, v.constraintParam)
	case EqualConstraint:
		return fmt.Sprintf("the value of `%s` field should be equal to %s", v.field, v.constraintParam)
	case LessThanFieldConstraint:
		return fmt.Sprintf("the value of `%s` field should be less than `%s` field", v.field, v.constraintParam)
	case MoreThanFieldConstraint:
		return fmt.Sprintf("the value of `%s` field should be more than `%s` field", v.field, v.constraintParam)
	case PositiveConstraint:
		return fmt.Sprintf("`%s` field is required and should be positive", v.field)
	default:
		return v.defaultMessage()
	}
}

func (v ValidationError) defaultMessage() string {
	message := fmt.Sprintf("%s: field `%s` constraint `%s` param `%s`", v.baseErr.Error(), v.field, v.constraint, v.constraintParam)
	if v.field != "" && v.constraint != "" && v.constraintParam == "" {
		message = fmt.Sprintf("constraint `%s` failed for field: `%s`", v.constraint, v.field)
	}
	if v.field != "" && v.constraint != "" && v.constraintParam != "" {
		message = fmt.Sprintf("constraint `%s` with param `%s` failed for field: `%s`", v.constraint, v.constraintParam, v.field)
	}

	return message
}

type ValidationErrorOptions func(err *ValidationError)

func WithFieldAndConstraint(field string, constraint string) ValidationErrorOptions {
	return func(err *ValidationError) {
		err.field = field
		err.constraint = constraint
		err.Message = err.getMessage()
	}

}

func WithFieldAndConstraintAndParam(field string, constraint string, param string) ValidationErrorOptions {
	return func(err *ValidationError) {
		err.field = field
		err.constraint = constraint
		err.constraintParam = param
		err.Message = err.getMessage()
	}
}

func NewValidationError(baseErr error, options ...ValidationErrorOptions) ValidationError {
	err := ValidationError{
		baseErr: baseErr,
	}

	for _, option := range options {
		option(&err)
	}
	err.Message = err.getMessage()

	return err
}

func ConvertToValidationError(errors validator.ValidationErrors) ValidationErrors {
	result := make([]ValidationError, 0, len(errors))
	for _, fieldError := range errors {
		result = append(result, getValidationError(fieldError))
	}

	return result
}

func getValidationError(fieldError validator.FieldError) ValidationError {
	return NewValidationError(fieldError, WithFieldAndConstraintAndParam(fieldError.Field(), fieldError.Tag(), fieldError.Param()))
}
