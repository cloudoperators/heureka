// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"testing"
)

func TestHelperFunctions(t *testing.T) {
	t.Run("NotFoundError", func(t *testing.T) {
		op := "issueHandler.GetIssue"
		entity := "Issue"
		id := "123"

		err := NotFoundError(op, entity, id)

		if err.Code != NotFound {
			t.Errorf("NotFoundError().Code = %q, want %q", err.Code, NotFound)
		}
		if err.Op != op {
			t.Errorf("NotFoundError().Op = %q, want %q", err.Op, op)
		}
		if err.Entity != entity {
			t.Errorf("NotFoundError().Entity = %q, want %q", err.Entity, entity)
		}
		if err.ID != id {
			t.Errorf("NotFoundError().ID = %q, want %q", err.ID, id)
		}
		if err.Message != "not found" {
			t.Errorf("NotFoundError().Message = %q, want %q", err.Message, "not found")
		}
	})

	t.Run("AlreadyExistsError", func(t *testing.T) {
		err := AlreadyExistsError("componentHandler.CreateComponent", "Component", "test-component")
		if err.Code != AlreadyExists {
			t.Errorf("AlreadyExistsError() produced error with code %q, want %q", err.Code, AlreadyExists)
		}
		if err.Message != "already exists" {
			t.Errorf("AlreadyExistsError().Message = %q, want %q", err.Message, "already exists")
		}
	})

	t.Run("InvalidArgumentError", func(t *testing.T) {
		err := InvalidArgumentError("serviceHandler.CreateService", "Service", "name is required")
		if err.Code != InvalidArgument {
			t.Errorf("InvalidArgumentError() produced error with code %q, want %q", err.Code, InvalidArgument)
		}
		if err.Message != "name is required" {
			t.Errorf("InvalidArgumentError().Message = %q, want %q", err.Message, "name is required")
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		cause := fmt.Errorf("database connection timeout")
		err := InternalError("SqlDatabase.Connect", "Database", "", cause)
		if err.Code != Internal {
			t.Errorf("InternalError() produced error with code %q, want %q", err.Code, Internal)
		}
		if err.Err != cause {
			t.Errorf("InternalError() did not properly wrap the cause error")
		}
	})
}

func TestErrorChecking(t *testing.T) {
	notFoundErr := NotFoundError("issueHandler.GetIssue", "Issue", "123")
	alreadyExistsErr := AlreadyExistsError("componentHandler.CreateComponent", "Component", "test-component")
	invalidArgErr := InvalidArgumentError("serviceHandler.CreateService", "Service", "name required")

	t.Run("Is", func(t *testing.T) {
		if !Is(notFoundErr, NotFound) {
			t.Errorf("Is() failed to match NotFound error")
		}
		if Is(notFoundErr, AlreadyExists) {
			t.Errorf("Is() incorrectly matched NotFound error with AlreadyExists code")
		}

		// Test with non-Error type
		stdErr := fmt.Errorf("standard error")
		if Is(stdErr, NotFound) {
			t.Errorf("Is() should return false for non-Error types")
		}
	})

	t.Run("IsNotFound", func(t *testing.T) {
		if !IsNotFound(notFoundErr) {
			t.Errorf("IsNotFound() failed to match NotFound error")
		}
		if IsNotFound(alreadyExistsErr) {
			t.Errorf("IsNotFound() incorrectly matched AlreadyExists error")
		}
	})

	t.Run("IsAlreadyExists", func(t *testing.T) {
		if !IsAlreadyExists(alreadyExistsErr) {
			t.Errorf("IsAlreadyExists() failed to match AlreadyExists error")
		}
		if IsAlreadyExists(notFoundErr) {
			t.Errorf("IsAlreadyExists() incorrectly matched NotFound error")
		}
	})

	t.Run("IsInvalidArgument", func(t *testing.T) {
		if !IsInvalidArgument(invalidArgErr) {
			t.Errorf("IsInvalidArgument() failed to match InvalidArgument error")
		}
		if IsInvalidArgument(notFoundErr) {
			t.Errorf("IsInvalidArgument() incorrectly matched NotFound error")
		}
	})

	t.Run("Match", func(t *testing.T) {
		if !Match(notFoundErr, "Issue", NotFound) {
			t.Errorf("Match() failed to match correct entity and code")
		}
		if Match(notFoundErr, "Component", NotFound) {
			t.Errorf("Match() incorrectly matched with wrong entity")
		}
		if Match(notFoundErr, "Issue", Internal) {
			t.Errorf("Match() incorrectly matched with wrong code")
		}
		if !Match(notFoundErr, "", NotFound) {
			t.Errorf("Match() should match when entity is empty string")
		}

		// Test with non-Error type
		stdErr := fmt.Errorf("database connection failed")
		if Match(stdErr, "Issue", NotFound) {
			t.Errorf("Match() should return false for non-Error types")
		}
	})
}

func TestCodeToHTTPStatus(t *testing.T) {
	tests := []struct {
		code           Code
		expectedStatus int
	}{
		{OK, 200},
		{InvalidArgument, 400},
		{Unauthenticated, 401},
		{PermissionDenied, 403},
		{NotFound, 404},
		{AlreadyExists, 409},
		{FailedPrecondition, 412},
		{ResourceExhausted, 429},
		{Internal, 500},
		{Unknown, 500},
		{Canceled, 500},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			status := CodeToHTTPStatus(tt.code)
			if status != tt.expectedStatus {
				t.Errorf("CodeToHTTPStatus(%s) = %d, want %d", tt.code, status, tt.expectedStatus)
			}
		})
	}
}

func TestEBuilder(t *testing.T) {
	t.Run("basic usage", func(t *testing.T) {
		err := E(Op("issueHandler.GetIssue"), "Issue", "123", NotFound, "issue not found")

		if err.Op != "issueHandler.GetIssue" {
			t.Errorf("E() Op = %q, want %q", err.Op, "issueHandler.GetIssue")
		}
		if err.Entity != "Issue" {
			t.Errorf("E() Entity = %q, want %q", err.Entity, "Issue")
		}
		if err.ID != "123" {
			t.Errorf("E() ID = %q, want %q", err.ID, "123")
		}
		if err.Code != NotFound {
			t.Errorf("E() Code = %q, want %q", err.Code, NotFound)
		}
		if err.Message != "issue not found" {
			t.Errorf("E() Message = %q, want %q", err.Message, "issue not found")
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		originalErr := fmt.Errorf("database connection failed")
		err := E(Op("SqlDatabase.Connect"), "Database", Internal, originalErr)

		if err.Err != originalErr {
			t.Errorf("E() should wrap the original error")
		}
	})

	t.Run("with fields", func(t *testing.T) {
		fields := map[string]interface{}{
			"retry_count": 3,
			"severity":    "high",
		}
		err := E(Op("SqlDatabase.Connect"), "Database", Internal, fields)

		if err.Fields["retry_count"] != 3 {
			t.Errorf("E() should include retry_count field")
		}
		if err.Fields["severity"] != "high" {
			t.Errorf("E() should include severity field")
		}
	})

	t.Run("copying existing error", func(t *testing.T) {
		originalErr := &Error{
			Code:    NotFound,
			Entity:  "IssueMatch",
			ID:      "456",
			Message: "issue match not found",
		}

		newErr := E(Op("issueMatchHandler.UpdateIssueMatch"), originalErr)

		if newErr.Op != "issueMatchHandler.UpdateIssueMatch" {
			t.Errorf("E() should set new operation")
		}
		if newErr.Code != NotFound {
			t.Errorf("E() should copy code from original error")
		}
		if newErr.Entity != "IssueMatch" {
			t.Errorf("E() should copy entity from original error")
		}
	})
}

func TestErrorf(t *testing.T) {
	err := Errorf(InvalidArgument, "invalid value: %s", "test")

	if err.Code != InvalidArgument {
		t.Errorf("Errorf() Code = %q, want %q", err.Code, InvalidArgument)
	}
	if err.Message != "invalid value: test" {
		t.Errorf("Errorf() Message = %q, want %q", err.Message, "invalid value: test")
	}
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := Wrap(originalErr, Internal, "internal error occurred")

	if wrappedErr.Code != Internal {
		t.Errorf("Wrap() Code = %q, want %q", wrappedErr.Code, Internal)
	}
	if wrappedErr.Message != "internal error occurred" {
		t.Errorf("Wrap() Message = %q, want %q", wrappedErr.Message, "internal error occurred")
	}
	if wrappedErr.Err != originalErr {
		t.Errorf("Wrap() should wrap the original error")
	}

	// Test wrapping nil error
	nilWrapped := Wrap(nil, Internal, "message")
	if nilWrapped != nil {
		t.Errorf("Wrap(nil) should return nil")
	}
}
