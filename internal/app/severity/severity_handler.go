// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package severity

import (
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/app/issue_variant"
	"github.wdf.sap.corp/cc/heureka/internal/database"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

type severityService struct {
	database            database.Database
	eventRegistry       event.EventRegistry
	issueVariantService issue_variant.IssueVariantService
}

func NewSeverityService(database database.Database, eventRegistry event.EventRegistry, ivs issue_variant.IssueVariantService) SeverityService {
	return &severityService{
		database:            database,
		eventRegistry:       eventRegistry,
		issueVariantService: ivs,
	}
}

type SeverityServiceError struct {
	message string
}

func NewSeverityServiceError(message string) *SeverityServiceError {
	return &SeverityServiceError{message: message}
}

func (e *SeverityServiceError) Error() string {
	return e.message
}

func (s *severityService) GetSeverity(filter *entity.SeverityFilter) (*entity.Severity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  GetSeverityEventName,
		"filter": filter,
	})

	issueVariantFilter := entity.IssueVariantFilter{
		IssueMatchId: filter.IssueMatchId,
		IssueId:      filter.IssueId,
	}

	opts := entity.ListOptions{}
	issueVariants, err := s.issueVariantService.ListEffectiveIssueVariants(&issueVariantFilter, &opts)

	if err != nil {
		l.Error(err)
		return nil, NewSeverityServiceError("Internal error while returning effective issueVariants.")

	}

	issueVariant := lo.MaxBy(issueVariants.Elements, func(item entity.IssueVariantResult, max entity.IssueVariantResult) bool {
		return item.Severity.Score > max.Severity.Score
	})

	if issueVariant.IssueVariant == nil {
		return nil, nil
	}

	s.eventRegistry.PushEvent(&GetSeverityEvent{
		Filter: filter,
		Result: &issueVariant.Severity,
	})

	return &issueVariant.Severity, nil
}
