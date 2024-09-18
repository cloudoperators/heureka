// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/pkg/util"
)

func PreparePagination(filter *entity.Paginated, options *entity.ListOptions) {
	if options.ShowPageInfo {
		first := 10
		if filter.First != nil {
			first = *filter.First
		}
		//@†odo add debug log entry
		listSize := first + 1
		filter.First = &listSize
	}
}

func EnsurePaginated(filter *entity.Paginated) {
	if filter.First == nil {
		first := 10
		filter.First = &first
	}
	if filter.After == nil {
		var after int64 = 0
		filter.After = &after
	}
}

func GetPages(firstCursor *string, ids []int64, pageSize int) ([]entity.Page, entity.Page) {
	var currentCursor = util.Ptr("0")
	var pages []entity.Page
	var currentPage entity.Page
	var i = 0
	var pN = 0
	var page entity.Page
	for _, id := range ids {
		i++
		if i == 1 {
			pN++
			page = entity.Page{
				After:     currentCursor,
				IsCurrent: false,
			}
		}
		if fmt.Sprintf("%d", id) == *firstCursor {
			page.IsCurrent = true
		}
		page.PageCount = util.Ptr(i)
		if i >= pageSize {
			currentCursor = util.Ptr(fmt.Sprintf("%d", id))
			page.PageNumber = util.Ptr(pN)
			pages = append(pages, page)
			i = 0
			if page.IsCurrent {
				currentPage = page
			}
		}
	}
	if len(ids)%pageSize != 0 {
		page.PageNumber = util.Ptr(pN)
		pages = append(pages, page)
		if page.IsCurrent {
			currentPage = page
		}
	}
	return pages, currentPage
}

func GetPageInfo[T entity.HasCursor](res []T, ids []int64, pageSize int, currentCursor int64) *entity.PageInfo {
	var nextPageAfter *string
	currentAfter := util.Ptr(fmt.Sprintf("%d", currentCursor))
	firstCursor := res[0].Cursor()
	if len(res) > 1 {
		nextPageAfter = res[len(res)-1].Cursor()
	} else {
		nextPageAfter = firstCursor
	}

	pages, currentPage := GetPages(firstCursor, ids, pageSize)

	return &entity.PageInfo{
		HasNextPage:     util.Ptr(currentPage.PageNumber != nil && *currentPage.PageNumber < len(pages)),
		HasPreviousPage: util.Ptr(currentPage.PageNumber != nil && *currentPage.PageNumber > 1),
		IsValidPage:     util.Ptr(*currentPage.After == *currentAfter),
		PageNumber:      currentPage.PageNumber,
		NextPageAfter:   nextPageAfter,
		Pages:           pages,
	}
}

func FinalizePagination[T entity.HasCursor](results []T, filter *entity.Paginated, options *entity.ListOptions) (*entity.PageInfo, []T) {
	var pageInfo entity.PageInfo
	count := len(results)
	if options.ShowPageInfo && len(results) > 0 {
		hasNextPage := len(results) == *filter.First
		if hasNextPage {
			results = results[:count-1]
		}
		firstCursor := results[0].Cursor()
		lastCursor := results[len(results)-1].Cursor()
		pageInfo = entity.PageInfo{
			HasNextPage: &hasNextPage,
			StartCursor: firstCursor,
			EndCursor:   lastCursor,
		}

	}
	return &pageInfo, results
}
