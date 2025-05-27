// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
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
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	When("Getting All User IDs", Label("GetAllUserIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllUserIds(nil)
				res = e2e_common.SubtractSystemUserId(res)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list of non-system users", func() {
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
					res = e2e_common.SubtractSystemUserId(res)

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
				res = e2e_common.SubtractSystemUsersEntity(res)

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
					res = e2e_common.SubtractSystemUsersEntity(res)
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
									Expect(r.UniqueUserID).Should(BeEquivalentTo(row.UniqueUserID.String), "Unique User ID matches")
									Expect(r.Type).Should(BeEquivalentTo(row.Type.Int64), "Type matches")
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

					filter := &entity.UserFilter{UniqueUserID: []*string{&row.UniqueUserID.String}}

					entries, err := db.GetUsers(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning expected number of results", func() {
						for _, entry := range entries {
							Expect(entry.UniqueUserID).To(BeEquivalentTo(row.UniqueUserID.String))
						}
					})
				})
				It("can filter by user type", func() {
					humanUserTypeFilter := &entity.UserFilter{Type: []entity.UserType{entity.HumanUserType}}
					humanUserEntries, cErr := db.GetUsers(humanUserTypeFilter)
					By("throwing no error when filtering human user type", func() {
						Expect(cErr).To(BeNil())
					})
					By("returning some results for human user type", func() {
						Expect(humanUserEntries).NotTo(BeEmpty())
					})
					By("returning expected number of human users", func() {
						for _, entry := range humanUserEntries {
							Expect(entry.Type).To(BeEquivalentTo(entity.HumanUserType))
						}
					})

					technicalUserTypeFilter := &entity.UserFilter{Type: []entity.UserType{entity.TechnicalUserType}}
					technicalUserEntries, tErr := db.GetUsers(technicalUserTypeFilter)
					By("throwing no error when filtering technical user type", func() {
						Expect(tErr).To(BeNil())
					})
					By("returning some results for technical user type", func() {
						Expect(technicalUserEntries).NotTo(BeEmpty())
					})
					By("returning expected number of technical users", func() {
						for _, entry := range technicalUserEntries {
							Expect(entry.Type).To(BeEquivalentTo(entity.TechnicalUserType))
						}
					})

					By("number of human and technical user types should match number of all users", func() {
						Expect(e2e_common.SubtractSystemUsers(len(humanUserEntries) + len(technicalUserEntries))).To(BeEquivalentTo(len(seedCollection.UserRows)))
					})
				})
			})
		})
	})
	When("Counting Users", Label("CountUsers"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountUsers(nil)
				c = e2e_common.SubtractSystemUsers(c)

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
					c = e2e_common.SubtractSystemUsers(c)

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
					c = e2e_common.SubtractSystemUsers(c)

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
					Expect(u[0].UniqueUserID).To(BeEquivalentTo(user.UniqueUserID))
					Expect(u[0].Type).To(BeEquivalentTo(user.Type))
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
					Expect(u[0].UniqueUserID).To(BeEquivalentTo(user.UniqueUserID))
					Expect(u[0].Type).To(BeEquivalentTo(user.Type))
				})
			})
			It("can update unique user id correctly", func() {
				user := seedCollection.UserRows[0].AsUser()

				user.UniqueUserID = "D13377331"
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
					Expect(u[0].UniqueUserID).To(BeEquivalentTo(user.UniqueUserID))
					Expect(u[0].Type).To(BeEquivalentTo(user.Type))
				})
			})
			It("can update user type correctly", func() {
				user := seedCollection.UserRows[0].AsUser()

				user.Type = entity.TechnicalUserType
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
					Expect(u[0].UniqueUserID).To(BeEquivalentTo(user.UniqueUserID))
					Expect(u[0].Type).To(BeEquivalentTo(user.Type))
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

				err := db.DeleteUser(user.Id, systemUserId)

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
	When("Getting UserNames", Label("GetUserNames"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetUserNames(nil)
				res = e2e_common.SubtractSystemUserNameVL(res)
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
					res, err := db.GetUserNames(nil)
					res = e2e_common.SubtractSystemUserNameVL(res)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ServiceRows)))
					})

					existingUserNames := lo.Map(seedCollection.UserRows, func(s mariadb.UserRow, index int) string {
						return s.Name.String
					})

					By("returning the correct names", func() {
						left, right := lo.Difference(res, existingUserNames)
						Expect(left).Should(BeEmpty())
						Expect(right).Should(BeEmpty())
					})
				})
			})
			Context("and using a UserNames filter", func() {

				var filter *entity.UserFilter
				var expectedUserNames []string
				BeforeEach(func() {
					namePointers := []*string{}

					name := "f1"
					namePointers = append(namePointers, &name)

					filter = &entity.UserFilter{
						Name: namePointers,
					}

					It("can fetch the filtered items correctly", func() {
						res, err := db.GetUserNames(filter)

						By("throwing no error", func() {
							Expect(err).Should(BeNil())
						})

						By("returning the correct number of results", func() {
							Expect(len(res)).Should(BeIdenticalTo(len(expectedUserNames)))
						})

						By("returning the correct names", func() {
							left, right := lo.Difference(res, expectedUserNames)
							Expect(left).Should(BeEmpty())
							Expect(right).Should(BeEmpty())
						})
					})
					It("and using another filter", func() {

						var anotherFilter *entity.UserFilter
						BeforeEach(func() {

							nonExistentUserName := "NonexistentUserName"

							nonExistentUserNames := []*string{&nonExistentUserName}

							anotherFilter = &entity.UserFilter{
								Name: nonExistentUserNames,
							}

							It("returns an empty list when no users match the filter", func() {
								res, err := db.GetUserNames(anotherFilter)
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
	When("Getting UniqueUserID", Label("GetUniqueUserID"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetUniqueUserIDs(nil)
				res = e2e_common.SubtractSystemUserUniqueUserIdVL(res)
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
					res, err := db.GetUniqueUserIDs(nil)
					res = e2e_common.SubtractSystemUserUniqueUserIdVL(res)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.UserRows)))
					})

					existingUniqueUserID := lo.Map(seedCollection.UserRows, func(s mariadb.UserRow, index int) string {
						return s.UniqueUserID.String
					})

					By("returning the correct UniqueUserID", func() {
						left, right := lo.Difference(res, existingUniqueUserID)
						Expect(left).Should(BeEmpty())
						Expect(right).Should(BeEmpty())
					})
				})
			})
			Context("and using a UniqueUserID filter", func() {

				var filter *entity.UserFilter
				var expectedUniqueUserIDs []string
				BeforeEach(func() {
					uuidPointers := []*string{}

					name := "f1"
					uuidPointers = append(uuidPointers, &name)

					filter = &entity.UserFilter{
						Name: uuidPointers,
					}

					It("can fetch the filtered items correctly", func() {
						res, err := db.GetUniqueUserIDs(filter)

						By("throwing no error", func() {
							Expect(err).Should(BeNil())
						})

						By("returning the correct number of results", func() {
							Expect(len(res)).Should(BeIdenticalTo(len(expectedUniqueUserIDs)))
						})

						By("returning the correct names", func() {
							left, right := lo.Difference(res, expectedUniqueUserIDs)
							Expect(left).Should(BeEmpty())
							Expect(right).Should(BeEmpty())
						})
					})
					It("and using another filter", func() {

						var anotherFilter *entity.UserFilter
						BeforeEach(func() {

							nonExistentUniqueUserID := "NonexistentUniqueUserIDs"

							nonExistentUniqueUserIDs := []*string{&nonExistentUniqueUserID}

							anotherFilter = &entity.UserFilter{
								Name: nonExistentUniqueUserIDs,
							}

							It("returns an empty list when no users match the filter", func() {
								res, err := db.GetUniqueUserIDs(anotherFilter)
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
