// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/machinebox/graphql"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/model"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb/test"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	testentity "github.wdf.sap.corp/cc/heureka/internal/entity/test"
	"github.wdf.sap.corp/cc/heureka/internal/server"
	"github.wdf.sap.corp/cc/heureka/internal/util"
	util2 "github.wdf.sap.corp/cc/heureka/pkg/util"
)

var _ = Describe("Getting Issues via API", Label("e2e", "Issues"), func() {

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
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "0")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				Issues model.IssueConnection `json:"Issues"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.Issues.TotalCount).To(Equal(0))
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
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/minimal.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 5)
					req.Var("after", "0")

					req.Header.Set("Cache-Control", "no-cache")
					ctx := context.Background()

					var respData struct {
						Issues model.IssueConnection `json:"Issues"`
					}
					if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
						logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
					}

					Expect(respData.Issues.TotalCount).To(Equal(len(seedCollection.IssueRows)))
					Expect(len(respData.Issues.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {

				var respData struct {
					Issues model.IssueConnection `json:"Issues"`
				}
				BeforeEach(func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/directRelations.graphql")

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
					Expect(respData.Issues.TotalCount).To(Equal(len(seedCollection.IssueRows)))
					Expect(len(respData.Issues.Edges)).To(Equal(5))
				})

				It("- returns the expected content", func() {
					//this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage

					for _, issue := range respData.Issues.Edges {
						Expect(issue.Node.PrimaryName).ToNot(BeNil(), "Name is set")
						Expect(issue.Node.Type).ToNot(BeNil(), "Type is set")

						for _, iv := range issue.Node.IssueVariants.Edges {
							Expect(iv.Node.ID).ToNot(BeNil(), "IssueVariant has a ID set")
							Expect(iv.Node.SecondaryName).ToNot(BeNil(), "IssueVariant has a name set")
							Expect(iv.Node.Severity).ToNot(BeNil(), "IssueVariant has a severity set")
							Expect(iv.Node.Severity.Score).ToNot(BeNil(), "severity has a score set")

							_, ivFound := lo.Find(seedCollection.IssueVariantRows, func(row mariadb.IssueVariantRow) bool {
								return fmt.Sprintf("%d", row.Id.Int64) == iv.Node.ID && // correct issueVariant
									fmt.Sprintf("%d", row.IssueId.Int64) == issue.Node.ID && // belongs actually to the issue
									fmt.Sprintf("%d", row.IssueRepositoryId.Int64) == iv.Node.IssueRepository.ID //references correct repository
							})
							Expect(ivFound).To(BeTrue(), "attached issueVariant does exist and belongs to issue and repository belongs to issueVariant")

							Expect(iv.Node.IssueRepository.Name).ToNot(BeNil(), "Repository name is set")
						}
						for _, im := range issue.Node.IssueMatches.Edges {
							_, issueMatchFound := lo.Find(seedCollection.IssueMatchRows, func(row mariadb.IssueMatchRow) bool {
								return fmt.Sprintf("%d", row.Id.Int64) == im.Node.ID && // ID Matches
									//@todo check and verify the date format comparison
									//row.TargetRemediationDate.Time.String() == *vm.Node.TargetRemediationDate && // target remediation date matches
									fmt.Sprintf("%d", row.IssueId.Int64) == issue.Node.ID && // issue match belongs to the respective issue
									fmt.Sprintf("%d", row.ComponentInstanceId.Int64) == im.Node.ComponentInstance.ID // correct component instance attached to issue match
							})
							Expect(issueMatchFound).To(BeTrue(), "attached IssueMatch is correct")

							_, componentInstanceFound := lo.Find(seedCollection.ComponentInstanceRows, func(row mariadb.ComponentInstanceRow) bool {
								return fmt.Sprintf("%d", row.Id.Int64) == im.Node.ComponentInstance.ID &&
									row.CCRN.String == *im.Node.ComponentInstance.Ccrn &&
									int(row.Count.Int16) == *im.Node.ComponentInstance.Count
							})
							Expect(componentInstanceFound).To(BeTrue(), "attached Component instance is correct")
						}
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.Issues.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.Issues.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.Issues.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.Issues.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
					Expect(*respData.Issues.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
		})
	})
})

var _ = Describe("Creating Issue via API", Label("e2e", "Issues"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var issue entity.Issue

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

		BeforeEach(func() {
			seeder.SeedDbWithNFakeData(10)
			issue = testentity.NewFakeIssueEntity()
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new issue", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"primaryName": issue.PrimaryName,
					"description": issue.Description,
					"type":        issue.Type.String(),
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Issue model.Issue `json:"createIssue"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Issue.PrimaryName).To(Equal(issue.PrimaryName))
				Expect(*respData.Issue.Description).To(Equal(issue.Description))
				Expect(respData.Issue.Type.String()).To(Equal(issue.Type.String()))
			})
		})
	})
})

var _ = Describe("Updating issue via API", Label("e2e", "Issues"), func() {

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
			It("updates issue", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				issue := seedCollection.IssueRows[0].AsIssue()
				issue.Description = "New Description"

				req.Var("id", fmt.Sprintf("%d", issue.Id))
				req.Var("input", map[string]string{
					"description": issue.Description,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Issue model.Issue `json:"updateIssue"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Issue.Description).To(Equal(issue.Description))
			})
		})
	})
})

var _ = Describe("Deleting Issue via API", Label("e2e", "Issues"), func() {

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
			It("deletes issue", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.ServiceRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteIssue"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})

var _ = Describe("Modifying relationship of ComponentVersion of Issue via API", Label("e2e", "ComponentVersionIssueRelationship"), func() {

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

		Context("and a mutation query is performed", func() {
			It("adds componentVersion to issue", Label("addComponentVersion.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/addComponentVersion.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				issue := seedCollection.IssueRows[0].AsIssue()
				// find all componentVersions that are assigned to the issue
				componentVersionIds := lo.FilterMap(seedCollection.ComponentVersionIssueRows, func(row mariadb.ComponentVersionIssueRow, _ int) (int64, bool) {
					if row.IssueId.Int64 == issue.Id {
						return row.ComponentVersionId.Int64, true
					}
					return 0, false
				})

				// find a componentVersion that is not assigned to the issue
				componentVersionRow, _ := lo.Find(seedCollection.ComponentVersionRows, func(row mariadb.ComponentVersionRow) bool {
					return !lo.Contains(componentVersionIds, row.Id.Int64)
				})

				req.Var("issueId", fmt.Sprintf("%d", issue.Id))
				req.Var("componentVersionId", fmt.Sprintf("%d", componentVersionRow.Id.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Issue model.Issue `json:"addComponentVersionToIssue"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Issue.ComponentVersions.Edges, func(edge *model.ComponentVersionEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", componentVersionRow.Id.Int64)
				})

				Expect(respData.Issue.ID).To(Equal(fmt.Sprintf("%d", issue.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes componentVersion from issue", Label("removeComponentVersion.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/removeComponentVersion.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				issue := seedCollection.IssueRows[0].AsIssue()

				// find a componentVersion that is assigned to the issue
				componentVersionRow, _ := lo.Find(seedCollection.ComponentVersionIssueRows, func(row mariadb.ComponentVersionIssueRow) bool {
					return row.IssueId.Int64 == issue.Id
				})

				req.Var("issueId", fmt.Sprintf("%d", issue.Id))
				req.Var("componentVersionId", fmt.Sprintf("%d", componentVersionRow.ComponentVersionId.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Issue model.Issue `json:"removeComponentVersionFromIssue"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Issue.ComponentVersions.Edges, func(edge *model.ComponentVersionEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", componentVersionRow.ComponentVersionId.Int64)
				})

				Expect(respData.Issue.ID).To(Equal(fmt.Sprintf("%d", issue.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})
