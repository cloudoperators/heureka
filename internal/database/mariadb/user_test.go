// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb/test"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

var _ = Describe("User", Label("database", "User"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Getting All User IDs", Label("GetAllUserIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllUserIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Users in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, u := range seedCollection.UserRows {
					ids = append(ids, u.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllUserIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.UserRows)))
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
				It("can filter by a single user id that does exist", func() {
					uId := ids[rand.Intn(len(ids))]
					filter := &entity.UserFilter{
						Id: []*int64{&uId},
					}

					entries, err := db.GetAllUserIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(uId))
					})
				})
			})
		})
	})

	When("Getting Users", Label("GetUsers"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetUsers(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 Users in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			Context("and using no filter", func() {

				It("can fetch the items correctly", func() {
					res, err := db.GetUsers(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.UserRows)))
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
							for _, row := range seedCollection.UserRows {
								if r.Id == row.Id.Int64 {
									Expect(r.Name).Should(BeEquivalentTo(row.Name.String), "Name matches")
									Expect(r.SapID).Should(BeEquivalentTo(row.SapID.String), "SAP ID matches")
									Expect(r.CreatedAt.Unix()).ShouldNot(BeEquivalentTo(row.CreatedAt.Time.Unix()), "CreatedAt got set")
									Expect(r.UpdatedAt.Unix()).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time.Unix()), "UpdatedAt got set")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single user id that does exist", func() {
					user := seedCollection.UserRows[rand.Intn(len(seedCollection.UserRows))]
					filter := &entity.UserFilter{
						Id: []*int64{&user.Id.Int64},
					}

					entries, err := db.GetUsers(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(user.Id.Int64))
					})
				})
				It("can filter by a single support group id that does exist", func() {
					sgu := seedCollection.SupportGroupUserRows[rand.Intn(len(seedCollection.SupportGroupUserRows))]
					filter := &entity.UserFilter{
						SupportGroupId: []*int64{&sgu.SupportGroupId.Int64},
					}

					var userIds []int64
					for _, e := range seedCollection.SupportGroupUserRows {
						if e.SupportGroupId.Int64 == sgu.SupportGroupId.Int64 {
							userIds = append(userIds, e.UserId.Int64)
						}
					}

					entries, err := db.GetUsers(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(userIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(userIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single service id that does exist", func() {
					owner := seedCollection.OwnerRows[rand.Intn(len(seedCollection.OwnerRows))]
					filter := &entity.UserFilter{
						ServiceId: []*int64{&owner.ServiceId.Int64},
					}

					var userIds []int64
					for _, e := range seedCollection.OwnerRows {
						if e.ServiceId.Int64 == owner.ServiceId.Int64 {
							userIds = append(userIds, e.UserId.Int64)
						}
					}

					entries, err := db.GetUsers(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(userIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(userIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single user name", func() {
					row := seedCollection.UserRows[rand.Intn(len(seedCollection.UserRows))]

					filter := &entity.UserFilter{Name: []*string{&row.Name.String}}

					entries, err := db.GetUsers(filter)

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
				It("can filter by a single sap id", func() {
					row := seedCollection.UserRows[rand.Intn(len(seedCollection.UserRows))]

					filter := &entity.UserFilter{SapID: []*string{&row.SapID.String}}

					entries, err := db.GetUsers(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning expected number of results", func() {
						for _, entry := range entries {
							Expect(entry.SapID).To(BeEquivalentTo(row.SapID.String))
						}
					})
				})
			})
		})
	})
	When("Counting Users", Label("CountUsers"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountUsers(nil)

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
			var sgRows []mariadb.UserRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				sgRows = seedCollection.UserRows
				count = len(sgRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountUsers(nil)

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
					filter := &entity.UserFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountUsers(filter)

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
	When("Insert User", Label("InsertUser"), func() {
		Context("and we have 10 Users in the database", func() {
			var newUserRow mariadb.UserRow
			var newUser entity.User
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newUserRow = test.NewFakeUser()
				newUser = newUserRow.AsUser()
			})
			It("can insert correctly", func() {
				user, err := db.CreateUser(&newUser)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets user id", func() {
					Expect(user).NotTo(BeEquivalentTo(0))
				})

				userFilter := &entity.UserFilter{
					Id: []*int64{&user.Id},
				}

				u, err := db.GetUsers(userFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning user", func() {
					Expect(len(u)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(u[0].Id).To(BeEquivalentTo(user.Id))
					Expect(u[0].Name).To(BeEquivalentTo(user.Name))
					Expect(u[0].SapID).To(BeEquivalentTo(user.SapID))
				})
			})
			It("does not insert user with existing sap id", func() {
				userRow := seedCollection.UserRows[0]
				user := userRow.AsUser()
				newUser, err := db.CreateUser(&user)

				By("throwing error", func() {
					Expect(err).ToNot(BeNil())
				})
				By("no user returned", func() {
					Expect(newUser).To(BeNil())
				})

			})
		})
	})
	When("Update User", Label("UpdateUser"), func() {
		Context("and we have 10 Users in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update name correctly", func() {
				user := seedCollection.UserRows[0].AsUser()

				user.Name = "Sauron"
				err := db.UpdateUser(&user)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				userFilter := &entity.UserFilter{
					Id: []*int64{&user.Id},
				}

				u, err := db.GetUsers(userFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning user", func() {
					Expect(len(u)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(u[0].Id).To(BeEquivalentTo(user.Id))
					Expect(u[0].Name).To(BeEquivalentTo(user.Name))
					Expect(u[0].SapID).To(BeEquivalentTo(user.SapID))
				})
			})
			It("can update sapId correctly", func() {
				user := seedCollection.UserRows[0].AsUser()

				user.SapID = "D13377331"
				err := db.UpdateUser(&user)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				userFilter := &entity.UserFilter{
					Id: []*int64{&user.Id},
				}

				u, err := db.GetUsers(userFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning user", func() {
					Expect(len(u)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(u[0].Id).To(BeEquivalentTo(user.Id))
					Expect(u[0].Name).To(BeEquivalentTo(user.Name))
					Expect(u[0].SapID).To(BeEquivalentTo(user.SapID))
				})
			})
		})
	})
	When("Delete User", Label("DeleteUser"), func() {
		Context("and we have 10 Users in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete user correctly", func() {
				user := seedCollection.UserRows[0].AsUser()

				err := db.DeleteUser(user.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				userFilter := &entity.UserFilter{
					Id: []*int64{&user.Id},
				}

				u, err := db.GetUsers(userFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no user", func() {
					Expect(len(u)).To(BeEquivalentTo(0))
				})
			})
		})
	})
})
