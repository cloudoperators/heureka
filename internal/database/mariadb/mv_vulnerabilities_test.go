// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/common"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		remediation.IssueId = seedCollection.IssueRows[0].Id
		remediation.Issue = seedCollection.IssueRows[0].PrimaryName
		remediation.ExpirationDate = sql.NullTime{Time: expirationDate, Valid: true}
		r := remediation.AsRemediation()
		newRemediation, err := db.CreateRemediation(&r)
		Expect(err).To(BeNil())
		return newRemediation
	}

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
		seedCollection, err = seeder.SeedForIssueCounts()
		Expect(err).To(BeNil())
		err = seeder.RefreshCountIssueRatings()
		Expect(err).To(BeNil())
	})
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	When("there are no remediations", Label("NoRemediations"), func() {
		It("returns the correct count for all issues", func() {
			severityCounts, err := test.LoadIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_severity.json"))
			Expect(err).To(BeNil())
			testIssueSeverityCount(nil, severityCounts)
		})
		It("returns the correct count for component version issues", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			Expect(err).To(BeNil())

			testComponentVersions(severityCounts)
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
	})
	When("there is an expired remediation for a service", Label("WithRemediations"), func() {
		BeforeEach(func() {
			expirationDate := time.Now().Add(-10 * 24 * time.Hour)
			insertRemediation(&seedCollection.ServiceRows[0], nil, expirationDate)
			seeder.RefreshCountIssueRatings()

		})
		It("returns the correct count for component version issues", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			Expect(err).To(BeNil())

			testComponentVersions(severityCounts)
		})
		It("returns the correct count for services", func() {
			seeder.RefreshCountIssueRatings()
			severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
			Expect(err).To(BeNil())

			testServices(severityCounts)
		})
		It("returns the correct count for supportgroup", func() {
			severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
			Expect(err).To(BeNil())

			testSupportGroups(severityCounts)
		})
	})
	When("there is an active remediation for a service", Label("WithRemediations"), func() {
		var serviceCounts entity.IssueSeverityCounts
		var serviceId string
		var remediation *entity.Remediation
		BeforeEach(func() {
			expirationDate := time.Now().Add(10 * 24 * time.Hour)
			remediation = insertRemediation(&seedCollection.ServiceRows[0], nil, expirationDate)
			seeder.RefreshCountIssueRatings()
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
			seeder.RefreshCountIssueRatings()

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
			seeder.RefreshCountIssueRatings()
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
	})
	When("there is a deleted remediation for a service", Label("WithRemediations"), func() {
		BeforeEach(func() {
			expirationDate := time.Now().Add(10 * 24 * time.Hour)
			createdRemediation := insertRemediation(&seedCollection.ServiceRows[0], nil, expirationDate)
			err := db.DeleteRemediation(createdRemediation.Id, common.SystemUserId)
			Expect(err).To(BeNil())
			err = seeder.RefreshCountIssueRatings()
			Expect(err).To(BeNil())
		})
		It("returns the correct count for component version issues", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			Expect(err).To(BeNil())

			testComponentVersions(severityCounts)
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
	})
	When("there is an active remediation for a component in a service", Label("WithRemediations"), func() {
		var serviceCounts entity.IssueSeverityCounts
		var serviceId string
		BeforeEach(func() {
			expirationDate := time.Now().Add(10 * 24 * time.Hour)
			r := insertRemediation(&seedCollection.ServiceRows[0], &seedCollection.ComponentRows[0], expirationDate)
			seeder.RefreshCountIssueRatings()
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
			seeder.RefreshCountIssueRatings()
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
	})

})
