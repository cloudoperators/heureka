// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/samber/lo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb/test"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

var _ = Describe("SupportGroup", Label("database", "SupportGroup"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Getting All SupportGroup IDs", Label("GetAllSupportGroupIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllSupportGroupIds(nil)

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

				for _, s := range seedCollection.ServiceRows {
					ids = append(ids, s.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllSupportGroupIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.SupportGroupRows)))
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
				It("can filter by a single support group id that does exist", func() {
					sId := ids[rand.Intn(len(ids))]
					filter := &entity.SupportGroupFilter{
						Id: []*int64{&sId},
					}

					entries, err := db.GetAllSupportGroupIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(sId))
					})
				})
			})
		})
	})

	When("Getting SupportGroups", Label("GetSupportGroups"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetSupportGroups(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 SupportGroups in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			Context("and using no filter", func() {

				It("can fetch the items correctly", func() {
					res, err := db.GetSupportGroups(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.SupportGroupRows)))
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
							for _, row := range seedCollection.SupportGroupRows {
								if r.Id == row.Id.Int64 {
									Expect(r.Name).Should(BeEquivalentTo(row.Name.String), "Name matches")
									Expect(r.CreatedAt.Unix()).ShouldNot(BeEquivalentTo(row.CreatedAt.Time.Unix()), "CreatedAt got set")
									Expect(r.UpdatedAt.Unix()).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time.Unix()), "UpdatedAt got set")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single service id that does exist", func() {
					sgs := seedCollection.SupportGroupServiceRows[rand.Intn(len(seedCollection.SupportGroupServiceRows))]
					filter := &entity.SupportGroupFilter{
						ServiceId: []*int64{&sgs.ServiceId.Int64},
					}

					var sgIds []int64
					for _, e := range seedCollection.SupportGroupServiceRows {
						if e.ServiceId.Int64 == sgs.ServiceId.Int64 {
							sgIds = append(sgIds, e.SupportGroupId.Int64)
						}
					}

					entries, err := db.GetSupportGroups(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(sgIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(sgIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single support group Id", func() {
					row := seedCollection.SupportGroupRows[rand.Intn(len(seedCollection.SupportGroupRows))]
					filter := &entity.SupportGroupFilter{Id: []*int64{&row.Id.Int64}}

					entries, err := db.GetSupportGroups(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entry", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(row.Id.Int64))
					})
				})
				It("can filter by a single user id that does exist", func() {
					sgu := seedCollection.SupportGroupUserRows[rand.Intn(len(seedCollection.SupportGroupUserRows))]
					filter := &entity.SupportGroupFilter{
						UserId: []*int64{&sgu.UserId.Int64},
					}

					var sgIds []int64
					for _, e := range seedCollection.SupportGroupUserRows {
						if e.UserId.Int64 == sgu.UserId.Int64 {
							sgIds = append(sgIds, e.SupportGroupId.Int64)
						}
					}

					entries, err := db.GetSupportGroups(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(sgIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(sgIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single support group name", func() {
					row := seedCollection.SupportGroupRows[rand.Intn(len(seedCollection.SupportGroupRows))]

					filter := &entity.SupportGroupFilter{Name: []*string{&row.Name.String}}

					entries, err := db.GetSupportGroups(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning expected number of results", func() {
						for _, entry := range entries {
							Expect(entry.Name).To(BeEquivalentTo(row.Name.String))
						}
					})
				})
			})
		})
	})
	When("Counting SupportGroups", Label("CountSupportGroups"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountSupportGroups(nil)

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
			var sgRows []mariadb.SupportGroupRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				sgRows = seedCollection.SupportGroupRows
				count = len(sgRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountSupportGroups(nil)

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
					filter := &entity.SupportGroupFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountSupportGroups(filter)

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
	When("Insert SupportGroup", Label("InsertSupportGroup"), func() {
		Context("and we have 10 SupportGroups in the database", func() {
			var newSupportGroupRow mariadb.SupportGroupRow
			var newSupportGroup entity.SupportGroup
			BeforeEach(func() {
				seeder.SeedDbWithNFakeData(10)
				newSupportGroupRow = test.NewFakeSupportGroup()
				newSupportGroup = newSupportGroupRow.AsSupportGroup()
			})
			It("can insert correctly", func() {
				supportGroup, err := db.CreateSupportGroup(&newSupportGroup)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets supportGroup id", func() {
					Expect(supportGroup).NotTo(BeEquivalentTo(0))
				})

				sgFilter := &entity.SupportGroupFilter{
					Id: []*int64{&supportGroup.Id},
				}

				sg, err := db.GetSupportGroups(sgFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning supportGroup", func() {
					Expect(len(sg)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(sg[0].Id).To(BeEquivalentTo(supportGroup.Id))
					Expect(sg[0].Name).To(BeEquivalentTo(supportGroup.Name))
				})
			})
		})
	})
	When("Update SupportGroup", Label("UpdateSupportGroup"), func() {
		Context("and we have 10 SupportGroups in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update name correctly", func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()

				supportGroup.Name = "Team Alone"
				err := db.UpdateSupportGroup(&supportGroup)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				supportGroupFilter := &entity.SupportGroupFilter{
					Id: []*int64{&supportGroup.Id},
				}

				sg, err := db.GetSupportGroups(supportGroupFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning supportGroup", func() {
					Expect(len(sg)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(sg[0].Id).To(BeEquivalentTo(supportGroup.Id))
					Expect(sg[0].Name).To(BeEquivalentTo(supportGroup.Name))
				})
			})
		})
	})
	When("Delete SupportGroup", Label("DeleteSupportGroup"), func() {
		Context("and we have 10 SupportGroups in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete supportGroup correctly", func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()

				err := db.DeleteSupportGroup(supportGroup.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				sgFilter := &entity.SupportGroupFilter{
					Id: []*int64{&supportGroup.Id},
				}

				sg, err := db.GetSupportGroups(sgFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no supportGroup", func() {
					Expect(len(sg)).To(BeEquivalentTo(0))
				})
			})
		})
	})
})
