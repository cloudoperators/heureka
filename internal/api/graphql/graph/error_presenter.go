// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// ErrorPresenter is a simplified custom error presenter for gqlgen
// that translates application errors to GraphQL errors
func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	// Try to extract our application error
	var appErr *appErrors.Error
	if errors.As(err, &appErr) {
		// Map the error code to a GraphQL error code
		code := mapErrorCodeToGraphQL(appErr.Code)
		
		// Create the GraphQL error with appropriate metadata
		gqlErr := &gqlerror.Error{
			Message: formatErrorMessage(appErr),
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
		
		// Add fields if they exist and this is not an internal error
		if appErr.Code != appErrors.Internal && appErr.Fields != nil && len(appErr.Fields) > 0 {
			gqlErr.Extensions["fields"] = appErr.Fields
		}
		
		return gqlErr
	}
	
	// For any other error, use the default behavior
	return graphql.DefaultErrorPresenter(ctx, err)
}

// mapErrorCodeToGraphQL maps application error codes to GraphQL error codes
func mapErrorCodeToGraphQL(code appErrors.Code) string {
	switch code {
	case appErrors.InvalidArgument:
		return "BAD_REQUEST"
	case appErrors.NotFound:
		return "NOT_FOUND"
	case appErrors.AlreadyExists:
		return "CONFLICT"
	case appErrors.PermissionDenied:
		return "FORBIDDEN"
	case appErrors.Unauthenticated:
		return "UNAUTHENTICATED"
	case appErrors.ResourceExhausted:
		return "TOO_MANY_REQUESTS"
	case appErrors.FailedPrecondition:
		return "PRECONDITION_FAILED"
	default:
		return "INTERNAL"
	}
}

// formatErrorMessage creates a user-friendly error message
func formatErrorMessage(appErr *appErrors.Error) string {
	// For internal errors, provide a generic message
	if appErr.Code == appErrors.Internal {
		return "An internal server error occurred"
	}
	
	// For other errors, use the error message or create one
	if appErr.Message != "" {
		return appErr.Message
	} else if appErr.Entity != "" {
		switch appErr.Code {
		case appErrors.NotFound:
			if appErr.ID != "" {
				return appErr.Entity + " with ID " + appErr.ID + " not found"
			}
			return appErr.Entity + " not found"
		case appErrors.AlreadyExists:
			if appErr.ID != "" {
				return appErr.Entity + " with ID " + appErr.ID + " already exists"
			}
			return appErr.Entity + " already exists"
		}
	}
	
	// If no specific message can be created, use the error string
	return appErr.Error()
}
