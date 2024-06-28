// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.wdf.sap.corp/cc/heureka/pkg/util"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getIssueResultsWithAggregations(filter *entity.IssueFilter) ([]entity.IssueResult, error) {
	var issueResults []entity.IssueResult
	issues, err := h.database.GetIssuesWithAggregations(filter)
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		cursor := fmt.Sprintf("%d", issue.Id)
		issueResults = append(issueResults, entity.IssueResult{
			WithCursor:        entity.WithCursor{Value: cursor},
			IssueAggregations: util.Ptr(issue.IssueAggregations),
			Issue:             util.Ptr(issue.Issue),
		})
	}

	return issueResults, nil
}

func (h *HeurekaApp) getIssueResults(filter *entity.IssueFilter) ([]entity.IssueResult, error) {
	var issueResults []entity.IssueResult
	issues, err := h.database.GetIssues(filter)
	if err != nil {
		return nil, err
	}
	for _, issue := range issues {
		cursor := fmt.Sprintf("%d", issue.Id)
		issueResults = append(issueResults, entity.IssueResult{
			WithCursor:        entity.WithCursor{Value: cursor},
			IssueAggregations: nil,
			Issue:             util.Ptr(issue),
		})
	}
	return issueResults, nil
}

func (h *HeurekaApp) ListIssues(filter *entity.IssueFilter, options *entity.ListOptions) (*entity.List[entity.IssueResult], error) {
	var count int64
	var pageInfo *entity.PageInfo
	var res []entity.IssueResult
	var err error

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListIssues",
		"filter": filter,
	})

	ensurePaginated(&filter.Paginated)

	if options.IncludeAggregations {
		res, err = h.getIssueResultsWithAggregations(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Internal error while retrieving list results with aggregations")
		}
	} else {
		res, err = h.getIssueResults(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Internal error while retrieving list results.")
		}
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllIssueIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountIssues(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of issues")
		}
	}

	return &entity.List[entity.IssueResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}
