// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"math/rand"

	"github.com/cloudoperators/heureka/pkg/util"
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
				res, err := db.GetIssues(nil)
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
					res, err := db.GetIssues(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueRows)))
					})

					By("returning the correct order", func() {
						var prev int64 = 0
						for _, r := range res {

							Expect(r.Id > prev).Should(BeTrue())
							prev = r.Id

						}
					})

					By("returning the correct fields", func() {
						for _, r := range res {
							for _, row := range seedCollection.IssueRows {
								if r.Id == row.Id.Int64 {
									Expect(r.PrimaryName).Should(BeEquivalentTo(row.PrimaryName.String), "Name should match")
									Expect(r.Type).Should(BeEquivalentTo(row.Type.String), "Type should match")
									Expect(r.Description).Should(BeEquivalentTo(row.Description.String), "Description should match")
									Expect(r.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
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
					filter := &entity.IssueFilter{ServiceName: []*string{&row.Name.String}}

					entries, err := db.GetIssues(filter)

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
					nonExistingName := util.GenerateRandomString(40, nil)
					filter := &entity.IssueFilter{ServiceName: []*string{&nonExistingName}}

					entries, err := db.GetIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning no results", func() {
						Expect(entries).To(BeEmpty())
					})
				})
				It("can filter by multiple existing service names", func() {
					serviceNames := make([]*string, len(seedCollection.ServiceRows))
					var expectedIssues []mariadb.IssueRow
					for i, row := range seedCollection.ServiceRows {
						x := row.Name.String
						expectedIssues = append(expectedIssues, seedCollection.GetIssueByService(&row)...)
						serviceNames[i] = &x
					}
					expectedIssues = lo.Uniq(expectedIssues)
					filter := &entity.IssueFilter{ServiceName: serviceNames}

					entries, err := db.GetIssues(filter)

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

					entries, err := db.GetIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning exactly 1 element", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning the expected element", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(row.Id.Int64))
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

					entries, err := db.GetIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						for _, entry := range entries {
							Expect(issueIds).To(ContainElement(entry.Id))
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

					entries, err := db.GetIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the expected elements", func() {
						for _, entry := range entries {
							Expect(issueIds).To(ContainElement(entry.Id))
						}
					})
				})
				It("can filter by a single issueVariant id", func() {
					// select an issueVariant
					issueVariantRow := seedCollection.IssueVariantRows[rand.Intn(len(seedCollection.IssueVariantRows))]

					filter := &entity.IssueFilter{IssueVariantId: []*int64{&issueVariantRow.Id.Int64}}

					entries, err := db.GetIssues(filter)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Id)
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

					entries, err := db.GetIssues(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					for _, entry := range entries {
						Expect(entry.Type).To(BeEquivalentTo(issueType))
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

					entries, err := db.GetIssues(filter)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Id)
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

					entries, err := db.GetIssues(filter)

					issueIds := []int64{}
					for _, entry := range entries {
						issueIds = append(issueIds, entry.Id)
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
					test.TestPaginationOfList(
						db.GetIssues,
						func(first *int, after *int64) *entity.IssueFilter {
							return &entity.IssueFilter{
								Paginated: entity.Paginated{First: first, After: after},
							}
						},
						func(entries []entity.Issue) *int64 { return &entries[len(entries)-1].Id },
						10,
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
		BeforeEach(func() {
			_ = seeder.SeedDbWithNFakeData(10)
		})
		Context("and and we have 10 elements in the database", func() {
			It("returns the issues with aggregations", func() {
				entriesWithAggregations, err := db.GetIssuesWithAggregations(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning some aggregations", func() {
					for _, entryWithAggregations := range entriesWithAggregations {
						Expect(entryWithAggregations).NotTo(
							BeEquivalentTo(entity.IssueAggregations{}))
					}
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
				Entry("when page size is 100", 100))
		})
		Context("and using a filter", func() {
			Context("and having 20 elements in the Database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(20)
				})
				It("does not influence the count when pagination is applied", func() {
					var first = 1
					var after int64 = 0
					filter := &entity.IssueFilter{
						Paginated: entity.Paginated{
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
					filter := &entity.IssueFilter{ServiceName: []*string{&row.Name.String}}

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

				i, err := db.GetIssues(issueFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issue", func() {
					Expect(len(i)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(i[0].PrimaryName).To(BeEquivalentTo(issue.PrimaryName))
					Expect(i[0].Type.String()).To(BeEquivalentTo(issue.Type.String()))
					Expect(i[0].Description).To(BeEquivalentTo(issue.Description))
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

				i, err := db.GetIssues(issueFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issue", func() {
					Expect(len(i)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(i[0].Description).To(BeEquivalentTo(issue.Description))
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

				err := db.DeleteIssue(issue.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter)
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
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can add component version correctly", func() {
				issue := seedCollection.IssueRows[0].AsIssue()
				componentVersion := seedCollection.ComponentVersionRows[0].AsComponentVersion()
				componentVersion.ComponentId = seedCollection.ComponentRows[0].Id.Int64

				err := db.AddComponentVersionToIssue(issue.Id, componentVersion.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueFilter := &entity.IssueFilter{
					Id: []*int64{&issue.Id},
				}

				i, err := db.GetIssues(issueFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issue", func() {
					Expect(len(i)).To(BeEquivalentTo(1))
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

				issues, err := db.GetIssues(issueFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				for _, issue := range issues {
					Expect(issue.Id).ToNot(BeEquivalentTo(componentVersionIssueRow.IssueId.Int64))
				}
			})
		})
	})
})
