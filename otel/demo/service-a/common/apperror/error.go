package apperror

import (
	"fmt"
	"net/http"
	"thanhldt060802/common/constant"

	"github.com/danielgtaylor/huma/v2"
)

type CustomError struct {
	error
	Status   int      `json:"status"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	ErrorMsg string   `json:"error,omitempty"`
	Details  []string `json:"details,omitempty"`
}

func NewCustomError(status int, code string, message string, errs ...error) huma.StatusError {
	details := make([]string, len(errs))
	for i, err := range errs {
		details[i] = err.Error()
	}
	errMsg := message
	if len(details) > 0 {
		errMsg = details[0]
	}
	return &CustomError{
		Status:   status,
		Code:     code,
		Message:  message,
		ErrorMsg: errMsg,
		Details:  details,
	}
}

func (e *CustomError) Error() string {
	if e.error != nil {
		return e.error.Error()
	}
	return e.Message
}

func (e *CustomError) GetStatus() int {
	return e.Status
}

func NewHumaError(status int, message string, errs ...error) huma.StatusError {
	var appError = &CustomError{
		Status:  status,
		Code:    message,
		Message: message,
	}

	if len(errs) > 0 {
		details := []string{}
		for _, err := range errs {
			details = append(details, err.Error())
		}
		appError.Details = details
	}
	return appError
}

func ErrServiceUnavailable(err error, message string, details ...string) huma.StatusError {
	return &CustomError{
		error:    err,
		Status:   http.StatusServiceUnavailable,
		Message:  message,
		Code:     string(constant.ERR_SERVICE_UNAVAILABLE),
		ErrorMsg: fmt.Sprintf("%s: %s", constant.ERR_SERVICE_UNAVAILABLE, message),
		Details:  details,
	}
}

func ErrBadRequest(message string, locs ...string) *CustomError {
	details := make([]string, len(locs))
	copy(details, locs)
	return &CustomError{
		Status:   http.StatusBadRequest,
		Message:  message,
		Code:     string(constant.ERR_BAD_REQUEST),
		ErrorMsg: fmt.Sprintf("%s: %s", constant.ERR_BAD_REQUEST, message),
		Details:  details,
	}
}

func ErrUnauthorized(err error, message string, details ...string) *CustomError {
	return &CustomError{
		error:    err,
		Status:   http.StatusUnauthorized,
		Message:  message,
		Code:     string(constant.ERR_UNAUTHORIZED),
		ErrorMsg: fmt.Sprintf("%s: %s", constant.ERR_UNAUTHORIZED, message),
		Details:  details,
	}
}

func ErrForbidden(err error, message string, details ...string) *CustomError {
	return &CustomError{
		error:    err,
		Status:   http.StatusForbidden,
		Message:  message,
		Code:     string(constant.ERR_FORBIDDEN),
		ErrorMsg: fmt.Sprintf("%s: %s", constant.ERR_FORBIDDEN, message),
		Details:  details,
	}
}

func ErrNotFound(message string, notFoundCode string, details ...string) *CustomError {
	return &CustomError{
		Status:   http.StatusNotFound,
		Message:  message,
		Code:     notFoundCode,
		ErrorMsg: message,
		Details:  details,
	}
}

func ErrConflict(message string, conflictCode string, details ...string) *CustomError {
	return &CustomError{
		Status:   http.StatusConflict,
		Message:  message,
		Code:     conflictCode,
		ErrorMsg: message,
		Details:  details,
	}
}

func ErrInternalServerError(err error, message string, internalServerErrorCode string, errs ...error) *CustomError {
	var details []string
	if len(errs) > 0 {
		for _, e := range errs {
			details = append(details, e.Error())
		}
	}
	if err != nil {
		details = append(details, err.Error())
	}
	return &CustomError{
		error:    err,
		Status:   http.StatusInternalServerError,
		Message:  message,
		Code:     internalServerErrorCode,
		ErrorMsg: message,
		Details:  details,
	}
}
