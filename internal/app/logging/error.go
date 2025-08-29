// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package logging

import (
	"errors"

	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/sirupsen/logrus"
)

// LogError logs errors using our internal error package with structured fields
func LogError(logger *logrus.Logger, err error, fields logrus.Fields) {
	var appErr *appErrors.Error
	if !errors.As(err, &appErr) {
		logger.WithError(err).WithFields(fields).Error("Unknown error")
		return
	}

	errorFields := logrus.Fields{
		"error_code": string(appErr.Code),
	}

	if appErr.Entity != "" {
		errorFields["entity"] = appErr.Entity
	}
	if appErr.ID != "" {
		errorFields["entity_id"] = appErr.ID
	}
	if appErr.Op != "" {
		errorFields["operation"] = appErr.Op
	}

	// Add any additional fields from the error
	for k, v := range appErr.Fields {
		errorFields[k] = v
	}

	// Add any passed-in fields
	for k, v := range fields {
		errorFields[k] = v
	}

	logger.WithFields(errorFields).WithError(appErr.Err).Error(appErr.Error())
}
