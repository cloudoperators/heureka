// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	. "github.com/onsi/gomega"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

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
