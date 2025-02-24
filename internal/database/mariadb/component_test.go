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

var _ = Describe("Component", Label("database", "Component"), func() {

	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Getting All Component IDs", Label("GetAllComponentIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllComponentIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Components in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, c := range seedCollection.ComponentRows {
					ids = append(ids, c.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllComponentIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentRows)))
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
				It("can filter by a single component id that does exist", func() {
					cId := ids[rand.Intn(len(ids))]
					filter := &entity.ComponentFilter{
						Id: []*int64{&cId},
					}

					entries, err := db.GetAllComponentIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(cId))
					})
				})
			})
		})
	})

	When("Getting Components", Label("GetComponents"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetComponents(nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 components in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetComponents(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentRows)))
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
							for _, row := range seedCollection.ComponentRows {
								if r.Id == row.Id.Int64 {
									Expect(r.CCRN).Should(BeEquivalentTo(row.CCRN.String), "CCRN should match")
									Expect(r.Type).Should(BeEquivalentTo(row.Type.String), "Type should match")
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
					row := seedCollection.ComponentRows[rand.Intn(len(seedCollection.ComponentRows))]
					filter := &entity.ComponentFilter{Id: []*int64{&row.Id.Int64}}

					entries, err := db.GetComponents(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the component ccrn", func() {
						for _, entry := range entries {
							Expect(entry.Id).To(BeEquivalentTo(row.Id.Int64))
						}
					})
				})
				It("can filter by a single ccrn", func() {
					row := seedCollection.ComponentRows[rand.Intn(len(seedCollection.ComponentRows))]
					filter := &entity.ComponentFilter{CCRN: []*string{&row.CCRN.String}}

					entries, err := db.GetComponents(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the component ccrn", func() {
						for _, entry := range entries {
							Expect(entry.CCRN).To(BeEquivalentTo(row.CCRN.String))
						}
					})
				})
				It("can filter by a random non existing component ccrn", func() {
					nonExistingCCRN := util.GenerateRandomString(40, nil)
					filter := &entity.ComponentFilter{CCRN: []*string{&nonExistingCCRN}}

					entries, err := db.GetComponents(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning no results", func() {
						Expect(entries).To(BeEmpty())
					})
				})
				It("can filter by all existing component ccrns", func() {
					componentCCRNs := make([]*string, len(seedCollection.ComponentRows))
					for i, row := range seedCollection.ComponentRows {
						x := row.CCRN.String
						componentCCRNs[i] = &x
					}
					filter := &entity.ComponentFilter{CCRN: componentCCRNs}

					entries, err := db.GetComponents(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(seedCollection.ComponentRows)))
					})
				})
			})
			Context("and using pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfList(
						db.GetComponents,
						func(first *int, after *int64) *entity.ComponentFilter {
							return &entity.ComponentFilter{
								Paginated: entity.Paginated{First: first, After: after},
							}
						},
						func(entries []entity.Component) *int64 { return &entries[len(entries)-1].Id },
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
	When("Counting Components", Label("CountComponents"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountComponents(nil)

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
			var componentRows []mariadb.ComponentRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				componentRows = seedCollection.ComponentRows
				count = len(componentRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountComponents(nil)

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
					filter := &entity.ComponentFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountComponents(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})

			Context("and using a filter", func() {
				DescribeTable("can count with a filter", func(pageSize int, filterMatches int) {
					componentRow := seedCollection.ComponentRows[rand.Intn(len(seedCollection.ComponentRows))]

					// collect all component ids that have the previously selected ccrn
					componentIds := []int64{}
					for _, cRow := range seedCollection.ComponentRows {
						if cRow.CCRN.String == componentRow.CCRN.String {
							componentIds = append(componentIds, cRow.Id.Int64)
						}
					}

					filter := &entity.ComponentFilter{
						Paginated: entity.Paginated{
							First: &pageSize,
							After: nil,
						},
						CCRN: []*string{&componentRow.CCRN.String},
					}
					entries, err := db.CountComponents(filter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(entries).To(BeEquivalentTo(len(componentIds)))
					})
				},
					Entry("and pageSize is 1 and it has 13 elements", 1, 13),
					Entry("and  pageSize is 20 and it has 5 elements", 20, 5),
					Entry("and  pageSize is 100 and it has 100 elements", 100, 100),
				)
			})
		})
		When("Insert Component", Label("InsertComponent"), func() {
			Context("and we have 10 Components in the database", func() {
				var newComponentRow mariadb.ComponentRow
				var newComponent entity.Component
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
					newComponentRow = test.NewFakeComponent()
					newComponent = newComponentRow.AsComponent()
				})
				It("can insert correctly", func() {
					component, err := db.CreateComponent(&newComponent)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("sets component id", func() {
						Expect(component).NotTo(BeEquivalentTo(0))
					})

					componentFilter := &entity.ComponentFilter{
						Id: []*int64{&component.Id},
					}

					c, err := db.GetComponents(componentFilter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning component", func() {
						Expect(len(c)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(c[0].Id).To(BeEquivalentTo(component.Id))
						Expect(c[0].CCRN).To(BeEquivalentTo(component.CCRN))
						Expect(c[0].Type).To(BeEquivalentTo(component.Type))
					})
				})
				It("does not insert component with existing ccrn", func() {
					componentRow := seedCollection.ComponentRows[0]
					component := componentRow.AsComponent()
					newComponent, err := db.CreateComponent(&component)

					By("throwing error", func() {
						Expect(err).ToNot(BeNil())
					})
					By("no component returned", func() {
						Expect(newComponent).To(BeNil())
					})

				})
			})
		})
		When("Update Component", Label("UpdateComponent"), func() {
			Context("and we have 10 Components in the database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
				})
				It("can update ccrn correctly", func() {
					component := seedCollection.ComponentRows[0].AsComponent()

					component.CCRN = "NewCCRN"
					err := db.UpdateComponent(&component)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					componentFilter := &entity.ComponentFilter{
						Id: []*int64{&component.Id},
					}

					c, err := db.GetComponents(componentFilter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning component", func() {
						Expect(len(c)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(c[0].Id).To(BeEquivalentTo(component.Id))
						Expect(c[0].CCRN).To(BeEquivalentTo(component.CCRN))
						Expect(c[0].Type).To(BeEquivalentTo(component.Type))
					})
				})
				It("can update type correctly", func() {
					component := seedCollection.ComponentRows[0].AsComponent()

					component.Type = "NewType"
					err := db.UpdateComponent(&component)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					componentFilter := &entity.ComponentFilter{
						Id: []*int64{&component.Id},
					}

					c, err := db.GetComponents(componentFilter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning component", func() {
						Expect(len(c)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(c[0].Id).To(BeEquivalentTo(component.Id))
						Expect(c[0].CCRN).To(BeEquivalentTo(component.CCRN))
						Expect(c[0].Type).To(BeEquivalentTo(component.Type))
					})
				})
			})
		})
		When("Delete Component", Label("DeleteComponent"), func() {
			Context("and we have 10 Components in the database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
				})
				It("can delete component correctly", func() {
					component := seedCollection.ComponentRows[0].AsComponent()

					err := db.DeleteComponent(component.Id, systemUserId)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					componentFilter := &entity.ComponentFilter{
						Id: []*int64{&component.Id},
					}

					c, err := db.GetComponents(componentFilter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning no component", func() {
						Expect(len(c)).To(BeEquivalentTo(0))
					})
				})
			})
		})
	})
})
