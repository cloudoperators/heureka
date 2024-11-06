// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"github.com/samber/lo"

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
		filter.First = lo.ToPtr(10)
	}
	if filter.After == nil {
		var after int64 = 0
		filter.After = &after
	}
	if filter.Cursor == nil {
		filter.Cursor = lo.ToPtr("")
	}
}

func GetCursorPages(firstCursor *string, cursors []string, pageSize int) ([]entity.Page, entity.Page) {
	var currentCursor = ""
	var pages []entity.Page
	var currentPage entity.Page
	var i = 0
	var pN = 0
	var page entity.Page
	for _, c := range cursors {
		i++
		if i == 1 {
			pN++
			page = entity.Page{
				After:     &currentCursor,
				IsCurrent: false,
			}
		}
		if c == *firstCursor {
			page.IsCurrent = true
		}
		page.PageCount = util.Ptr(i)
		if i >= pageSize {
			currentCursor = c
			page.PageNumber = util.Ptr(pN)
			pages = append(pages, page)
			i = 0
			if page.IsCurrent {
				currentPage = page
			}
		}
	}
	if len(cursors)%pageSize != 0 {
		page.PageNumber = util.Ptr(pN)
		pages = append(pages, page)
		if page.IsCurrent {
			currentPage = page
		}
	}
	return pages, currentPage
}

func GetCursorPageInfo[T entity.HasCursor](res []T, cursors []string, pageSize int, currentCursor *string) *entity.PageInfo {

	var nextPageAfter *string
	currentAfter := currentCursor
	firstCursor := res[0].Cursor()

	if len(res) > 1 {
		nextPageAfter = res[len(res)-1].Cursor()
	} else {
		nextPageAfter = firstCursor
	}

	pages, currentPage := GetCursorPages(firstCursor, cursors, pageSize)

	return &entity.PageInfo{
		HasNextPage:     util.Ptr(currentPage.PageNumber != nil && *currentPage.PageNumber < len(pages)),
		HasPreviousPage: util.Ptr(currentPage.PageNumber != nil && *currentPage.PageNumber > 1),
		IsValidPage:     util.Ptr(currentPage.After != nil && currentAfter != nil && *currentPage.After == *currentAfter),
		PageNumber:      currentPage.PageNumber,
		NextPageAfter:   nextPageAfter,
		Pages:           pages,
	}
}

// Deprecated do not use for new code
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

// Deprecated do not use for new code
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
