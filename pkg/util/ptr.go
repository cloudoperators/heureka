// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

// takes abitary value and returns a pointer to the value
func Ptr[T any](v T) *T {
	return &v
}

// takes abitary pointer and returns value of the pointer
func Drf[T any](p *T) T {
	return *p
}
