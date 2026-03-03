// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Getting SupportGroups via API", Label("e2e", "SupportGroups"), func() {
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
				SupportGroups model.SupportGroupConnection `json:"SupportGroups"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/supportGroup/minimal.graphql",
				map[string]any{
					"filter": map[string]string{},
					"first":  10,
					"after":  "",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.SupportGroups.TotalCount).To(Equal(0))
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
						SupportGroups model.SupportGroupConnection `json:"SupportGroups"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/supportGroup/minimal.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  5,
							"after":  "",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())
					Expect(respData.SupportGroups.TotalCount).To(Equal(len(seedCollection.SupportGroupRows)))
					Expect(len(respData.SupportGroups.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {
				respData := struct {
					SupportGroups model.SupportGroupConnection `json:"SupportGroups"`
				}{}
				BeforeEach(func() {
					resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						SupportGroups model.SupportGroupConnection `json:"SupportGroups"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/supportGroup/directRelations.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  5,
							"after":  "",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())

					respData = resp
				})

				It("- returns the correct result count", func() {
					Expect(respData.SupportGroups.TotalCount).To(Equal(len(seedCollection.SupportGroupRows)))
					Expect(len(respData.SupportGroups.Edges)).To(Equal(5))
				})

				It("- returns the expected content", func() {
					// this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage

					for _, sg := range respData.SupportGroups.Edges {
						Expect(sg.Node.ID).ToNot(BeNil(), "supportGroup has a ID set")
						Expect(sg.Node.Ccrn).ToNot(BeNil(), "supportGroup has a ccrn set")

						for _, service := range sg.Node.Services.Edges {
							Expect(service.Node.ID).ToNot(BeNil(), "Service has a ID set")
							Expect(service.Node.Ccrn).ToNot(BeNil(), "Service has a ccrn set")

							_, serviceFound := lo.Find(seedCollection.SupportGroupServiceRows, func(row mariadb.SupportGroupServiceRow) bool {
								return fmt.Sprintf("%d", row.SupportGroupId.Int64) == sg.Node.ID && // correct support group
									fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // references correct service
							})
							Expect(serviceFound).To(BeTrue(), "attached service does exist and belongs to supportGroup")
						}

						for _, user := range sg.Node.Users.Edges {
							Expect(user.Node.ID).ToNot(BeNil(), "user has a ID set")
							Expect(user.Node.Name).ToNot(BeNil(), "user has a name set")
							Expect(user.Node.UniqueUserID).ToNot(BeNil(), "user has a uniqueUserId set")
							Expect(user.Node.Type).ToNot(BeNil(), "user has a type set")

							_, userFound := lo.Find(seedCollection.SupportGroupUserRows, func(row mariadb.SupportGroupUserRow) bool {
								return fmt.Sprintf("%d", row.SupportGroupId.Int64) == sg.Node.ID && // correct support group
									fmt.Sprintf("%d", row.UserId.Int64) == user.Node.ID // references correct user
							})
							Expect(userFound).To(BeTrue(), "attached user does exist and belongs to supportGroup")

						}
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.SupportGroups.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.SupportGroups.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.SupportGroups.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.SupportGroups.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
					Expect(*respData.SupportGroups.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
		})
		Context("and we use order", Label("withOrder.graphql"), func() {
			c := collate.New(language.English)

			It("can order by ccrn", Label("withOrder.graphql"), func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SupportGroups model.SupportGroupConnection `json:"SupportGroups"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/supportGroup/withOrder.graphql",
					map[string]any{
						"filter": map[string]string{},
						"first":  10,
						"after":  "",
						"orderBy": []map[string]string{
							{"by": "ccrn", "direction": "asc"},
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				By("- returns the correct result count", func() {
					Expect(respData.SupportGroups.TotalCount).To(Equal(len(seedCollection.SupportGroupRows)))
					Expect(len(respData.SupportGroups.Edges)).To(Equal(10))
				})

				By("- returns the expected content in order", func() {
					var prev string = ""
					for _, im := range respData.SupportGroups.Edges {
						Expect(c.CompareString(*im.Node.Ccrn, prev)).Should(BeNumerically(">=", 0))
						prev = *im.Node.Ccrn
					}
				})
			})
		})
	})
})

var _ = Describe("Creating SupportGroup via API", Label("e2e", "SupportGroups"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var supportGroup entity.SupportGroup
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
			supportGroup = testentity.NewFakeSupportGroupEntity()
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new supportGroup", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SupportGroup model.SupportGroup `json:"createSupportGroup"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/supportGroup/create.graphql",
					map[string]any{
						"input": map[string]string{
							"ccrn": supportGroup.CCRN,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.SupportGroup.Ccrn).To(Equal(supportGroup.CCRN))
			})
		})
	})
})

var _ = Describe("Updating SupportGroup via API", Label("e2e", "SupportGroups"), func() {
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
			It("updates supportGroup", func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()
				supportGroup.CCRN = "Team Alone"

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SupportGroup model.SupportGroup `json:"updateSupportGroup"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/supportGroup/update.graphql",
					map[string]any{
						"id": supportGroup.Id,
						"input": map[string]string{
							"ccrn": supportGroup.CCRN,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.SupportGroup.Ccrn).To(Equal(supportGroup.CCRN))
			})
		})
	})
})

var _ = Describe("Deleting SupportGroup via API", Label("e2e", "SupportGroups"), func() {
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
			It("deletes supportGroup", func() {
				id := fmt.Sprintf("%d", seedCollection.SupportGroupRows[0].Id.Int64)

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Id string `json:"deleteSupportGroup"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/supportGroup/delete.graphql",
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

var _ = Describe("Modifying Services of SupportGroup via API", Label("e2e", "SupportGroups"), func() {
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

		Context("and a mutation query is performed", func() {
			It("adds service to supportGroup", Label("addService.graphql"), func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()
				// find all services that are attached to the supportGroup
				serviceIds := lo.FilterMap(seedCollection.SupportGroupServiceRows, func(row mariadb.SupportGroupServiceRow, _ int) (int64, bool) {
					if row.SupportGroupId.Int64 == supportGroup.Id {
						return row.ServiceId.Int64, true
					}
					return 0, false
				})

				// find a service that is not attached to the supportGroup
				serviceRow, _ := lo.Find(seedCollection.ServiceRows, func(row mariadb.BaseServiceRow) bool {
					return !lo.Contains(serviceIds, row.Id.Int64)
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SupportGroup model.SupportGroup `json:"addServiceToSupportGroup"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/supportGroup/addService.graphql",
					map[string]any{
						"supportGroupId": fmt.Sprintf("%d", supportGroup.Id),
						"serviceId":      fmt.Sprintf("%d", serviceRow.Id.Int64),
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				_, found := lo.Find(respData.SupportGroup.Services.Edges, func(edge *model.ServiceEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", serviceRow.Id.Int64)
				})

				Expect(respData.SupportGroup.ID).To(Equal(fmt.Sprintf("%d", supportGroup.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes service from supportGroup", Label("removeService.graphql"), func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()

				// find a service that is attached to the supportGroup
				serviceRow, _ := lo.Find(seedCollection.SupportGroupServiceRows, func(row mariadb.SupportGroupServiceRow) bool {
					return row.SupportGroupId.Int64 == supportGroup.Id
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SupportGroup model.SupportGroup `json:"removeServiceFromSupportGroup"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/supportGroup/removeService.graphql",
					map[string]any{
						"supportGroupId": fmt.Sprintf("%d", supportGroup.Id),
						"serviceId":      fmt.Sprintf("%d", serviceRow.ServiceId.Int64),
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				_, found := lo.Find(respData.SupportGroup.Services.Edges, func(edge *model.ServiceEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", serviceRow.ServiceId.Int64)
				})

				Expect(respData.SupportGroup.ID).To(Equal(fmt.Sprintf("%d", supportGroup.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})

var _ = Describe("Modifying Users of SupportGroup via API", Label("e2e", "SupportGroups"), func() {
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

		Context("and a mutation query is performed", func() {
			It("adds user to supportGroup", Label("addUser.graphql"), func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()
				// find all users that are attached to the supportGroup
				userIds := lo.FilterMap(seedCollection.SupportGroupUserRows, func(row mariadb.SupportGroupUserRow, _ int) (int64, bool) {
					if row.SupportGroupId.Int64 == supportGroup.Id {
						return row.UserId.Int64, true
					}
					return 0, false
				})

				// find a user that is not attached to the supportGroup
				userRow, _ := lo.Find(seedCollection.UserRows, func(row mariadb.UserRow) bool {
					return !lo.Contains(userIds, row.Id.Int64)
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SupportGroup model.SupportGroup `json:"addUserToSupportGroup"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/supportGroup/addUser.graphql",
					map[string]any{
						"supportGroupId": fmt.Sprintf("%d", supportGroup.Id),
						"userId":         fmt.Sprintf("%d", userRow.Id.Int64),
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				_, found := lo.Find(respData.SupportGroup.Users.Edges, func(edge *model.UserEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", userRow.Id.Int64)
				})

				Expect(respData.SupportGroup.ID).To(Equal(fmt.Sprintf("%d", supportGroup.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes user from supportGroup", Label("removeUser.graphql"), func() {
				supportGroup := seedCollection.SupportGroupRows[0].AsSupportGroup()

				// find a user that is attached to the supportGroup
				userRow, _ := lo.Find(seedCollection.SupportGroupUserRows, func(row mariadb.SupportGroupUserRow) bool {
					return row.SupportGroupId.Int64 == supportGroup.Id
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SupportGroup model.SupportGroup `json:"removeUserFromSupportGroup"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/supportGroup/removeUser.graphql",
					map[string]any{
						"supportGroupId": fmt.Sprintf("%d", supportGroup.Id),
						"userId":         fmt.Sprintf("%d", userRow.UserId.Int64),
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				_, found := lo.Find(respData.SupportGroup.Users.Edges, func(edge *model.UserEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", userRow.UserId.Int64)
				})

				Expect(respData.SupportGroup.ID).To(Equal(fmt.Sprintf("%d", supportGroup.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})
