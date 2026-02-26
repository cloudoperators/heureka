// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Patch", Label("database", "Patch"), func() {
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

	When("Getting Patches", Label("GetPatches"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetPatches(nil, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 patches in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetPatches(nil, nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.PatchRows)))
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
							for _, row := range seedCollection.PatchRows {
								if r.Id == row.Id.Int64 {
									Expect(r.ServiceId).Should(BeEquivalentTo(row.ServiceId.Int64), "ServiceId should match")
									Expect(r.ServiceName).Should(BeEquivalentTo(row.ServiceName.String), "ServiceName should match")
									Expect(r.ComponentVersionId).Should(BeEquivalentTo(row.ComponentVersionId.Int64), "ComponentVersionId should match")
									Expect(r.ComponentVersionName).Should(BeEquivalentTo(row.ComponentVersionName.String), "ComponentVersionName should match")
									Expect(r.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single id", func() {
					row := seedCollection.PatchRows[rand.Intn(len(seedCollection.PatchRows))]
					filter := &entity.PatchFilter{Id: []*int64{&row.Id.Int64}}

					entries, err := db.GetPatches(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning one result", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})
					By("returned entry includes the id", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(row.Id.Int64))
					})
				})
				It("can filter by a service id", func() {
					row := seedCollection.PatchRows[rand.Intn(len(seedCollection.PatchRows))]
					filter := &entity.PatchFilter{ServiceId: []*int64{&row.ServiceId.Int64}}

					entries, err := db.GetPatches(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the service id", func() {
						for _, entry := range entries {
							Expect(entry.ServiceId).To(BeEquivalentTo(row.ServiceId.Int64))
						}
					})
				})
				It("can filter by a service name", func() {
					row := seedCollection.PatchRows[rand.Intn(len(seedCollection.PatchRows))]
					filter := &entity.PatchFilter{ServiceName: []*string{&row.ServiceName.String}}

					entries, err := db.GetPatches(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the service name", func() {
						for _, entry := range entries {
							Expect(entry.ServiceName).To(BeEquivalentTo(row.ServiceName.String))
						}
					})
				})
				It("can filter by a component version id", func() {
					row := seedCollection.PatchRows[rand.Intn(len(seedCollection.PatchRows))]
					filter := &entity.PatchFilter{ComponentVersionId: []*int64{&row.ComponentVersionId.Int64}}

					entries, err := db.GetPatches(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the component version id", func() {
						for _, entry := range entries {
							Expect(entry.ComponentVersionId).To(BeEquivalentTo(row.ComponentVersionId.Int64))
						}
					})
				})
				It("can filter by a component version name", func() {
					row := seedCollection.PatchRows[rand.Intn(len(seedCollection.PatchRows))]
					filter := &entity.PatchFilter{ComponentVersionName: []*string{&row.ComponentVersionName.String}}

					entries, err := db.GetPatches(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the component version name", func() {
						for _, entry := range entries {
							Expect(entry.ComponentVersionName).To(BeEquivalentTo(row.ComponentVersionName.String))
						}
					})
				})
			})
			Context("and using pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfListWithOrder(
						db.GetPatches,
						func(first *int, after *int64, afterX *string) *entity.PatchFilter {
							return &entity.PatchFilter{
								PaginatedX: entity.PaginatedX{First: first, After: afterX},
							}
						},
						[]entity.Order{},
						func(entries []entity.PatchResult) string {
							after, _ := mariadb.EncodeCursor(mariadb.WithPatch([]entity.Order{}, *entries[len(entries)-1].Patch))
							return after
						},
						len(seedCollection.PatchRows),
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
	When("Counting Patches", Label("CountPatches"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountPatches(nil)

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
			var patchRows []mariadb.PatchRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				patchRows = seedCollection.PatchRows
				count = len(patchRows)
			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountPatches(nil)

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
					after := ""
					filter := &entity.PatchFilter{
						PaginatedX: entity.PaginatedX{
							First: &f,
							After: &after,
						},
					}
					c, err := db.CountPatches(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})
		})
	})
})
