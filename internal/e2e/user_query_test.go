// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/server"
)

var _ = Describe("Getting Users via API", Label("e2e", "Users"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				Users model.UserConnection `json:"Users"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/user/minimal.graphql",
				map[string]any{
					"filter": map[string]string{},
					"first":  10,
					"after":  "0",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			e2e_common.ExpectNonSystemUserCount(respData.Users.TotalCount, 0)
		})
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context(", no additional filters are present", func() {
			Context("and  a minimal query is performed", Label("minimal.graphql"), func() {
				It("returns correct result count", func() {
					respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						Users model.UserConnection `json:"Users"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/user/minimal.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  5,
							"after":  "0",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())
					e2e_common.ExpectNonSystemUserCount(respData.Users.TotalCount, len(seedCollection.UserRows))
					Expect(len(respData.Users.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {
				respData := struct {
					Users model.UserConnection `json:"Users"`
				}{}
				BeforeEach(func() {
					resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						Users model.UserConnection `json:"Users"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/user/directRelations.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  5,
							"after":  "0",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())

					respData = resp
				})

				It("- returns the correct result count", func() {
					e2e_common.ExpectNonSystemUserCount(respData.Users.TotalCount, len(seedCollection.UserRows))
					Expect(len(respData.Users.Edges)).To(Equal(5))
				})

				It("- returns the expected content", func() {
					// this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage

					for _, user := range respData.Users.Edges {
						Expect(user.Node.ID).ToNot(BeNil(), "user has a ID set")
						Expect(user.Node.Name).ToNot(BeNil(), "user has a name set")
						Expect(user.Node.UniqueUserID).ToNot(BeNil(), "user has a uniqueUserId set")
						Expect(user.Node.Type).ToNot(BeNil(), "user has a type set")
						Expect(user.Node.Email).ToNot(BeNil(), "user has a email set")

						for _, service := range user.Node.Services.Edges {
							Expect(service.Node.ID).ToNot(BeNil(), "Service has a ID set")
							Expect(service.Node.Ccrn).ToNot(BeNil(), "Service has a name set")

							_, serviceFound := lo.Find(seedCollection.OwnerRows, func(row mariadb.OwnerRow) bool {
								return fmt.Sprintf("%d", row.UserId.Int64) == user.Node.ID && // correct user
									fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // references correct service
							})
							Expect(serviceFound).To(BeTrue(), "attached service does exist and belongs to user")
						}

						for _, sg := range user.Node.SupportGroups.Edges {
							Expect(sg.Node.ID).ToNot(BeNil(), "supportGroup has a ID set")
							Expect(sg.Node.Ccrn).ToNot(BeNil(), "supportGroup has a ccrn set")

							_, sgFound := lo.Find(seedCollection.SupportGroupUserRows, func(row mariadb.SupportGroupUserRow) bool {
								return fmt.Sprintf("%d", row.SupportGroupId.Int64) == sg.Node.ID && // correct support group
									fmt.Sprintf("%d", row.UserId.Int64) == user.Node.ID // references correct user
							})
							Expect(sgFound).To(BeTrue(), "attached supportGroup does exist and belongs to user")

						}
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.Users.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.Users.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.Users.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.Users.PageInfo.Pages)).To(Equal(3), "Correct amount of pages")
					Expect(*respData.Users.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
		})
	})
})

var _ = Describe("Creating User via API", Label("e2e", "Users"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var user entity.User
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		BeforeEach(func() {
			seeder.SeedDbWithNFakeData(10)
			user = testentity.NewFakeUserEntity()
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new user", func() {
				respUser := e2e_common.QueryCreateUser(cfg.Port, e2e_common.User{UniqueUserID: user.UniqueUserID, Type: user.Type, Name: user.Name, Email: user.Email})
				Expect(*respUser.Name).To(Equal(user.Name))
				Expect(*respUser.UniqueUserID).To(Equal(user.UniqueUserID))
				Expect(entity.UserType(respUser.Type)).To(Equal(user.Type))
				Expect(*respUser.Email).To(Equal(user.Email))
			})
		})
	})
})

var _ = Describe("Updating User via API", Label("e2e", "Users"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates user", func() {
				user := seedCollection.UserRows[0].AsUser()
				user.Name = "Sauron"
				respUser := e2e_common.QueryUpdateUser(cfg.Port, e2e_common.User{UniqueUserID: user.UniqueUserID, Name: user.Name, Type: user.Type, Email: user.Email}, fmt.Sprintf("%d", user.Id))
				Expect(*respUser.Name).To(Equal(user.Name))
				Expect(*respUser.UniqueUserID).To(Equal(user.UniqueUserID))
				Expect(entity.UserType(respUser.Type)).To(Equal(user.Type))
				Expect(*respUser.Email).To(Equal(user.Email))
			})
		})
	})
})

var _ = Describe("Deleting User via API", Label("e2e", "Users"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes user", func() {
				id := fmt.Sprintf("%d", seedCollection.UserRows[0].Id.Int64)

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Id string `json:"deleteUser"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/user/delete.graphql",
					map[string]any{
						"id": id,
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
