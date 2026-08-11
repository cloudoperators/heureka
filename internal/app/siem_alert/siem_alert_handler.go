// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package siem_alert

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/app/comment"
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
	commentHandler      comment.CommentHandler
	cache               cache.Cache
}

func NewSIEMAlertHandler(
	handlerContext common.HandlerContext,
	imh issue_match.IssueMatchHandler,
	ivh issue_variant.IssueVariantHandler,
	ih issue.IssueHandler,
	ch comment.CommentHandler,
) SIEMAlertHandler {
	return &siemAlertHandler{
		issueMatchHandler:   imh,
		issueVariantHandler: ivh,
		issueHandler:        ih,
		commentHandler:      ch,
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

func (h *siemAlertHandler) AcknowledgeSIEMAlert(ctx context.Context, id int64) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"siemAlertId": id,
	})

	securityEvent := entity.IssueTypeSecurityEvent.String()

	matches, err := h.issueMatchHandler.ListIssueMatches(ctx, &entity.IssueMatchFilter{
		Id:        []*int64{&id},
		IssueType: []*string{&securityEvent},
	}, entity.NewListOptions())
	if err != nil || matches == nil || len(matches.Elements) == 0 {
		l.Error(err)

		return nil, NewSIEMAlertHandlerError("Internal error while resolving SIEM alert.")
	}

	im := matches.Elements[0].IssueMatch
	im.Acknowledged = true

	updated, err := h.issueMatchHandler.UpdateIssueMatch(ctx, im)
	if err != nil {
		l.Error(err)
		return nil, NewSIEMAlertHandlerError("Internal error while acknowledging SIEM alert.")
	}

	return updated, nil
}

func (h *siemAlertHandler) UpdateSIEMAlert(ctx context.Context, id int64, input entity.UpdateIssueMatchInput) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"siemAlertId": id,
	})

	if input.Comment == "" {
		return nil, NewSIEMAlertHandlerError("A comment is required when updating a SIEM alert.")
	}

	matches, err := h.issueMatchHandler.ListIssueMatches(ctx, &entity.IssueMatchFilter{Id: []*int64{&id}}, entity.NewListOptions())
	if err != nil {
		l.Error(err)
		return nil, NewSIEMAlertHandlerError("Internal error while resolving SIEM alert.")
	}

	if matches == nil || len(matches.Elements) == 0 {
		return nil, NewSIEMAlertHandlerError("SIEM alert not found.")
	}

	if matches.Elements[0].IssueMatch == nil {
		return nil, NewSIEMAlertHandlerError("Internal error while resolving SIEM alert.")
	}

	issueMatch := *matches.Elements[0].IssueMatch
	if input.Status != nil {
		issueMatch.Status = *input.Status
	}

	if input.UserId != nil {
		issueMatch.UserId = *input.UserId
	}

	_, err = h.commentHandler.CreateComment(ctx, &entity.Comment{
		IssueMatchId: id,
		Text:         input.Comment,
	})
	if err != nil {
		l.Error(err)
		return nil, NewSIEMAlertHandlerError("Internal error while creating comment for SIEM alert.")
	}

	updated, err := h.issueMatchHandler.UpdateIssueMatch(ctx, &issueMatch)
	if err != nil {
		l.Error(err)
		return nil, NewSIEMAlertHandlerError("Internal error while updating SIEM alert.")
	}

	if h.cache != nil {
		_ = h.cache.InvalidateByMatch(func(decodedKey string) bool {
			isIssueMatchQuery := strings.Contains(decodedKey, "GetIssueMatches") ||
				strings.Contains(decodedKey, "GetAllIssueMatchCursors") ||
				strings.Contains(decodedKey, "CountIssueMatches")

			if !isIssueMatchQuery {
				return false
			}

			return strings.Contains(decodedKey, fmt.Sprintf("\"id\":[%d]", id))
		})
	}

	return updated, nil
}
