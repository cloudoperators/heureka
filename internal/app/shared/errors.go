// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/cloudoperators/heureka/internal/database"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
)

// FromDatabaseError converts database errors to application errors
// This function follows clean architecture principles by keeping database error translation
// in the application layer, not the database layer.
func FromDatabaseError(op string, entity string, id string, err error) error {
	if err == nil {
		return nil
	}

	// Check for no rows error
	if errors.Is(err, sql.ErrNoRows) {
		return appErrors.NotFoundError(op, entity, id)
	}

	// Check for duplicate entry error
	var dupErr *database.DuplicateEntryDatabaseError
	if errors.As(err, &dupErr) {
		return appErrors.AlreadyExistsError(op, entity, id)
	}

	// Check for specific MySQL/MariaDB errors
	if strings.Contains(err.Error(), "Error 1062") || // Duplicate entry
		strings.Contains(err.Error(), "Duplicate entry") {
		return appErrors.AlreadyExistsError(op, entity, id)
	}

	// Check for foreign key constraint violations
	if strings.Contains(strings.ToLower(err.Error()), "foreign key constraint") {
		return appErrors.InvalidArgumentError(op, entity, "invalid reference")
	}

	// Default to internal error
	return appErrors.InternalError(op, entity, id, err)
}
