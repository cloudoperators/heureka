// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"
	"time"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Getting IssueMatches via API", Label("e2e", "IssueMatches"), func() {
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
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			// create a queryCollection (safe to share across requests)
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatch/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.IssueMatches.TotalCount).To(Equal(0))
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
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatch/minimal.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 5)
					req.Var("after", "")

					req.Header.Set("Cache-Control", "no-cache")
					ctx := context.Background()

					var respData struct {
						IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
					}
					if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
						logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
					}

					Expect(respData.IssueMatches.TotalCount).To(Equal(len(seedCollection.IssueMatchRows)))
					Expect(len(respData.IssueMatches.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {
				var respData struct {
					IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
				}
				BeforeEach(func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatch/directRelations.graphql")

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
					Expect(respData.IssueMatches.TotalCount).To(Equal(len(seedCollection.IssueMatchRows)))
					Expect(len(respData.IssueMatches.Edges)).To(Equal(5))
				})

				It("- returns the expected content", func() {
					// this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage

					for _, im := range respData.IssueMatches.Edges {
						Expect(im.Node.ID).ToNot(BeNil(), "issueMatch has a ID set")
						Expect(im.Node.Status).ToNot(BeNil(), "issueMatch has a status set")
						Expect(im.Node.RemediationDate).ToNot(BeNil(), "issueMatch has a remediation date set")
						Expect(im.Node.TargetRemediationDate).ToNot(BeNil(), "issueMatch has a target remediation date set")

						if im.Node.Severity != nil {
							Expect(im.Node.Severity.Value).ToNot(BeNil(), "issueMatch has a severity value set")
							Expect(im.Node.Severity.Score).ToNot(BeNil(), "issueMatch has a severity score set")
						}

						for _, eiv := range im.Node.EffectiveIssueVariants.Edges {
							Expect(eiv.Node.ID).ToNot(BeNil(), "effectiveIssueVariant has a ID set")
							Expect(eiv.Node.Description).ToNot(BeNil(), "effectiveIssueVariant has a description set")
							Expect(eiv.Node.SecondaryName).ToNot(BeNil(), "effectiveIssueVariant has a name set")
						}

						issue := im.Node.Issue
						Expect(issue.ID).ToNot(BeNil(), "issue has a ID set")
						Expect(issue.LastModified).ToNot(BeNil(), "issue has a lastModified set")
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.IssueMatches.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.IssueMatches.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.IssueMatches.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.IssueMatches.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
					Expect(*respData.IssueMatches.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
			Context("we use ordering", Label("withOrder.graphql"), func() {
				var respData struct {
					IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
				}

				It("can order by primaryName", Label("withOrder.graphql"), func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatch/withOrder.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 10)
					req.Var("after", "")
					req.Var("orderBy", []map[string]string{
						{"by": "primaryName", "direction": "asc"},
					})

					req.Header.Set("Cache-Control", "no-cache")

					ctx := context.Background()

					err = client.Run(ctx, req, &respData)

					Expect(err).To(BeNil(), "Error while unmarshaling")

					By("- returns the correct result count", func() {
						Expect(respData.IssueMatches.TotalCount).To(Equal(len(seedCollection.IssueMatchRows)))
						Expect(len(respData.IssueMatches.Edges)).To(Equal(10))
					})

					By("- returns the expected content in order", func() {
						var prev string = ""
						for _, im := range respData.IssueMatches.Edges {
							Expect(*im.Node.Issue.PrimaryName >= prev).Should(BeTrue())
							prev = *im.Node.Issue.PrimaryName
						}
					})
				})

				It("can order by primaryName and targetRemediationDate", Label("withOrder.graphql"), func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatch/withOrder.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 10)
					req.Var("after", "")
					req.Var("orderBy", []map[string]string{
						{"by": "primaryName", "direction": "asc"},
						{"by": "targetRemediationDate", "direction": "desc"},
					})

					req.Header.Set("Cache-Control", "no-cache")

					ctx := context.Background()

					err = client.Run(ctx, req, &respData)

					Expect(err).To(BeNil(), "Error while unmarshaling")

					By("- returns the correct result count", func() {
						Expect(respData.IssueMatches.TotalCount).To(Equal(len(seedCollection.IssueMatchRows)))
						Expect(len(respData.IssueMatches.Edges)).To(Equal(10))
					})

					By("- returns the expected content in order", func() {
						var prevPn string = ""
						var prevTrd time.Time = time.Now()
						for _, im := range respData.IssueMatches.Edges {
							if *im.Node.Issue.PrimaryName == prevPn {
								trd, err := time.Parse("2006-01-02T15:04:05Z", *im.Node.TargetRemediationDate)
								Expect(err).To(BeNil())
								Expect(trd.Before(prevTrd)).Should(BeTrue())
								prevTrd = trd
							} else {
								Expect(*im.Node.Issue.PrimaryName > prevPn).To(BeTrue())
								prevTrd = time.Now()
							}
							prevPn = *im.Node.Issue.PrimaryName
						}
					})
				})
			})
		})
	})
})

var _ = Describe("Creating IssueMatch via API", Label("e2e", "IssueMatches"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var issueMatch entity.IssueMatch
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
		dbm.TestTearDown(db)
	})

	// use only 1 entry to make sure that all relations are resolved correctly
	When("the database has 1 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(1)
			issueMatch = testentity.NewFakeIssueMatch()
			issueMatch.ComponentInstanceId = seedCollection.ComponentInstanceRows[0].Id.Int64

			issueMatch.IssueId = seedCollection.IssueRows[0].Id.Int64
			issueMatch.UserId = seedCollection.UserRows[0].Id.Int64
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new issueMatch", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatch/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]interface{}{
					"status":                issueMatch.Status,
					"userId":                issueMatch.UserId,
					"componentInstanceId":   issueMatch.ComponentInstanceId,
					"issueId":               fmt.Sprintf("%d", issueMatch.IssueId),
					"remediationDate":       issueMatch.RemediationDate.Format(time.RFC3339),
					"targetRemediationDate": issueMatch.TargetRemediationDate.Format(time.RFC3339),
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueMatch model.IssueMatch `json:"createIssueMatch"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.IssueMatch.Status.String()).To(Equal(issueMatch.Status.String()))
				Expect(*respData.IssueMatch.IssueID).To(Equal(fmt.Sprintf("%d", issueMatch.IssueId)))
				Expect(*respData.IssueMatch.UserID).To(Equal(fmt.Sprintf("%d", issueMatch.UserId)))
				Expect(*respData.IssueMatch.ComponentInstanceID).To(Equal(fmt.Sprintf("%d", issueMatch.ComponentInstanceId)))
				Expect(*respData.IssueMatch.RemediationDate).To(Equal(issueMatch.RemediationDate.Format(time.RFC3339)))
				Expect(*respData.IssueMatch.TargetRemediationDate).To(Equal(issueMatch.TargetRemediationDate.Format(time.RFC3339)))
			})
		})
	})
})

var _ = Describe("Updating issueMatch via API", Label("e2e", "IssueMatches"), func() {
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
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates issueMatch", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatch/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				issueMatch := seedCollection.IssueMatchRows[0].AsIssueMatch()
				issueMatch.RemediationDate = issueMatch.RemediationDate.Add(time.Hour * 24 * 7)

				req.Var("id", fmt.Sprintf("%d", issueMatch.Id))
				req.Var("input", map[string]string{
					"remediationDate": issueMatch.RemediationDate.Format(time.RFC3339),
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueMatch model.IssueMatch `json:"updateIssueMatch"`
				}

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.IssueMatch.RemediationDate).To(Equal(issueMatch.RemediationDate.Format(time.RFC3339)))
			})
		})
	})
})

var _ = Describe("Deleting IssueMatch via API", Label("e2e", "IssueMatches"), func() {
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
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes issuematch", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatch/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.IssueVariantRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteIssueMatch"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
