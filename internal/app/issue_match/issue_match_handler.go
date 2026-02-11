// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/severity"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

var CacheTtlGetIssueMatches = 12 * time.Hour
var CacheTtlGetAllIssueMatchCursors = 12 * time.Hour
var CacheTtlCountIssueMatches = 12 * time.Hour

type issueMatchHandler struct {
	database        database.Database
	eventRegistry   event.EventRegistry
	cache           cache.Cache
	authz           openfga.Authorization
	severityHandler severity.SeverityHandler
}

func NewIssueMatchHandler(handlerContext common.HandlerContext, ss severity.SeverityHandler) IssueMatchHandler {
	return &issueMatchHandler{
		database:        handlerContext.DB,
		eventRegistry:   handlerContext.EventReg,
		cache:           handlerContext.Cache,
		authz:           handlerContext.Authz,
		severityHandler: ss,
	}
}

type IssueMatchHandlerError struct {
	message string
}

func NewIssueMatchHandlerError(message string) *IssueMatchHandlerError {
	return &IssueMatchHandlerError{message: message}
}

func (e *IssueMatchHandlerError) Error() string {
	return e.message
}

func (im *issueMatchHandler) GetIssueMatch(issueMatchId int64) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetIssueMatchEventName,
		"id":    issueMatchId,
	})

	// get current user id
	currentUserId, err := common.GetCurrentUserId(im.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Error while getting current user id")
	}

	// Authorization check
	hasPermission, err := im.authz.CheckPermission(openfga.RelationInput{
		UserType:   openfga.TypeUser,
		UserId:     openfga.UserId(fmt.Sprint(currentUserId)),
		Relation:   openfga.RelCanView,
		ObjectType: openfga.TypeIssueMatch,
		ObjectId:   openfga.ObjectId(fmt.Sprint(issueMatchId)),
	})
	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Error while checking permission for user")
	}
	if !hasPermission {
		return nil, NewIssueMatchHandlerError("User does not have permission to view this issue match")
	}

	issueMatchFilter := entity.IssueMatchFilter{Id: []*int64{&issueMatchId}}
	options := entity.ListOptions{Order: []entity.Order{}}
	issueMatches, err := im.ListIssueMatches(&issueMatchFilter, &options)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Internal error while retrieving issueMatches.")
	}

	if len(issueMatches.Elements) != 1 {
		return nil, NewIssueMatchHandlerError(fmt.Sprintf("IssueMatch %d not found.", issueMatchId))
	}

	im.eventRegistry.PushEvent(&GetIssueMatchEvent{
		IssueMatchID: issueMatchId,
		Result:       issueMatches.Elements[0].IssueMatch,
	})

	return issueMatches.Elements[0].IssueMatch, nil
}

func (im *issueMatchHandler) ListIssueMatches(filter *entity.IssueMatchFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginatedX(&filter.PaginatedX)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListIssueMatchesEventName,
		"filter": filter,
	})

	// get current user id
	currentUserId, err := common.GetCurrentUserId(im.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Error while getting current user id")
	}

	// Authorization check
	accessibleCompInstIds, err := im.authz.GetListOfAccessibleObjectIds(openfga.UserId(fmt.Sprint(currentUserId)), openfga.TypeComponentInstance)
	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Error while listing accessible issue matches for user")
	}

	// Update the filter.ComponentInstanceId based on accessibleCompInstIds
	filter.ComponentInstanceId = common.CombineFilterWithAccesibleIds(filter.ComponentInstanceId, accessibleCompInstIds)

	res, err := cache.CallCached[[]entity.IssueMatchResult](
		im.cache,
		CacheTtlGetIssueMatches,
		"GetIssueMatches",
		im.database.GetIssueMatches,
		filter,
		options.Order,
	)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Error while filtering for Issue Matches")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				im.cache,
				CacheTtlGetAllIssueMatchCursors,
				"GetAllIssueMatchCursors",
				im.database.GetAllIssueMatchCursors,
				filter,
				options.Order,
			)
			if err != nil {
				l.Error(err)
				return nil, NewIssueMatchHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			im.cache,
			CacheTtlCountIssueMatches,
			"CountIssueMatches",
			im.database.CountIssueMatches,
			filter,
		)
		if err != nil {
			l.Error(err)
			return nil, NewIssueMatchHandlerError("Error while total count of Issue Matches")
		}
	}

	ret := &entity.List[entity.IssueMatchResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	im.eventRegistry.PushEvent(&ListIssueMatchesEvent{
		Filter:  filter,
		Options: options,
		Results: ret,
	})

	return ret, nil
}

func (im *issueMatchHandler) CreateIssueMatch(issueMatch *entity.IssueMatch) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateIssueMatchEventName,
		"object": issueMatch,
	})

	var err error
	issueMatch.CreatedBy, err = common.GetCurrentUserId(im.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Internal error while creating issueMatch (GetUserId).")
	}
	issueMatch.UpdatedBy = issueMatch.CreatedBy

	severityFilter := &entity.SeverityFilter{
		IssueId: []*int64{&issueMatch.IssueId},
	}

	//@todo discuss: may be moved to somewhere else?
	effectiveSeverity, err := im.severityHandler.GetSeverity(severityFilter)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Internal error while retrieving effective severity.")
	}

	issueMatch.Severity = *effectiveSeverity

	newIssueMatch, err := im.database.CreateIssueMatch(issueMatch)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Internal error while creating issueMatch.")
	}

	im.eventRegistry.PushEvent(&CreateIssueMatchEvent{
		IssueMatch: newIssueMatch,
	})

	return newIssueMatch, nil
}

func (im *issueMatchHandler) UpdateIssueMatch(issueMatch *entity.IssueMatch) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateIssueMatchEventName,
		"object": issueMatch,
	})

	var err error
	issueMatch.UpdatedBy, err = common.GetCurrentUserId(im.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Internal error while updating issueMatch (GetUserId).")
	}

	err = im.database.UpdateIssueMatch(issueMatch)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Internal error while updating issueMatch.")
	}

	im.eventRegistry.PushEvent(&UpdateIssueMatchEvent{
		IssueMatch: issueMatch,
	})

	return im.GetIssueMatch(issueMatch.Id)
}

func (im *issueMatchHandler) DeleteIssueMatch(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteIssueMatchEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(im.database)
	if err != nil {
		l.Error(err)
		return NewIssueMatchHandlerError("Internal error while deleting issueMatch (GetUserId).")
	}

	err = im.database.DeleteIssueMatch(id, userId)

	if err != nil {
		l.Error(err)
		return NewIssueMatchHandlerError("Internal error while deleting issueMatch.")
	}

	im.eventRegistry.PushEvent(&DeleteIssueMatchEvent{
		IssueMatchID: id,
	})

	return nil
}

func (im *issueMatchHandler) AddEvidenceToIssueMatch(issueMatchId, evidenceId int64) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":        AddEvidenceToIssueMatchEventName,
		"issueMatchId": issueMatchId,
		"evidenceId":   evidenceId,
	})

	err := im.database.AddEvidenceToIssueMatch(issueMatchId, evidenceId)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Internal error while adding evidence to issueMatch.")
	}

	im.eventRegistry.PushEvent(&AddEvidenceToIssueMatchEvent{
		IssueMatchID: issueMatchId,
		EvidenceID:   evidenceId,
	})

	return im.GetIssueMatch(issueMatchId)
}

func (im *issueMatchHandler) RemoveEvidenceFromIssueMatch(issueMatchId, evidenceId int64) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":        RemoveEvidenceFromIssueMatchEventName,
		"issueMatchId": issueMatchId,
		"evidenceId":   evidenceId,
	})

	err := im.database.RemoveEvidenceFromIssueMatch(issueMatchId, evidenceId)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchHandlerError("Internal error while removing evidence from issueMatch.")
	}

	im.eventRegistry.PushEvent(&RemoveEvidenceFromIssueMatchEvent{
		IssueMatchID: issueMatchId,
	})

	return im.GetIssueMatch(issueMatchId)
}
