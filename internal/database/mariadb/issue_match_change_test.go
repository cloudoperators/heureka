// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IssueMatchChange", Label("database", "IssueMatchChange"), func() {
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

	When("Getting All IssueMatchChange IDs", Label("GetAllIssueMatchChangeIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllIssueMatchChangeIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 IssueMatchChanges in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, vmc := range seedCollection.IssueMatchChangeRows {
					ids = append(ids, vmc.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllIssueMatchChangeIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueMatchChangeRows)))
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
				It("can filter by a single issueMatchChange id that does exist", func() {
					imcId := ids[rand.Intn(len(ids))]
					filter := &entity.IssueMatchChangeFilter{
						Id: []*int64{&imcId},
					}

					entries, err := db.GetAllIssueMatchChangeIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(imcId))
					})
				})
			})
		})
	})

	When("Getting IssueMatchChanges", Label("GetIssueMatchChanges"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetIssueMatchChanges(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 IssueMatchChanges in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			Context("and using no filter", func() {

				It("can fetch the items correctly", func() {
					res, err := db.GetIssueMatchChanges(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueMatchChangeRows)))
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
							for _, row := range seedCollection.IssueMatchChangeRows {
								if r.Id == row.Id.Int64 {
									Expect(r.Action).Should(BeEquivalentTo(row.Action.String), "Action matches")
									Expect(r.CreatedAt.Unix()).ShouldNot(BeEquivalentTo(row.CreatedAt.Time.Unix()), "CreatedAt got set")
									Expect(r.UpdatedAt.Unix()).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time.Unix()), "UpdatedAt got set")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single issue match change id that does exist", func() {
					imc := seedCollection.IssueMatchChangeRows[rand.Intn(len(seedCollection.IssueMatchChangeRows))]
					filter := &entity.IssueMatchChangeFilter{
						Id: []*int64{&imc.Id.Int64},
					}

					entries, err := db.GetIssueMatchChanges(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(imc.Id.Int64))
					})
				})
				It("can filter by a single issue match id that does exist", func() {
					vmc := seedCollection.IssueMatchChangeRows[rand.Intn(len(seedCollection.IssueMatchChangeRows))]
					filter := &entity.IssueMatchChangeFilter{
						Paginated:    entity.Paginated{},
						IssueMatchId: []*int64{&vmc.IssueMatchId.Int64},
					}

					var imcIds []int64
					for _, e := range seedCollection.IssueMatchChangeRows {
						if e.IssueMatchId.Int64 == vmc.IssueMatchId.Int64 {
							imcIds = append(imcIds, e.Id.Int64)
						}
					}

					entries, err := db.GetIssueMatchChanges(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(imcIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(imcIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single activity id that does exist", func() {
					imc := seedCollection.IssueMatchChangeRows[rand.Intn(len(seedCollection.IssueMatchChangeRows))]
					filter := &entity.IssueMatchChangeFilter{
						Paginated:  entity.Paginated{},
						ActivityId: []*int64{&imc.ActivityId.Int64},
					}

					var imcIds []int64
					for _, e := range seedCollection.IssueMatchChangeRows {
						if e.ActivityId.Int64 == imc.ActivityId.Int64 {
							imcIds = append(imcIds, e.Id.Int64)
						}
					}

					entries, err := db.GetIssueMatchChanges(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(imcIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(imcIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single action that does exist", func() {
					imc := seedCollection.IssueMatchChangeRows[rand.Intn(len(seedCollection.IssueMatchChangeRows))]
					filter := &entity.IssueMatchChangeFilter{
						Paginated: entity.Paginated{},
						Action:    []*string{&imc.Action.String},
					}

					var imcIds []int64
					for _, e := range seedCollection.IssueMatchChangeRows {
						if e.Action.String == imc.Action.String {
							imcIds = append(imcIds, e.Id.Int64)
						}
					}

					entries, err := db.GetIssueMatchChanges(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(imcIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(imcIds, entry.Id)).To(BeTrue())
						}
					})
				})
				Context("and and we use Pagination", func() {
					DescribeTable("can correctly paginate ", func(pageSize int) {
						test.TestPaginationOfList(
							db.GetIssueMatchChanges,
							func(first *int, after *int64) *entity.IssueMatchChangeFilter {
								return &entity.IssueMatchChangeFilter{
									Paginated: entity.Paginated{
										First: first,
										After: after,
									},
								}
							},
							func(entries []entity.IssueMatchChange) *int64 { return &entries[len(entries)-1].Id },
							len(seedCollection.IssueMatchChangeRows),
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
	})
	When("Counting IssueMatchChanges", Label("CountIssueMatchChanges"), func() {
		Context("and using no filter", func() {
			DescribeTable("it returns correct count", func(x int) {
				_ = seeder.SeedDbWithNFakeData(x)
				res, err := db.CountIssueMatchChanges(nil)

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
					filter := &entity.IssueMatchChangeFilter{
						Paginated: entity.Paginated{
							First: &first,
							After: &after,
						},
					}
					res, err := db.CountIssueMatchChanges(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(res).To(BeEquivalentTo(20))
					})
				})
				It("does show the correct amount when filtering for a issue match", func() {
					vmc := seedCollection.IssueMatchChangeRows[rand.Intn(len(seedCollection.IssueMatchChangeRows))]
					filter := &entity.IssueMatchChangeFilter{
						Paginated:    entity.Paginated{},
						IssueMatchId: []*int64{&vmc.IssueMatchId.Int64},
					}

					var imcIds []int64
					for _, e := range seedCollection.IssueMatchChangeRows {
						if e.IssueMatchId.Int64 == vmc.IssueMatchId.Int64 {
							imcIds = append(imcIds, e.Id.Int64)
						}
					}
					count, err := db.CountIssueMatchChanges(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(count).To(BeEquivalentTo(len(imcIds)))
					})
				})
			})
		})
	})
	When("Insert IssueMatchChange", Label("InsertIssueMatchChange"), func() {
		Context("and we have 10 IssueMatchChanges in the database", func() {
			var newIssueMatchChangeRow mariadb.IssueMatchChangeRow
			var newIssueMatchChange entity.IssueMatchChange
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newIssueMatchChangeRow = test.NewFakeIssueMatchChange()
				newIssueMatchChange = newIssueMatchChangeRow.AsIssueMatchChange()
				newIssueMatchChange.ActivityId = seedCollection.ActivityRows[0].Id.Int64
				newIssueMatchChange.IssueMatchId = seedCollection.IssueMatchRows[0].Id.Int64
			})
			It("can insert correctly", func() {
				issueMatchChange, err := db.CreateIssueMatchChange(&newIssueMatchChange)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets issueMatchChange id", func() {
					Expect(issueMatchChange).NotTo(BeEquivalentTo(0))
				})

				issueMatchChangeFilter := &entity.IssueMatchChangeFilter{
					Id: []*int64{&issueMatchChange.Id},
				}

				imcs, err := db.GetIssueMatchChanges(issueMatchChangeFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issueMatchChange", func() {
					Expect(len(imcs)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(imcs[0].Action).To(BeEquivalentTo(issueMatchChange.Action))
					Expect(imcs[0].ActivityId).To(BeEquivalentTo(issueMatchChange.ActivityId))
					Expect(imcs[0].IssueMatchId).To(BeEquivalentTo(issueMatchChange.IssueMatchId))
				})
			})
		})
	})
	When("Update IssueMatchChange", Label("UpdateIssueMatchChange"), func() {
		Context("and we have 10 IssueMatchChanges in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update issueMatchChange action correctly", func() {
				imc := seedCollection.IssueMatchChangeRows[0].AsIssueMatchChange()

				if imc.Action == entity.IssueMatchChangeActionAdd.String() {
					imc.Action = entity.IssueMatchChangeActionRemove.String()
				} else {
					imc.Action = entity.IssueMatchChangeActionAdd.String()
				}

				err := db.UpdateIssueMatchChange(&imc)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				imcFilter := &entity.IssueMatchChangeFilter{
					Id: []*int64{&imc.Id},
				}

				imcs, err := db.GetIssueMatchChanges(imcFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issueMatchChanges", func() {
					Expect(len(imcs)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(imcs[0].Action).To(BeEquivalentTo(imc.Action))
				})
			})
		})
	})
	When("Delete IssueMatchChange", Label("DeleteIssueMatchChange"), func() {
		Context("and we have 10 IssueMatchChanges in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete issueMatchChange correctly", func() {
				imc := seedCollection.IssueMatchChangeRows[0].AsIssueMatchChange()

				err := db.DeleteIssueMatchChange(imc.Id, util.SystemUserId)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				imcFilter := &entity.IssueMatchChangeFilter{
					Id: []*int64{&imc.Id},
				}

				imcs, err := db.GetIssueMatchChanges(imcFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no issueMatchChanges", func() {
					Expect(len(imcs)).To(BeEquivalentTo(0))
				})
			})
		})
	})
})
