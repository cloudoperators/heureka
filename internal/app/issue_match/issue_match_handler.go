// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/severity"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
)

type issueMatchHandler struct {
	common.BaseHandler[entity.IssueMatchResult, *entity.IssueMatchFilter]
	severityHandler severity.SeverityHandler
}

func NewIssueMatchHandler(
	handlerContext common.HandlerContext,
	ss severity.SeverityHandler,
) IssueMatchHandler {
	return &issueMatchHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.IssueMatchResult, *entity.IssueMatchFilter]{
			Op:              appErrors.Op("issueMatchHandler"),
			Entity:          "IssueMatches",
			CursorEntity:    "IssueMatchCursors",
			CountEntity:     "IssueMatchCount",
			GetFn:           handlerContext.DB.GetIssueMatches,
			CursorsFn:       handlerContext.DB.GetAllIssueMatchCursors,
			CountFn:         handlerContext.DB.CountIssueMatches,
			Authz:           handlerContext.Authz,
			AuthzObjectType: openfga.TypeComponentInstance,
			AuthzApplyFn: func(f *entity.IssueMatchFilter, ids []*int64) {
				f.ComponentInstanceId = common.CombineFilterWithAccessibleIds(f.ComponentInstanceId, ids)
			},
			ListEventFn: func(f *entity.IssueMatchFilter, o *entity.ListOptions, r *entity.List[entity.IssueMatchResult]) any {
				return &ListIssueMatchesEvent{Filter: f, Options: o, Results: r}
			},
			DeleteFn:      handlerContext.DB.DeleteIssueMatch,
			DeleteEventFn: func(id int64) any { return &DeleteIssueMatchEvent{IssueMatchID: id} },
		}),
		severityHandler: ss,
	}
}

func (im *issueMatchHandler) GetIssueMatch(
	ctx context.Context,
	issueMatchId int64,
) (*entity.IssueMatch, error) {
	op := appErrors.CallerOp()

	currentUserId, err := common.GetCurrentUserId(ctx, im.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueMatch", fmt.Sprint(issueMatchId), err)
	}

	hasPermission, err := im.Authz().CheckPermission(openfga.RelationInput{
		UserType:   openfga.TypeUser,
		UserId:     openfga.UserId(fmt.Sprint(currentUserId)),
		Relation:   openfga.RelCanView,
		ObjectType: openfga.TypeIssueMatch,
		ObjectId:   openfga.ObjectId(fmt.Sprint(issueMatchId)),
	})
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueMatch", fmt.Sprint(issueMatchId), err)
	}

	if !hasPermission {
		return nil, appErrors.PermissionDeniedError(string(op), "IssueMatch", fmt.Sprint(issueMatchId))
	}

	result, err := im.ListIssueMatches(
		ctx,
		&entity.IssueMatchFilter{Id: []*int64{&issueMatchId}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueMatch", fmt.Sprint(issueMatchId), err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.InternalError(string(op), "IssueMatch", fmt.Sprint(issueMatchId),
			fmt.Errorf("expected 1, got %d", len(result.Elements)))
	}

	im.PushEvent(&GetIssueMatchEvent{
		IssueMatchID: issueMatchId,
		Result:       result.Elements[0].IssueMatch,
	})

	return result.Elements[0].IssueMatch, nil
}

func (im *issueMatchHandler) ListIssueMatches(
	ctx context.Context,
	filter *entity.IssueMatchFilter,
	options *entity.ListOptions,
) (*entity.List[entity.IssueMatchResult], error) {
	return im.List(ctx, appErrors.CallerOp(), filter, options)
}

func (im *issueMatchHandler) CreateIssueMatch(
	ctx context.Context,
	issueMatch *entity.IssueMatch,
) (*entity.IssueMatch, error) {
	op := appErrors.CallerOp()

	var err error

	issueMatch.CreatedBy, err = common.GetCurrentUserId(ctx, im.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueMatch", "", err)
	}

	issueMatch.UpdatedBy = issueMatch.CreatedBy

	effectiveSeverity, err := im.severityHandler.GetSeverity(
		ctx,
		&entity.SeverityFilter{IssueId: []*int64{&issueMatch.IssueId}},
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueMatch", "", err)
	}

	issueMatch.Severity = *effectiveSeverity

	newIssueMatch, err := im.DB().CreateIssueMatch(issueMatch)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueMatch", "", err)
	}

	im.PushEvent(&CreateIssueMatchEvent{IssueMatch: newIssueMatch})

	return newIssueMatch, nil
}

func (im *issueMatchHandler) UpdateIssueMatch(
	ctx context.Context,
	issueMatch *entity.IssueMatch,
) (*entity.IssueMatch, error) {
	op := appErrors.CallerOp()
	id := strconv.FormatInt(issueMatch.Id, 10)

	var err error

	issueMatch.UpdatedBy, err = common.GetCurrentUserId(ctx, im.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueMatch", id, err)
	}

	if err = im.DB().UpdateIssueMatch(issueMatch); err != nil {
		return nil, appErrors.InternalError(string(op), "IssueMatch", id, err)
	}

	im.PushEvent(&UpdateIssueMatchEvent{IssueMatch: issueMatch})

	return im.GetIssueMatch(ctx, issueMatch.Id)
}

func (im *issueMatchHandler) DeleteIssueMatch(ctx context.Context, id int64) error {
	return im.Delete(ctx, id)
}
