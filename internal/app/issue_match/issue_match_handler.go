// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match

import (
	"fmt"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/app/severity"
	"github.wdf.sap.corp/cc/heureka/internal/database"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/pkg/util"
)

type issueMatchService struct {
	database        database.Database
	eventRegistry   event.EventRegistry
	severityService severity.SeverityService
}

func NewIssueMatchService(database database.Database, eventRegistry event.EventRegistry, ss severity.SeverityService) IssueMatchService {
	return &issueMatchService{
		database:        database,
		eventRegistry:   eventRegistry,
		severityService: ss,
	}
}

type IssueMatchServiceError struct {
	message string
}

func NewIssueMatchServiceError(message string) *IssueMatchServiceError {
	return &IssueMatchServiceError{message: message}
}

func (e *IssueMatchServiceError) Error() string {
	return e.message
}

func (h *issueMatchService) getIssueMatchResults(filter *entity.IssueMatchFilter) ([]entity.IssueMatchResult, error) {
	var results []entity.IssueMatchResult
	ims, err := h.database.GetIssueMatches(filter)
	if err != nil {
		return nil, err
	}
	for _, im := range ims {
		cursor := fmt.Sprintf("%d", im.Id)
		results = append(results, entity.IssueMatchResult{
			WithCursor: entity.WithCursor{Value: cursor},
			IssueMatch: util.Ptr(im),
		})
	}

	return results, nil
}

func (im *issueMatchService) GetIssueMatch(issueMatchId int64) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetIssueMatchEventName,
		"id":    issueMatchId,
	})
	issueMatchFilter := entity.IssueMatchFilter{Id: []*int64{&issueMatchId}}
	issueMatches, err := im.ListIssueMatches(&issueMatchFilter, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchServiceError("Internal error while retrieving issueMatches.")
	}

	if len(issueMatches.Elements) != 1 {
		return nil, NewIssueMatchServiceError(fmt.Sprintf("IssueMatch %d not found.", issueMatchId))
	}

	im.eventRegistry.PushEvent(&GetIssueMatchEvent{
		IssueMatchID: issueMatchId,
		Result:       issueMatches.Elements[0].IssueMatch,
	})

	return issueMatches.Elements[0].IssueMatch, nil
}

func (im *issueMatchService) ListIssueMatches(filter *entity.IssueMatchFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListIssueMatchesEventName,
		"filter": filter,
	})

	res, err := im.getIssueMatchResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchServiceError("Error while filtering for Issue Matches")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := im.database.GetAllIssueMatchIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewIssueMatchServiceError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = im.database.CountIssueMatches(filter)
		if err != nil {
			l.Error(err)
			return nil, NewIssueMatchServiceError("Error while total count of Issue Matches")
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

func (im *issueMatchService) CreateIssueMatch(issueMatch *entity.IssueMatch) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateIssueMatchEventName,
		"object": issueMatch,
	})

	severityFilter := &entity.SeverityFilter{
		IssueId: []*int64{&issueMatch.IssueId},
	}

	//@todo discuss: may be moved to somewhere else?
	effectiveSeverity, err := im.severityService.GetSeverity(severityFilter)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchServiceError("Internal error while retrieving effective severity.")
	}

	issueMatch.Severity = *effectiveSeverity

	newIssueMatch, err := im.database.CreateIssueMatch(issueMatch)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchServiceError("Internal error while creating issueMatch.")
	}

	im.eventRegistry.PushEvent(&CreateIssueMatchEvent{
		IssueMatch: newIssueMatch,
	})

	return newIssueMatch, nil
}

func (im *issueMatchService) UpdateIssueMatch(issueMatch *entity.IssueMatch) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateIssueMatchEventName,
		"object": issueMatch,
	})

	err := im.database.UpdateIssueMatch(issueMatch)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchServiceError("Internal error while updating issueMatch.")
	}

	im.eventRegistry.PushEvent(&UpdateIssueMatchEvent{
		IssueMatch: issueMatch,
	})

	return im.GetIssueMatch(issueMatch.Id)
}

func (im *issueMatchService) DeleteIssueMatch(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteIssueMatchEventName,
		"id":    id,
	})

	err := im.database.DeleteIssueMatch(id)

	if err != nil {
		l.Error(err)
		return NewIssueMatchServiceError("Internal error while deleting issueMatch.")
	}

	im.eventRegistry.PushEvent(&DeleteIssueMatchEvent{
		IssueMatchID: id,
	})

	return nil
}

func (im *issueMatchService) AddEvidenceToIssueMatch(issueMatchId, evidenceId int64) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":        AddEvidenceToIssueMatchEventName,
		"issueMatchId": issueMatchId,
		"evidenceId":   evidenceId,
	})

	err := im.database.AddEvidenceToIssueMatch(issueMatchId, evidenceId)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchServiceError("Internal error while adding evidence to issueMatch.")
	}

	im.eventRegistry.PushEvent(&AddEvidenceToIssueMatchEvent{
		IssueMatchID: issueMatchId,
		EvidenceID:   evidenceId,
	})

	return im.GetIssueMatch(issueMatchId)
}

func (im *issueMatchService) RemoveEvidenceFromIssueMatch(issueMatchId, evidenceId int64) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":        RemoveEvidenceFromIssueMatchEventName,
		"issueMatchId": issueMatchId,
		"evidenceId":   evidenceId,
	})

	err := im.database.RemoveEvidenceFromIssueMatch(issueMatchId, evidenceId)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchServiceError("Internal error while removing evidence from issueMatch.")
	}

	im.eventRegistry.PushEvent(&RemoveEvidenceFromIssueMatchEvent{
		IssueMatchID: issueMatchId,
	})

	return im.GetIssueMatch(issueMatchId)
}
