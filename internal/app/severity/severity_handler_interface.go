// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package severity

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type SeverityHandler interface {
	GetSeverity(*entity.SeverityFilter) (*entity.Severity, error)
}
