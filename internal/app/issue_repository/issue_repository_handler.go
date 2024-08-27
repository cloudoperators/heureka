// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_repository

import (
	"fmt"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

type issueRepositoryHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewIssueRepositoryHandler(database database.Database, eventRegistry event.EventRegistry) IssueRepositoryHandler {
	return &issueRepositoryHandler{
		database:      database,
		eventRegistry: eventRegistry,
	}
}

type IssueRepositoryHandlerError struct {
	message string
}

func NewIssueRepositoryHandlerError(message string) *IssueRepositoryHandlerError {
	return &IssueRepositoryHandlerError{message: message}
}

func (e *IssueRepositoryHandlerError) Error() string {
	return e.message
}

func (ir *issueRepositoryHandler) getIssueRepositoryResults(filter *entity.IssueRepositoryFilter) ([]entity.IssueRepositoryResult, error) {
	var irResults []entity.IssueRepositoryResult
	issueRepositories, err := ir.database.GetIssueRepositories(filter)
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

func (ir *issueRepositoryHandler) ListIssueRepositories(filter *entity.IssueRepositoryFilter, options *entity.ListOptions) (*entity.List[entity.IssueRepositoryResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)
	l := logrus.WithFields(logrus.Fields{
		"event":  ListIssueRepositoriesEventName,
		"filter": filter,
	})

	res, err := ir.getIssueRepositoryResults(filter)

	if err != nil {
		return nil, NewIssueRepositoryHandlerError("Error while filtering for IssueRepositories")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := ir.database.GetAllIssueRepositoryIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewIssueRepositoryHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = ir.database.CountIssueRepositories(filter)
		if err != nil {
			l.Error(err)
			return nil, NewIssueRepositoryHandlerError("Error while total count of IssueRepositories")
		}
	}

	ret := &entity.List[entity.IssueRepositoryResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	ir.eventRegistry.PushEvent(&ListIssueRepositoriesEvent{Filter: filter, Options: options, Results: ret})

	return ret, nil
}

func (ir *issueRepositoryHandler) CreateIssueRepository(issueRepository *entity.IssueRepository) (*entity.IssueRepository, error) {
	f := &entity.IssueRepositoryFilter{
		Name: []*string{&issueRepository.Name},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateIssueRepositoryEventName,
		"object": issueRepository,
		"filter": f,
	})

	issueRepositories, err := ir.ListIssueRepositories(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueRepositoryHandlerError("Internal error while creating issueRepository.")
	}

	if len(issueRepositories.Elements) > 0 {
		l.Error(err)
		return nil, NewIssueRepositoryHandlerError(fmt.Sprintf("Duplicated entry %s for name.", issueRepository.Name))
	}

	newIssueRepository, err := ir.database.CreateIssueRepository(issueRepository)

	if err != nil {
		l.Error(err)
		return nil, NewIssueRepositoryHandlerError("Internal error while creating issueRepository.")
	}

	ir.eventRegistry.PushEvent(&CreateIssueRepositoryEvent{IssueRepository: newIssueRepository})

	return newIssueRepository, nil
}

func (ir *issueRepositoryHandler) UpdateIssueRepository(issueRepository *entity.IssueRepository) (*entity.IssueRepository, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateIssueRepositoryEventName,
		"object": issueRepository,
	})

	err := ir.database.UpdateIssueRepository(issueRepository)

	if err != nil {
		l.Error(err)
		return nil, NewIssueRepositoryHandlerError("Internal error while updating issueRepository.")
	}

	issueRepositoryResult, err := ir.ListIssueRepositories(&entity.IssueRepositoryFilter{Id: []*int64{&issueRepository.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueRepositoryHandlerError("Internal error while retrieving updated issueRepository.")
	}

	if len(issueRepositoryResult.Elements) != 1 {
		l.Error(err)
		return nil, NewIssueRepositoryHandlerError("Multiple issueRepositories found.")
	}

	ir.eventRegistry.PushEvent(&UpdateIssueRepositoryEvent{IssueRepository: issueRepository})

	return issueRepositoryResult.Elements[0].IssueRepository, nil
}

func (ir *issueRepositoryHandler) DeleteIssueRepository(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteIssueRepositoryEventName,
		"id":    id,
	})

	err := ir.database.DeleteIssueRepository(id)

	if err != nil {
		l.Error(err)
		return NewIssueRepositoryHandlerError("Internal error while updating issueRepository.")
	}

	ir.eventRegistry.PushEvent(&DeleteIssueRepositoryEvent{IssueRepositoryID: id})

	return nil
}
