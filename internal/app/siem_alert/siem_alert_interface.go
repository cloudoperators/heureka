// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package siem_alert

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type SIEMAlertHandler interface {
	DeleteSIEMAlert(ctx context.Context, id int64) error
	AcknowledgeSIEMAlert(ctx context.Context, id int64) (*entity.IssueMatch, error)
}
