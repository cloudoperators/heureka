// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/samber/lo"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/common"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	AfterEach(func() {
		dbm.TestTearDown(db)
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
				res, err := db.GetSupportGroups(nil, nil)

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
					res, err := db.GetSupportGroups(nil, nil)

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
									Expect(r.CCRN).Should(BeEquivalentTo(row.CCRN.String), "CCRN matches")
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

					entries, err := db.GetSupportGroups(filter, nil)

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

					entries, err := db.GetSupportGroups(filter, nil)

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

					entries, err := db.GetSupportGroups(filter, nil)

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
				It("can filter by a single support group ccrn", func() {
					row := seedCollection.SupportGroupRows[rand.Intn(len(seedCollection.SupportGroupRows))]

					filter := &entity.SupportGroupFilter{CCRN: []*string{&row.CCRN.String}}

					entries, err := db.GetSupportGroups(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning expected number of results", func() {
						for _, entry := range entries {
							Expect(entry.CCRN).To(BeEquivalentTo(row.CCRN.String))
						}
					})
				})
			})
			Context("and using ordering", func() {
				c := collate.New(language.English)
				var testOrder = func(
					order []entity.Order,
					verifyFunc func(res []entity.SupportGroupResult),
				) {
					res, err := db.GetSupportGroups(nil, order)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.SupportGroupRows)))
					})

					By("returning the correct order", func() {
						verifyFunc(res)
					})
				}

				Context("and using asc order", func() {
					It("can order by ccrn", func() {
						order := []entity.Order{
							{
								By:        entity.SupportGroupCcrn,
								Direction: entity.OrderDirectionAsc,
							},
						}
						testOrder(order, func(res []entity.SupportGroupResult) {
							var prev string = ""
							for _, r := range res {
								Expect(c.CompareString(r.SupportGroup.CCRN, prev)).Should(BeNumerically(">=", 0))
								prev = r.SupportGroup.CCRN
							}
						})

					})

				})
				Context("and using desc order", func() {
					It("can order by ccrn", func() {
						order := []entity.Order{
							{
								By:        entity.SupportGroupCcrn,
								Direction: entity.OrderDirectionDesc,
							},
						}
						testOrder(order, func(res []entity.SupportGroupResult) {
							var prev string = "\U0010FFFF"
							for _, r := range res {
								Expect(c.CompareString(r.SupportGroup.CCRN, prev)).Should(BeNumerically("<=", 0))
								prev = r.SupportGroup.CCRN
							}
						})
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
					after := ""
					filter := &entity.SupportGroupFilter{
						PaginatedX: entity.PaginatedX{
							First: &f,
							After: &after,
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

				sg, err := db.GetSupportGroups(sgFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning supportGroup", func() {
					Expect(len(sg)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(sg[0].Id).To(BeEquivalentTo(supportGroup.Id))
					Expect(sg[0].CCRN).To(BeEquivalentTo(supportGroup.CCRN))
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
			It("can update ccrn correctly", func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()

				supportGroup.CCRN = "Team Alone"
				err := db.UpdateSupportGroup(&supportGroup)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				supportGroupFilter := &entity.SupportGroupFilter{
					Id: []*int64{&supportGroup.Id},
				}

				sg, err := db.GetSupportGroups(supportGroupFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning supportGroup", func() {
					Expect(len(sg)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(sg[0].Id).To(BeEquivalentTo(supportGroup.Id))
					Expect(sg[0].CCRN).To(BeEquivalentTo(supportGroup.CCRN))
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

				err := db.DeleteSupportGroup(supportGroup.Id, common.SystemUserId)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				sgFilter := &entity.SupportGroupFilter{
					Id: []*int64{&supportGroup.Id},
				}

				sg, err := db.GetSupportGroups(sgFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no supportGroup", func() {
					Expect(len(sg)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Add Service To SupportGroup", Label("AddServiceToSupportGroup"), func() {
		Context("and we have 10 SupportGroups in the database", func() {
			var seedCollection *test.SeedCollection
			var newServiceRow mariadb.ServiceRow
			var newService entity.Service
			var service *entity.Service
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newServiceRow = test.NewFakeService()
				newService = newServiceRow.AsService()
				service, _ = db.CreateService(&newService)
			})
			It("can add service correctly", func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()

				err := db.AddServiceToSupportGroup(supportGroup.Id, service.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				supportGroupFilter := &entity.SupportGroupFilter{
					ServiceId: []*int64{&service.Id},
				}

				sg, err := db.GetSupportGroups(supportGroupFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning supportGroup", func() {
					Expect(len(sg)).To(BeEquivalentTo(1))
				})
			})
		})
	})
	When("Remove Service From SupportGroup", Label("RemoveServiceFromSupportGroup"), func() {
		Context("and we have 10 SupportGroups in the database", func() {
			var seedCollection *test.SeedCollection
			var supportGroupServiceRow mariadb.SupportGroupServiceRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				supportGroupServiceRow = seedCollection.SupportGroupServiceRows[0]
			})
			It("can remove service correctly", func() {
				err := db.RemoveServiceFromSupportGroup(supportGroupServiceRow.SupportGroupId.Int64, supportGroupServiceRow.ServiceId.Int64)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				supportGroupFilter := &entity.SupportGroupFilter{
					ServiceId: []*int64{&supportGroupServiceRow.ServiceId.Int64},
				}

				supportGroups, err := db.GetSupportGroups(supportGroupFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				for _, sg := range supportGroups {
					Expect(sg.Id).ToNot(BeEquivalentTo(supportGroupServiceRow.SupportGroupId.Int64))
				}
			})
		})
	})
	When("Add User To SupportGroup", Label("AddUserToSupportGroup"), func() {
		Context("and we have 10 SupportGroups in the database", func() {
			var seedCollection *test.SeedCollection
			var newUserRow mariadb.UserRow
			var newUser entity.User
			var user *entity.User
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newUserRow = test.NewFakeUser()
				newUser = newUserRow.AsUser()
				user, _ = db.CreateUser(&newUser)
			})
			It("can add user correctly", func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()

				err := db.AddUserToSupportGroup(supportGroup.Id, user.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				supportGroupFilter := &entity.SupportGroupFilter{
					UserId: []*int64{&user.Id},
				}

				sg, err := db.GetSupportGroups(supportGroupFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning supportGroup", func() {
					Expect(len(sg)).To(BeEquivalentTo(1))
				})
			})
		})
	})
	When("Remove Service From SupportGroup", Label("RemoveUserFromSupportGroup"), func() {
		Context("and we have 10 SupportGroups in the database", func() {
			var seedCollection *test.SeedCollection
			var supportGroupUserRow mariadb.SupportGroupUserRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				supportGroupUserRow = seedCollection.SupportGroupUserRows[0]
			})
			It("can remove user correctly", func() {
				err := db.RemoveUserFromSupportGroup(supportGroupUserRow.SupportGroupId.Int64, supportGroupUserRow.UserId.Int64)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				supportGroupFilter := &entity.SupportGroupFilter{
					UserId: []*int64{&supportGroupUserRow.UserId.Int64},
				}

				supportGroups, err := db.GetSupportGroups(supportGroupFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				for _, sg := range supportGroups {
					Expect(sg.Id).ToNot(BeEquivalentTo(supportGroupUserRow.SupportGroupId.Int64))
				}
			})
		})
	})
	When("Getting SupportGroupCcrns", Label("GetSupportGroupCcrns"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetSupportGroupCcrns(nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 services in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetSupportGroupCcrns(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.SupportGroupRows)))
					})

					existingSupportGroupCcrns := lo.Map(seedCollection.SupportGroupRows, func(s mariadb.SupportGroupRow, index int) string {
						return s.CCRN.String
					})

					By("returning the correct ccrns", func() {
						left, right := lo.Difference(res, existingSupportGroupCcrns)
						Expect(left).Should(BeEmpty())
						Expect(right).Should(BeEmpty())
					})
				})
			})
			Context("and using a SupportGroupCcrns filter", func() {

				var filter *entity.SupportGroupFilter
				var expectedSupportGroupCcrns []string
				BeforeEach(func() {
					ccrnPointers := []*string{}

					ccrn := "f1"
					ccrnPointers = append(ccrnPointers, &ccrn)

					filter = &entity.SupportGroupFilter{
						CCRN: ccrnPointers,
					}

					It("can fetch the filtered items correctly", func() {
						res, err := db.GetSupportGroupCcrns(filter)

						By("throwing no error", func() {
							Expect(err).Should(BeNil())
						})

						By("returning the correct number of results", func() {
							Expect(len(res)).Should(BeIdenticalTo(len(expectedSupportGroupCcrns)))
						})

						By("returning the correct ccrns", func() {
							left, right := lo.Difference(res, expectedSupportGroupCcrns)
							Expect(left).Should(BeEmpty())
							Expect(right).Should(BeEmpty())
						})
					})
					It("and using another filter", func() {

						var anotherFilter *entity.SupportGroupFilter
						BeforeEach(func() {

							nonExistentSupportGroupCcrn := "NonexistentService"

							nonExistentSupportGroupCcrns := []*string{&nonExistentSupportGroupCcrn}

							anotherFilter = &entity.SupportGroupFilter{
								CCRN: nonExistentSupportGroupCcrns,
							}

							It("returns an empty list when no supportGroup match the filter", func() {
								res, err := db.GetSupportGroupCcrns(anotherFilter)
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
