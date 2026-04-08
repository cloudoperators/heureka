// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/cloudoperators/heureka/internal/entity"
)

func EnsurePaginated(filter *entity.Paginated) {
	if filter.First == nil {
		first := 10
		filter.First = &first
	}

	if filter.After == nil {
		after := ""

		filter.After = &after
	}
}

func GetPages(firstCursor *string, cursors []string, pageSize int) ([]entity.Page, entity.Page) {
	var (
		pages       []entity.Page
		currentPage entity.Page
	)

	i := 0
	pN := 0

	var isCurrent bool
	// declare variable so it can be used outside the loop
	var c string
	for _, c = range cursors {
		i++
		if i == 1 {
			pN++
			isCurrent = false
		}

		if c == *firstCursor {
			isCurrent = true
		}

		if i >= pageSize {
			pages = append(pages, entity.Page{
				After:      new(c),
				IsCurrent:  isCurrent,
				PageCount:  new(i),
				PageNumber: new(pN),
			})

			i = 0

			if isCurrent {
				currentPage = pages[len(pages)-1]
			}
		}
	}

	if len(cursors)%pageSize != 0 {
		pages = append(pages, entity.Page{
			After:      new(c),
			IsCurrent:  isCurrent,
			PageCount:  new(i),
			PageNumber: new(pN),
		})
		if isCurrent {
			currentPage = pages[len(pages)-1]
		}
	}

	return pages, currentPage
}

func GetPageInfo[T entity.HasCursor](
	res []T,
	cursors []string,
	pageSize int,
	currentCursor *string,
) *entity.PageInfo {
	var nextPageAfter *string

	firstCursor := res[0].Cursor()

	if len(res) > 1 {
		nextPageAfter = res[len(res)-1].Cursor()
	} else {
		nextPageAfter = firstCursor
	}

	pages, currentPage := GetPages(firstCursor, cursors, pageSize)

	return &entity.PageInfo{
		HasNextPage:     new(currentPage.PageNumber != nil && *currentPage.PageNumber < len(pages)),
		HasPreviousPage: new(currentPage.PageNumber != nil && *currentPage.PageNumber > 1),
		IsValidPage: new(
			currentPage.After != nil && currentCursor != nil &&
				*currentPage.After == *currentCursor,
		),
		PageNumber:    currentPage.PageNumber,
		NextPageAfter: nextPageAfter,
		Pages:         pages,
	}
}
