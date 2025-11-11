// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type EvidenceHandler interface {
	ListEvidences(*entity.EvidenceFilter, *entity.ListOptions) (*entity.List[entity.EvidenceResult], error)
	CreateEvidence(context.Context, *entity.Evidence) (*entity.Evidence, error)
	UpdateEvidence(context.Context, *entity.Evidence) (*entity.Evidence, error)
	DeleteEvidence(context.Context, int64) error
}
