// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"errors"
	"strings"

	appErrors "github.com/cloudoperators/heureka/internal/errors"
)

// GraphQLError represents a standard GraphQL error response
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Error implements the error interface
func (e *GraphQLError) Error() string {
	return e.Message
}

// GraphQL error codes
const (
	// Standard GraphQL error codes
	ErrorCodeBadRequest      = "BAD_REQUEST"
	ErrorCodeNotFound        = "NOT_FOUND"
	ErrorCodeInternal        = "INTERNAL"
	ErrorCodeUnauthenticated = "UNAUTHENTICATED"
	ErrorCodeForbidden       = "FORBIDDEN"
	ErrorCodeConflict        = "CONFLICT"
	ErrorCodeValidation      = "VALIDATION"
)

// NewGraphQLError creates a new GraphQL error with the given message and code
func NewGraphQLError(message string, code string) *GraphQLError {
	return &GraphQLError{
		Message: message,
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
}

// ToGraphQLError converts an application error to a GraphQL error
func ToGraphQLError(err error) *GraphQLError {
	if err == nil {
		return nil
	}

	// Try to extract our application error
	var appErr *appErrors.Error
	if !errors.As(err, &appErr) {
		// Handle non-app errors
		return &GraphQLError{
			Message: "An unexpected error occurred",
			Extensions: map[string]interface{}{
				"code": ErrorCodeInternal,
			},
		}
	}

	// Map the error code to a GraphQL error code
	code := appErrorCodeToGraphQL(appErr.Code)
	message := sanitizeErrorMessage(appErr)

	// Create GraphQL error with extensions
	gqlErr := &GraphQLError{
		Message: message,
		Extensions: map[string]interface{}{
			"code": code,
		},
	}

	// Add entity information if available
	if appErr.Entity != "" {
		gqlErr.Extensions["entity"] = appErr.Entity
	}

	// Add ID information if available
	if appErr.ID != "" {
		gqlErr.Extensions["id"] = appErr.ID
	}

	// Add operation information for debugging (not exposing to client)
	if appErr.Op != "" && appErr.Code == appErrors.Internal {
		// Only include operation info for internal errors and only in development
		// This could be controlled by environment variables
		gqlErr.Extensions["operation"] = appErr.Op
	}

	// Add fields if they exist and this is not an internal error
	if appErr.Code != appErrors.Internal && appErr.Fields != nil && len(appErr.Fields) > 0 {
		gqlErr.Extensions["fields"] = appErr.Fields
	}

	return gqlErr
}

// appErrorCodeToGraphQL maps application error codes to GraphQL error codes
func appErrorCodeToGraphQL(code appErrors.Code) string {
	switch code {
	case appErrors.InvalidArgument:
		return ErrorCodeBadRequest
	case appErrors.NotFound:
		return ErrorCodeNotFound
	case appErrors.AlreadyExists:
		return ErrorCodeConflict
	case appErrors.PermissionDenied:
		return ErrorCodeForbidden
	case appErrors.Unauthenticated:
		return ErrorCodeUnauthenticated
	case appErrors.ResourceExhausted:
		return ErrorCodeBadRequest
	case appErrors.FailedPrecondition:
		return ErrorCodeBadRequest
	default:
		return ErrorCodeInternal
	}
}

// sanitizeErrorMessage creates a client-friendly error message
func sanitizeErrorMessage(appErr *appErrors.Error) string {
	// For internal errors, provide a generic message
	if appErr.Code == appErrors.Internal {
		return "An internal server error occurred"
	}

	// For other errors, create a user-friendly message
	var message string

	switch appErr.Code {
	case appErrors.NotFound:
		if appErr.Entity != "" {
			if appErr.ID != "" {
				message = appErr.Entity + " with ID " + appErr.ID + " not found"
			} else {
				message = appErr.Entity + " not found"
			}
		} else {
			message = "Resource not found"
		}

	case appErrors.AlreadyExists:
		if appErr.Entity != "" {
			message = appErr.Entity + " already exists"
			if appErr.Message != "" {
				message += ": " + appErr.Message
			}
		} else {
			message = "Resource already exists"
		}

	case appErrors.InvalidArgument:
		if appErr.Message != "" {
			message = appErr.Message
		} else if appErr.Entity != "" {
			message = "Invalid " + strings.ToLower(appErr.Entity)
		} else {
			message = "Invalid input"
		}

	case appErrors.PermissionDenied:
		message = "Permission denied"
	
	case appErrors.Unauthenticated:
		message = "Authentication required"

	default:
		// Use the original message if available
		if appErr.Message != "" {
			message = appErr.Message
		} else {
			message = appErr.Error()
		}
	}

	return message
}
