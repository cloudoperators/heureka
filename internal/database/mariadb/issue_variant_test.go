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
)

var _ = Describe("IssueVariant - ", Label("database", "IssueVariant"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Getting All IssueVariant IDs", Label("GetAllIssueVariantIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllIssueVariantIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Issue Variants in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, a := range seedCollection.IssueVariantRows {
					ids = append(ids, a.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllIssueVariantIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueVariantRows)))
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
				It("can filter by a single issue variant id that does exist", func() {
					ivId := ids[rand.Intn(len(ids))]
					filter := &entity.IssueVariantFilter{
						Id: []*int64{&ivId},
					}

					entries, err := db.GetAllIssueVariantIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(ivId))
					})
				})
			})
		})
	})

	When("Getting IssueVariants", Label("GetIssueVariants"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetIssueVariants(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 issue variants in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			Context("and using no filter", func() {

				It("can fetch the items correctly", func() {
					res, err := db.GetIssueVariants(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueVariantRows)))
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
							for _, row := range seedCollection.IssueVariantRows {
								if r.Id == row.Id.Int64 {

									Expect(r.SecondaryName).Should(BeEquivalentTo(row.SecondaryName.String), "Name matches")
									Expect(r.Description).Should(BeEquivalentTo(row.Description.String), "Description matches")
									Expect(r.Severity.Cvss.Vector).Should(BeEquivalentTo(row.Vector.String), "Vector matches")
									Expect(r.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single issue variant id that does exist", func() {
					issueVariant := seedCollection.IssueVariantRows[rand.Intn(len(seedCollection.IssueVariantRows))]
					filter := &entity.IssueVariantFilter{
						Id: []*int64{&issueVariant.Id.Int64},
					}

					entries, err := db.GetIssueVariants(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(issueVariant.Id.Int64))
					})
				})
				It("can filter by a single issue id", func() {
					rnd := seedCollection.IssueVariantRows[rand.Intn(len(seedCollection.IssueVariantRows))]
					issueId := rnd.IssueId.Int64

					issueVariants := seedCollection.GetIssueVariantsByIssueId(issueId)
					filter := &entity.IssueVariantFilter{
						Paginated: entity.Paginated{},
						IssueId:   []*int64{&issueId},
					}

					entries, err := db.GetIssueVariants(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(issueVariants)))
					})

					By("returning expected elements", func() {
						for _, issueVariant := range issueVariants {
							Expect(lo.ContainsBy(entries, func(a entity.IssueVariant) bool { return issueVariant.Id.Int64 == a.Id })).To(BeTrue())
						}
					})
				})
				It("can filter by multiple issue ids ", func() {
					rnd1 := seedCollection.IssueRows[rand.Intn(len(seedCollection.IssueRows))]
					rnd2 := seedCollection.IssueRows[rand.Intn(len(seedCollection.IssueRows))]

					issueId1 := rnd1.Id.Int64
					issueId2 := rnd2.Id.Int64

					issueVariants := seedCollection.GetIssueVariantsByIssueId(issueId1)
					issueVariants = append(issueVariants, seedCollection.GetIssueVariantsByIssueId(issueId2)...)
					issueVariants = lo.UniqBy(issueVariants, func(e mariadb.IssueVariantRow) int64 { return e.Id.Int64 })
					filter := &entity.IssueVariantFilter{
						Paginated: entity.Paginated{},
						IssueId:   []*int64{&issueId1, &issueId2},
					}

					entries, err := db.GetIssueVariants(filter)

					By("throwing no Error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(issueVariants)))
					})

					By("returning expected elements", func() {
						for _, issueVariant := range issueVariants {
							Expect(lo.ContainsBy(entries, func(a entity.IssueVariant) bool { return issueVariant.Id.Int64 == a.Id })).To(BeTrue())
						}
					})
				})
				It("can filter by a single issue repository id", func() {
					ir := seedCollection.IssueVariantRows[rand.Intn(len(seedCollection.IssueVariantRows))]
					filter := &entity.IssueVariantFilter{
						Paginated:         entity.Paginated{},
						IssueRepositoryId: []*int64{&ir.IssueRepositoryId.Int64},
					}

					issueVariants, err := db.GetIssueVariants(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(issueVariants).NotTo(BeEmpty())
					})

					By("returning expected elements", func() {
						for _, issueVariant := range issueVariants {
							Expect(issueVariant.IssueRepositoryId).To(BeEquivalentTo(ir.IssueRepositoryId.Int64))
						}
					})
				})
				It("can filter by a single service id", func() {
					service := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]
					issueVariants := seedCollection.GetIssueVariantsByService(&service)

					filter := &entity.IssueVariantFilter{
						Paginated: entity.Paginated{},
						ServiceId: []*int64{&service.Id.Int64},
					}

					entries, err := db.GetIssueVariants(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(issueVariants)))
					})

					By("returning expected elements", func() {
						for _, issueVariant := range issueVariants {
							Expect(lo.ContainsBy(entries, func(a entity.IssueVariant) bool { return issueVariant.Id.Int64 == a.Id })).To(BeTrue())
						}
					})
				})
				It("can filter by a single issue match id", func() {
					im := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					issueVariants := seedCollection.GetIssueVariantsByIssueMatch(&im)

					filter := &entity.IssueVariantFilter{
						Paginated:    entity.Paginated{},
						IssueMatchId: []*int64{&im.Id.Int64},
					}

					entries, err := db.GetIssueVariants(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(issueVariants)))
					})

					By("returning expected elements", func() {
						for _, issueVariant := range issueVariants {
							Expect(lo.ContainsBy(entries, func(a entity.IssueVariant) bool { return issueVariant.Id.Int64 == a.Id })).To(BeTrue())
						}
					})
				})
				It("can filter by a secondary name", func() {
					iv := seedCollection.IssueVariantRows[rand.Intn(len(seedCollection.IssueVariantRows))]

					filter := &entity.IssueVariantFilter{
						SecondaryName: []*string{&iv.SecondaryName.String},
					}

					entries, err := db.GetIssueVariants(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(entries).To(HaveLen(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].SecondaryName).To(BeEquivalentTo(iv.SecondaryName.String))
					})
				})
			})
			Context("and using Pagination", func() {
				DescribeTable("can correctly paginate", func(pageSize int) {
					test.TestPaginationOfList(
						db.GetIssueVariants,
						func(first *int, after *int64) *entity.IssueVariantFilter {
							return &entity.IssueVariantFilter{
								Paginated: entity.Paginated{
									First: first,
									After: after,
								},
							}
						},
						func(entries []entity.IssueVariant) *int64 { return &entries[len(entries)-1].Id },
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
	When("Counting IssueVariants", Label("CountIssueVariants"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountIssueVariants(nil)

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
			var ivRows []mariadb.IssueVariantRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				ivRows = seedCollection.IssueVariantRows
				count = len(ivRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountIssueVariants(nil)

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
					filter := &entity.IssueVariantFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountIssueVariants(filter)

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

					rnd := seedCollection.IssueVariantRows[rand.Intn(len(seedCollection.IssueVariantRows))]
					issueId := rnd.IssueId.Int64

					issueVariants := seedCollection.GetIssueVariantsByIssueId(issueId)

					filter := &entity.IssueVariantFilter{
						Paginated: entity.Paginated{
							First: &pageSize,
							After: nil,
						},
						IssueId: []*int64{&issueId},
					}
					entries, err := db.CountIssueVariants(filter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(entries).To(BeEquivalentTo(len(issueVariants)))
					})
				},
					Entry("and pageSize is 1 and it has 13 elements", 1, 13),
					Entry("and pageSize is 20 and it has 5 elements", 20, 5),
					Entry("and pageSize is 100 and it has 100 elements", 100, 100),
				)
			})
		})
		When("Insert IssueVariant", Label("InsertIssueVariant"), func() {
			Context("and we have 10 IssueVariants in the database", func() {
				var newIssueVariantRow mariadb.IssueVariantRow
				var newIssueVariant entity.IssueVariant
				var seedCollection *test.SeedCollection
				var issueRepository entity.IssueRepository
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
					newIssueVariantRow = test.NewFakeIssueVariant(seedCollection.IssueRepositoryRows, seedCollection.IssueRows)
					issueRepository = seedCollection.IssueRepositoryRows[0].AsIssueRepository()
					newIssueVariant = newIssueVariantRow.AsIssueVariant(&issueRepository)
				})
				It("can insert correctly", func() {
					issueVariant, err := db.CreateIssueVariant(&newIssueVariant)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("sets issueVariant id", func() {
						Expect(issueVariant).NotTo(BeEquivalentTo(0))
					})

					issueVariantFilter := &entity.IssueVariantFilter{
						Id: []*int64{&issueVariant.Id},
					}

					iv, err := db.GetIssueVariants(issueVariantFilter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning issueVariant", func() {
						Expect(len(iv)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(iv[0].Id).To(BeEquivalentTo(issueVariant.Id))
						Expect(iv[0].SecondaryName).To(BeEquivalentTo(issueVariant.SecondaryName))
						Expect(iv[0].Description).To(BeEquivalentTo(issueVariant.Description))
						Expect(iv[0].IssueRepositoryId).To(BeEquivalentTo(issueVariant.IssueRepositoryId))
						Expect(iv[0].IssueId).To(BeEquivalentTo(issueVariant.IssueId))
						Expect(iv[0].Severity.Cvss.Vector).To(BeEquivalentTo(issueVariant.Severity.Cvss.Vector))
						Expect(iv[0].Severity.Score).To(BeEquivalentTo(issueVariant.Severity.Score))
						Expect(iv[0].Severity.Value).To(BeEquivalentTo(issueVariant.Severity.Value))
					})
				})
				It("does not insert issueVariant with existing name", func() {
					ir := seedCollection.IssueRepositoryRows[0].AsIssueRepository()
					issueVariantRow := seedCollection.IssueVariantRows[0]
					issueVariant := issueVariantRow.AsIssueVariant(&ir)
					newIssueVariant, err := db.CreateIssueVariant(&issueVariant)

					By("throwing error", func() {
						Expect(err).ToNot(BeNil())
					})
					By("no issueVariant returned", func() {
						Expect(newIssueVariant).To(BeNil())
					})
				})
			})
		})
		When("Update IssueVariant", Label("UpdateIssueVariant"), func() {
			Context("and we have 10 IssueVariants in the database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
				})
				It("can update name correctly", func() {
					ir := seedCollection.IssueRepositoryRows[0].AsIssueRepository()
					issueVariant := seedCollection.IssueVariantRows[0].AsIssueVariant(&ir)

					issueVariant.SecondaryName = "NewName"
					err := db.UpdateIssueVariant(&issueVariant)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					issueVariantFilter := &entity.IssueVariantFilter{
						Id: []*int64{&issueVariant.Id},
					}

					iv, err := db.GetIssueVariants(issueVariantFilter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning issueVariant", func() {
						Expect(len(iv)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(iv[0].Id).To(BeEquivalentTo(issueVariant.Id))
						Expect(iv[0].SecondaryName).To(BeEquivalentTo(issueVariant.SecondaryName))
						Expect(iv[0].Description).To(BeEquivalentTo(issueVariant.Description))
						Expect(iv[0].IssueRepositoryId).To(BeEquivalentTo(issueVariant.IssueRepositoryId))
						Expect(iv[0].IssueId).To(BeEquivalentTo(issueVariant.IssueId))
						Expect(iv[0].Severity.Cvss.Vector).To(BeEquivalentTo(issueVariant.Severity.Cvss.Vector))
						Expect(iv[0].Severity.Score).To(BeEquivalentTo(issueVariant.Severity.Score))
						Expect(iv[0].Severity.Value).To(BeEquivalentTo(issueVariant.Severity.Value))
					})
				})
			})
		})
		When("Delete IssueVariant", Label("DeleteIssueVariant"), func() {
			Context("and we have 10 IssueVariants in the database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
				})
				It("can delete issue variants correctly", func() {
					ir := seedCollection.IssueRepositoryRows[0].AsIssueRepository()
					issueVariant := seedCollection.IssueVariantRows[0].AsIssueVariant(&ir)

					err := db.DeleteIssueVariant(issueVariant.Id, systemUserId)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					issueVariantFilter := &entity.IssueVariantFilter{
						Id: []*int64{&issueVariant.Id},
					}

					iv, err := db.GetIssueVariants(issueVariantFilter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning no component", func() {
						Expect(len(iv)).To(BeEquivalentTo(0))
					})
				})
			})
		})
	})
})
