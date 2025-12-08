// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	pkg_util "github.com/cloudoperators/heureka/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"math/rand"
)

var _ = Describe("Issue", Label("database", "Issue"), func() {

	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {

		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	When("Getting All Issue IDs", Label("GetAllIssueIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllIssueIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, issue := range seedCollection.IssueRows {
					ids = append(ids, issue.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllIssueIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueRows)))
					})

					By("returning the correct order", func() {
						var prev int64 = 0
						for _, r := range res {

							Expect(r > prev).Should(BeTrue())
							prev = r

						}
					})

					By("returning the correct fields", func() {
						for _, r := range res {
							Expect(lo.Contains(ids, r)).To(BeTrue())
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single issue id that does exist", func() {
					issueId := ids[rand.Intn(len(ids))]
					filter := &entity.IssueFilter{
						Id: []*int64{&issueId},
					}

					entries, err := db.GetAllIssueIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(issueId))
					})
				})
			})
		})
	})

	When("Getting Issues", Label("GetIssues"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetIssues(nil, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 issues in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {

				It("can fetch the items correctly", func() {
					res, err := db.GetIssues(nil, nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueRows)))
					})

					By("returning the correct order", func() {
						var prev int64 = 0
						for _, r := range res {

							Expect(r.Issue.Id > prev).Should(BeTrue())
							prev = r.Issue.Id

						}
					})

					By("returning the correct fields", func() {
						for _, r := range res {
							for _, row := range seedCollection.IssueRows {
								if r.Issue.Id == row.Id.Int64 {
									Expect(r.Issue.PrimaryName).Should(BeEquivalentTo(row.PrimaryName.String), "Name should match")
									Expect(r.Issue.Type).Should(BeEquivalentTo(row.Type.String), "Type should match")
									Expect(r.Issue.Description).Should(BeEquivalentTo(row.Description.String), "Description should match")
									Expect(r.Issue.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.Issue.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single service name", func() {
					var row mariadb.BaseServiceRow
					searchingRow := true
					var issueRows []mariadb.IssueRow

					//get a service that should return at least 1 issue
					for searchingRow {
						row = seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]
						issueRows = seedCollection.GetIssueByService(&row)
						searchingRow = len(issueRows) == 0
					}
					filter := &entity.IssueFilter{ServiceCCRN: []*string{&row.CCRN.String}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(issueRows)))
					})
				})
				It("can filter a non existing service name", func() {
					nonExistingName := pkg_util.GenerateRandomString(40, nil)
					filter := &entity.IssueFilter{ServiceCCRN: []*string{&nonExistingName}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning no results", func() {
						Expect(entries).To(BeEmpty())
					})
				})
				It("can filter by multiple existing service names", func() {
					serviceCcrns := make([]*string, len(seedCollection.ServiceRows))
					var expectedIssues []mariadb.IssueRow
					for i, row := range seedCollection.ServiceRows {
						x := row.CCRN.String
						expectedIssues = append(expectedIssues, seedCollection.GetIssueByService(&row)...)
						serviceCcrns[i] = &x
					}
					expectedIssues = lo.Uniq(expectedIssues)
					filter := &entity.IssueFilter{ServiceCCRN: serviceCcrns}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(expectedIssues)))
					})

				})
				It("can filter by a single issue Id", func() {
					row := seedCollection.IssueRows[rand.Intn(len(seedCollection.IssueRows))]
					filter := &entity.IssueFilter{Id: []*int64{&row.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning exactly 1 element", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning the expected element", func() {
						Expect(entries[0].Issue.Id).To(BeEquivalentTo(row.Id.Int64))
					})
				})
				It("can filter by a single service Id", func() {
					serviceRow := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]
					ciIds := lo.FilterMap(seedCollection.ComponentInstanceRows, func(c mariadb.ComponentInstanceRow, _ int) (int64, bool) {
						return c.Id.Int64, serviceRow.Id.Int64 == c.ServiceId.Int64
					})
					issueIds := lo.FilterMap(seedCollection.IssueMatchRows, func(im mariadb.IssueMatchRow, _ int) (int64, bool) {
						return im.IssueId.Int64, lo.Contains(ciIds, im.ComponentInstanceId.Int64)
					})

					filter := &entity.IssueFilter{ServiceId: []*int64{&serviceRow.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected element", func() {
						for _, entry := range entries {
							Expect(lo.Contains(issueIds, entry.Issue.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single support group ccrn", func() {
					sgRow := seedCollection.SupportGroupRows[rand.Intn(len(seedCollection.SupportGroupRows))]
					serviceIds := lo.FilterMap(seedCollection.SupportGroupServiceRows, func(sgs mariadb.SupportGroupServiceRow, _ int) (int64, bool) {
						return sgs.ServiceId.Int64, sgRow.Id.Int64 == sgs.SupportGroupId.Int64
					})
					ciIds := lo.FilterMap(seedCollection.ComponentInstanceRows, func(c mariadb.ComponentInstanceRow, _ int) (int64, bool) {
						return c.Id.Int64, lo.Contains(serviceIds, c.ServiceId.Int64)
					})
					issueIds := lo.FilterMap(seedCollection.IssueMatchRows, func(im mariadb.IssueMatchRow, _ int) (int64, bool) {
						return im.IssueId.Int64, lo.Contains(ciIds, im.ComponentInstanceId.Int64)
					})

					filter := &entity.IssueFilter{SupportGroupCCRN: []*string{&sgRow.CCRN.String}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected element", func() {
						for _, entry := range entries {
							Expect(lo.Contains(issueIds, entry.Issue.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single activity id", func() {
					// select an activity
					activityRow := seedCollection.ActivityRows[rand.Intn(len(seedCollection.ActivityRows))]

					// collect all issue ids that belong to the activity
					issueIds := []int64{}
					for _, ahiRow := range seedCollection.ActivityHasIssueRows {
						if ahiRow.ActivityId.Int64 == activityRow.Id.Int64 {
							issueIds = append(issueIds, ahiRow.IssueId.Int64)
						}
					}

					filter := &entity.IssueFilter{ActivityId: []*int64{&activityRow.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						for _, entry := range entries {
							Expect(issueIds).To(ContainElement(entry.Issue.Id))
						}
					})
				})
				It("can filter by a single component version id", func() {
					// select a componentVersion
					cvRow := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]

					// collect all issue ids that belong to the component version
					issueIds := []int64{}
					for _, cvvRow := range seedCollection.ComponentVersionIssueRows {
						if cvvRow.ComponentVersionId.Int64 == cvRow.Id.Int64 {
							issueIds = append(issueIds, cvvRow.IssueId.Int64)
						}
					}

					filter := &entity.IssueFilter{ComponentVersionId: []*int64{&cvRow.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						for _, entry := range entries {
							Expect(issueIds).To(ContainElement(entry.Issue.Id))
						}
					})
				})
				It("can filter by a single component id", func() {
					// select a component
					cRow := seedCollection.ComponentRows[rand.Intn(len(seedCollection.ComponentRows))]

					// collect all componentVersion ids that belong to the component
					cvIds := []int64{}
					for _, cvRow := range seedCollection.ComponentVersionRows {
						if cvRow.ComponentId.Int64 == cRow.Id.Int64 {
							cvIds = append(cvIds, cvRow.Id.Int64)
						}
					}

					// collect all issue ids that belong to the component version ids
					issueIds := []int64{}
					for _, cviRow := range seedCollection.ComponentVersionIssueRows {
						if lo.Contains(cvIds, cviRow.ComponentVersionId.Int64) {
							issueIds = append(issueIds, cviRow.IssueId.Int64)
						}
					}

					filter := &entity.IssueFilter{ComponentId: []*int64{&cRow.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						for _, entry := range entries {
							Expect(issueIds).To(ContainElement(entry.Issue.Id))
						}
					})
				})
				It("can filter by a single issueVariant id", func() {
					// select an issueVariant
					issueVariantRow := seedCollection.IssueVariantRows[rand.Intn(len(seedCollection.IssueVariantRows))]

					filter := &entity.IssueFilter{IssueVariantId: []*int64{&issueVariantRow.Id.Int64}}

					entries, err := db.GetIssues(filter, nil)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Issue.Id)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						Expect(issueIds).To(ContainElement(issueVariantRow.IssueId.Int64))
					})
				})
				It("can filter by a issueType", func() {
					issueType := entity.AllIssueTypes[rand.Intn(len(entity.AllIssueTypes))]

					filter := &entity.IssueFilter{Type: []*string{&issueType}}

					entries, err := db.GetIssues(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					for _, entry := range entries {
						Expect(entry.Type).To(BeEquivalentTo(issueType))
					}
				})
				It("can filter by hasIssueMatches", func() {
					filter := &entity.IssueFilter{HasIssueMatches: true}

					entries, err := db.GetIssues(filter, nil)

					Expect(err).To(BeNil())
					for _, entry := range entries {
						hasMatch := lo.ContainsBy(seedCollection.IssueMatchRows, func(im mariadb.IssueMatchRow) bool {
							return im.IssueId.Int64 == entry.Issue.Id
						})
						Expect(hasMatch).To(BeTrue(), "Entry should have at least one matching IssueMatchRow")
					}

				})
				It("can filter by issueMatch severity", func() {
					for _, severity := range entity.AllSeverityValues {
						issueIds := lo.FilterMap(seedCollection.IssueMatchRows, func(im mariadb.IssueMatchRow, _ int) (int64, bool) {
							return im.IssueId.Int64, im.Rating.String == severity.String()
						})

						filter := &entity.IssueFilter{IssueMatchSeverity: []*string{lo.ToPtr(severity.String())}}

						entries, err := db.GetIssues(filter, nil)

						Expect(err).To(BeNil())
						for _, entry := range entries {
							Expect(lo.Contains(issueIds, entry.Issue.Id)).To(BeTrue(), "Entry should have severity %s", severity.String())
						}
					}

				})
				It("can filter issue PrimaryName using wild card search", func() {
					row := seedCollection.IssueRows[rand.Intn(len(seedCollection.IssueRows))]

					const charactersToRemoveFromBeginning = 2
					const charactersToRemoveFromEnd = 2
					const minimalCharactersToKeep = 5

					start := charactersToRemoveFromBeginning
					end := len(row.PrimaryName.String) - charactersToRemoveFromEnd

					Expect(start+minimalCharactersToKeep < end).To(BeTrue())

					searchStr := row.PrimaryName.String[start:end]
					filter := &entity.IssueFilter{Search: []*string{&searchStr}}

					entries, err := db.GetIssues(filter, nil)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Issue.Id)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("at least one element was discarded (filtered)", func() {
						Expect(len(seedCollection.IssueRows) > len(issueIds)).To(BeTrue())
					})

					By("returning the expected elements", func() {
						Expect(issueIds).To(ContainElement(row.Id.Int64))
					})
				})
				It("can filter issue variant SecondaryName using wild card search", func() {
					// select an issueVariant
					issueVariantRow := seedCollection.IssueVariantRows[rand.Intn(len(seedCollection.IssueVariantRows))]

					const charactersToRemoveFromBeginning = 2
					const charactersToRemoveFromEnd = 2
					const minimalCharactersToKeep = 5

					start := charactersToRemoveFromBeginning
					end := len(issueVariantRow.SecondaryName.String) - charactersToRemoveFromEnd

					Expect(start+minimalCharactersToKeep < end).To(BeTrue())

					searchStr := issueVariantRow.SecondaryName.String[start:end]
					filter := &entity.IssueFilter{Search: []*string{&searchStr}}

					entries, err := db.GetIssues(filter, nil)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Issue.Id)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						Expect(issueIds).To(ContainElement(issueVariantRow.IssueId.Int64))
					})
				})
			})
			Context("and using pagination", func() {
				DescribeTable("can correctly paginate", func(pageSize int) {
					test.TestPaginationOfListWithOrder(
						db.GetIssues,
						func(first *int, after *int64, afterX *string) *entity.IssueFilter {
							return &entity.IssueFilter{
								PaginatedX: entity.PaginatedX{First: first, After: afterX},
							}
						},
						[]entity.Order{},
						func(entries []entity.IssueResult) string {
							after, _ := mariadb.EncodeCursor(mariadb.WithIssue([]entity.Order{}, *entries[len(entries)-1].Issue, 0))
							return after
						},
						len(seedCollection.IssueRows),
						pageSize,
					)
				},
					Entry("when pageSize is 1", 1),
					Entry("when pageSize is 3", 3),
					Entry("when pageSize is 5", 5),
					Entry("when pageSize is 11", 11),
					Entry("when pageSize is 100", 100),
				)
			})
		})
	})
	When("Getting Issues with Aggregations", Label("GetIssuesWithAggregations"), func() {
		Context("and the database contains service without aggregations", func() {
			BeforeEach(func() {
				newIssueRow := test.NewFakeIssue()
				newIssue := newIssueRow.AsIssue()
				db.CreateIssue(&newIssue)
			})
			It("returns the issues with aggregations", func() {
				entriesWithAggregations, err := db.GetIssuesWithAggregations(nil, nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning some aggregations", func() {
					for _, entryWithAggregations := range entriesWithAggregations {
						Expect(entryWithAggregations).NotTo(
							BeEquivalentTo(entity.IssueAggregations{}))
						Expect(entryWithAggregations.IssueAggregations.Activities).To(BeEquivalentTo(0))
						Expect(entryWithAggregations.IssueAggregations.IssueMatches).To(BeEquivalentTo(0))
						Expect(entryWithAggregations.IssueAggregations.AffectedServices).To(BeEquivalentTo(0))
						Expect(entryWithAggregations.IssueAggregations.AffectedComponentInstances).To(BeEquivalentTo(0))
						Expect(entryWithAggregations.IssueAggregations.ComponentVersions).To(BeEquivalentTo(0))
					}
				})
				By("returning all issues", func() {
					Expect(len(entriesWithAggregations)).To(BeEquivalentTo(1))
				})
			})
		})
		Context("and and we have 10 elements in the database", func() {
			BeforeEach(func() {
				_ = seeder.SeedDbWithNFakeData(10)
			})
			It("returns the issues with aggregations", func() {
				entriesWithAggregations, err := db.GetIssuesWithAggregations(nil, nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning some aggregations", func() {
					for _, entryWithAggregations := range entriesWithAggregations {
						Expect(entryWithAggregations).NotTo(
							BeEquivalentTo(entity.IssueAggregations{}))
					}
				})
				By("returning all ld constraints exclude all Go files inservices", func() {
					Expect(len(entriesWithAggregations)).To(BeEquivalentTo(10))
				})
			})
			It("returns correct aggregation values", func() {
				//Should be filled with a check for each aggregation value,
				// this is currently skipped due to the complexity of the test implementation
				// as we would need to implement for each of the aggregations a manual aggregation
				// based on the seederCollection.
				//
				// This tests should therefore only get implemented in case we encourage errors in this area to test against
				// possible regressions
			})
		})
	})
	When("Counting Issues", Label("CountIssues"), func() {
		Context("and using no filter", func() {
			DescribeTable("it returns correct count", func(x int) {
				_ = seeder.SeedDbWithNFakeData(x)
				res, err := db.CountIssues(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning the correct count", func() {
					Expect(res).To(BeEquivalentTo(x))
				})

			},
				Entry("when page size is 0", 0),
				Entry("when page size is 1", 1),
				Entry("when page size is 11", 11),
				Entry("when page size is 100", 100),
			)
			Context("and counting issue types", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(20)
				})
				It("returns the correct count for each issue type", func() {
					vulnerabilityCount := 0
					policyViolationCount := 0
					securityEventCount := 0

					for _, issue := range seedCollection.IssueRows {
						switch issue.Type.String {
						case entity.IssueTypeVulnerability.String():
							vulnerabilityCount++
						case entity.IssueTypePolicyViolation.String():
							policyViolationCount++
						case entity.IssueTypeSecurityEvent.String():
							securityEventCount++
						}
					}

					issueTypeCounts, err := db.CountIssueTypes(nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct counts", func() {
						Expect(issueTypeCounts.VulnerabilityCount).To(BeEquivalentTo(vulnerabilityCount))
						Expect(issueTypeCounts.PolicyViolationCount).To(BeEquivalentTo(policyViolationCount))
						Expect(issueTypeCounts.SecurityEventCount).To(BeEquivalentTo(securityEventCount))
					})

				})
			})
		})
		Context("and using a filter", func() {
			Context("and having 20 elements in the Database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(20)
				})
				It("does not influence the count when pagination is applied", func() {
					var first = 1
					var after string = ""
					filter := &entity.IssueFilter{
						PaginatedX: entity.PaginatedX{
							First: &first,
							After: &after,
						},
					}
					res, err := db.CountIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(res).To(BeEquivalentTo(20))
					})
				})
				It("does show the correct amount when filtering for a service name", func() {
					var row mariadb.BaseServiceRow
					searchingRow := true
					var issueRows []mariadb.IssueRow

					//get a service that should return at least 1 issue
					for searchingRow {
						row = seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]
						issueRows = seedCollection.GetIssueByService(&row)
						searchingRow = len(issueRows) > 0
					}
					filter := &entity.IssueFilter{ServiceCCRN: []*string{&row.CCRN.String}}

					count, err := db.CountIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(count).To(BeEquivalentTo(len(issueRows)))
					})
				})
				It("does show the correct amount when filtering for a service id", func() {
					var row mariadb.BaseServiceRow
					searchingRow := true
					var issueRows []mariadb.IssueRow

					//get a service that should return at least 1 issue
					for searchingRow {
						row = seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]
						issueRows = seedCollection.GetIssueByService(&row)
						searchingRow = len(issueRows) > 0
					}
					filter := &entity.IssueFilter{ServiceId: []*int64{&row.Id.Int64}}

					count, err := db.CountIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(count).To(BeEquivalentTo(len(issueRows)))
					})
				})
			})
		})
	})
	When("IssueCounts by Severity", Label("IssueCounts"), func() {
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
		Context("and counting issue severities", func() {
			var seedCollection *test.SeedCollection
			var err error
			BeforeEach(func() {
				seedCollection, err = seeder.SeedForIssueCounts()
				Expect(err).To(BeNil())
				err = seeder.RefreshCountIssueRatings()
				Expect(err).To(BeNil())
			})

			It("returns the correct count for all services", func() {
				severityCounts, err := test.LoadIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_severity.json"))
				Expect(err).To(BeNil())

				filter := &entity.IssueFilter{
					AllServices: true,
				}

				testIssueSeverityCount(filter, severityCounts)
			})
			It("returns the correct count for services in support goups", func() {
				severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
				Expect(err).To(BeNil())

				for _, sg := range seedCollection.SupportGroupRows {

					filter := &entity.IssueFilter{
						AllServices:      true,
						SupportGroupCCRN: []*string{&sg.CCRN.String},
					}

					strId := fmt.Sprintf("%d", sg.Id.Int64)

					testIssueSeverityCount(filter, severityCounts[strId])
				}
			})
			It("returns the correct count for component version issues", func() {
				severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
				Expect(err).To(BeNil())

				for _, cv := range seedCollection.ComponentVersionRows {
					filter := &entity.IssueFilter{
						ComponentVersionId: []*int64{&cv.Id.Int64},
					}

					strId := fmt.Sprintf("%d", cv.Id.Int64)

					testIssueSeverityCount(filter, severityCounts[strId])
				}
			})
			It("returns the correct count for services", func() {
				severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_service.json"))
				Expect(err).To(BeNil())

				for _, service := range seedCollection.ServiceRows {
					filter := &entity.IssueFilter{
						ServiceId: []*int64{&service.Id.Int64},
					}

					strId := fmt.Sprintf("%d", service.Id.Int64)

					testIssueSeverityCount(filter, severityCounts[strId])
				}
			})
			It("returns the correct count for supportgroup", func() {
				severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
				Expect(err).To(BeNil())

				for _, sg := range seedCollection.SupportGroupRows {

					filter := &entity.IssueFilter{
						SupportGroupCCRN: []*string{&sg.CCRN.String},
					}

					strId := fmt.Sprintf("%d", sg.Id.Int64)

					testIssueSeverityCount(filter, severityCounts[strId])
				}
			})
			It("returns the correct count for unique filter", Label("ABCDEF"), func() {
				severityCounts, err := test.LoadIssueCounts(test.GetTestDataPath("../mariadb/testdata/issue_counts/issue_counts_per_severity.json"))
				Expect(err).To(BeNil())
				// Create a new IM that attaches an existing issue to a different component instance
				im := test.NewFakeIssueMatch()
				im.ComponentInstanceId = sql.NullInt64{Int64: 1, Valid: true}
				im.IssueId = sql.NullInt64{Int64: 3, Valid: true}
				im.UserId = sql.NullInt64{Int64: util.SystemUserId, Valid: true}
				_, err = seeder.InsertFakeIssueMatch(im)
				Expect(err).To(BeNil())

				filter := &entity.IssueFilter{
					AllServices: true,
					Unique:      true,
				}

				testIssueSeverityCount(filter, severityCounts)
			})
		})
	})
	When("Insert Issue", Label("InsertIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var newIssueRow mariadb.IssueRow
			var newIssue entity.Issue
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newIssueRow = test.NewFakeIssue()
				newIssue = newIssueRow.AsIssue()
			})
			It("can insert correctly", func() {
				issue, err := db.CreateIssue(&newIssue)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets issue id", func() {
					Expect(issue).NotTo(BeEquivalentTo(0))
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issue", func() {
					Expect(len(i)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(i[0].Issue.PrimaryName).To(BeEquivalentTo(issue.PrimaryName))
					Expect(i[0].Issue.Type.String()).To(BeEquivalentTo(issue.Type.String()))
					Expect(i[0].Issue.Description).To(BeEquivalentTo(issue.Description))
				})
			})
			It("does not insert issue with existing primary name", func() {
				issueRow := seedCollection.IssueRows[0]
				issue := issueRow.AsIssue()
				newIssue, err := db.CreateIssue(&issue)

				By("throwing error", func() {
					Expect(err).ToNot(BeNil())
				})
				By("no issue returned", func() {
					Expect(newIssue).To(BeNil())
				})

			})
		})
	})
	When("Update Issue", Label("UpdateIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update issue description correctly", func() {
				issue := seedCollection.IssueRows[0].AsIssue()

				issue.Description = "New Description"
				err := db.UpdateIssue(&issue)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issue", func() {
					Expect(len(i)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(i[0].Issue.Description).To(BeEquivalentTo(issue.Description))
				})
			})
		})
	})
	When("Delete Issue", Label("DeleteIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete issue correctly", func() {
				issue := seedCollection.IssueRows[0].AsIssue()

				err := db.DeleteIssue(issue.Id, util.SystemUserId)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no issue", func() {
					Expect(len(i)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Add Component Version to Issue", Label("AddComponentVersionToIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			var newComponentVersionRow mariadb.ComponentVersionRow
			var newComponentVersion entity.ComponentVersion
			var componentVersion *entity.ComponentVersion
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newComponentVersionRow = test.NewFakeComponentVersion()
				newComponentVersionRow.ComponentId = seedCollection.ComponentRows[0].Id
				newComponentVersion = newComponentVersionRow.AsComponentVersion()
				componentVersion, _ = db.CreateComponentVersion(&newComponentVersion)
			})
			It("can add component version correctly", func() {
				issue := seedCollection.IssueRows[0].AsIssue()

				err := db.AddComponentVersionToIssue(issue.Id, componentVersion.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issue", func() {
					Expect(i).To(HaveLen(1))
				})
			})
		})
	})
	When("Remove Component Version from Issue", Label("RemoveComponentVersionFromIssue"), func() {
		Context("and we have 10 Issues in the database", func() {
			var seedCollection *test.SeedCollection
			var componentVersionIssueRow mariadb.ComponentVersionIssueRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				componentVersionIssueRow = seedCollection.ComponentVersionIssueRows[0]
			})
			It("can remove component version correctly", func() {
				err := db.RemoveComponentVersionFromIssue(componentVersionIssueRow.IssueId.Int64, componentVersionIssueRow.ComponentVersionId.Int64)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					ComponentVersionId: []*int64{&componentVersionIssueRow.ComponentVersionId.Int64},
				}

				issues, err := db.GetIssues(issueFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				for _, issue := range issues {
					Expect(issue.Issue.Id).ToNot(BeEquivalentTo(componentVersionIssueRow.IssueId.Int64))
				}
			})
		})
	})
})

var _ = Describe("Ordering Issues", Label("IssueOrder"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	var seedCollection *test.SeedCollection

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	var testOrder = func(
		order []entity.Order,
		verifyFunc func(res []entity.IssueResult),
	) {
		res, err := db.GetIssues(nil, order)

		By("throwing no error", func() {
			Expect(err).Should(BeNil())
		})

		By("returning the correct number of results", func() {
			Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueRows)))
		})

		By("returning the correct order", func() {
			verifyFunc(res)
		})
	}

	When("with ASC order", Label("IssueASCOrder"), func() {

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			seedCollection.GetValidIssueMatchRows()
		})

		It("can order by id", func() {
			sort.Slice(seedCollection.IssueRows, func(i, j int) bool {
				return seedCollection.IssueRows[i].Id.Int64 < seedCollection.IssueRows[j].Id.Int64
			})

			order := []entity.Order{
				{By: entity.IssueId, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				for i, r := range res {
					Expect(r.Issue.Id).Should(BeEquivalentTo(seedCollection.IssueRows[i].Id.Int64))
				}
			})
		})

		It("can order by primaryName", func() {
			order := []entity.Order{
				{By: entity.IssuePrimaryName, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				var prev string = ""
				for _, r := range res {
					Expect(r).ShouldNot(BeNil())
					Expect(r.PrimaryName >= prev).Should(BeTrue())
					prev = r.PrimaryName
				}
			})
		})

		It("can order by rating", func() {

			order := []entity.Order{
				{By: entity.IssueVariantRating, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				prev := -10
				for _, r := range res {
					variants := seedCollection.GetIssueVariantsByIssueId(r.Issue.Id)
					ratings := lo.Map(variants, func(iv mariadb.IssueVariantRow, _ int) int {
						return test.SeverityToNumerical(iv.Rating.String)
					})
					highestRating := lo.Max(ratings)
					Expect(highestRating >= prev).Should(BeTrue())
				}
			})
		})

	})

	When("with DESC order", Label("IssueDESCOrder"), func() {

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		It("can order by id", func() {
			sort.Slice(seedCollection.IssueRows, func(i, j int) bool {
				return seedCollection.IssueRows[i].Id.Int64 > seedCollection.IssueRows[j].Id.Int64
			})

			order := []entity.Order{
				{By: entity.IssueId, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				for i, r := range res {
					Expect(r.Issue.Id).Should(BeEquivalentTo(seedCollection.IssueRows[i].Id.Int64))
				}
			})
		})

		It("can order by primaryName", func() {
			order := []entity.Order{
				{By: entity.IssuePrimaryName, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				var prev string = "\U0010FFFF"
				for _, r := range res {
					Expect(r).ShouldNot(BeNil())
					Expect(r.PrimaryName <= prev).Should(BeTrue())
					prev = r.PrimaryName
				}
			})
		})

		It("can order by rating", func() {
			order := []entity.Order{
				{By: entity.IssueVariantRating, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.IssueResult) {
				prev := 9999
				for _, r := range res {
					variants := seedCollection.GetIssueVariantsByIssueId(r.Issue.Id)
					ratings := lo.Map(variants, func(iv mariadb.IssueVariantRow, _ int) int {
						return test.SeverityToNumerical(iv.Rating.String)
					})
					highestRating := lo.Max(ratings)
					Expect(highestRating <= prev).Should(BeTrue())
				}
			})
		})
	})

})
