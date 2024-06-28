// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) GetSeverity(filter *entity.SeverityFilter) (*entity.Severity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.GetSeverity",
		"filter": filter,
	})

	issueVariantFilter := entity.IssueVariantFilter{
		IssueMatchId: filter.IssueMatchId,
		IssueId:      filter.IssueId,
	}

	opts := entity.ListOptions{}
	issueVariants, err := h.ListEffectiveIssueVariants(&issueVariantFilter, &opts)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while returning effective issueVariants.")

	}

	issueVariant := lo.MaxBy(issueVariants.Elements, func(item entity.IssueVariantResult, max entity.IssueVariantResult) bool {
		return item.Severity.Score > max.Severity.Score
	})

	if issueVariant.IssueVariant == nil {
		return nil, nil
	}

	return &issueVariant.Severity, nil
}
