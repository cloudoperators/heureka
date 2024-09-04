// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package severity

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type severityHandler struct {
	database            database.Database
	eventRegistry       event.EventRegistry
	issueVariantHandler issue_variant.IssueVariantHandler
}

func NewSeverityHandler(database database.Database, eventRegistry event.EventRegistry, ivs issue_variant.IssueVariantHandler) SeverityHandler {
	return &severityHandler{
		database:            database,
		eventRegistry:       eventRegistry,
		issueVariantHandler: ivs,
	}
}

type SeverityHandlerError struct {
	message string
}

func NewSeverityHandlerError(message string) *SeverityHandlerError {
	return &SeverityHandlerError{message: message}
}

func (e *SeverityHandlerError) Error() string {
	return e.message
}

func (s *severityHandler) GetSeverity(filter *entity.SeverityFilter) (*entity.Severity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  GetSeverityEventName,
		"filter": filter,
	})

	issueVariantFilter := entity.IssueVariantFilter{
		IssueMatchId: filter.IssueMatchId,
		IssueId:      filter.IssueId,
	}

	opts := entity.ListOptions{}
	issueVariants, err := s.issueVariantHandler.ListEffectiveIssueVariants(&issueVariantFilter, &opts)

	if err != nil {
		l.Error(err)
		return nil, NewSeverityHandlerError("Internal error while returning effective issueVariants.")

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
