// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_variant

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var CacheTtlGetIssueVariants = 12 * time.Hour
var CacheTtlGetAllIssueVariantIds = 12 * time.Hour
var CacheTtlCountIssueVariants = 12 * time.Hour

type issueVariantHandler struct {
	database          database.Database
	eventRegistry     event.EventRegistry
	repositoryService issue_repository.IssueRepositoryHandler
	cache             cache.Cache
}

func NewIssueVariantHandler(database database.Database, eventRegistry event.EventRegistry, rs issue_repository.IssueRepositoryHandler, cache cache.Cache) IssueVariantHandler {
	return &issueVariantHandler{
		database:          database,
		eventRegistry:     eventRegistry,
		repositoryService: rs,
		cache:             cache,
	}
}

type IssueVariantHandlerError struct {
	message string
}

func NewIssueVariantHandlerError(message string) *IssueVariantHandlerError {
	return &IssueVariantHandlerError{message: message}
}

func (e *IssueVariantHandlerError) Error() string {
	return e.message
}

func (iv *issueVariantHandler) getIssueVariantResults(filter *entity.IssueVariantFilter) ([]entity.IssueVariantResult, error) {
	var ivResults []entity.IssueVariantResult
	issueVariants, err := cache.CallCached[[]entity.IssueVariant](
		iv.cache,
		CacheTtlGetIssueVariants,
		"GetIssueVariants",
		iv.database.GetIssueVariants,
		filter)

	if err != nil {
		return nil, err
	}
	for _, iv := range issueVariants {
		issueVariant := iv
		cursor := fmt.Sprintf("%d", issueVariant.Id)
		ivResults = append(ivResults, entity.IssueVariantResult{
			WithCursor:               entity.WithCursor{Value: cursor},
			IssueVariantAggregations: nil,
			IssueVariant:             &issueVariant,
		})
	}
	return ivResults, nil
}

func (iv *issueVariantHandler) ListIssueVariants(filter *entity.IssueVariantFilter, options *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)
	l := logrus.WithFields(logrus.Fields{
		"event":  ListIssueVariantsEventName,
		"filter": filter,
	})

	res, err := iv.getIssueVariantResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Error while filtering for IssueVariants")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := cache.CallCached[[]int64](
				iv.cache,
				CacheTtlGetAllIssueVariantIds,
				"GetAllIssueVariantIds",
				iv.database.GetAllIssueVariantIds,
				filter,
			)
			if err != nil {
				l.Error(err)
				return nil, NewIssueVariantHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			iv.cache,
			CacheTtlCountIssueVariants,
			"CountIssueVariants",
			iv.database.CountIssueVariants,
			filter,
		)
		if err != nil {
			l.Error(err)
			return nil, NewIssueVariantHandlerError("Error while total count of IssueVariants")
		}
	}

	ret := &entity.List[entity.IssueVariantResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	iv.eventRegistry.PushEvent(&ListIssueVariantsEvent{
		Filter:  filter,
		Options: options,
		Results: ret,
	})

	return ret, nil
}

func (iv *issueVariantHandler) ListEffectiveIssueVariants(filter *entity.IssueVariantFilter, options *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListEffectiveIssueVariantsEventName,
		"filter": filter,
	})

	issueVariants, err := iv.ListIssueVariants(filter, options)

	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Internal error while returning issueVariants.")
	}

	repositoryIds := lo.Map(issueVariants.Elements, func(item entity.IssueVariantResult, _ int) *int64 {
		return &item.IssueRepositoryId
	})

	repositoryFilter := entity.IssueRepositoryFilter{
		Id: repositoryIds,
	}

	opts := entity.ListOptions{}
	repositories, err := iv.repositoryService.ListIssueRepositories(&repositoryFilter, &opts)

	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Internal error while returning issue repositories.")
	}

	maxPriorityIr := lo.MaxBy(repositories.Elements, func(item entity.IssueRepositoryResult, max entity.IssueRepositoryResult) bool {
		return item.Priority > max.Priority
	})

	maxRepositoryIds := lo.FilterMap(repositories.Elements, func(item entity.IssueRepositoryResult, index int) (int64, bool) {
		if item.Priority == maxPriorityIr.Priority {
			return item.Id, true
		}
		return 0, false
	})

	maxPriorityVariants := lo.Filter(issueVariants.Elements, func(item entity.IssueVariantResult, _ int) bool {
		return lo.Contains(maxRepositoryIds, item.IssueRepositoryId)
	})

	ret := &entity.List[entity.IssueVariantResult]{
		Elements: maxPriorityVariants,
	}

	iv.eventRegistry.PushEvent(&ListEffectiveIssueVariantsEvent{
		Filter:  filter,
		Options: options,
		Results: ret,
	})

	return ret, nil
}

func (iv *issueVariantHandler) CreateIssueVariant(issueVariant *entity.IssueVariant) (*entity.IssueVariant, error) {
	f := &entity.IssueVariantFilter{
		SecondaryName: []*string{&issueVariant.SecondaryName},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateIssueVariantEventName,
		"object": issueVariant,
		"filter": f,
	})

	var err error
	issueVariant.CreatedBy, err = common.GetCurrentUserId(iv.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Internal error while creating issueVariant (GetUserId).")
	}
	issueVariant.UpdatedBy = issueVariant.CreatedBy

	issueVariants, err := iv.ListIssueVariants(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Internal error while creating issueVariant.")
	}

	if len(issueVariants.Elements) > 0 {
		l.Error(err)
		return nil, NewIssueVariantHandlerError(fmt.Sprintf("Duplicated entry %s for name.", issueVariant.SecondaryName))
	}

	newIv, err := iv.database.CreateIssueVariant(issueVariant)

	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Internal error while creating issueVariant.")
	}

	iv.eventRegistry.PushEvent(&CreateIssueVariantEvent{IssueVariant: newIv})

	return newIv, nil
}

func (iv *issueVariantHandler) UpdateIssueVariant(issueVariant *entity.IssueVariant) (*entity.IssueVariant, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateIssueVariantEventName,
		"object": issueVariant,
	})

	var err error
	issueVariant.UpdatedBy, err = common.GetCurrentUserId(iv.database)
	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Internal error while updating issueVariant (GetUserId).")
	}

	err = iv.database.UpdateIssueVariant(issueVariant)

	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Internal error while updating issueVariant.")
	}

	ivResult, err := iv.ListIssueVariants(&entity.IssueVariantFilter{Id: []*int64{&issueVariant.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Internal error while retrieving updated issueVariant.")
	}

	if len(ivResult.Elements) != 1 {
		l.Error(err)
		return nil, NewIssueVariantHandlerError("Multiple issueVariants found.")
	}

	iv.eventRegistry.PushEvent(&UpdateIssueVariantEvent{IssueVariant: ivResult.Elements[0].IssueVariant})

	return ivResult.Elements[0].IssueVariant, nil
}

func (iv *issueVariantHandler) DeleteIssueVariant(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteIssueVariantEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(iv.database)
	if err != nil {
		l.Error(err)
		return NewIssueVariantHandlerError("Internal error while deleting issueVariant (GetUserId).")
	}

	err = iv.database.DeleteIssueVariant(id, userId)

	if err != nil {
		l.Error(err)
		return NewIssueVariantHandlerError("Internal error while deleting issueVariant.")
	}

	iv.eventRegistry.PushEvent(&DeleteIssueVariantEvent{IssueVariantID: id})

	return nil
}
