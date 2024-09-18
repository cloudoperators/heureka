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

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Getting IssueMatchChanges via API", Label("e2e", "IssueMatchChanges"), func() {

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
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "0")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				IssueMatchChanges model.IssueMatchChangeConnection `json:"IssueMatchChanges"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.IssueMatchChanges.TotalCount).To(Equal(0))
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
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/minimal.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 5)
					req.Var("after", "0")

					req.Header.Set("Cache-Control", "no-cache")
					ctx := context.Background()

					var respData struct {
						IssueMatchChanges model.IssueMatchChangeConnection `json:"IssueMatchChanges"`
					}
					if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
						logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
					}

					Expect(respData.IssueMatchChanges.TotalCount).To(Equal(len(seedCollection.IssueMatchChangeRows)))
					Expect(len(respData.IssueMatchChanges.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {

				var respData struct {
					IssueMatchChanges model.IssueMatchChangeConnection `json:"IssueMatchChanges"`
				}
				BeforeEach(func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/directRelations.graphql")

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
					Expect(respData.IssueMatchChanges.TotalCount).To(Equal(len(seedCollection.IssueMatchChangeRows)))
					Expect(len(respData.IssueMatchChanges.Edges)).To(Equal(5))
				})

				It("- returns the expected content", func() {
					//this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage

					for _, imc := range respData.IssueMatchChanges.Edges {
						Expect(imc.Node.ID).ToNot(BeNil(), "issueMatchChange has a ID set")
						Expect(imc.Node.Action).ToNot(BeNil(), "issueMatchChange has an action set")
						Expect(imc.Node.ActivityID).ToNot(BeNil(), "issueMatchChange has an activityId set")
						Expect(imc.Node.IssueMatchID).ToNot(BeNil(), "issueMatchChange has a issueMatchId set")

						activity := imc.Node.Activity
						Expect(activity.ID).ToNot(BeNil(), "activity has a ID set")

						im := imc.Node.IssueMatch
						Expect(im.ID).ToNot(BeNil(), "issueMatch has a ID set")
						Expect(im.Status).ToNot(BeNil(), "issueMatch has a status set")
						Expect(im.RemediationDate).ToNot(BeNil(), "issueMatch has a remediationDate set")
						Expect(im.DiscoveryDate).ToNot(BeNil(), "issueMatch has a discoveryDate set")
						Expect(im.TargetRemediationDate).ToNot(BeNil(), "issueMatch has a targetRemediationDate set")
						Expect(im.IssueID).ToNot(BeNil(), "issueMatch has a issueId set")
						Expect(im.ComponentInstanceID).ToNot(BeNil(), "issueMatch has a componentInstanceId set")
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.IssueMatchChanges.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.IssueMatchChanges.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.IssueMatchChanges.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.IssueMatchChanges.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
					Expect(*respData.IssueMatchChanges.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
		})
	})
})

var _ = Describe("Creating IssueMatchChange via API", Label("e2e", "IssueMatchChanges"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var issueMatchChange entity.IssueMatchChange

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

	// use only 1 entry to make sure that all relations are resolved correctly
	When("the database has 1 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(1)
			issueMatchChange = testentity.NewFakeIssueMatchChange()
			issueMatchChange.IssueMatchId = seedCollection.ActivityRows[0].Id.Int64
			issueMatchChange.ActivityId = seedCollection.IssueMatchRows[0].Id.Int64
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new issueMatchChange", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]interface{}{
					"action":       issueMatchChange.Action,
					"activityId":   issueMatchChange.ActivityId,
					"issueMatchId": issueMatchChange.IssueMatchId,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueMatchChange model.IssueMatchChange `json:"createIssueMatchChange"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.IssueMatchChange.Action.String()).To(Equal(issueMatchChange.Action))
				Expect(*respData.IssueMatchChange.ActivityID).To(Equal(fmt.Sprintf("%d", issueMatchChange.ActivityId)))
				Expect(*respData.IssueMatchChange.IssueMatchID).To(Equal(fmt.Sprintf("%d", issueMatchChange.IssueMatchId)))
			})
		})
	})
})

var _ = Describe("Updating issueMatchChange via API", Label("e2e", "IssueMatchChangees"), func() {

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
			It("updates issueMatchChange", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				issueMatchChange := seedCollection.IssueMatchChangeRows[0].AsIssueMatchChange()

				if issueMatchChange.Action == entity.IssueMatchChangeActionAdd.String() {
					issueMatchChange.Action = entity.IssueMatchChangeActionRemove.String()
				} else {
					issueMatchChange.Action = entity.IssueMatchChangeActionAdd.String()
				}

				req.Var("id", fmt.Sprintf("%d", issueMatchChange.Id))
				req.Var("input", map[string]string{
					"action": issueMatchChange.Action,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueMatchChange model.IssueMatchChange `json:"updateIssueMatchChange"`
				}

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.IssueMatchChange.Action.String()).To(Equal(issueMatchChange.Action))
			})
		})
	})
})

var _ = Describe("Deleting IssueMatchChange via API", Label("e2e", "IssueMatchChanges"), func() {

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
			It("deletes issueMatchChange", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.IssueVariantRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteIssueMatchChange"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
