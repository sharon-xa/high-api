package utils

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(code int, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

var (
	ErrNotFound     = NewAPIError(http.StatusNotFound, "Requested resource not found.")
	ErrUnauthorized = NewAPIError(http.StatusUnauthorized, "Invalid credentials.")
	ErrBadRequest   = NewAPIError(http.StatusBadRequest, "Invalid request data.")
	ErrInternal     = NewAPIError(
		http.StatusInternalServerError,
		"Something went wrong, please try again later.",
	)
	ErrStolenToken   = NewAPIError(http.StatusUnauthorized, "kharab allah")
	ErrHeaderMissing = func(headerName string) *APIError {
		return NewAPIError(http.StatusBadRequest, fmt.Sprintf("%s header is missing", headerName))
	}

	// DB Errors
	ErrForeignKeyViolation = func(columnName string) *APIError {
		return NewAPIError(http.StatusBadRequest, fmt.Sprintf("No such %s", columnName))
	}
	ErrUniqueViolation = func(columnName string) *APIError {
		return NewAPIError(http.StatusConflict, fmt.Sprintf("%s already exists", columnName))
	}

	// Token Errors
	ErrTokenExpired       = NewAPIError(http.StatusUnauthorized, jwt.ErrTokenExpired.Error())
	ErrTokenInvalidClaims = NewAPIError(http.StatusUnauthorized, jwt.ErrTokenInvalidClaims.Error())
	ErrParsingToken       = NewAPIError(http.StatusUnauthorized, "unable to parse token")
	ErrInvalidToken       = NewAPIError(http.StatusUnauthorized, "invalid token")

	// Role Errors
	ErrRoleNotAllowed = NewAPIError(http.StatusForbidden, "role not allowed")
)
