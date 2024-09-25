// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package database

import "fmt"

type DuplicateEntryDatabaseError struct {
	msg string
}

func (e *DuplicateEntryDatabaseError) Error() string {
	return fmt.Sprintf("Database entry already exists: %s", e.msg)
}

func NewDuplicateEntryDatabaseError(msg string) *DuplicateEntryDatabaseError {
	return &DuplicateEntryDatabaseError{msg: msg}
}
