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
