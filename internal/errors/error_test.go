// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	stdErrors "errors"
	"fmt"
	"testing"
)

func TestErrorString(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "complete error",
			err: &Error{
				Code:    NotFound,
				Entity:  "Issue",
				ID:      "123",
				Op:      "issueHandler.GetIssue",
				Message: "issue does not exist",
			},
			expected: "issueHandler.GetIssue: Issue(123): NOT_FOUND: issue does not exist",
		},
		{
			name: "error without ID",
			err: &Error{
				Code:    InvalidArgument,
				Entity:  "Component",
				Op:      "componentHandler.CreateComponent",
				Message: "name is required",
			},
			expected: "componentHandler.CreateComponent: Component: INVALID_ARGUMENT: name is required",
		},
		{
			name: "error with wrapped error",
			err: &Error{
				Code:   Internal,
				Entity: "Database",
				Op:     "SqlDatabase.Connect",
				Err:    fmt.Errorf("connection refused"),
			},
			expected: "SqlDatabase.Connect: Database: INTERNAL: connection refused",
		},
		{
			name: "minimal error",
			err: &Error{
				Message: "something went wrong",
			},
			expected: "something went wrong",
		},
		{
			name: "error with only code",
			err: &Error{
				Code: NotFound,
			},
			expected: "NOT_FOUND",
		},
		{
			name: "error with operation only",
			err: &Error{
				Op:      "serviceHandler.ValidateService",
				Message: "validation failed",
			},
			expected: "serviceHandler.ValidateService: validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := &Error{
		Code:    Internal,
		Message: "wrapped error",
		Err:     originalErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test error without wrapped error
	errNoWrap := &Error{
		Code:    NotFound,
		Message: "not wrapped",
	}

	if errNoWrap.Unwrap() != nil {
		t.Errorf("Unwrap() should return nil for error without wrapped error")
	}
}

func TestErrorWith(t *testing.T) {
	err := &Error{
		Code:    NotFound,
		Entity:  "IssueMatch",
		Message: "not found",
	}

	fields := map[string]interface{}{
		"issue_id":    123,
		"service_ccrn": "ccrn:test:service:123",
	}

	newErr := err.With(fields)

	// Original error should be unchanged
	if err.Fields != nil {
		t.Errorf("Original error fields should be nil")
	}

	// New error should have fields
	if newErr.Fields["issue_id"] != 123 {
		t.Errorf("New error should have issue_id field")
	}
	if newErr.Fields["service_ccrn"] != "ccrn:test:service:123" {
		t.Errorf("New error should have service_ccrn field")
	}

	// Test adding to existing fields
	moreFields := map[string]interface{}{
		"severity": "high",
	}

	finalErr := newErr.With(moreFields)
	if len(finalErr.Fields) != 3 {
		t.Errorf("Final error should have 3 fields, got %d", len(finalErr.Fields))
	}
}

func TestErrorIs(t *testing.T) {
	// Test with Error types
	err1 := &Error{Code: NotFound}
	err2 := &Error{Code: NotFound}
	err3 := &Error{Code: Internal}

	if !err1.Is(err2) {
		t.Errorf("Errors with same code should match")
	}

	if err1.Is(err3) {
		t.Errorf("Errors with different codes should not match")
	}

	// Test with wrapped standard error
	stdErr := fmt.Errorf("standard error")
	errWithWrap := &Error{
		Code: Internal,
		Err:  stdErr,
	}

	if !errWithWrap.Is(stdErr) {
		t.Errorf("Error should match wrapped standard error")
	}

	// Test with stdErrors.Is
	if !stdErrors.Is(errWithWrap, stdErr) {
		t.Errorf("stdErrors.Is should work with wrapped error")
	}
}

func TestErrorAs(t *testing.T) {
	originalErr := &Error{
		Code:    NotFound,
		Entity:  "Component",
		ID:      "123",
		Message: "component not found",
	}

	// Test errors.As
	var targetErr *Error
	if !stdErrors.As(originalErr, &targetErr) {
		t.Errorf("stdErrors.As() failed to convert error to *Error")
	}

	if targetErr.Code != NotFound {
		t.Errorf("stdErrors.As() produced error with code %q, want %q", targetErr.Code, NotFound)
	}

	// Test with wrapped error
	wrappedErr := &Error{
		Code: Internal,
		Err:  originalErr,
	}

	var innerErr *Error
	if !stdErrors.As(wrappedErr, &innerErr) {
		t.Errorf("stdErrors.As() should find inner Error")
	}

	if innerErr.Code != Internal {
		t.Errorf("Inner error should have Internal code, got %s", innerErr.Code)
	}
}

func TestCodeString(t *testing.T) {
	tests := []struct {
		code     Code
		expected string
	}{
		{OK, "OK"},
		{NotFound, "NOT_FOUND"},
		{AlreadyExists, "ALREADY_EXISTS"},
		{InvalidArgument, "INVALID_ARGUMENT"},
		{Internal, "INTERNAL"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if tt.code.String() != tt.expected {
				t.Errorf("Code.String() = %q, want %q", tt.code.String(), tt.expected)
			}
		})
	}
}
