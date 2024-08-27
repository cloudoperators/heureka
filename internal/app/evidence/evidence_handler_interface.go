// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type EvidenceService interface {
	ListEvidences(*entity.EvidenceFilter, *entity.ListOptions) (*entity.List[entity.EvidenceResult], error)
	CreateEvidence(*entity.Evidence) (*entity.Evidence, error)
	UpdateEvidence(*entity.Evidence) (*entity.Evidence, error)
	DeleteEvidence(int64) error
}
