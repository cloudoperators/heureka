// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/pkg/util"
)

func (h *HeurekaApp) getIssueMatchChangeResults(filter *entity.IssueMatchChangeFilter) ([]entity.IssueMatchChangeResult, error) {
	var results []entity.IssueMatchChangeResult
	vmcs, err := h.database.GetIssueMatchChanges(filter)
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

func (h *HeurekaApp) ListIssueMatchChanges(filter *entity.IssueMatchChangeFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchChangeResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListIssueMatchChanges",
		"filter": filter,
	})

	res, err := h.getIssueMatchChangeResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for IssueMatchChanges")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllIssueMatchChangeIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountIssueMatchChanges(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of IssueMatchChanges")
		}
	}

	return &entity.List[entity.IssueMatchChangeResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}
