// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("ComponentInstance - ", Label("database", "ComponentInstance"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Getting All ComponentInstance IDs", Label("GetAllComponentInstanceIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllComponentInstanceIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Services in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, ci := range seedCollection.ComponentInstanceRows {
					ids = append(ids, ci.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllComponentInstanceIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentInstanceRows)))
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
				It("can filter by a single componentInstance id that does exist", func() {
					ciId := ids[rand.Intn(len(ids))]
					filter := &entity.ComponentInstanceFilter{
						Id: []*int64{&ciId},
					}

					entries, err := db.GetAllComponentInstanceIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(ciId))
					})
				})
				It("can filter by a single componentVersion id that does exist", func() {
					// select a component version
					cvRow := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]

					// collect all componentInstance ids that belong to the component version
					ciIds := []int64{}
					for _, ciRow := range seedCollection.ComponentInstanceRows {
						if ciRow.ComponentVersionId.Int64 == cvRow.Id.Int64 {
							ciIds = append(ciIds, ciRow.ServiceId.Int64)
						}
					}

					filter := &entity.ComponentInstanceFilter{
						ComponentVersionId: []*int64{&cvRow.Id.Int64},
					}

					entries, err := db.GetAllComponentInstanceIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(ciIds)))
					})
				})
			})
		})
	})

	When("Getting ComponentInstances", Label("GetComponentInstance"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetComponentInstances(nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 component instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			Context("and using no filter", func() {

				It("can fetch the items correctly", func() {
					res, err := db.GetComponentInstances(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentInstanceRows)))
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
							for _, row := range seedCollection.ComponentInstanceRows {
								if r.Id == row.Id.Int64 {
									Expect(r.CCRN).Should(BeEquivalentTo(row.CCRN.String), "CCRN matches")
									Expect(r.Count).Should(BeEquivalentTo(row.Count.Int16), "Count matches")
									Expect(r.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single component instance id that does exist", func() {
					ci := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]
					filter := &entity.ComponentInstanceFilter{
						Id: []*int64{&ci.Id.Int64},
					}

					entries, err := db.GetComponentInstances(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(ci.Id.Int64))
					})
				})
				It("can filter by a single issue match id that does exist", func() {
					//get a service that should return at least one issue
					rnd := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					ciId := rnd.ComponentInstanceId.Int64
					filter := &entity.ComponentInstanceFilter{
						Paginated:    entity.Paginated{},
						IssueMatchId: []*int64{&rnd.Id.Int64},
					}

					entries, err := db.GetComponentInstances(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(ciId))
					})

				})
				It("can filter by a single service id that does exist", func() {
					cir := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]
					filter := &entity.ComponentInstanceFilter{
						Paginated: entity.Paginated{},
						ServiceId: []*int64{&cir.ServiceId.Int64},
					}

					entries, err := db.GetComponentInstances(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(Not(BeZero()))
					})

					By("returning expected elements", func() {
						for i := range entries {
							Expect(entries[i].ServiceId).To(BeEquivalentTo(cir.ServiceId.Int64))
						}
					})

				})
				It("can filter by all existing issue match ids ", func() {
					expectedComponentInstances, ids := seedCollection.GetComponentInstanceByIssueMatches(seedCollection.IssueMatchRows)
					filter := &entity.ComponentInstanceFilter{IssueMatchId: ids}

					entries, err := db.GetComponentInstances(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected result count", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(expectedComponentInstances)))
					})
				})
			})
			Context("and using Pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfList(
						db.GetComponentInstances,
						func(first *int, after *int64) *entity.ComponentInstanceFilter {
							return &entity.ComponentInstanceFilter{
								Paginated: entity.Paginated{
									First: first,
									After: after,
								},
							}
						},
						func(entries []entity.ComponentInstance) *int64 { return &entries[len(entries)-1].Id },
						10,
						pageSize,
					)
				},
					Entry("when x is 1", 1),
					Entry("when x is 3", 3),
					Entry("when x is 5", 5),
					Entry("when x is 11", 11),
					Entry("when x is 100", 100),
				)
			})
		})
	})
	When("Counting ComponentInstances", Label("CountComponentInstance"), func() {
		Context("and the database is empty", func() {
			It("returns a correct totalCount without an error", func() {
				c, err := db.CountComponentInstances(nil)

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
			var componentInstanceRows []mariadb.ComponentInstanceRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				componentInstanceRows = seedCollection.GetValidComponentInstanceRows()
				count = len(componentInstanceRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountComponentInstances(nil)

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
					filter := &entity.ComponentInstanceFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
						IssueMatchId: nil,
					}
					c, err := db.CountComponentInstances(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})
			Context("and using a filter", func() {
				DescribeTable("does return totalCount of applied filter", func(pageSize int, filterMatches int) {

					imCol := seedCollection.IssueMatchRows[:filterMatches]
					expectedComponentInstances, ids := seedCollection.GetComponentInstanceByIssueMatches(imCol)

					filter := &entity.ComponentInstanceFilter{
						Paginated: entity.Paginated{
							First: &pageSize,
							After: nil,
						},
						IssueMatchId: ids,
					}
					entries, err := db.CountComponentInstances(filter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(entries).To(BeEquivalentTo(len(expectedComponentInstances)))
					})
					Expect(err).To(BeNil(), "No error should be thrown")
				},
					Entry("when pageSize is 1 and it has 13 elements", 1, 13),
					Entry("when pageSize is 20 and it has 5 elements", 20, 5),
					Entry("when pageSize is 100 and it has 100 elements", 100, 100),
				)
			})
		})

	})
	When("Insert ComponentInstance", Label("InsertComponentInstance"), func() {
		Context("and we have 10 ComponentInstances in the database", func() {
			var newComponentInstanceRow mariadb.ComponentInstanceRow
			var newComponentInstance entity.ComponentInstance
			var componentVersion entity.ComponentVersion
			var service entity.Service
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newComponentInstanceRow = test.NewFakeComponentInstance()
				componentVersion = seedCollection.ComponentVersionRows[0].AsComponentVersion()
				service = seedCollection.ServiceRows[0].AsService()
				newComponentInstance = newComponentInstanceRow.AsComponentInstance()
				newComponentInstance.ComponentVersionId = componentVersion.Id
				newComponentInstance.ServiceId = service.Id
			})
			It("can insert correctly", func() {
				componentInstance, err := db.CreateComponentInstance(&newComponentInstance)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets componentInstance id", func() {
					Expect(componentInstance).NotTo(BeEquivalentTo(0))
				})

				componentInstanceFilter := &entity.ComponentInstanceFilter{
					Id: []*int64{&componentInstance.Id},
				}

				ci, err := db.GetComponentInstances(componentInstanceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning componentInstance", func() {
					Expect(len(ci)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(ci[0].CCRN).To(BeEquivalentTo(componentInstance.CCRN))
					Expect(ci[0].Count).To(BeEquivalentTo(componentInstance.Count))
					Expect(ci[0].ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
					Expect(ci[0].ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
				})
			})
		})
	})
	When("Update ComponentInstance", Label("UpdateComponentInstance"), func() {
		Context("and we have 10 ComponentInstances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update componentInstance count correctly", func() {
				componentInstance := seedCollection.ComponentInstanceRows[0].AsComponentInstance()

				componentInstance.Count = componentInstance.Count + 1
				err := db.UpdateComponentInstance(&componentInstance)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				componentInstanceFilter := &entity.ComponentInstanceFilter{
					Id: []*int64{&componentInstance.Id},
				}

				ci, err := db.GetComponentInstances(componentInstanceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning componentInstance", func() {
					Expect(len(ci)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(ci[0].CCRN).To(BeEquivalentTo(componentInstance.CCRN))
					Expect(ci[0].Count).To(BeEquivalentTo(componentInstance.Count))
					Expect(ci[0].ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
					Expect(ci[0].ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
				})
			})
		})
	})
	When("Delete ComponentInstance", Label("DeleteComponentInstance"), func() {
		Context("and we have 10 ComponentInstances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete componentInstance correctly", func() {
				componentInstance := seedCollection.ComponentInstanceRows[0].AsComponentInstance()

				err := db.DeleteComponentInstance(componentInstance.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				componentInstanceFilter := &entity.ComponentInstanceFilter{
					Id: []*int64{&componentInstance.Id},
				}

				ci, err := db.GetComponentInstances(componentInstanceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no service", func() {
					Expect(len(ci)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Getting CCRN", Label("GetCCRN"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetCcrn(nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 CCRNs in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetCcrn(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentInstanceRows)))
					})

					existingCCRN := lo.Map(seedCollection.ComponentInstanceRows, func(s mariadb.ComponentInstanceRow, index int) string {
						return s.CCRN.String
					})

					By("returning the correct CCRN", func() {
						left, right := lo.Difference(res, existingCCRN)
						Expect(left).Should(BeEmpty())
						Expect(right).Should(BeEmpty())
					})
				})
			})
			Context("and using a CCRN filter", func() {

				var filter *entity.ComponentInstanceFilter
				var expectedCCRN []string
				BeforeEach(func() {
					namePointers := []*string{}

					ccrnValue := "ca9d963d-b441-4167-b08d-086e76186653"
					namePointers = append(namePointers, &ccrnValue)

					filter = &entity.ComponentInstanceFilter{
						CCRN: namePointers,
					}

					It("can fetch the filtered items correctly", func() {
						res, err := db.GetCcrn(filter)

						By("throwing no error", func() {
							Expect(err).Should(BeNil())
						})

						By("returning the correct number of results", func() {
							Expect(len(res)).Should(BeIdenticalTo(len(expectedCCRN)))
						})

						By("returning the correct names", func() {
							left, right := lo.Difference(res, expectedCCRN)
							Expect(left).Should(BeEmpty())
							Expect(right).Should(BeEmpty())
						})
					})
					It("and using another filter", func() {

						var anotherFilter *entity.ComponentInstanceFilter
						BeforeEach(func() {

							nonExistentCCRN := "NonexistentCCRN"

							nonExistentCCRNs := []*string{&nonExistentCCRN}

							anotherFilter = &entity.ComponentInstanceFilter{
								CCRN: nonExistentCCRNs,
							}

							It("returns an empty list when no CCRN match the filter", func() {
								res, err := db.GetCcrn(anotherFilter)
								Expect(err).Should(BeNil())
								Expect(res).Should(BeEmpty())

								By("throwing no error", func() {
									Expect(err).Should(BeNil())
								})

								By("returning an empty list", func() {
									Expect(res).Should(BeEmpty())
								})
							})
						})
					})
				})
			})
		})
	})
})
