// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package remediation

import "github.com/cloudoperators/heureka/internal/entity"

type RemediationHandler interface {
	ListRemediations(*entity.RemediationFilter, *entity.ListOptions) (*entity.List[entity.RemediationResult], error)
	CreateRemediation(*entity.Remediation) (*entity.Remediation, error)
	UpdateRemediation(*entity.Remediation) (*entity.Remediation, error)
	DeleteRemediation(int64) error
}
