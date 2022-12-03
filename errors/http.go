package errors

import (
	"errors"
	"net/http"
)

// ErrNotFound not found error
var ErrNotFound = &Error{
	HTTPCode: http.StatusNotFound,
	Err:      errors.New("not found"),
	Code:     "NOT_FOUND",
}

// ErrForbidden access is forbidden
var ErrForbidden = &Error{
	HTTPCode: http.StatusForbidden,
	Err:      errors.New("forbidden"),
	Code:     "ACCESS_IS_FORBIDDEN",
}

// ErrInternalServerError internal server error
var ErrInternalServerError = &Error{
	HTTPCode: http.StatusInternalServerError,
	Err:      errors.New("internal server error"),
	Code:     "INTERNAL_SERVER_ERROR",
}

// ErrDatabaseServerError internal server error
var ErrDatabaseServerError = &Error{
	HTTPCode: http.StatusInternalServerError,
	Err:      errors.New("operation database failed"),
	Code:     "INTERNAL_SERVER_ERROR",
}

// ErrUnauthorized 未认证/未授权
var ErrUnauthorized = &Error{
	HTTPCode: http.StatusUnauthorized,
	Err:      errors.New("unauthorized"),
	Code:     "UNAUTHORIZED",
}

// ErrInvalidArguments 无效的参数
var ErrInvalidArguments = &Error{
	HTTPCode: http.StatusBadRequest,
	Err:      errors.New("Bad Request"),
	Code:     "INVALID_ARGUMENTS",
}

// ErrRequestEntityTooLarge 请求实体太大
var ErrRequestEntityTooLarge = &Error{
	HTTPCode: http.StatusRequestEntityTooLarge,
	Err:      errors.New("Request Entity TooLarge"),
	Code:     "REQUEST_ENTITY_TOO_LARGE",
}
