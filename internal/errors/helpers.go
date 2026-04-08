// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Op represents an operation in the call chain
type Op string

// E creates a new Error from the provided arguments.
// Each argument can be an Op, a string, a Code, an error, or a map[string]any.
func E(args ...any) *Error {
	e := &Error{
		Code:   Unknown,
		Fields: make(map[string]any),
	}

	for _, arg := range args {
		switch a := arg.(type) {
		case Op:
			e.Op = string(a)
		case string:
			// If no entity is set, treat string as entity
			if e.Entity == "" {
				e.Entity = a
			} else if e.ID == "" {
				// If entity is set but ID is not, treat string as ID
				e.ID = a
			} else if e.Message == "" {
				// If both entity and ID are set, treat as message
				e.Message = a
			}
		case Code:
			e.Code = a
		case *Error:
			// Copy the error, but only if not overriding existing values
			copy := *a
			if e.Code == Unknown {
				e.Code = copy.Code
			}

			if e.Entity == "" {
				e.Entity = copy.Entity
			}

			if e.ID == "" {
				e.ID = copy.ID
			}

			if e.Op == "" {
				e.Op = copy.Op
			}

			if e.Message == "" {
				e.Message = copy.Message
			}

			if e.Err == nil {
				e.Err = copy.Err
			}
			// Merge fields
			for k, v := range copy.Fields {
				if e.Fields == nil {
					e.Fields = make(map[string]any)
				}

				if _, exists := e.Fields[k]; !exists {
					e.Fields[k] = v
				}
			}
		case error:
			e.Err = a
		case map[string]any:
			for k, v := range a {
				if e.Fields == nil {
					e.Fields = make(map[string]any)
				}

				e.Fields[k] = v
			}
		}
	}

	return e
}

// Errorf creates a new Error with a formatted message
func Errorf(code Code, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Fields:  make(map[string]any),
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
		Fields:  make(map[string]any),
	}
}

// NotFoundError creates a new NotFound error
func NotFoundError(op, entity, id string) *Error {
	return E(Op(op), entity, id, NotFound, "not found")
}

// AlreadyExistsError creates a new AlreadyExists error
func AlreadyExistsError(op, entity, id string) *Error {
	return E(Op(op), entity, id, AlreadyExists, "already exists")
}

// PermissionDeniedError creates a new PermissionDenied error
func PermissionDeniedError(op, entity, id string) *Error {
	return E(Op(op), entity, id, PermissionDenied, "permission denied")
}

// InvalidArgumentError creates a new InvalidArgument error
func InvalidArgumentError(op, entity, reason string) *Error {
	return &Error{
		Op:      op,
		Entity:  entity,
		Code:    InvalidArgument,
		Message: reason,
		Fields:  make(map[string]any),
	}
}

// InternalError creates a new Internal error with an underlying cause
func InternalError(op, entity, id string, err error) *Error {
	return E(Op(op), entity, id, Internal, err)
}

// Is checks if an error has a specific code
func Is(err error, code Code) bool {
	if e, ok := errors.AsType[*Error](err); ok {
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
