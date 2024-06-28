// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/pkg/util"
)

func (h *HeurekaApp) getIssueMatchResults(filter *entity.IssueMatchFilter) ([]entity.IssueMatchResult, error) {
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

func (h *HeurekaApp) ListIssueMatches(filter *entity.IssueMatchFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListIssueMatches",
		"filter": filter,
	})

	res, err := h.getIssueMatchResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for Issue Matches")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllIssueMatchIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountIssueMatches(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of Issue Matches")
		}
	}

	return &entity.List[entity.IssueMatchResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateIssueMatch(issueMatch *entity.IssueMatch) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateIssueMatch",
		"object": issueMatch,
	})

	severityFilter := &entity.SeverityFilter{
		IssueId: []*int64{&issueMatch.IssueId},
	}

	effectiveSeverity, err := h.GetSeverity(severityFilter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving effective severity.")
	}

	issueMatch.Severity = *effectiveSeverity

	newIssueMatch, err := h.database.CreateIssueMatch(issueMatch)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating issueMatch.")
	}

	return newIssueMatch, nil
}

func (h *HeurekaApp) UpdateIssueMatch(issueMatch *entity.IssueMatch) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateIssueMatch",
		"object": issueMatch,
	})

	err := h.database.UpdateIssueMatch(issueMatch)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating issueMatch.")
	}

	issueMatchResult, err := h.ListIssueMatches(&entity.IssueMatchFilter{Id: []*int64{&issueMatch.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated issueMatch.")
	}

	if len(issueMatchResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple issueMatches found.")
	}

	return issueMatchResult.Elements[0].IssueMatch, nil
}

func (h *HeurekaApp) DeleteIssueMatch(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteIssueMatch",
		"id":    id,
	})

	err := h.database.DeleteIssueMatch(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting issueMatch.")
	}

	return nil
}
