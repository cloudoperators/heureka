// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/gomega"
)

// Temporary used until order is used in all entities
func TestPaginationOfListWithOrder[F entity.HeurekaFilter, E entity.HeurekaEntity](
	listFunction func(*F, []entity.Order) ([]E, error),
	filterFunction func(*int, *int64) *F,
	order []entity.Order,
	getAfterFunction func([]E) *int64,
	elementCount int,
	pageSize int,
) {
	quotient, remainder := elementCount/pageSize, elementCount%pageSize
	expectedPages := quotient
	if remainder > 0 {
		expectedPages = expectedPages + 1
	}

	var after *int64
	for i := expectedPages; i > 0; i-- {
		entries, err := listFunction(filterFunction(&pageSize, after), order)

		Expect(err).To(BeNil())

		if i == 1 && remainder > 0 {
			Expect(len(entries)).To(BeEquivalentTo(remainder), "on the last page we expect")
		} else {
			if pageSize > elementCount {
				Expect(len(entries)).To(BeEquivalentTo(elementCount), "on a page with a higher pageSize then element count we expect")
			} else {
				Expect(len(entries)).To(BeEquivalentTo(pageSize), "on a normal page we expect the element count to be equal to the page size")

			}
		}
		after = getAfterFunction(entries)

	}
}

func TestPaginationOfList[F entity.HeurekaFilter, E entity.HeurekaEntity](
	listFunction func(*F) ([]E, error),
	filterFunction func(*int, *int64) *F,
	getAfterFunction func([]E) *int64,
	elementCount int,
	pageSize int,
) {
	quotient, remainder := elementCount/pageSize, elementCount%pageSize
	expectedPages := quotient
	if remainder > 0 {
		expectedPages = expectedPages + 1
	}

	var after *int64
	for i := expectedPages; i > 0; i-- {
		entries, err := listFunction(filterFunction(&pageSize, after))

		Expect(err).To(BeNil())

		if i == 1 && remainder > 0 {
			Expect(len(entries)).To(BeEquivalentTo(remainder), "on the last page we expect")
		} else {
			if pageSize > elementCount {
				Expect(len(entries)).To(BeEquivalentTo(elementCount), "on a page with a higher pageSize then element count we expect")
			} else {
				Expect(len(entries)).To(BeEquivalentTo(pageSize), "on a normal page we expect the element count to be equal to the page size")

			}
		}
		after = getAfterFunction(entries)

	}
}

// DB stores rating as enum
// entity.Severity.Score is based on CVSS vector and has a range between x and y
// This means a rating "Low" can have a Score 3.1, 3.3, ...
// Ordering is done based on enum on DB layer, so Score can't be used for checking order
// and needs a numerical translation
func SeverityToNumerical(s string) int {
	rating := map[string]int{
		"None":     0,
		"Low":      1,
		"Medium":   2,
		"High":     3,
		"Critical": 4,
	}
	if val, ok := rating[s]; ok {
		return val
	} else {
		return -1
	}
}
