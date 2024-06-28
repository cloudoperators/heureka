// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.wdf.sap.corp/cc/heureka/internal/app"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/internal/entity/test"
	"github.wdf.sap.corp/cc/heureka/internal/mocks"
)

func evidenceFilter() *entity.EvidenceFilter {
	return &entity.EvidenceFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
	}
}

func evidenceListOptions() *entity.ListOptions {
	return &entity.ListOptions{
		ShowTotalCount:      false,
		ShowPageInfo:        false,
		IncludeAggregations: false,
	}
}

var _ = Describe("When listing Evidences", Label("app", "ListEvidences"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		filter  *entity.EvidenceFilter
		options *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = evidenceListOptions()
		filter = evidenceFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetEvidences", filter).Return([]entity.Evidence{}, nil)
			db.On("CountEvidences", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			heureka = app.NewHeurekaApp(db)
			res, err := heureka.ListEvidences(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.TotalCount).Should(BeEquivalentTo(int64(1337)), "return correct Totalcount")
		})
	})

	When("the list option does include the PageInfo", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
		})
		DescribeTable("pagination information is correct", func(pageSize int, dbElements int, resElements int, hasNextPage bool) {
			filter.First = &pageSize
			evidences := test.NNewFakeEvidences(resElements)

			var ids = lo.Map(evidences, func(e entity.Evidence, _ int) int64 { return e.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetEvidences", filter).Return(evidences, nil)
			db.On("GetAllEvidenceIds", filter).Return(ids, nil)
			heureka = app.NewHeurekaApp(db)
			res, err := heureka.ListEvidences(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.PageInfo.HasNextPage).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
			Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
			Expect(len(res.PageInfo.Pages)).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct  number of pages")
		},
			Entry("When  pageSize is 1 and the database was returning 2 elements", 1, 2, 1, true),
			Entry("When  pageSize is 10 and the database was returning 9 elements", 10, 9, 9, false),
			Entry("When  pageSize is 10 and the database was returning 11 elements", 10, 11, 10, true),
		)
	})
})

var _ = Describe("When creating Evidence", Label("app", "CreateEvidence"), func() {
	var (
		db       *mocks.MockDatabase
		heureka  app.Heureka
		evidence entity.Evidence
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		evidence = test.NewFakeEvidenceEntity()
	})

	It("creates evidence", func() {
		db.On("CreateEvidence", &evidence).Return(&evidence, nil)
		heureka = app.NewHeurekaApp(db)
		newEvidence, err := heureka.CreateEvidence(&evidence)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newEvidence.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newEvidence.Description).To(BeEquivalentTo(evidence.Description))
			Expect(newEvidence.UserId).To(BeEquivalentTo(evidence.UserId))
			Expect(newEvidence.ActivityId).To(BeEquivalentTo(evidence.ActivityId))
			Expect(newEvidence.Type).To(BeEquivalentTo(evidence.Type))
			Expect(newEvidence.Severity.Cvss.Vector).To(BeEquivalentTo(evidence.Severity.Cvss.Vector))
			Expect(newEvidence.RaaEnd.Unix()).To(BeEquivalentTo(evidence.RaaEnd.Unix()))
		})
	})
})

var _ = Describe("When updating Evidence", Label("app", "UpdateEvidence"), func() {
	var (
		db       *mocks.MockDatabase
		heureka  app.Heureka
		evidence entity.Evidence
		filter   *entity.EvidenceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		evidence = test.NewFakeEvidenceEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.EvidenceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates evidence", func() {
		db.On("UpdateEvidence", &evidence).Return(nil)
		heureka = app.NewHeurekaApp(db)
		evidence.Description = "New Description"
		filter.Id = []*int64{&evidence.Id}
		db.On("GetEvidences", filter).Return([]entity.Evidence{evidence}, nil)
		updatedEvidence, err := heureka.UpdateEvidence(&evidence)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedEvidence.Description).To(BeEquivalentTo(evidence.Description))
			Expect(updatedEvidence.UserId).To(BeEquivalentTo(evidence.UserId))
			Expect(updatedEvidence.ActivityId).To(BeEquivalentTo(evidence.ActivityId))
			Expect(updatedEvidence.Type).To(BeEquivalentTo(evidence.Type))
			Expect(updatedEvidence.Severity.Cvss.Vector).To(BeEquivalentTo(evidence.Severity.Cvss.Vector))
			Expect(updatedEvidence.RaaEnd.Unix()).To(BeEquivalentTo(evidence.RaaEnd.Unix()))
		})
	})
})

var _ = Describe("When deleting Evidence", Label("app", "DeleteEvidence"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		id      int64
		filter  *entity.EvidenceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.EvidenceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes evidence", func() {
		db.On("DeleteEvidence", id).Return(nil)
		heureka = app.NewHeurekaApp(db)
		db.On("GetEvidences", filter).Return([]entity.Evidence{}, nil)
		err := heureka.DeleteEvidence(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		evidences, err := heureka.ListEvidences(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(evidences.Elements).To(BeEmpty(), "no error should be thrown")
	})
})
