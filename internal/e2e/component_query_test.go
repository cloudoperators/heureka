// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"

	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
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
		cfg.Port = util2.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		_ = dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			// create a queryCollection (safe to share across requests)
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/component/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				Component model.ComponentConnection `json:"Component"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

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
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/component/minimal.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "")

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Components model.ComponentConnection `json:"Components"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Components.TotalCount).To(Equal(len(seedCollection.ComponentRows)))
				Expect(len(respData.Components.Edges)).To(Equal(5))
			})
		})
		Context("and we query to resolve levels of relations", Label("directRelations.graphql"), func() {
			var respData struct {
				Components model.ComponentConnection `json:"Components"`
			}
			BeforeEach(func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/component/directRelations.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "")

				req.Header.Set("Cache-Control", "no-cache")

				ctx := context.Background()

				err = client.Run(ctx, req, &respData)

				Expect(err).To(BeNil(), "Error while unmarshaling")
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
		cfg.Port = util2.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		_ = dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		BeforeEach(func() {
			seeder.SeedDbWithNFakeData(10)
			component = testentity.NewFakeComponentEntity()
			component.Type = "virtualMachineImage"
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new component", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/component/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"type":         component.Type,
					"ccrn":         component.CCRN,
					"repository":   component.Repository,
					"organization": component.Organization,
					"url":          component.Url,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Component model.Component `json:"createComponent"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

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
		cfg.Port = util2.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		_ = dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates component", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/component/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				component := seedCollection.ComponentRows[0].AsComponent()
				component.CCRN = "NewCCRN"

				req.Var("id", fmt.Sprintf("%d", component.Id))
				req.Var("input", map[string]string{
					"ccrn": component.CCRN,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Component model.Component `json:"updateComponent"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

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
		cfg.Port = util2.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		_ = dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes component", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/component/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.ComponentRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteComponent"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
