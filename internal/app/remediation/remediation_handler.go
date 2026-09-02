// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package remediation

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

type remediationHandler struct {
	common.BaseHandler[entity.RemediationResult, *entity.RemediationFilter]
}

func NewRemediationHandler(handlerContext common.HandlerContext) RemediationHandler {
	return &remediationHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.RemediationResult, *entity.RemediationFilter]{
			Op:           appErrors.Op("remediationHandler"),
			Entity:       "Remediations",
			CursorEntity: "RemediationCursors",
			CountEntity:  "RemediationCount",
			GetFn:        handlerContext.DB.GetRemediations,
			CursorsFn:    handlerContext.DB.GetAllRemediationCursors,
			CountFn:      handlerContext.DB.CountRemediations,
			Authz:        handlerContext.Authz,
			ListEventFn: func(f *entity.RemediationFilter, o *entity.ListOptions, r *entity.List[entity.RemediationResult]) any {
				return &ListRemediationsEvent{Filter: f, Options: o, Remediations: r}
			},
			DeleteFn:      handlerContext.DB.DeleteRemediation,
			DeleteEventFn: func(id int64) any { return &DeleteRemediationEvent{RemediationID: id} },
		}),
	}
}

func (rh *remediationHandler) ListRemediations(
	ctx context.Context,
	filter *entity.RemediationFilter,
	options *entity.ListOptions,
) (*entity.List[entity.RemediationResult], error) {
	return rh.List(ctx, appErrors.CallerOp(), filter, options)
}

func validateFilteredRemediationDescription(description string, op appErrors.Op) error {
	if description == "" {
		return appErrors.E(op, "Remediation", appErrors.InvalidArgument, "Description is required for filtered remediation")
	}

	return nil
}

func (rh *remediationHandler) validateFilteredRemediationIssue(
	ctx context.Context,
	issueID int64,
	id string,
	op appErrors.Op,
) error {
	issues, err := rh.DB().GetIssues(ctx, &entity.IssueFilter{
		Id: []*int64{&issueID},
	}, nil)
	if err != nil {
		return appErrors.InternalError(string(op), "Remediation", id, err)
	}

	if len(issues) != 1 || issues[0].Type != entity.IssueTypeSecurityEvent {
		return appErrors.E(op, "Remediation", appErrors.InvalidArgument, "filtered remediation can only be applied to SIEM alerts")
	}

	return nil
}

func (rh *remediationHandler) CreateRemediation(
	ctx context.Context,
	remediation *entity.Remediation,
) (*entity.Remediation, error) {
	op := appErrors.CallerOp()

	if remediation == nil {
		return nil, appErrors.E(op, "Remediation", appErrors.InvalidArgument, "remediation cannot be nil")
	}

	if remediation.Service == "" {
		return nil, appErrors.E(op, "Remediation", appErrors.InvalidArgument, "Service is required")
	}

	if err := validateByType(remediation); err != nil {
		applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"remediation": remediation})

		return nil, appErrors.E(op, "Remediation", appErrors.InvalidArgument, err.Error())
	}

	if remediation.Type == entity.RemediationTypeFiltered {
		if err := validateFilteredRemediationDescription(remediation.Description, op); err != nil {
			applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"remediation": remediation})
			return nil, err
		}

		if err := rh.validateFilteredRemediationIssue(ctx, remediation.IssueId, "", op); err != nil {
			applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"remediation": remediation})
			return nil, err
		}
	}

	var err error

	remediation.CreatedBy, err = common.GetCurrentUserId(ctx, rh.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", "", err)
	}

	hasPermission, err := rh.Authz().CheckPermission(openfga.RelationInput{
		UserType:   openfga.TypeUser,
		UserId:     openfga.UserId(fmt.Sprint(remediation.CreatedBy)),
		Relation:   openfga.RelCanWrite,
		ObjectType: openfga.TypeService,
		ObjectId:   openfga.ObjectId(fmt.Sprint(remediation.ServiceId)),
	})
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", "", err)
	}

	if !hasPermission {
		return nil, appErrors.PermissionDeniedError(string(op), "Service", fmt.Sprint(remediation.ServiceId))
	}

	remediation.UpdatedBy = remediation.CreatedBy

	if remediation.RemediatedBy == "" {
		remediation.RemediatedBy, err = common.GetCurrentUniqueUserId(ctx)
		if err != nil {
			return nil, appErrors.InternalError(string(op), "Remediation", "", err)
		}
	}

	remediation.RemediatedById, err = common.GetUserIdByUniqueId(ctx, rh.DB(), remediation.RemediatedBy)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", "", err)
	}

	if remediation.Assignee != "" {
		remediation.AssigneeId, err = common.GetUserIdByUniqueId(
			ctx,
			rh.DB(),
			remediation.Assignee,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "Remediation", "", err)
			applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{
				"remediation": remediation,
			})

			return nil, wrappedErr
		}

		if remediation.AssigneeId == 0 {
			err := appErrors.E(
				op, "Remediation", appErrors.InvalidArgument,
				fmt.Sprintf("assignee %q not found", remediation.Assignee),
			)
			applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"remediation": remediation})

			return nil, err
		}
	}

	// Check for existing remediation
	existingRemediations, err := rh.DB().GetRemediations(ctx, &entity.RemediationFilter{
		ServiceId: []*int64{&remediation.ServiceId},
		IssueId:   []*int64{&remediation.IssueId},
		State:     []entity.StateFilterType{entity.Active},
	}, nil)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", "", err)
	}

	for _, er := range existingRemediations {
		if er.Remediation == nil {
			continue
		}

		sameComponent := (remediation.ComponentId <= 0 && er.ComponentId <= 0) ||
			(remediation.ComponentId == er.ComponentId)
		if !sameComponent {
			continue
		}

		if !er.ExpirationDate.IsZero() && er.ExpirationDate.Before(time.Now()) {
			continue
		}

		if er.Type != remediation.Type {
			continue
		}

		existingIsOpen := er.ExpirationDate.IsZero()

		newIsLater := !remediation.ExpirationDate.IsZero() && remediation.ExpirationDate.After(er.ExpirationDate)
		if existingIsOpen || !newIsLater {
			return nil, appErrors.E(op, "Remediation", appErrors.InvalidArgument,
				"A remediation of this type is already in progress; the new expiration date must be later than the existing one.")
		}
	}

	newRemediation, err := rh.DB().CreateRemediation(remediation)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", "", err)
	}

	if rh.Cache() != nil {
		if err := rh.Cache().InvalidateByMatch(func(decodedKey string) bool {
			return (strings.Contains(decodedKey, fmt.Sprintf("\"issue_id\":[%d]", newRemediation.IssueId)) ||
				strings.Contains(decodedKey, fmt.Sprintf("\"id\":[%d]", newRemediation.IssueId))) &&
				(strings.Contains(decodedKey, "GetIssuesWithAggregations") || strings.Contains(decodedKey, "GetIssues") ||
					strings.Contains(decodedKey, "GetAllIssueCursors") || strings.Contains(decodedKey, "GetIssueVariants") ||
					strings.Contains(decodedKey, "GetIssueMatches"))
		}); err != nil {
			// non-fatal: log and continue
			applog.LogError(logrus.StandardLogger(), appErrors.InternalError(string(op), "CacheInvalidation", "", err), logrus.Fields{})
		}
	}

	rh.PushEvent(&CreateRemediationEvent{Remediation: newRemediation})

	return newRemediation, nil
}

func (rh *remediationHandler) UpdateRemediation(
	ctx context.Context,
	remediation *entity.Remediation,
) (*entity.Remediation, error) {
	op := appErrors.CallerOp()

	if remediation == nil {
		return nil, appErrors.E(op, "Remediation", appErrors.InvalidArgument, "remediation cannot be nil")
	}

	if remediation.Id <= 0 {
		return nil, appErrors.E(op, "Remediation", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", remediation.Id))
	}

	id := strconv.FormatInt(remediation.Id, 10)

	var err error

	remediation.UpdatedBy, err = common.GetCurrentUserId(ctx, rh.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", id, err)
	}

	existing, err := rh.ListRemediations(ctx, &entity.RemediationFilter{Id: []*int64{&remediation.Id}}, entity.NewListOptions())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", id, err)
	}

	if len(existing.Elements) != 1 {
		return nil, appErrors.E(op, "Remediation", id, appErrors.Internal,
			fmt.Sprintf("unexpected result count: %d", len(existing.Elements)))
	}

	existingRemediation := existing.Elements[0].Remediation

	if existingRemediation == nil {
		err := appErrors.E(
			op,
			"Remediation",
			strconv.FormatInt(remediation.Id, 10),
			appErrors.Internal,
			"existing remediation record has a nil pointer",
		)
		applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"id": remediation.Id})

		return nil, err
	}

	finalType := existingRemediation.Type
	if remediation.Type != "" {
		finalType = remediation.Type
	}

	finalURL := existingRemediation.URL
	if remediation.URL != "" {
		finalURL = remediation.URL
	}

	finalDescription := existingRemediation.Description
	if remediation.Description != "" {
		finalDescription = remediation.Description
	}

	scratch := &entity.Remediation{Type: finalType, URL: finalURL, Description: finalDescription}

	if err := validateByType(scratch); err != nil {
		applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"remediation": remediation})

		return nil, appErrors.E(op, "Remediation", appErrors.InvalidArgument, err.Error())
	}

	if finalType == entity.RemediationTypeFiltered {
		finalDescription := existingRemediation.Description
		if remediation.Description != "" {
			finalDescription = remediation.Description
		}

		if err := validateFilteredRemediationDescription(finalDescription, op); err != nil {
			applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"remediation": remediation})
			return nil, err
		}

		if err := rh.validateFilteredRemediationIssue(ctx, existingRemediation.IssueId, strconv.FormatInt(remediation.Id, 10), op); err != nil {
			applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"remediation": remediation})
			return nil, err
		}
	}

	remediation.URL = finalURL
	remediation.Type = finalType
	remediation.Description = finalDescription

	if remediation.Assignee != "" {
		remediation.AssigneeId, err = common.GetUserIdByUniqueId(
			ctx,
			rh.DB(),
			remediation.Assignee,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(
				string(op),
				"Remediation",
				strconv.FormatInt(remediation.Id, 10),
				err,
			)
			applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{"remediation": remediation})

			return nil, wrappedErr
		}

		if remediation.AssigneeId == 0 {
			err := appErrors.E(
				op, "Remediation", appErrors.InvalidArgument,
				fmt.Sprintf("assignee %q not found", remediation.Assignee),
			)
			applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"remediation": remediation})

			return nil, err
		}
	}

	if err = rh.DB().UpdateRemediation(remediation); err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", id, err)
	}

	result, err := rh.ListRemediations(ctx, &entity.RemediationFilter{Id: []*int64{&remediation.Id}}, entity.NewListOptions())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Remediation", id, err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.E(op, "Remediation", id, appErrors.Internal,
			fmt.Sprintf("unexpected result count: %d", len(result.Elements)))
	}

	updatedRemediation := result.Elements[0].Remediation
	rh.PushEvent(&UpdateRemediationEvent{Remediation: updatedRemediation})

	return updatedRemediation, nil
}

func (rh *remediationHandler) DeleteRemediation(ctx context.Context, id int64) error {
	if id <= 0 {
		return appErrors.E(appErrors.CallerOp(), "Remediation", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", id))
	}

	return rh.Delete(ctx, id)
}

func validateExternalURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid external URL: %q", rawURL)
	}

	return nil
}

func validateRiskAccepted(r *entity.Remediation) error {
	if r.URL == "" {
		return fmt.Errorf("URL is required for risk_accepted remediation")
	}

	return validateExternalURL(r.URL)
}

func validateEscalation(r *entity.Remediation) error {
	if r.Description == "" {
		return fmt.Errorf("description is required for escalation remediation")
	}

	if r.URL == "" {
		return fmt.Errorf("URL is required for escalation remediation")
	}

	return validateExternalURL(r.URL)
}

func validateByType(r *entity.Remediation) error {
	switch r.Type {
	case entity.RemediationTypeRiskAccepted:
		return validateRiskAccepted(r)
	case entity.RemediationTypeEscalation:
		return validateEscalation(r)
	}

	return nil
}
