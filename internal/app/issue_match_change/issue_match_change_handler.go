// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match_change

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/pkg/util"
	"github.com/sirupsen/logrus"
)

type issueMatchChangeHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewIssueMatchChangeHandler(db database.Database, er event.EventRegistry) IssueMatchChangeHandler {
	return &issueMatchChangeHandler{
		database:      db,
		eventRegistry: er,
	}
}

type IssueMatchChangeHandlerError struct {
	msg string
}

func (e *IssueMatchChangeHandlerError) Error() string {
	return fmt.Sprintf("IssueMatchChangeHandlerError: %s", e.msg)
}

func NewIssueMatchChangeHandlerError(msg string) *IssueMatchChangeHandlerError {
	return &IssueMatchChangeHandlerError{msg: msg}
}

func (imc *issueMatchChangeHandler) getIssueMatchChangeResults(filter *entity.IssueMatchChangeFilter) ([]entity.IssueMatchChangeResult, error) {
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

func (imc *issueMatchChangeHandler) ListIssueMatchChanges(filter *entity.IssueMatchChangeFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchChangeResult], error) {
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
		return nil, NewIssueMatchChangeHandlerError("Error while filtering for IssueMatchChanges")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := imc.database.GetAllIssueMatchChangeIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewIssueMatchChangeHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = imc.database.CountIssueMatchChanges(filter)
		if err != nil {
			l.Error(err)
			return nil, NewIssueMatchChangeHandlerError("Error while total count of IssueMatchChanges")
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

func (imc *issueMatchChangeHandler) CreateIssueMatchChange(issueMatchChange *entity.IssueMatchChange) (*entity.IssueMatchChange, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateIssueMatchChangeEventName,
		"object": issueMatchChange,
	})

	var err error
	issueMatchChange.CreatedBy, err = common.GetCurrentUserId(imc.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeHandlerError("Internal error while creating issueMatchChange (GetUserId).")
	}
	issueMatchChange.UpdatedBy = issueMatchChange.CreatedBy

	newIssueMatchChange, err := imc.database.CreateIssueMatchChange(issueMatchChange)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeHandlerError("Internal error while creating issueMatchChange.")
	}

	imc.eventRegistry.PushEvent(&CreateIssueMatchChangeEvent{
		IssueMatchChange: newIssueMatchChange,
	})

	return newIssueMatchChange, nil
}

func (imc *issueMatchChangeHandler) UpdateIssueMatchChange(issueMatchChange *entity.IssueMatchChange) (*entity.IssueMatchChange, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateIssueMatchChangeEventName,
		"object": issueMatchChange,
	})

	var err error
	issueMatchChange.UpdatedBy, err = common.GetCurrentUserId(imc.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeHandlerError("Internal error while updating issueMatchChange (GetUserId).")
	}

	err = imc.database.UpdateIssueMatchChange(issueMatchChange)

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeHandlerError("Internal error while updating issueMatchChange.")
	}

	imcResult, err := imc.ListIssueMatchChanges(&entity.IssueMatchChangeFilter{Id: []*int64{&issueMatchChange.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueMatchChangeHandlerError("Internal error while retrieving updated issueMatchChange.")
	}

	if len(imcResult.Elements) != 1 {
		l.Error(err)
		return nil, NewIssueMatchChangeHandlerError("Multiple issueMatchChanges found.")
	}

	imc.eventRegistry.PushEvent(&UpdateIssueMatchChangeEvent{
		IssueMatchChange: imcResult.Elements[0].IssueMatchChange,
	})

	return imcResult.Elements[0].IssueMatchChange, nil
}

func (imc *issueMatchChangeHandler) DeleteIssueMatchChange(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteIssueMatchChangeEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(imc.database)
	if err != nil {
		l.Error(err)
		return NewIssueMatchChangeHandlerError("Internal error while deleting issueMatchChange (GetUserId).")
	}

	err = imc.database.DeleteIssueMatchChange(id, userId)

	if err != nil {
		l.Error(err)
		return NewIssueMatchChangeHandlerError("Internal error while deleting issueMatchChange.")
	}

	imc.eventRegistry.PushEvent(&DeleteIssueMatchChangeEvent{
		IssueMatchChangeID: id,
	})

	return nil
}
