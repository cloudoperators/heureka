// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_variant

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type IssueVariantService interface {
	ListIssueVariants(*entity.IssueVariantFilter, *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error)
	ListEffectiveIssueVariants(*entity.IssueVariantFilter, *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error)
	CreateIssueVariant(*entity.IssueVariant) (*entity.IssueVariant, error)
	UpdateIssueVariant(*entity.IssueVariant) (*entity.IssueVariant, error)
	DeleteIssueVariant(int64) error
}
