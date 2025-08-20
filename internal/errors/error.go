// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	stdErrors "errors"
	"strings"
)

// Error is the fundamental error type for the application
type Error struct {
	// Code is one of the error codes defined in code.go
	Code Code

	// Entity is the type of entity being operated on (e.g., "Issue")
	Entity string

	// ID is the ID of the entity (if applicable)
	ID string

	// Op is the operation being performed (e.g., "issueHandler.GetIssue")
	Op string

	// Message is a human-readable description of the error
	Message string

	// Err is the underlying error that triggered this one
	Err error

	// Fields contains additional structured metadata about the error
	Fields map[string]interface{}
}

// Error implements the error interface
func (e *Error) Error() string {
	var parts []string

	if e.Op != "" {
		parts = append(parts, e.Op)
	}

	if e.Entity != "" {
		entity := e.Entity
		if e.ID != "" {
			entity += "(" + e.ID + ")"
		}
		parts = append(parts, entity)
	}

	if e.Code != Unknown && e.Code != "" {
		parts = append(parts, string(e.Code))
	}

	var msg string
	if len(parts) > 0 {
		msg = strings.Join(parts, ": ")
	}

	if e.Message != "" {
		if msg != "" {
			msg += ": " + e.Message
		} else {
			msg = e.Message
		}
	} else if e.Err != nil {
		if msg != "" {
			msg += ": " + e.Err.Error()
		} else {
			msg = e.Err.Error()
		}
	}

	return msg
}

// Unwrap implements the errors.Wrapper interface
func (e *Error) Unwrap() error {
	return e.Err
}

// With adds additional context fields to the error
func (e *Error) With(fields map[string]interface{}) *Error {
	// Create a copy of the error
	result := *e

	// Initialize fields if nil
	if result.Fields == nil {
		result.Fields = make(map[string]interface{})
	} else {
		// Create a new fields map with existing values
		newFields := make(map[string]interface{}, len(e.Fields)+len(fields))
		for k, v := range e.Fields {
			newFields[k] = v
		}
		result.Fields = newFields
	}

	// Add new fields
	for k, v := range fields {
		result.Fields[k] = v
	}

	return &result
}

// Is checks if this error matches a target error
// This enables the use of errors.Is with this error type
func (e *Error) Is(target error) bool {
	// If target is an *Error, compare codes
	if t, ok := target.(*Error); ok {
		return e.Code == t.Code
	}

	// If target is a standard error, compare with the wrapped error
	if e.Err != nil {
		return stdErrors.Is(e.Err, target)
	}

	return false
}

// As implements the errors.As interface for error type conversion
func (e *Error) As(target interface{}) bool {
	// If target is an *Error pointer, set it to the current error
	if t, ok := target.(**Error); ok {
		*t = e
		return true
	}

	// Try the wrapped error, if any
	if e.Err != nil {
		return stdErrors.As(e.Err, target)
	}

	return false
}
