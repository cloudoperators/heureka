// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb/test"
	"github.wdf.sap.corp/cc/heureka/internal/entity"

	"math/rand"
)

var _ = Describe("IssueRepository", Label("database", "IssueRepository"), func() {

	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Getting All IssueRepository IDs", Label("GetAllIssueRepositoryIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllIssueRepositoryIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 IssueRepositories in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, ar := range seedCollection.IssueRepositoryRows {
					ids = append(ids, ar.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllIssueRepositoryIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueRepositoryRows)))
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
				It("can filter by a single issueRepository id that does exist", func() {
					irId := ids[rand.Intn(len(ids))]
					filter := &entity.IssueRepositoryFilter{
						Id: []*int64{&irId},
					}

					entries, err := db.GetAllIssueRepositoryIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(irId))
					})
				})
			})
		})
	})

	When("Getting IssueRepositories", Label("GetIssueRepositories"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetIssueRepositories(nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 issue repositories in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetIssueRepositories(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueRepositoryRows)))
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
							for _, row := range seedCollection.IssueRepositoryRows {
								if r.Id == row.Id.Int64 {
									Expect(r.Name).Should(BeEquivalentTo(row.Name.String), "Name should match")
									Expect(r.Url).Should(BeEquivalentTo(row.Url.String), "URL should match")
									Expect(r.BaseIssueRepository.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.BaseIssueRepository.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single name", func() {
					row := seedCollection.IssueRepositoryRows[rand.Intn(len(seedCollection.IssueRepositoryRows))]
					filter := &entity.IssueRepositoryFilter{Name: []*string{&row.Name.String}}

					entries, err := db.GetIssueRepositories(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the issue repository name", func() {
						for _, entry := range entries {
							Expect(entry.Name).To(BeEquivalentTo(row.Name.String))
						}
					})
				})
				It("can filter by a single service name", func() {
					// select a service
					sRow := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					// collect all issue repository ids that belong to the service
					irIds := []int64{}
					for _, irsRow := range seedCollection.IssueRepositoryServiceRows {
						if irsRow.ServiceId.Int64 == sRow.Id.Int64 {
							irIds = append(irIds, irsRow.IssueRepositoryId.Int64)
						}
					}

					filter := &entity.IssueRepositoryFilter{ServiceName: []*string{&sRow.Name.String}}

					entries, err := db.GetIssueRepositories(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entries", func() {
						for _, entry := range entries {
							Expect(irIds).To(ContainElement(entry.Id))
						}
					})

				})
				It("can filter by a single id", func() {
					// select a service
					sRow := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					// collect all issue repository ids that belong to the service
					irIds := []int64{}
					for _, irsRow := range seedCollection.IssueRepositoryServiceRows {
						if irsRow.ServiceId.Int64 == sRow.Id.Int64 {
							irIds = append(irIds, irsRow.IssueRepositoryId.Int64)
						}
					}

					filter := &entity.IssueRepositoryFilter{ServiceId: []*int64{&sRow.Id.Int64}}

					entries, err := db.GetIssueRepositories(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entries", func() {
						for _, entry := range entries {
							Expect(irIds).To(ContainElement(entry.Id))
						}
					})

				})
			})
			Context("and using pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfList(
						db.GetIssueRepositories,
						func(first *int, after *int64) *entity.IssueRepositoryFilter {
							return &entity.IssueRepositoryFilter{
								Paginated: entity.Paginated{First: first, After: after},
							}
						},
						func(entries []entity.IssueRepository) *int64 { return &entries[len(entries)-1].Id },
						len(seedCollection.IssueRepositoryRows),
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
	When("Counting IssueRepositories", Label("CountIssueRepositories"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountIssueRepositories(nil)

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
			var irRows []mariadb.BaseIssueRepositoryRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				irRows = seedCollection.IssueRepositoryRows
				count = len(irRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountIssueRepositories(nil)

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
					filter := &entity.IssueRepositoryFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountIssueRepositories(filter)

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
					// select a service
					sRow := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					// collect all issue repository ids that belong to the service
					irIds := []int64{}
					for _, irsRow := range seedCollection.IssueRepositoryServiceRows {
						if irsRow.ServiceId.Int64 == sRow.Id.Int64 {
							irIds = append(irIds, irsRow.IssueRepositoryId.Int64)
						}
					}

					filter := &entity.IssueRepositoryFilter{
						Paginated: entity.Paginated{
							First: &pageSize,
							After: nil,
						},
						ServiceName: []*string{&sRow.Name.String},
					}
					entries, err := db.CountIssueRepositories(filter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(entries).To(BeEquivalentTo(len(irIds)))
					})
				},
					Entry("and pageSize is 1 and it has 13 elements", 1, 13),
					Entry("and pageSize is 20 and it has 5 elements", 20, 5),
					Entry("and pageSize is 100 and it has 100 elements", 100, 100),
				)
			})
		})
	})
	When("Insert IssueRepository", Label("InsertIssueRepository"), func() {
		Context("and we have 10 IssueRepositories in the database", func() {
			var newIssueRepositoryRow mariadb.IssueRepositoryRow
			var newIssueRepository entity.IssueRepository
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newIssueRepositoryRow = test.NewFakeIssueRepository()
				newIssueRepository = newIssueRepositoryRow.AsIssueRepository()
			})
			It("can insert correctly", func() {
				issueRepository, err := db.CreateIssueRepository(&newIssueRepository)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets issueRepository id", func() {
					Expect(issueRepository).NotTo(BeEquivalentTo(0))
				})

				issueRepositoryFilter := &entity.IssueRepositoryFilter{
					Id: []*int64{&issueRepository.Id},
				}

				ir, err := db.GetIssueRepositories(issueRepositoryFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issueRepository", func() {
					Expect(len(ir)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(ir[0].Id).To(BeEquivalentTo(issueRepository.Id))
					Expect(ir[0].Name).To(BeEquivalentTo(issueRepository.Name))
					Expect(ir[0].Url).To(BeEquivalentTo(issueRepository.Url))
				})
			})
			It("does not insert issueRepository with existing name", func() {
				issueRepositoryRow := seedCollection.IssueRepositoryRows[0]
				issueRepository := issueRepositoryRow.AsIssueRepository()
				newIssueRepository, err := db.CreateIssueRepository(&issueRepository)

				By("throwing error", func() {
					Expect(err).ToNot(BeNil())
				})
				By("no issueRepository returned", func() {
					Expect(newIssueRepository).To(BeNil())
				})

			})
		})
	})
	When("Update IssueRepository", Label("UpdateIssueRepository"), func() {
		Context("and we have 10 IssueRepositories in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update name correctly", func() {
				issueRepository := seedCollection.IssueRepositoryRows[0].AsIssueRepository()

				issueRepository.Name = "NewName"
				err := db.UpdateIssueRepository(&issueRepository)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueRepositoryFilter := &entity.IssueRepositoryFilter{
					Id: []*int64{&issueRepository.Id},
				}

				ir, err := db.GetIssueRepositories(issueRepositoryFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issueRepository", func() {
					Expect(len(ir)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(ir[0].Id).To(BeEquivalentTo(issueRepository.Id))
					Expect(ir[0].Name).To(BeEquivalentTo(issueRepository.Name))
					Expect(ir[0].Url).To(BeEquivalentTo(issueRepository.Url))
				})
			})
			It("can update url correctly", func() {
				issueRepository := seedCollection.IssueRepositoryRows[0].AsIssueRepository()

				issueRepository.Url = "NewType"
				err := db.UpdateIssueRepository(&issueRepository)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueRepositoryFilter := &entity.IssueRepositoryFilter{
					Id: []*int64{&issueRepository.Id},
				}

				ir, err := db.GetIssueRepositories(issueRepositoryFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issueRepository", func() {
					Expect(len(ir)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(ir[0].Id).To(BeEquivalentTo(issueRepository.Id))
					Expect(ir[0].Name).To(BeEquivalentTo(issueRepository.Name))
					Expect(ir[0].Url).To(BeEquivalentTo(issueRepository.Url))
				})
			})
		})
	})
	When("Delete IssueRepository", Label("DeleteIssueRepository"), func() {
		Context("and we have 10 IssueRepositories in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete issueRepository correctly", func() {
				issueRepository := seedCollection.IssueRepositoryRows[0].AsIssueRepository()

				err := db.DeleteIssueRepository(issueRepository.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueRepositoryFilter := &entity.IssueRepositoryFilter{
					Id: []*int64{&issueRepository.Id},
				}

				ir, err := db.GetIssueRepositories(issueRepositoryFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no issueRepository", func() {
					Expect(len(ir)).To(BeEquivalentTo(0))
				})
			})
		})
	})
})
