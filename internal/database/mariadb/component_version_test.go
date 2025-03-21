// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"fmt"
	"math/rand"

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
			})
		})
	})

	When("Getting ComponentVersions", Label("GetComponentVersions"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetComponentVersions(nil)
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
					res, err := db.GetComponentVersions(nil)

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

					entries, err := db.GetComponentVersions(filter)

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

					entries, err := db.GetComponentVersions(filter)

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

					entries, err := db.GetComponentVersions(filter)

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

					entries, err := db.GetComponentVersions(filter)

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

					entries, err := db.GetComponentVersions(filter)

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

					entries, err := db.GetComponentVersions(filter)

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

					entries, err := db.GetComponentVersions(filter)

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
					// First, find a component version to update with a known tag
					cv := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]
					componentVersion := cv.AsComponentVersion()

					// Set a specific test tag
					testTag := "specific-test-tag-for-filtering"
					componentVersion.Tag = testTag

					// Update the existing component version with our test tag
					err := db.UpdateComponentVersion(&componentVersion)
					Expect(err).To(BeNil())

					// Now filter by the specific tag we just set
					filter := &entity.ComponentVersionFilter{Tag: []*string{&testTag}}
					entries, err := db.GetComponentVersions(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning at least one result", func() {
						Expect(entries).NotTo(BeEmpty())
					})

					By("ensuring all returned entries have the correct tag", func() {
						for _, entry := range entries {
							Expect(entry.Tag).To(Equal(testTag))
						}
					})

					By("including our updated component version", func() {
						found := false
						for _, entry := range entries {
							if entry.Id == componentVersion.Id {
								found = true
								break
							}
						}
						Expect(found).To(BeTrue())
					})
				})
			})

			Context("and using pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfList(
						db.GetComponentVersions,
						func(first *int, after *int64) *entity.ComponentVersionFilter {
							return &entity.ComponentVersionFilter{
								Paginated: entity.Paginated{First: first, After: after},
							}
						},
						func(entries []entity.ComponentVersion) *int64 { return &entries[len(entries)-1].Id },
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
						Paginated: entity.Paginated{
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

					cv, err := db.GetComponentVersions(componentVersionFilter)
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

					cv, err := db.GetComponentVersions(componentVersionFilter)
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

					cv, err := db.GetComponentVersions(componentVersionFilter)
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

				It("can update tag correctly", func() {
					// Get an existing component version to update
					componentVersion := seedCollection.ComponentVersionRows[0].AsComponentVersion()

					// Set a unique updated tag value
					updatedTag := "updated-tag-" + fmt.Sprintf("%d", rand.Int())
					componentVersion.Tag = updatedTag

					// Perform the update
					err := db.UpdateComponentVersion(&componentVersion)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					// Retrieve the updated component version
					componentVersionFilter := &entity.ComponentVersionFilter{
						Id: []*int64{&componentVersion.Id},
					}

					cv, err := db.GetComponentVersions(componentVersionFilter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning componentVersion", func() {
						Expect(len(cv)).To(BeEquivalentTo(1))
					})

					By("setting the tag field correctly", func() {
						Expect(cv[0].Tag).To(BeEquivalentTo(updatedTag))
					})

					By("preserving other fields", func() {
						Expect(cv[0].Id).To(BeEquivalentTo(componentVersion.Id))
						Expect(cv[0].Version).To(BeEquivalentTo(componentVersion.Version))
						Expect(cv[0].ComponentId).To(BeEquivalentTo(componentVersion.ComponentId))
					})
				})
			})
		})

	})
})
