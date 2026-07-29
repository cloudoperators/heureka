// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package siem_alert

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/issue"
	"github.com/cloudoperators/heureka/internal/app/issue_match"
	"github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

type siemAlertHandler struct {
	issueMatchHandler   issue_match.IssueMatchHandler
	issueVariantHandler issue_variant.IssueVariantHandler
	issueHandler        issue.IssueHandler
	cache               cache.Cache
}

func NewSIEMAlertHandler(
	handlerContext common.HandlerContext,
	imh issue_match.IssueMatchHandler,
	ivh issue_variant.IssueVariantHandler,
	ih issue.IssueHandler,
) SIEMAlertHandler {
	return &siemAlertHandler{
		issueMatchHandler:   imh,
		issueVariantHandler: ivh,
		issueHandler:        ih,
		cache:               handlerContext.Cache,
	}
}

type SIEMAlertHandlerError struct {
	message string
}

func NewSIEMAlertHandlerError(message string) *SIEMAlertHandlerError {
	return &SIEMAlertHandlerError{message: message}
}

func (e *SIEMAlertHandlerError) Error() string {
	return e.message
}

func (h *siemAlertHandler) DeleteSIEMAlert(ctx context.Context, id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"siemAlertId": id,
	})

	matches, err := h.issueMatchHandler.ListIssueMatches(ctx, &entity.IssueMatchFilter{Id: []*int64{&id}}, entity.NewListOptions())
	if err != nil || matches == nil || len(matches.Elements) == 0 {
		l.Error(err)
		return NewSIEMAlertHandlerError("Internal error while resolving SIEM alert.")
	}

	issueId := matches.Elements[0].IssueId

	if err := h.issueMatchHandler.DeleteIssueMatch(ctx, id); err != nil {
		l.Error(err)
		return NewSIEMAlertHandlerError("Internal error while deleting SIEM alert.")
	}

	if h.cache != nil {
		_ = h.cache.InvalidateByMatch(func(decodedKey string) bool {
			isIssueMatchQuery := strings.Contains(decodedKey, "GetIssueMatches") ||
				strings.Contains(decodedKey, "GetAllIssueMatchCursors") ||
				strings.Contains(decodedKey, "CountIssueMatches")

			if !isIssueMatchQuery {
				return false
			}

			return strings.Contains(decodedKey, fmt.Sprintf("\"issue_id\":[%d]", issueId)) ||
				strings.Contains(decodedKey, fmt.Sprintf("\"id\":[%d]", id)) ||
				strings.Contains(decodedKey, "\"issue_type\":[\"SecurityEvent\"]")
		})
	}

	remaining, err := h.issueMatchHandler.ListIssueMatches(ctx, &entity.IssueMatchFilter{IssueId: []*int64{&issueId}}, entity.NewListOptions())
	if err != nil {
		l.Error(err)
		return NewSIEMAlertHandlerError("Internal error while checking for orphaned issue.")
	}

	remainingCount := 0
	if remaining != nil {
		remainingCount = len(remaining.Elements)
	}

	l.WithField("remainingIssueMatches", remainingCount).Info("Orphan check after IssueMatch deletion")

	if remainingCount == 0 {
		variants, err := h.issueVariantHandler.ListIssueVariants(ctx, &entity.IssueVariantFilter{IssueId: []*int64{&issueId}}, entity.NewListOptions())
		if err != nil {
			l.Error(err)

			return NewSIEMAlertHandlerError("Internal error while fetching orphaned issue variants.")
		}

		if variants != nil {
			l.WithField("variantCount", len(variants.Elements)).Info("Deleting orphaned IssueVariants")

			for _, v := range variants.Elements {
				if err := h.issueVariantHandler.DeleteIssueVariant(ctx, v.Id); err != nil {
					l.WithField("variantId", v.Id).Error(err)

					return NewSIEMAlertHandlerError("Internal error while deleting orphaned issue variant.")
				}
			}
		}

		l.WithField("issueId", issueId).Info("Deleting orphaned Issue")

		if err := h.issueHandler.DeleteIssue(ctx, issueId); err != nil {
			l.WithField("issueId", issueId).Error(err)
			return NewSIEMAlertHandlerError("Internal error while deleting orphaned issue.")
		}
	}

	return nil
}
