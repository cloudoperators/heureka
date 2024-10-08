// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import "github.com/cloudoperators/heureka/internal/entity"

type EvidenceHandler interface {
	ListEvidences(*entity.EvidenceFilter, *entity.ListOptions) (*entity.List[entity.EvidenceResult], error)
	CreateEvidence(*entity.Evidence) (*entity.Evidence, error)
	UpdateEvidence(*entity.Evidence) (*entity.Evidence, error)
	DeleteEvidence(int64) error
}
