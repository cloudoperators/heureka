// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getIssueVariantResults(filter *entity.IssueVariantFilter) ([]entity.IssueVariantResult, error) {
	var ivResults []entity.IssueVariantResult
	issueVariants, err := h.database.GetIssueVariants(filter)
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

func (h *HeurekaApp) ListIssueVariants(filter *entity.IssueVariantFilter, options *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListIssueVariants",
		"filter": filter,
	})

	res, err := h.getIssueVariantResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for IssueVariants")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllIssueVariantIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountIssueVariants(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of IssueVariants")
		}
	}

	return &entity.List[entity.IssueVariantResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) ListEffectiveIssueVariants(filter *entity.IssueVariantFilter, options *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListEffectiveIssueVariants",
		"filter": filter,
	})

	issueVariants, err := h.ListIssueVariants(filter, options)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while returning issueVariants.")
	}

	repositoryIds := lo.Map(issueVariants.Elements, func(item entity.IssueVariantResult, _ int) *int64 {
		return &item.IssueRepositoryId
	})

	repositoryFilter := entity.IssueRepositoryFilter{
		Id: repositoryIds,
	}

	opts := entity.ListOptions{}
	repositories, err := h.ListIssueRepositories(&repositoryFilter, &opts)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while returning issue repositories.")
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

	return &entity.List[entity.IssueVariantResult]{
		Elements: maxPriorityVariants,
	}, nil
}

func (h *HeurekaApp) CreateIssueVariant(issueVariant *entity.IssueVariant) (*entity.IssueVariant, error) {
	f := &entity.IssueVariantFilter{
		SecondaryName: []*string{&issueVariant.SecondaryName},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateIssueVariant",
		"object": issueVariant,
		"filter": f,
	})

	issueVariants, err := h.ListIssueVariants(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating issueVariant.")
	}

	if len(issueVariants.Elements) > 0 {
		l.Error(err)
		return nil, heurekaError(fmt.Sprintf("Duplicated entry %s for name.", issueVariant.SecondaryName))
	}

	newIv, err := h.database.CreateIssueVariant(issueVariant)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating issueVariant.")
	}

	return newIv, nil
}

func (h *HeurekaApp) UpdateIssueVariant(issueVariant *entity.IssueVariant) (*entity.IssueVariant, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateIssueVariant",
		"object": issueVariant,
	})

	err := h.database.UpdateIssueVariant(issueVariant)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating issueVariant.")
	}

	ivResult, err := h.ListIssueVariants(&entity.IssueVariantFilter{Id: []*int64{&issueVariant.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated issueVariant.")
	}

	if len(ivResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple issueVariants found.")
	}

	return ivResult.Elements[0].IssueVariant, nil
}

func (h *HeurekaApp) DeleteIssueVariant(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteIssueVariant",
		"id":    id,
	})

	err := h.database.DeleteIssueVariant(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting issueVariant.")
	}

	return nil
}
