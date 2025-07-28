// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// Package errors provides a structured error handling system
// based on Google's error handling best practices.
package errors

// Code defines the type of error
type Code string

// Application-wide error codes
const (
	// OK indicates success (no error)
	OK Code = "OK"

	// Canceled indicates the operation was canceled
	Canceled Code = "CANCELED"

	// Unknown indicates an unknown or unclassified error
	Unknown Code = "UNKNOWN"

	// InvalidArgument indicates the caller specified an invalid argument
	InvalidArgument Code = "INVALID_ARGUMENT"

	// NotFound indicates the requested entity was not found
	NotFound Code = "NOT_FOUND"

	// AlreadyExists indicates an attempt to create an entity that already exists
	AlreadyExists Code = "ALREADY_EXISTS"

	// PermissionDenied indicates the caller does not have permission
	PermissionDenied Code = "PERMISSION_DENIED"

	// Unauthenticated indicates the request lacks valid authentication credentials
	Unauthenticated Code = "UNAUTHENTICATED"

	// ResourceExhausted indicates resource quota or limits were exceeded
	ResourceExhausted Code = "RESOURCE_EXHAUSTED"

	// FailedPrecondition indicates the system is not in a state required for the operation
	FailedPrecondition Code = "FAILED_PRECONDITION"

	// Internal indicates an internal error has occurred
	Internal Code = "INTERNAL"
)

// String returns the string representation of the error code
func (c Code) String() string {
	return string(c)
}
