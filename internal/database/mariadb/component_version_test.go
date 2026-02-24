// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("ComponentVersion", Label("database", "ComponentVersion"), func() {
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

	When("Getting All ComponentVersion IDs", Label("GetAllComponentVersionIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllComponentVersionIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 ComponentVersions in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, cv := range seedCollection.ComponentVersionRows {
					ids = append(ids, cv.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllComponentVersionIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentVersionRows)))
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
				It("can filter by a single componentVersion id that does exist", func() {
					cvId := ids[rand.Intn(len(ids))]
					filter := &entity.ComponentVersionFilter{
						Id: []*int64{&cvId},
					}

					entries, err := db.GetAllComponentVersionIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(cvId))
					})
				})
				It("can filter by end of life as true value", func() {
					endOfLifeAsTrue := true

					filter := &entity.ComponentVersionFilter{
						EndOfLife: []*bool{&endOfLifeAsTrue},
					}

					ids, err := db.GetAllComponentVersionIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning versions without end of life", func() {
						Expect(ids).ToNot(BeEmpty())
					})

					cvFilterIDs := make([]*int64, 0, len(ids))
					for _, id := range ids {
						cvFilterIDs = append(cvFilterIDs, &id)
					}

					res, err := db.GetComponentVersions(&entity.ComponentVersionFilter{
						Id: cvFilterIDs,
					}, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					for _, r := range res {
						Expect(r.EndOfLife).To(BeTrue())
					}
				})
				It("can filter by end of life as false value", func() {
					endOfLifeAsFalse := false

					filter := &entity.ComponentVersionFilter{
						EndOfLife: []*bool{&endOfLifeAsFalse},
					}

					ids, err := db.GetAllComponentVersionIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning versions without end of life", func() {
						Expect(ids).ToNot(BeEmpty())
					})

					cvFilterIDs := make([]*int64, 0, len(ids))
					for _, id := range ids {
						cvFilterIDs = append(cvFilterIDs, &id)
					}

					res, err := db.GetComponentVersions(&entity.ComponentVersionFilter{
						Id: cvFilterIDs,
					}, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					for _, r := range res {
						Expect(r.EndOfLife).To(BeFalse())
					}
				})
			})
		})
	})

	When("Getting ComponentVersions", Label("GetComponentVersions"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetComponentVersions(nil, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 component versions in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetComponentVersions(nil, nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentVersionRows)))
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
							for _, row := range seedCollection.ComponentVersionRows {
								if r.Id == row.Id.Int64 {
									Expect(r.Version).Should(BeEquivalentTo(row.Version.String), "Name should match")
									Expect(r.Tag).Should(BeEquivalentTo(row.Tag.String), "Tag matches")
									Expect(r.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single component version id that does exist", func() {
					cv := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]
					filter := &entity.ComponentVersionFilter{
						Id: []*int64{&cv.Id.Int64},
					}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(cv.Id.Int64))
					})
				})
				It("can filter by an issue id", func() {
					issueRow := seedCollection.IssueRows[rand.Intn(len(seedCollection.IssueRows))]

					// collect all component version ids that belong to the issues
					componentVersionIds := []int64{}
					for _, cvvRow := range seedCollection.ComponentVersionIssueRows {
						if cvvRow.IssueId.Int64 == issueRow.Id.Int64 {
							componentVersionIds = append(componentVersionIds, cvvRow.ComponentVersionId.Int64)
						}
					}

					filter := &entity.ComponentVersionFilter{IssueId: []*int64{&issueRow.Id.Int64}}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(componentVersionIds).To(ContainElement(entry.Id))
						}
					})
				})
				It("can filter by a component id", func() {
					// select a component
					componentRow := seedCollection.ComponentRows[rand.Intn(len(seedCollection.ComponentRows))]

					// collect all activity ids that belong to the component
					componentVersionIds := []int64{}
					for _, cvRow := range seedCollection.ComponentVersionRows {
						if cvRow.ComponentId.Int64 == componentRow.Id.Int64 {
							componentVersionIds = append(componentVersionIds, cvRow.Id.Int64)
						}
					}

					filter := &entity.ComponentVersionFilter{ComponentId: []*int64{&componentRow.Id.Int64}}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(componentVersionIds).To(ContainElement(entry.Id))
						}
					})
				})
				It("can filter by a version", func() {
					cv := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]

					filter := &entity.ComponentVersionFilter{Version: []*string{&cv.Version.String}}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(entry.Version).To(BeEquivalentTo(cv.Version.String))
						}
					})
				})
				It("can filter by a version and component", func() {
					cv := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]

					componentCCRN := ""
					for _, cr := range seedCollection.ComponentRows {
						if cr.Id.Int64 == cv.ComponentId.Int64 {
							componentCCRN = cr.CCRN.String
						}
					}

					filter := &entity.ComponentVersionFilter{
						Version:       []*string{&cv.Version.String},
						ComponentCCRN: []*string{&componentCCRN},
					}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(entry.Version).To(BeEquivalentTo(cv.Version.String))
							Expect(entry.ComponentId).To(BeEquivalentTo(cv.ComponentId.Int64))
						}
					})
				})

				It("can filter by a service id", func() {
					s := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					cvs := []int64{}
					for _, ci := range seedCollection.ComponentInstanceRows {
						if ci.ServiceId.Int64 == s.Id.Int64 {
							cvs = append(cvs, ci.ComponentVersionId.Int64)
						}
					}

					filter := &entity.ComponentVersionFilter{ServiceId: []*int64{&s.Id.Int64}}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(cvs, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a service ccrn", func() {
					s := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					cvs := []int64{}
					for _, ci := range seedCollection.ComponentInstanceRows {
						if ci.ServiceId.Int64 == s.Id.Int64 {
							cvs = append(cvs, ci.ComponentVersionId.Int64)
						}
					}

					filter := &entity.ComponentVersionFilter{ServiceCCRN: []*string{&s.CCRN.String}}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(cvs, entry.Id)).To(BeTrue())
						}
					})
				})

				It("can filter by tag", func() {
					// Get an existing component version from the fixtures
					cv := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]

					// Get the tag value directly from the fixture
					tagToFilterBy := cv.Tag.String

					// Create a filter using the existing tag value
					filter := &entity.ComponentVersionFilter{Tag: []*string{&tagToFilterBy}}

					// Execute the query
					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning at least one result", func() {
						Expect(entries).NotTo(BeEmpty())
					})

					By("ensuring all returned entries have the correct tag", func() {
						for _, entry := range entries {
							Expect(entry.Tag).To(Equal(tagToFilterBy))
						}
					})

					By("including our expected component version", func() {
						found := false
						for _, entry := range entries {
							if entry.Id == cv.Id.Int64 {
								found = true
								break
							}
						}
						Expect(found).To(BeTrue())
					})
				})

				It("can filter by a repository", func() {
					cv := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]
					filter := &entity.ComponentVersionFilter{Repository: []*string{&cv.Repository.String}}
					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(entry.Repository).To(BeEquivalentTo(cv.Repository.String))
						}
					})
				})

				It("can filter by an organization", func() {
					cv := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]
					filter := &entity.ComponentVersionFilter{Organization: []*string{&cv.Organization.String}}
					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(entry.Organization).To(BeEquivalentTo(cv.Organization.String))
						}
					})
				})

				It("can filter by end of life as false value", func() {
					endOfLifeAsFalse := false

					filter := &entity.ComponentVersionFilter{
						EndOfLife: []*bool{&endOfLifeAsFalse},
					}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected values for EndOfLife", func() {
						for _, entry := range entries {
							Expect(entry.EndOfLife).To(BeFalse())
						}
					})
				})

				It("can filter by end of life as true value", func() {
					endOfLifeAsTrue := true

					filter := &entity.ComponentVersionFilter{
						EndOfLife: []*bool{&endOfLifeAsTrue},
					}

					entries, err := db.GetComponentVersions(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected values for EndOfLife", func() {
						for _, entry := range entries {
							Expect(entry.EndOfLife).To(BeTrue())
						}
					})
				})
			})

			Context("and using pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfListWithOrder(
						db.GetComponentVersions,
						func(first *int, after *int64, afterX *string) *entity.ComponentVersionFilter {
							return &entity.ComponentVersionFilter{
								PaginatedX: entity.PaginatedX{First: first, After: afterX},
							}
						},
						[]entity.Order{},
						func(entries []entity.ComponentVersionResult) string {
							after, _ := mariadb.EncodeCursor(mariadb.WithComponentVersion([]entity.Order{}, *entries[len(entries)-1].ComponentVersion, entity.IssueSeverityCounts{}))
							return after
						},
						len(seedCollection.ComponentVersionRows),
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
	When("Counting ComponentVersions", Label("CountComponentVersions"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountComponentVersions(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning the correct count", func() {
					Expect(c).To(BeEquivalentTo(0))
				})
			})
		})
		Context("and the database has 100 entries", func() {
			var seedCollection *test.SeedCollection
			var cvRows []mariadb.ComponentVersionRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				cvRows = seedCollection.ComponentVersionRows
				count = len(cvRows)
			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountComponentVersions(nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})
			Context("and using pagination", func() {
				It("can count", func() {
					f := 10
					filter := &entity.ComponentVersionFilter{
						PaginatedX: entity.PaginatedX{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountComponentVersions(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})
		})
		When("Insert ComponentVersion", Label("InsertComponentVersion"), func() {
			Context("and we have 10 ComponentVersions in the database", func() {
				var newComponentVersionRow mariadb.ComponentVersionRow
				var newComponentVersion entity.ComponentVersion
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seeder.SeedDbWithNFakeData(10)
					seedCollection = seeder.SeedDbWithNFakeData(10)
					newComponentVersionRow = test.NewFakeComponentVersion()
					newComponentVersionRow.ComponentId = seedCollection.ComponentRows[0].Id

					// Set a specific tag value on the row
					testTag := "insert-test-tag"
					newComponentVersionRow.Tag.String = testTag
					newComponentVersionRow.Tag.Valid = true

					newComponentVersion = newComponentVersionRow.AsComponentVersion()

					// Ensure the entity also has the tag set
					Expect(newComponentVersion.Tag).To(Equal(testTag))
				})
				It("can insert correctly", func() {
					componentVersion, err := db.CreateComponentVersion(&newComponentVersion)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("sets componentVersion id", func() {
						Expect(componentVersion).NotTo(BeEquivalentTo(0))
					})

					componentVersionFilter := &entity.ComponentVersionFilter{
						Id: []*int64{&componentVersion.Id},
					}

					cv, err := db.GetComponentVersions(componentVersionFilter, nil)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning componentVersion", func() {
						Expect(len(cv)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(cv[0].Id).To(BeEquivalentTo(componentVersion.Id))
						Expect(cv[0].Tag).To(BeEquivalentTo(componentVersion.Tag))
						Expect(cv[0].Version).To(BeEquivalentTo(componentVersion.Version))
					})
				})
			})
		})
		When("Update ComponentVersion", Label("UpdateComponentVersion"), func() {
			Context("and we have 10 ComponentVersions in the database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
				})
				It("can update version correctly", func() {
					componentVersion := seedCollection.ComponentVersionRows[0].AsComponentVersion()

					componentVersion.Version = "1.3.3.7"
					err := db.UpdateComponentVersion(&componentVersion)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					componentVersionFilter := &entity.ComponentVersionFilter{
						Id: []*int64{&componentVersion.Id},
					}

					cv, err := db.GetComponentVersions(componentVersionFilter, nil)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning componentVersions", func() {
						Expect(len(cv)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(cv[0].Id).To(BeEquivalentTo(componentVersion.Id))
						Expect(cv[0].Version).To(BeEquivalentTo(componentVersion.Version))
						Expect(cv[0].ComponentId).To(BeEquivalentTo(componentVersion.ComponentId))
					})
				})
				It("can update componentId correctly", func() {
					componentVersion := seedCollection.ComponentVersionRows[0].AsComponentVersion()

					componentVersion.ComponentId = seedCollection.ComponentRows[1].Id.Int64
					err := db.UpdateComponentVersion(&componentVersion)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					componentVersionFilter := &entity.ComponentVersionFilter{
						Id: []*int64{&componentVersion.Id},
					}

					cv, err := db.GetComponentVersions(componentVersionFilter, nil)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning componentVersion", func() {
						Expect(len(cv)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(cv[0].Id).To(BeEquivalentTo(componentVersion.Id))
						Expect(cv[0].Version).To(BeEquivalentTo(componentVersion.Version))
						Expect(cv[0].ComponentId).To(BeEquivalentTo(componentVersion.ComponentId))
					})
				})

				It("can update tag correctly", Label("UpdateComponentVersion", "GetComponentVersions"), func() {
					// Get an existing component version to update
					componentVersion := seedCollection.ComponentVersionRows[0].AsComponentVersion()

					// Store the original values for comparison
					originalId := componentVersion.Id
					originalVersion := componentVersion.Version
					originalComponentId := componentVersion.ComponentId

					// Set a unique updated tag value
					updatedTag := "updated-tag-" + fmt.Sprintf("%d", rand.Int())
					componentVersion.Tag = updatedTag

					// Perform the update
					err := db.UpdateComponentVersion(&componentVersion)

					By("throwing no error during update", func() {
						Expect(err).To(BeNil())
					})

					// Retrieve all component versions and find our updated one manually
					// This avoids relying on the filter functionality
					allVersions, err := db.GetComponentVersions(nil, nil)

					By("throwing no error during retrieval", func() {
						Expect(err).To(BeNil())
					})

					By("being able to find the updated version", func() {
						found := false
						var updatedCV entity.ComponentVersion

						for _, cv := range allVersions {
							if cv.Id == originalId {
								found = true
								updatedCV = *cv.ComponentVersion
								break
							}
						}

						Expect(found).To(BeTrue(), "Updated component version should be retrievable")

						if found {
							Expect(updatedCV.Tag).To(BeEquivalentTo(updatedTag), "Tag should be updated")
							Expect(updatedCV.Id).To(BeEquivalentTo(originalId), "ID should be preserved")
							Expect(updatedCV.Version).To(BeEquivalentTo(originalVersion), "Version should be preserved")
							Expect(updatedCV.ComponentId).To(BeEquivalentTo(originalComponentId), "ComponentId should be preserved")
						}
					})
				})
			})
		})
	})
})

var _ = Describe("Ordering ComponentVersions", func() {
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

	testOrder := func(
		order []entity.Order,
		verifyFunc func(res []entity.ComponentVersionResult),
	) {
		res, err := db.GetComponentVersions(nil, order)

		By("throwing no error", func() {
			Expect(err).Should(BeNil())
		})

		By("returning the correct number of results", func() {
			Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentVersionRows)))
		})

		By("returning the correct order", func() {
			verifyFunc(res)
		})
	}

	loadTestData := func() ([]mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, error) {
		issueVariants, err := test.LoadIssueVariants(test.GetTestDataPath("testdata/component_version_order/issue_variant.json"))
		if err != nil {
			return nil, nil, err
		}
		cvIssues, err := test.LoadComponentVersionIssues(test.GetTestDataPath("testdata/component_version_order/component_version_issue.json"))
		if err != nil {
			return nil, nil, err
		}
		return issueVariants, cvIssues, nil
	}

	When("order by count is used", func() {
		BeforeEach(func() {
			seeder.SeedIssueRepositories()
			seeder.SeedIssues(10)
			components := seeder.SeedComponents(1)
			seeder.SeedComponentVersions(10, components)
			issueVariants, componentVersionIssues, err := loadTestData()
			Expect(err).To(BeNil())
			// Important: the order need to be preserved
			for _, iv := range issueVariants {
				_, err := seeder.InsertFakeIssueVariant(iv)
				Expect(err).To(BeNil())
			}
			for _, cvi := range componentVersionIssues {
				_, err := seeder.InsertFakeComponentVersionIssue(cvi)
				Expect(err).To(BeNil())
			}
		})
		It("can order desc by critical, high, medium, low and none", func() {
			order := []entity.Order{
				{By: entity.CriticalCount, Direction: entity.OrderDirectionDesc},
				{By: entity.HighCount, Direction: entity.OrderDirectionDesc},
				{By: entity.MediumCount, Direction: entity.OrderDirectionDesc},
				{By: entity.LowCount, Direction: entity.OrderDirectionDesc},
				{By: entity.NoneCount, Direction: entity.OrderDirectionDesc},
			}
			cvs, err := db.GetComponentVersions(nil, order)
			Expect(err).To(BeNil())
			Expect(cvs[0].Id).To(BeEquivalentTo(3))
			Expect(cvs[1].Id).To(BeEquivalentTo(8))
			Expect(cvs[2].Id).To(BeEquivalentTo(2))
			Expect(cvs[3].Id).To(BeEquivalentTo(7))
			Expect(cvs[4].Id).To(BeEquivalentTo(1))
			Expect(cvs[5].Id).To(BeEquivalentTo(6))
			Expect(cvs[6].Id).To(BeEquivalentTo(5))
			Expect(cvs[7].Id).To(BeEquivalentTo(4))
			Expect(cvs[8].Id).To(BeEquivalentTo(10))
			Expect(cvs[9].Id).To(BeEquivalentTo(9))
		})
		It("can order asc by critical, high, medium, low and none", func() {
			order := []entity.Order{
				{By: entity.CriticalCount, Direction: entity.OrderDirectionAsc},
				{By: entity.HighCount, Direction: entity.OrderDirectionAsc},
				{By: entity.MediumCount, Direction: entity.OrderDirectionAsc},
				{By: entity.LowCount, Direction: entity.OrderDirectionAsc},
				{By: entity.NoneCount, Direction: entity.OrderDirectionAsc},
			}
			cvs, err := db.GetComponentVersions(nil, order)
			Expect(err).To(BeNil())
			Expect(cvs[0].Id).To(BeEquivalentTo(9))
			Expect(cvs[1].Id).To(BeEquivalentTo(10))
			Expect(cvs[2].Id).To(BeEquivalentTo(4))
			Expect(cvs[3].Id).To(BeEquivalentTo(5))
			Expect(cvs[4].Id).To(BeEquivalentTo(6))
			Expect(cvs[5].Id).To(BeEquivalentTo(1))
			Expect(cvs[6].Id).To(BeEquivalentTo(7))
			Expect(cvs[7].Id).To(BeEquivalentTo(2))
			Expect(cvs[8].Id).To(BeEquivalentTo(8))
			Expect(cvs[9].Id).To(BeEquivalentTo(3))
		})
	})

	When("with ASC order", Label("ComponentVersionASCOrder"), func() {
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		It("can order by id", func() {
			sort.Slice(seedCollection.ComponentVersionRows, func(i, j int) bool {
				return seedCollection.ComponentVersionRows[i].Id.Int64 < seedCollection.ComponentVersionRows[j].Id.Int64
			})

			order := []entity.Order{
				{By: entity.ComponentVersionId, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.ComponentVersionResult) {
				for i, r := range res {
					Expect(r.Id).Should(BeEquivalentTo(seedCollection.ComponentVersionRows[i].Id.Int64))
				}
			})
		})

		It("can order by repository", func() {
			sort.Slice(seedCollection.ComponentVersionRows, func(i, j int) bool {
				return seedCollection.ComponentVersionRows[i].Repository.String < seedCollection.ComponentVersionRows[j].Repository.String
			})

			order := []entity.Order{
				{By: entity.ComponentVersionRepository, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.ComponentVersionResult) {
				for i, r := range res {
					Expect(r.Id).Should(BeEquivalentTo(seedCollection.ComponentVersionRows[i].Id.Int64))
				}
			})
		})
	})

	When("with DESC order", Label("ComponentVersionDESCOrder"), func() {
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		It("can order by id", func() {
			sort.Slice(seedCollection.ComponentVersionRows, func(i, j int) bool {
				return seedCollection.ComponentVersionRows[i].Id.Int64 > seedCollection.ComponentVersionRows[j].Id.Int64
			})

			order := []entity.Order{
				{By: entity.ComponentVersionId, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.ComponentVersionResult) {
				for i, r := range res {
					Expect(r.Id).Should(BeEquivalentTo(seedCollection.ComponentVersionRows[i].Id.Int64))
				}
			})
		})

		It("can order by repository", func() {
			sort.Slice(seedCollection.ComponentVersionRows, func(i, j int) bool {
				return seedCollection.ComponentVersionRows[i].Repository.String > seedCollection.ComponentVersionRows[j].Repository.String
			})

			order := []entity.Order{
				{By: entity.ComponentVersionRepository, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.ComponentVersionResult) {
				for i, r := range res {
					Expect(r.Id).Should(BeEquivalentTo(seedCollection.ComponentVersionRows[i].Id.Int64))
				}
			})
		})
	})
})
