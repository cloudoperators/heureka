// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package severity

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type SeverityHandler interface {
	GetSeverity(context.Context, *entity.SeverityFilter) (*entity.Severity, error)
}
