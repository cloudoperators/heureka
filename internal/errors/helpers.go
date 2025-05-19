// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

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
// Each argument can be an Op, a string, a Code, an error, or a map[string]interface{}.
func E(args ...interface{}) *Error {
	e := &Error{
		Code:   Unknown,
		Fields: make(map[string]interface{}),
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
					e.Fields = make(map[string]interface{})
				}
				if _, exists := e.Fields[k]; !exists {
					e.Fields[k] = v
				}
			}
		case error:
			e.Err = a
		case map[string]interface{}:
			for k, v := range a {
				if e.Fields == nil {
					e.Fields = make(map[string]interface{})
				}
				e.Fields[k] = v
			}
		}
	}

	return e
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

// NotFound creates a new NotFound error
func NotFound(op, entity, id string) *Error {
	return E(Op(op), entity, id, NotFound, "not found")
}

// AlreadyExists creates a new AlreadyExists error
func AlreadyExists(op, entity, id string) *Error {
	return E(Op(op), entity, id, AlreadyExists, "already exists")
}

// InvalidArgument creates a new InvalidArgument error
func InvalidArgument(op, entity, reason string) *Error {
	return E(Op(op), entity, InvalidArgument, reason)
}

// InternalError creates a new Internal error with an underlying cause
func InternalError(op, entity, id string, err error) *Error {
	return E(Op(op), entity, id, Internal, err)
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
