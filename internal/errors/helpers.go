// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Op represents an operation in the call chain
type Op string

// NewError creates a new Error with the specified code
func NewError(code Code) *Error {
	return &Error{
		Code:    code,
		Fields:  make(map[string]interface{}),
	}
}

// Errorf creates a new Error with a formatted message
func Errorf(code Code, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Fields:  make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code Code, message string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
		Fields:  make(map[string]interface{}),
	}
}

// NotFoundError creates a new NotFound error
func NotFoundError(op, entity, id string) *Error {
	return &Error{
		Code:    NotFound,
		Entity:  entity,
		ID:      id,
		Op:      op,
		Message: "not found",
		Fields:  make(map[string]interface{}),
	}
}

// AlreadyExistsError creates a new AlreadyExists error
func AlreadyExistsError(op, entity, id string) *Error {
	return &Error{
		Code:    AlreadyExists,
		Entity:  entity,
		ID:      id,
		Op:      op,
		Message: "already exists",
		Fields:  make(map[string]interface{}),
	}
}

// InvalidArgumentError creates a new InvalidArgument error
func InvalidArgumentError(op, entity, reason string) *Error {
	return &Error{
		Code:    InvalidArgument,
		Entity:  entity,
		Op:      op,
		Message: reason,
		Fields:  make(map[string]interface{}),
	}
}

// InternalError creates a new Internal error with an underlying cause
func InternalError(op, entity, id string, err error) *Error {
	return &Error{
		Code:    Internal,
		Entity:  entity,
		ID:      id,
		Op:      op,
		Err:     err,
		Fields:  make(map[string]interface{}),
	}
}

// Is checks if an error has a specific code
func Is(err error, code Code) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

// IsNotFound checks if an error indicates a "not found" condition
func IsNotFound(err error) bool {
	return Is(err, NotFound)
}

// IsAlreadyExists checks if an error indicates an "already exists" condition
func IsAlreadyExists(err error) bool {
	return Is(err, AlreadyExists)
}

// IsInvalidArgument checks if an error indicates an "invalid argument" condition
func IsInvalidArgument(err error) bool {
	return Is(err, InvalidArgument)
}

// Match reports whether the error matches all the given criteria
func Match(err error, entity string, code Code) bool {
	var e *Error
	if !errors.As(err, &e) {
		return false
	}
	if entity != "" && e.Entity != entity {
		return false
	}
	return e.Code == code
}

// CodeToHTTPStatus maps error codes to HTTP status codes
func CodeToHTTPStatus(code Code) int {
	switch code {
	case OK:
		return http.StatusOK
	case InvalidArgument:
		return http.StatusBadRequest
	case NotFound:
		return http.StatusNotFound
	case AlreadyExists:
		return http.StatusConflict
	case PermissionDenied:
		return http.StatusForbidden
	case Unauthenticated:
		return http.StatusUnauthorized
	case ResourceExhausted:
		return http.StatusTooManyRequests
	case FailedPrecondition:
		return http.StatusPreconditionFailed
	default:
		return http.StatusInternalServerError
	}
}
