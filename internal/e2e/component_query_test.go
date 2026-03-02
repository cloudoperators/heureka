// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"

	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Getting Components via API", Label("e2e", "Components"), func() {
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
				Component model.ComponentConnection `json:"Component"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/component/minimal.graphql",
				map[string]any{
					"filter": map[string]string{},
					"first":  10,
					"after":  "",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.Component.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct result count", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Components model.ComponentConnection `json:"Components"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/component/minimal.graphql",
					map[string]any{
						"filter": map[string]string{},
						"first":  5,
						"after":  "",
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Components.TotalCount).To(Equal(len(seedCollection.ComponentRows)))
				Expect(len(respData.Components.Edges)).To(Equal(5))
			})
		})
		Context("and we query to resolve levels of relations", Label("directRelations.graphql"), func() {
			respData := struct {
				Components model.ComponentConnection `json:"Components"`
			}{}
			BeforeEach(func() {
				resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Components model.ComponentConnection `json:"Components"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/component/directRelations.graphql",
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
				Expect(respData.Components.TotalCount).To(Equal(len(seedCollection.ComponentRows)))
				Expect(len(respData.Components.Edges)).To(Equal(5))
			})

			It("- returns the expected content", func() {
				// this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
				// a complete verification
				// additional checks are added based on bugs discovered during usage

				for _, component := range respData.Components.Edges {
					Expect(component.Node.ID).ToNot(BeNil(), "component has a ID set")
					Expect(component.Node.Ccrn).ToNot(BeNil(), "component has a CCRN set")
					Expect(component.Node.Repository).ToNot(BeNil(), "component has a Repository set")
					Expect(component.Node.Organization).ToNot(BeNil(), "component has an Organization set")
					Expect(component.Node.URL).ToNot(BeNil(), "component has a URL set")
					Expect(component.Node.Type).ToNot(BeNil(), "component has a Type set")

					for _, cv := range component.Node.ComponentVersions.Edges {
						Expect(cv.Node.ID).ToNot(BeNil(), "componentVersion has a ID set")
						Expect(cv.Node.Version).ToNot(BeNil(), "componentVersion has a version set")
						Expect(cv.Node.ComponentID).ToNot(BeNil(), "componentVersion has a componentId set")

						_, cvFound := lo.Find(seedCollection.ComponentVersionRows, func(row mariadb.ComponentVersionRow) bool {
							return fmt.Sprintf("%d", row.Id.Int64) == cv.Node.ID && // correct component version
								fmt.Sprintf("%d", row.ComponentId.Int64) == component.Node.ID // belongs actually to the component
						})
						Expect(cvFound).To(BeTrue(), "attached componentVersion does exist and belongs to component")
					}
				}
			})
			It("- returns the expected PageInfo", func() {
				Expect(*respData.Components.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.Components.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.Components.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.Components.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.Components.PageInfo.PageNumber).To(Equal(1), "Correct page number")
			})
		})
	})
})

var _ = Describe("Creating Component via API", Label("e2e", "Components"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var component entity.Component
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
			component = testentity.NewFakeComponentEntity()
			component.Type = "virtualMachineImage"
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new component", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Component model.Component `json:"createComponent"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/component/create.graphql",
					map[string]any{
						"input": map[string]string{
							"type":         component.Type,
							"ccrn":         component.CCRN,
							"repository":   component.Repository,
							"organization": component.Organization,
							"url":          component.Url,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.Component.Ccrn).To(Equal(component.CCRN))
				Expect(respData.Component.Type.String()).To(Equal(component.Type))
			})
		})
	})
})

var _ = Describe("Updating Component via API", Label("e2e", "Components"), func() {
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
			It("updates component", func() {
				component := seedCollection.ComponentRows[0].AsComponent()
				component.CCRN = "NewCCRN"

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Component model.Component `json:"updateComponent"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/component/update.graphql",
					map[string]any{
						"id": fmt.Sprintf("%d", component.Id),
						"input": map[string]string{
							"ccrn": component.CCRN,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.Component.Ccrn).To(Equal(component.CCRN))
			})
		})
	})
})

var _ = Describe("Deleting Component via API", Label("e2e", "Components"), func() {
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
			It("deletes component", func() {
				id := fmt.Sprintf("%d", seedCollection.ComponentRows[0].Id.Int64)

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Id string `json:"deleteComponent"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/component/delete.graphql",
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
