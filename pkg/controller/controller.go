package controller

import (
	"net/http"

	"blockbook/pkg/errors"
	"blockbook/pkg/logging"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

const (
	LoggerName = "logger"

	validationErrorMessage = "Could not validate your request payload"
	internalErrorMessage   = "An internal error has been happened while processing your request"
	oKMessage              = "OK"

	validationErrorType = "ValidationError"
	internalErrorType   = "InternalError"
)

type Controller interface {
	PathPrefix() string
	RegisterHandlers(engine *gin.RouterGroup)
}

//nolint:wrapcheck
func BindBody[T any](c *gin.Context) (T, error) {
	var result T
	err := c.ShouldBindJSON(&result)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return result, errors.ConvertToValidationError(validationErrs)
		}

		return result, ErrMalformedRequest
	}

	return result, nil
}

//nolint:wrapcheck
func BindUri[T any](c *gin.Context) (T, error) {
	var result T
	err := c.ShouldBindUri(&result)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return result, errors.ConvertToValidationError(validationErrs)
		}

		return result, ErrMalformedRequest
	}

	return result, nil
}

func GetLogger(c *gin.Context) *zap.Logger {
	rawLogger, ok := c.Get(LoggerName)
	if !ok {
		panic("logger is not set on the context")
	}
	logger, ok := rawLogger.(*zap.Logger)
	if !ok {
		panic("logger is not set on the context")
	}

	return logger
}

func WriteError(err error, c *gin.Context) {
	logger := GetLogger(c)
	statusCode, response := getGinErrorResponse(logger, err)
	c.JSON(statusCode, response)
}

func WriteSuccess(result any, c *gin.Context) {
	c.JSON(http.StatusOK, getGinResponse(true, result, oKMessage))
}

func getGinErrorResponse(logger *zap.Logger, err error) (int, gin.H) {
	var validationErrs errors.ValidationErrors
	if errors.As(err, &validationErrs) {
		return http.StatusBadRequest, getGinResponse(false, gin.H{
			"errors": validationErrs,
			"type":   validationErrorType,
		}, validationErrorMessage)
	}

	var validationErr errors.ValidationError
	if errors.As(err, &validationErr) {
		return http.StatusBadRequest, getGinResponse(false, gin.H{
			"errors": errors.ValidationErrors{validationErr},
			"type":   validationErrorType,
		}, validationErrorMessage)
	}

	var cErr errors.Error

	// if error is not our custom error type, or it's not a user error, return internal error.
	if !errors.As(err, &cErr) || cErr.StatusCode > 499 || cErr.StatusCode < 400 {
		logger.With(zap.Error(err)).Error("error occurred in api")

		status := http.StatusInternalServerError
		if cErr.StatusCode != 0 {
			status = cErr.StatusCode
		}

		if logging.IsDebug(logger) {
			return status, getGinResponse(false, gin.H{
				"type": cErr.Type,
			}, err.Error())
		}

		return status, getGinResponse(false, gin.H{
			"type": internalErrorType,
		}, internalErrorMessage)
	}

	status := http.StatusBadRequest
	if cErr.StatusCode != 0 {
		status = cErr.StatusCode
	}

	return status, getGinResponse(false, gin.H{
		"type": cErr.Type,
	}, err.Error())
}

func getGinResponse(success bool, result any, message string) gin.H {
	return gin.H{
		"success": success,
		"message": message,
		"result":  result,
	}
}
