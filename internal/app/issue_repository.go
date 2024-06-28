// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getIssueRepositoryResults(filter *entity.IssueRepositoryFilter) ([]entity.IssueRepositoryResult, error) {
	var irResults []entity.IssueRepositoryResult
	issueRepositories, err := h.database.GetIssueRepositories(filter)
	if err != nil {
		return nil, err
	}
	for _, ir := range issueRepositories {
		issueRepository := ir
		cursor := fmt.Sprintf("%d", ir.Id)
		irResults = append(irResults, entity.IssueRepositoryResult{
			WithCursor:                  entity.WithCursor{Value: cursor},
			IssueRepositoryAggregations: nil,
			IssueRepository:             &issueRepository,
		})
	}
	return irResults, nil
}

func (h *HeurekaApp) ListIssueRepositories(filter *entity.IssueRepositoryFilter, options *entity.ListOptions) (*entity.List[entity.IssueRepositoryResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListIssueRepositories",
		"filter": filter,
	})

	res, err := h.getIssueRepositoryResults(filter)

	if err != nil {
		return nil, heurekaError("Error while filtering for IssueRepositories")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllIssueRepositoryIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountIssueRepositories(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of IssueRepositories")
		}
	}

	return &entity.List[entity.IssueRepositoryResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateIssueRepository(issueRepository *entity.IssueRepository) (*entity.IssueRepository, error) {
	f := &entity.IssueRepositoryFilter{
		Name: []*string{&issueRepository.Name},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateIssueRepository",
		"object": issueRepository,
		"filter": f,
	})

	issueRepositories, err := h.ListIssueRepositories(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating issueRepository.")
	}

	if len(issueRepositories.Elements) > 0 {
		l.Error(err)
		return nil, heurekaError(fmt.Sprintf("Duplicated entry %s for name.", issueRepository.Name))
	}

	newIssueRepository, err := h.database.CreateIssueRepository(issueRepository)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating issueRepository.")
	}

	return newIssueRepository, nil
}

func (h *HeurekaApp) UpdateIssueRepository(issueRepository *entity.IssueRepository) (*entity.IssueRepository, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateIssueRepository",
		"object": issueRepository,
	})

	err := h.database.UpdateIssueRepository(issueRepository)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating issueRepository.")
	}

	issueRepositoryResult, err := h.ListIssueRepositories(&entity.IssueRepositoryFilter{Id: []*int64{&issueRepository.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated issueRepository.")
	}

	if len(issueRepositoryResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple issueRepositories found.")
	}

	return issueRepositoryResult.Elements[0].IssueRepository, nil
}

func (h *HeurekaApp) DeleteIssueRepository(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteIssueRepository",
		"id":    id,
	})

	err := h.database.DeleteIssueRepository(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while updating issueRepository.")
	}

	return nil
}
