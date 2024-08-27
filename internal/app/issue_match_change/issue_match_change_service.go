// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match_change

import (
	"fmt"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/pkg/util"
)

type issueMatchChangeService struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewIssueMatchChangeService(db database.Database, er event.EventRegistry) IssueMatchChangeService {
	return &issueMatchChangeService{
		database:      db,
		eventRegistry: er,
	}
}

type IssueMatchChangeServiceError struct {
	msg string
}

func (e *IssueMatchChangeServiceError) Error() string {
	return fmt.Sprintf("IssueMatchChangeServiceError: %s", e.msg)
}

func NewIssueMatchChangeServiceError(msg string) *IssueMatchChangeServiceError {
	return &IssueMatchChangeServiceError{msg: msg}
}

func (imc *issueMatchChangeService) getIssueMatchChangeResults(filter *entity.IssueMatchChangeFilter) ([]entity.IssueMatchChangeResult, error) {
	var results []entity.IssueMatchChangeResult
	vmcs, err := imc.database.GetIssueMatchChanges(filter)
	if err != nil {
		return nil, err
	}
	for _, vmc := range vmcs {
		cursor := fmt.Sprintf("%d", vmc.Id)
		results = append(results, entity.IssueMatchChangeResult{
			WithCursor:       entity.WithCursor{Value: cursor},
			IssueMatchChange: util.Ptr(vmc),
		})
	}

	return results, nil
}

func (imc *issueMatchChangeService) ListIssueMatchChanges(filter *entity.IssueMatchChangeFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchChangeResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListIssueMatchChangesEventName,
		"filter": filter,
	})

	res, err := imc.getIssueMatchChangeResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeServiceError("Error while filtering for IssueMatchChanges")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := imc.database.GetAllIssueMatchChangeIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewIssueMatchChangeServiceError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = imc.database.CountIssueMatchChanges(filter)
		if err != nil {
			l.Error(err)
			return nil, NewIssueMatchChangeServiceError("Error while total count of IssueMatchChanges")
		}
	}

	ret := &entity.List[entity.IssueMatchChangeResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	imc.eventRegistry.PushEvent(&ListIssueMatchChangesEvent{
		Filter:  filter,
		Options: options,
		Results: ret,
	})

	return ret, nil
}

func (imc *issueMatchChangeService) CreateIssueMatchChange(issueMatchChange *entity.IssueMatchChange) (*entity.IssueMatchChange, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateIssueMatchChangeEventName,
		"object": issueMatchChange,
	})

	newIssueMatchChange, err := imc.database.CreateIssueMatchChange(issueMatchChange)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeServiceError("Internal error while creating issueMatchChange.")
	}

	imc.eventRegistry.PushEvent(&CreateIssueMatchChangeEvent{
		IssueMatchChange: newIssueMatchChange,
	})

	return newIssueMatchChange, nil
}

func (imc *issueMatchChangeService) UpdateIssueMatchChange(issueMatchChange *entity.IssueMatchChange) (*entity.IssueMatchChange, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateIssueMatchChangeEventName,
		"object": issueMatchChange,
	})

	err := imc.database.UpdateIssueMatchChange(issueMatchChange)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeServiceError("Internal error while updating issueMatchChange.")
	}

	imcResult, err := imc.ListIssueMatchChanges(&entity.IssueMatchChangeFilter{Id: []*int64{&issueMatchChange.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeServiceError("Internal error while retrieving updated issueMatchChange.")
	}

	if len(imcResult.Elements) != 1 {
		l.Error(err)
		return nil, NewIssueMatchChangeServiceError("Multiple issueMatchChanges found.")
	}

	imc.eventRegistry.PushEvent(&UpdateIssueMatchChangeEvent{
		IssueMatchChange: imcResult.Elements[0].IssueMatchChange,
	})

	return imcResult.Elements[0].IssueMatchChange, nil
}

func (imc *issueMatchChangeService) DeleteIssueMatchChange(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteIssueMatchChangeEventName,
		"id":    id,
	})

	err := imc.database.DeleteIssueMatchChange(id)

	if err != nil {
		l.Error(err)
		return NewIssueMatchChangeServiceError("Internal error while deleting issueMatchChange.")
	}

	imc.eventRegistry.PushEvent(&DeleteIssueMatchChangeEvent{
		IssueMatchChangeID: id,
	})

	return nil
}
