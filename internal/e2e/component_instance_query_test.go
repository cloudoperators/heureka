// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"

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

var _ = Describe("Getting ComponentInstances via API", Label("e2e", "ComponentInstances"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			// create a queryCollection (safe to share across requests)
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstance/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "0")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ComponentInstances model.ComponentInstanceConnection `json:"ComponentInstances"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ComponentInstances.TotalCount).To(Equal(0))
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

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstance/minimal.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "0")

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ComponentInstances model.ComponentInstanceConnection `json:"ComponentInstances"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.ComponentInstances.TotalCount).To(Equal(len(seedCollection.ComponentInstanceRows)))
				Expect(len(respData.ComponentInstances.Edges)).To(Equal(5))
			})

		})
		Context("and we query to resolve levels of relations", Label("directRelations.graphql"), func() {

			var respData struct {
				ComponentInstances model.ComponentInstanceConnection `json:"ComponentInstances"`
			}
			BeforeEach(func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstance/directRelations.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "0")

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				err = client.Run(ctx, req, &respData)

				Expect(err).To(BeNil(), "Error while unmarshaling")
			})

			It("- returns the correct result count", func() {
				Expect(respData.ComponentInstances.TotalCount).To(Equal(len(seedCollection.ComponentInstanceRows)))
				Expect(len(respData.ComponentInstances.Edges)).To(Equal(5))
			})

			It("- returns the expected content", func() {
				//this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
				// a complete verification
				// additional checks are added based on bugs discovered during usage

				for _, ci := range respData.ComponentInstances.Edges {
					Expect(ci.Node.ID).ToNot(BeNil(), "componentInstance has a ID set")
					Expect(ci.Node.Ccrn).ToNot(BeNil(), "componentInstance has a ccrn set")
					Expect(ci.Node.Count).ToNot(BeNil(), "componentInstance has a count set")

					cv := ci.Node.ComponentVersion
					Expect(cv.ID).ToNot(BeNil(), "componentVersion has a ID set")
					Expect(cv.Version).ToNot(BeNil(), "componentVersion has a version set")

					_, cvFound := lo.Find(seedCollection.ComponentVersionRows, func(row mariadb.ComponentVersionRow) bool {
						return fmt.Sprintf("%d", row.Id.Int64) == cv.ID
					})
					Expect(cvFound).To(BeTrue(), "attached componentVersion does exist")

					service := ci.Node.Service
					Expect(service.ID).ToNot(BeNil(), "service has a ID set")
					Expect(service.Name).ToNot(BeNil(), "service has a name set")

					_, serviceFound := lo.Find(seedCollection.ServiceRows, func(row mariadb.BaseServiceRow) bool {
						return fmt.Sprintf("%d", row.Id.Int64) == service.ID
					})
					Expect(serviceFound).To(BeTrue(), "attached service does exist")

					for _, im := range ci.Node.IssueMatches.Edges {
						Expect(im.Node.ID).ToNot(BeNil(), "issueMatch has a ID set")
						Expect(im.Node.Status).ToNot(BeNil(), "issueMatch has a status set")
						Expect(im.Node.DiscoveryDate).ToNot(BeNil(), "issueMatch has a discovery date set")
						Expect(im.Node.TargetRemediationDate).ToNot(BeNil(), "issueMatch has a target remediation date set")

						_, issueMatchFound := lo.Find(seedCollection.IssueMatchRows, func(row mariadb.IssueMatchRow) bool {
							return fmt.Sprintf("%d", row.Id.Int64) == im.Node.ID && // ID Matches
								fmt.Sprintf("%d", row.ComponentInstanceId.Int64) == ci.Node.ID // correct component instance attached to issue match
						})
						Expect(issueMatchFound).To(BeTrue(), "attached IssueMatch is correct")
					}
				}
			})
			It("- returns the expected PageInfo", func() {
				Expect(*respData.ComponentInstances.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.ComponentInstances.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.ComponentInstances.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.ComponentInstances.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.ComponentInstances.PageInfo.PageNumber).To(Equal(1), "Correct page number")
			})
		})
	})
})

var _ = Describe("Creating ComponentInstance via API", Label("e2e", "ComponentInstances"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var componentInstance entity.ComponentInstance

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database has 10 entries", func() {

		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			componentInstance = testentity.NewFakeComponentInstanceEntity()
			componentInstance.ComponentVersionId = seedCollection.ComponentVersionRows[0].Id.Int64
			componentInstance.ServiceId = seedCollection.ServiceRows[0].Id.Int64
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new componentInstance", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstance/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"ccrn":               componentInstance.CCRN,
					"count":              fmt.Sprintf("%d", componentInstance.Count),
					"componentVersionId": fmt.Sprintf("%d", componentInstance.ComponentVersionId),
					"serviceId":          fmt.Sprintf("%d", componentInstance.ServiceId),
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ComponentInstance model.ComponentInstance `json:"createComponentInstance"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.ComponentInstance.Ccrn).To(Equal(componentInstance.CCRN))
				Expect(*respData.ComponentInstance.Count).To(Equal(int(componentInstance.Count)))
				Expect(*respData.ComponentInstance.ComponentVersionID).To(Equal(fmt.Sprintf("%d", componentInstance.ComponentVersionId)))
				Expect(*respData.ComponentInstance.ServiceID).To(Equal(fmt.Sprintf("%d", componentInstance.ServiceId)))
			})
		})
	})
})

var _ = Describe("Updating componentInstance via API", Label("e2e", "ComponentInstances"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates componentInstance", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstance/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				componentInstance := seedCollection.ComponentInstanceRows[0].AsComponentInstance()
				componentInstance.CCRN = "NewCCRN"

				req.Var("id", fmt.Sprintf("%d", componentInstance.Id))
				req.Var("input", map[string]string{
					"ccrn": componentInstance.CCRN,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ComponentInstance model.ComponentInstance `json:"updateComponentInstance"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.ComponentInstance.Ccrn).To(Equal(componentInstance.CCRN))
				Expect(*respData.ComponentInstance.Count).To(Equal(int(componentInstance.Count)))
				Expect(*respData.ComponentInstance.ComponentVersionID).To(Equal(fmt.Sprintf("%d", componentInstance.ComponentVersionId)))
				Expect(*respData.ComponentInstance.ServiceID).To(Equal(fmt.Sprintf("%d", componentInstance.ServiceId)))
			})
		})
	})
})

var _ = Describe("Deleting ComponentInstance via API", Label("e2e", "ComponentInstances"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes componentInstance", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstance/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.ComponentInstanceRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteComponentInstance"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
