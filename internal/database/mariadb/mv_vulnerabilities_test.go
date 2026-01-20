// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Counting Issues by Severity", Label("IssueCounts"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	var seedCollection *test.SeedCollection

	var testIssueSeverityCount = func(filter *entity.IssueFilter, counts entity.IssueSeverityCounts) {
		issueSeverityCounts, err := db.CountIssueRatings(filter)

		By("throwing no error", func() {
			Expect(err).To(BeNil())
		})

		By("returning the correct counts", func() {
			Expect(issueSeverityCounts.Critical).To(BeEquivalentTo(counts.Critical))
			Expect(issueSeverityCounts.High).To(BeEquivalentTo(counts.High))
			Expect(issueSeverityCounts.Medium).To(BeEquivalentTo(counts.Medium))
			Expect(issueSeverityCounts.Low).To(BeEquivalentTo(counts.Low))
			Expect(issueSeverityCounts.None).To(BeEquivalentTo(counts.None))
			Expect(issueSeverityCounts.Total).To(BeEquivalentTo(counts.Total))
		})
	}

	var testServices = func(counts map[string]entity.IssueSeverityCounts) {
		for _, service := range seedCollection.ServiceRows {
			serviceId := service.Id.Int64

			filter := &entity.IssueFilter{
				ServiceId: []*int64{&serviceId},
			}

			strId := fmt.Sprintf("%d", serviceId)

			testIssueSeverityCount(filter, counts[strId])
		}
	}

	var testComponentVersions = func(counts map[string]entity.IssueSeverityCounts) {
		for _, cvi := range seedCollection.ComponentVersionIssueRows {
			cvId := cvi.ComponentVersionId.Int64
			filter := &entity.IssueFilter{
				ComponentVersionId: []*int64{&cvId},
			}

			strId := fmt.Sprintf("%d", cvId)

			testIssueSeverityCount(filter, counts[strId])
		}
	}

	var testSupportGroups = func(counts map[string]entity.IssueSeverityCounts) {
		for _, sg := range seedCollection.SupportGroupRows {
			filter := &entity.IssueFilter{
				SupportGroupCCRN: []*string{&sg.CCRN.String},
			}

			strId := fmt.Sprintf("%d", sg.Id.Int64)

			testIssueSeverityCount(filter, counts[strId])
		}
	}

	var testServicesTotalCountWithSupportGroup = func(counts map[string]entity.IssueSeverityCounts) {
		for _, sg := range seedCollection.SupportGroupRows {
			filter := &entity.IssueFilter{
				SupportGroupCCRN: []*string{&sg.CCRN.String},
				AllServices:      true,
			}

			strId := fmt.Sprintf("%d", sg.Id.Int64)

			testIssueSeverityCount(filter, counts[strId])
		}
	}

	var testServicesTotalCountWithUnique = func(counts entity.IssueSeverityCounts) {
		filter := &entity.IssueFilter{
			Unique:      true,
			AllServices: true,
		}

		testIssueSeverityCount(filter, counts)
	}

	var testServicesTotalCount = func(counts entity.IssueSeverityCounts) {
		filter := &entity.IssueFilter{
			AllServices: true,
		}

		testIssueSeverityCount(filter, counts)
	}

	var insertRemediation = func(serviceRow *mariadb.BaseServiceRow, componentRow *mariadb.ComponentRow, expirationDate time.Time) *entity.Remediation {
		remediation := test.NewFakeRemediation()
		if serviceRow != nil {
			remediation.ServiceId = serviceRow.Id
			remediation.Service = serviceRow.CCRN
		}
		if componentRow != nil {
			remediation.ComponentId = componentRow.Id
			remediation.Component = componentRow.CCRN
		}
		remediation.IssueId = sql.NullInt64{Int64: 1, Valid: true}
		remediation.Issue = seedCollection.IssueRows[0].PrimaryName
		remediation.ExpirationDate = sql.NullTime{Time: expirationDate, Valid: true}
		r := remediation.AsRemediation()
		newRemediation, err := db.CreateRemediation(&r)
		Expect(err).To(BeNil())
		return newRemediation
	}

	var testNoActiveRemediation = func() {
		It("returns the correct count for component version issues", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			Expect(err).To(BeNil())

			testComponentVersions(severityCounts)
		})
		It("returns the correct count for component version issues with service ccrn filter", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			Expect(err).To(BeNil())
			for _, cvi := range seedCollection.ComponentVersionIssueRows {
				cvId := cvi.ComponentVersionId.Int64
				filter := &entity.IssueFilter{
					ComponentVersionId: []*int64{&cvId},
				}

				serviceIds := lo.FilterMap(seedCollection.ComponentInstanceRows, func(s mariadb.ComponentInstanceRow, _ int) (*int64, bool) {
					if s.ComponentVersionId.Int64 == cvId {
						return &s.ServiceId.Int64, true
					}
					return nil, false
				})

				serviceCcrns := lo.FilterMap(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, _ int) (*string, bool) {
					if lo.Contains(serviceIds, &s.Id.Int64) {
						return &s.CCRN.String, true
					}
					return nil, false
				})

				filter.ServiceCCRN = serviceCcrns

				strId := fmt.Sprintf("%d", cvId)

				testIssueSeverityCount(filter, severityCounts[strId])
			}
		})
		It("returns the correct count for services", func() {
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())

			testServices(severityCounts)
		})
		It("returns the correct count for supportgroup", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			Expect(err).To(BeNil())

			testSupportGroups(severityCounts)
		})
		It("return the total count for all services with support group filter", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			Expect(err).To(BeNil())

			testServicesTotalCountWithSupportGroup(severityCounts)
		})
		It("return the total count for all services unique filter", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			Expect(err).To(BeNil())
			totalCounts := entity.IssueSeverityCounts{}
			for _, count := range severityCounts {
				totalCounts.Critical += count.Critical
				totalCounts.High += count.High
				totalCounts.Medium += count.Medium
				totalCounts.Low += count.Low
				totalCounts.None += count.None
				totalCounts.Total += count.Total
			}

			iv := test.NewFakeIssueVariant(seedCollection.IssueRepositoryRows, seedCollection.IssueRows)
			seeder.InsertFakeIssueVariant(iv)

			testServicesTotalCountWithUnique(totalCounts)
		})
		It("return the total count for all services without support group filter", func() {
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())

			totalCount := entity.IssueSeverityCounts{}
			for _, count := range severityCounts {
				totalCount.Critical += count.Critical
				totalCount.High += count.High
				totalCount.Medium += count.Medium
				totalCount.Low += count.Low
				totalCount.None += count.None
				totalCount.Total += count.Total
			}

			testServicesTotalCount(totalCount)
		})
		It("can filter by service ccrn", func() {
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())

			testServices(severityCounts)
		})
		It("can filter by support group ccrn and service ccrn", func() {
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())

			for _, service := range seedCollection.ServiceRows {
				serviceId := service.Id.Int64
				sgId, found := lo.Find(seedCollection.SupportGroupServiceRows, func(sgs mariadb.SupportGroupServiceRow) bool {
					return sgs.ServiceId.Int64 == serviceId
				})

				Expect(found).To(BeTrue(), "Support group for service should be found")

				sg, found := lo.Find(seedCollection.SupportGroupRows, func(sg mariadb.SupportGroupRow) bool {
					return sg.Id.Int64 == sgId.SupportGroupId.Int64
				})

				Expect(found).To(BeTrue(), "Support group should be found")

				filter := &entity.IssueFilter{
					ServiceCCRN:      []*string{&service.CCRN.String},
					SupportGroupCCRN: []*string{&sg.CCRN.String},
				}

				strId := fmt.Sprintf("%d", serviceId)

				testIssueSeverityCount(filter, severityCounts[strId])
			}
		})
	}

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
		seedCollection, err = seeder.SeedForIssueCounts()
		Expect(err).To(BeNil())
		Expect(seeder.RefreshCountIssueRatings()).To(BeNil())
	})
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	When("there are no remediations", Label("NoRemediations"), func() {
		testNoActiveRemediation()
	})

	When("there is an expired remediation for a service", Label("WithRemediations"), func() {
		BeforeEach(func() {
			expirationDate := time.Now().Add(-10 * 24 * time.Hour)
			insertRemediation(&seedCollection.ServiceRows[0], nil, expirationDate)
			Expect(seeder.RefreshCountIssueRatings()).To(BeNil())
		})
		testNoActiveRemediation()
	})
	When("there is a deleted remediation for a service", Label("WithRemediations"), func() {
		BeforeEach(func() {
			expirationDate := time.Now().Add(10 * 24 * time.Hour)
			createdRemediation := insertRemediation(&seedCollection.ServiceRows[0], nil, expirationDate)
			err := db.DeleteRemediation(createdRemediation.Id, util.SystemUserId)
			Expect(err).To(BeNil())
			Expect(seeder.RefreshCountIssueRatings()).To(BeNil())
			Expect(err).To(BeNil())
		})
		testNoActiveRemediation()
	})
	When("there is an active remediation for a service", Label("WithRemediations"), func() {
		var serviceCounts entity.IssueSeverityCounts
		var serviceId string
		var remediation *entity.Remediation
		BeforeEach(func() {
			expirationDate := time.Now().Add(10 * 24 * time.Hour)
			remediation = insertRemediation(&seedCollection.ServiceRows[0], nil, expirationDate)
			Expect(seeder.RefreshCountIssueRatings()).To(BeNil())
			// remediation for previously critical issue
			serviceCounts = entity.IssueSeverityCounts{
				Critical: 0,
				High:     0,
				Medium:   0,
				Low:      1,
				None:     0,
				Total:    1,
			}
			serviceId = fmt.Sprintf("%d", remediation.ServiceId)
		})
		It("returns the correct count for component version issues", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			severityCounts["1"] = entity.IssueSeverityCounts{}
			Expect(err).To(BeNil())

			testComponentVersions(severityCounts)
		})
		It("return the correct count for component version used in two services", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			Expect(err).To(BeNil())

			cv := seedCollection.ComponentVersionRows[0]
			// create new component instance with component version that has been remediated in another service
			ci := test.NewFakeComponentInstance()
			ci.ComponentVersionId = seedCollection.ComponentVersionRows[0].Id
			ci.ServiceId = seedCollection.ServiceRows[3].Id
			componentInstance := ci.AsComponentInstance()
			newCi, err := db.CreateComponentInstance(&componentInstance)
			Expect(err).To(BeNil())
			Expect(seeder.RefreshCountIssueRatings()).To(BeNil())

			counts, err := db.CountIssueRatings(&entity.IssueFilter{
				ComponentVersionId: []*int64{&cv.Id.Int64},
				ServiceId:          []*int64{&remediation.ServiceId},
			})

			Expect(err).To(BeNil())

			Expect(counts.Critical).To(BeEquivalentTo(0))
			Expect(counts.High).To(BeEquivalentTo(0))
			Expect(counts.Medium).To(BeEquivalentTo(0))
			Expect(counts.Low).To(BeEquivalentTo(0))
			Expect(counts.None).To(BeEquivalentTo(0))
			Expect(counts.Total).To(BeEquivalentTo(0))

			countsEmpty, err := db.CountIssueRatings(&entity.IssueFilter{
				ComponentVersionId: []*int64{&cv.Id.Int64},
				ServiceId:          []*int64{&newCi.ServiceId},
			})
			Expect(err).To(BeNil())

			cvId := fmt.Sprintf("%d", cv.Id.Int64)
			Expect(countsEmpty.Critical).To(BeEquivalentTo(severityCounts[cvId].Critical))
			Expect(countsEmpty.High).To(BeEquivalentTo(severityCounts[cvId].High))
			Expect(countsEmpty.Medium).To(BeEquivalentTo(severityCounts[cvId].Medium))
			Expect(countsEmpty.Low).To(BeEquivalentTo(severityCounts[cvId].Low))
			Expect(countsEmpty.None).To(BeEquivalentTo(severityCounts[cvId].None))
			Expect(countsEmpty.Total).To(BeEquivalentTo(severityCounts[cvId].Total))
		})
		It("returns the correct count for services", func() {
			Expect(seeder.RefreshCountIssueRatings()).To(BeNil())
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())
			severityCounts[serviceId] = serviceCounts

			testServices(severityCounts)
		})
		It("returns the correct count for supportgroup", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			severityCounts["1"] = entity.IssueSeverityCounts{
				Critical: 1,
				Medium:   1,
				Low:      2,
				None:     1,
				Total:    5,
			}
			Expect(err).To(BeNil())

			testSupportGroups(severityCounts)
		})
		It("return the total count for all services with support group filter", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			Expect(err).To(BeNil())
			severityCounts["1"] = entity.IssueSeverityCounts{
				Critical: 1,
				Medium:   1,
				Low:      2,
				None:     1,
				Total:    5,
			}

			testServicesTotalCountWithSupportGroup(severityCounts)
		})
		It("return the total count for all services without support group filter", func() {
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())

			totalCount := entity.IssueSeverityCounts{}
			for _, count := range severityCounts {
				totalCount.Critical += count.Critical
				totalCount.High += count.High
				totalCount.Medium += count.Medium
				totalCount.Low += count.Low
				totalCount.None += count.None
				totalCount.Total += count.Total
			}
			// remediation for one critical issue
			totalCount.Critical -= 1
			totalCount.Total -= 1

			testServicesTotalCount(totalCount)
		})
	})
	When("there is an active remediation for a component in a service", Label("WithRemediations"), func() {
		var serviceCounts entity.IssueSeverityCounts
		var serviceId string
		BeforeEach(func() {
			expirationDate := time.Now().Add(10 * 24 * time.Hour)
			r := insertRemediation(&seedCollection.ServiceRows[0], &seedCollection.ComponentRows[0], expirationDate)
			Expect(seeder.RefreshCountIssueRatings()).To(BeNil())
			// remediation for previously critical issue
			serviceCounts = entity.IssueSeverityCounts{
				Critical: 0,
				High:     0,
				Medium:   0,
				Low:      1,
				None:     0,
				Total:    1,
			}
			serviceId = fmt.Sprintf("%d", r.ServiceId)
		})
		It("returns the correct count for component version issues", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			severityCounts["1"] = entity.IssueSeverityCounts{}
			Expect(err).To(BeNil())

			testComponentVersions(severityCounts)
		})
		It("returns the correct count for services", func() {
			Expect(seeder.RefreshCountIssueRatings()).To(BeNil())
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())
			severityCounts[serviceId] = serviceCounts

			testServices(severityCounts)
		})
		It("returns the correct count for supportgroup", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			severityCounts["1"] = entity.IssueSeverityCounts{
				Critical: 1,
				Medium:   1,
				Low:      2,
				None:     1,
				Total:    5,
			}
			Expect(err).To(BeNil())

			testSupportGroups(severityCounts)
		})
		It("return the total count for all services with support group filter", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			Expect(err).To(BeNil())
			severityCounts["1"] = entity.IssueSeverityCounts{
				Critical: 1,
				Medium:   1,
				Low:      2,
				None:     1,
				Total:    5,
			}

			testServicesTotalCountWithSupportGroup(severityCounts)
		})
		It("return the total count for all services without support group filter", func() {
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())

			totalCount := entity.IssueSeverityCounts{}
			for _, count := range severityCounts {
				totalCount.Critical += count.Critical
				totalCount.High += count.High
				totalCount.Medium += count.Medium
				totalCount.Low += count.Low
				totalCount.None += count.None
				totalCount.Total += count.Total
			}
			// remediation for one critical issue
			totalCount.Critical -= 1
			totalCount.Total -= 1

			testServicesTotalCount(totalCount)
		})
	})
	When("there is an active remediation for a vulnerability only in one service", Label("WithRemediations"), func() {
		BeforeEach(func() {
			expirationDate := time.Now().Add(10 * 24 * time.Hour)
			r := insertRemediation(&seedCollection.ServiceRows[0], &seedCollection.ComponentRows[0], expirationDate)
			im := test.NewFakeIssueMatch()
			im.ComponentInstanceId = sql.NullInt64{Int64: 3, Valid: true}
			im.IssueId = sql.NullInt64{Int64: r.IssueId, Valid: true}
			im.UserId = sql.NullInt64{Int64: util.SystemUserId, Valid: true}
			_, err := seeder.InsertFakeIssueMatch(im)
			Expect(err).To(BeNil())
			Expect(seeder.RefreshCountIssueRatings()).To(BeNil())
		})
		It("returns the total count for all services with support group filter", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			Expect(err).To(BeNil())

			testServicesTotalCountWithSupportGroup(severityCounts)
		})
	})

})
