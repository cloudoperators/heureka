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

var _ = Describe("Getting IssueVariants via API", Label("e2e", "IssueVariants"), func() {
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
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueVariant/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "0")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				IssueVariants model.IssueVariantConnection `json:"IssueVariants"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.IssueVariants.TotalCount).To(Equal(0))
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
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueVariant/minimal.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "0")

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueVariants model.IssueVariantConnection `json:"IssueVariants"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.IssueVariants.TotalCount).To(Equal(len(seedCollection.IssueVariantRows)))
				Expect(len(respData.IssueVariants.Edges)).To(Equal(5))
			})

		})
		Context("and we query to resolve levels of relations", Label("directRelations.graphql"), func() {

			var respData struct {
				IssueVariants model.IssueVariantConnection `json:"IssueVariants"`
			}
			BeforeEach(func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueVariant/directRelations.graphql")

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
				Expect(respData.IssueVariants.TotalCount).To(Equal(len(seedCollection.IssueVariantRows)))
				Expect(len(respData.IssueVariants.Edges)).To(Equal(5))
			})

			It("- returns the expected content", func() {
				//this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
				// a complete verification
				// additional checks are added based on bugs discovered during usage

				for _, issueVariant := range respData.IssueVariants.Edges {
					Expect(issueVariant.Node.ID).ToNot(BeNil(), "issueVariant has a ID set")
					Expect(issueVariant.Node.SecondaryName).ToNot(BeNil(), "issueVariant has a name set")
					Expect(issueVariant.Node.Description).ToNot(BeNil(), "issueVariant has a description set")
					Expect(issueVariant.Node.Severity.Value).ToNot(BeNil(), "issueVariant has a severity value set")
					Expect(issueVariant.Node.Severity.Score).ToNot(BeNil(), "issueVariant has a severity score set")

					issue := issueVariant.Node.Issue
					Expect(issue.ID).ToNot(BeNil(), "issue has a ID set")
					Expect(issue.LastModified).ToNot(BeNil(), "issue has a lastModified set")

					_, issueFound := lo.Find(seedCollection.IssueRows, func(row mariadb.IssueRow) bool {
						return fmt.Sprintf("%d", row.Id.Int64) == issue.ID
					})
					Expect(issueFound).To(BeTrue(), "attached issue does exist")

					ir := issueVariant.Node.IssueRepository
					Expect(ir.ID).ToNot(BeNil(), "issueRepository has a ID set")
					Expect(ir.Name).ToNot(BeNil(), "issueRepository has a name set")
					Expect(ir.URL).ToNot(BeNil(), "issueRepository has a url set")

					_, irFound := lo.Find(seedCollection.IssueRepositoryRows, func(row mariadb.BaseIssueRepositoryRow) bool {
						return fmt.Sprintf("%d", row.Id.Int64) == ir.ID
					})
					Expect(irFound).To(BeTrue(), "attached issueRepository does exist")
				}
			})
			It("- returns the expected PageInfo", func() {
				Expect(*respData.IssueVariants.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.IssueVariants.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.IssueVariants.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.IssueVariants.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.IssueVariants.PageInfo.PageNumber).To(Equal(1), "Correct page number")
			})
		})
	})
})

var _ = Describe("Creating IssueVariant via API", Label("e2e", "IssueVariants"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var issueVariant entity.IssueVariant

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
			issueVariant = testentity.NewFakeIssueVariantEntity(nil)
			issueVariant.IssueRepositoryId = seedCollection.IssueRepositoryRows[0].Id.Int64
			issueVariant.IssueId = seedCollection.IssueRows[0].Id.Int64
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new issueVariant with Vector", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueVariant/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]interface{}{
					"secondaryName":     issueVariant.SecondaryName,
					"description":       issueVariant.Description,
					"issueRepositoryId": fmt.Sprintf("%d", issueVariant.IssueRepositoryId),
					"issueId":           fmt.Sprintf("%d", issueVariant.IssueId),
					"severity": map[string]string{
						"vector": issueVariant.Severity.Cvss.Vector,
					},
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueVariant model.IssueVariant `json:"createIssueVariant"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.IssueVariant.SecondaryName).To(Equal(issueVariant.SecondaryName))
				Expect(*respData.IssueVariant.Description).To(Equal(issueVariant.Description))
				Expect(*respData.IssueVariant.IssueRepositoryID).To(Equal(fmt.Sprintf("%d", issueVariant.IssueRepositoryId)))
				Expect(*respData.IssueVariant.IssueID).To(Equal(fmt.Sprintf("%d", issueVariant.IssueId)))
				Expect(*respData.IssueVariant.Severity.Cvss.Vector).To(Equal(issueVariant.Severity.Cvss.Vector))
			})
			It("creates new issueVariant with Rating", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueVariant/createWithRating.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]interface{}{
					"secondaryName":     issueVariant.SecondaryName,
					"description":       issueVariant.Description,
					"issueRepositoryId": fmt.Sprintf("%d", issueVariant.IssueRepositoryId),
					"issueId":           fmt.Sprintf("%d", issueVariant.IssueId),
					"severity": map[string]string{
						"rating": issueVariant.Severity.Value,
					},
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueVariant model.IssueVariant `json:"createIssueVariant"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.IssueVariant.SecondaryName).To(Equal(issueVariant.SecondaryName))
				Expect(*respData.IssueVariant.Description).To(Equal(issueVariant.Description))
				Expect(*respData.IssueVariant.IssueRepositoryID).To(Equal(fmt.Sprintf("%d", issueVariant.IssueRepositoryId)))
				Expect(*respData.IssueVariant.IssueID).To(Equal(fmt.Sprintf("%d", issueVariant.IssueId)))
				Expect(string(*respData.IssueVariant.Severity.Value)).To(Equal(issueVariant.Severity.Value))
			})
		})
	})
})

var _ = Describe("Updating issueVariant via API", Label("e2e", "IssueVariants"), func() {

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
			It("updates issueVariant", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueVariant/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				ir := seedCollection.IssueRepositoryRows[0].AsIssueRepository()
				issueVariant := seedCollection.IssueVariantRows[0].AsIssueVariant(&ir)
				issueVariant.SecondaryName = "SecretIssueVariant"

				req.Var("id", fmt.Sprintf("%d", issueVariant.Id))
				req.Var("input", map[string]interface{}{
					"secondaryName": issueVariant.SecondaryName,
					"severity": model.SeverityInput{
						Rating: &model.AllSeverityValues[0],
						Vector: &issueVariant.Severity.Cvss.Vector,
					},
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueVariant model.IssueVariant `json:"updateIssueVariant"`
				}

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.IssueVariant.SecondaryName).To(Equal(issueVariant.SecondaryName))
			})
		})
	})
})

var _ = Describe("Deleting IssueVariant via API", Label("e2e", "IssueVariants"), func() {

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
			It("deletes issue variant", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueVariant/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.IssueVariantRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteIssueVariant"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
