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

var _ = Describe("Getting ComponentVersions via API", Label("e2e", "ComponentVersions"), func() {
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
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentVersion/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ComponentVersion model.ComponentVersionConnection `json:"ComponentVersion"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ComponentVersion.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {

		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and  no additional filters are present", func() {
			It("returns correct result count", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentVersion/minimal.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "")

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ComponentVersions model.ComponentVersionConnection `json:"ComponentVersions"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.ComponentVersions.TotalCount).To(Equal(len(seedCollection.ComponentVersionRows)))
				Expect(len(respData.ComponentVersions.Edges)).To(Equal(5))
			})

		})
		Context("and we query to resolve levels of relations", Label("directRelations.graphql"), func() {

			var respData struct {
				ComponentVersions model.ComponentVersionConnection `json:"ComponentVersions"`
			}
			BeforeEach(func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentVersion/directRelations.graphql")

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
				Expect(respData.ComponentVersions.TotalCount).To(Equal(len(seedCollection.ComponentVersionRows)))
				Expect(len(respData.ComponentVersions.Edges)).To(Equal(5))
			})

			It("- returns the expected content", func() {
				//this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
				// a complete verification
				// additional checks are added based on bugs discovered during usage

				for _, cv := range respData.ComponentVersions.Edges {
					for _, issue := range cv.Node.Issues.Edges {
						Expect(issue.Node.ID).ToNot(BeNil(), "issue has a ID set")
						Expect(issue.Node.LastModified).ToNot(BeNil(), "issue has lastModified set")

						_, issueFound := lo.Find(seedCollection.ComponentVersionIssueRows, func(row mariadb.ComponentVersionIssueRow) bool {
							return fmt.Sprintf("%d", row.IssueId.Int64) == issue.Node.ID && // correct issue
								fmt.Sprintf("%d", row.ComponentVersionId.Int64) == cv.Node.ID // belongs actually to the componentVersion
						})
						Expect(issueFound).To(BeTrue(), "attached issue does exist and belongs to componentVersion")
					}

					for _, ci := range cv.Node.ComponentInstances.Edges {
						Expect(ci.Node.ID).ToNot(BeNil(), "componentInstance has a ID set")
						Expect(ci.Node.Ccrn).ToNot(BeNil(), "componentInstance has ccrn set")

						Expect(*ci.Node.ComponentVersionID).To(BeEquivalentTo(cv.Node.ID))

					}

					if cv.Node.Component != nil {
						Expect(cv.Node.Component.ID).ToNot(BeNil(), "component has a ID set")
						Expect(cv.Node.Component.Ccrn).ToNot(BeNil(), "component has a ccrn set")
						Expect(cv.Node.Component.Type).ToNot(BeNil(), "component has a type set")
					}

					if cv.Node.Tag != nil {
						// If there's a tag value in the database, verify it matches
						for _, row := range seedCollection.ComponentVersionRows {
							if fmt.Sprintf("%d", row.Id.Int64) == cv.Node.ID && row.Tag.Valid {
								Expect(*cv.Node.Tag).To(Equal(row.Tag.String))
							}
						}
					}
				}
			})
			It("- returns the expected PageInfo", func() {
				Expect(*respData.ComponentVersions.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.ComponentVersions.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.ComponentVersions.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.ComponentVersions.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.ComponentVersions.PageInfo.PageNumber).To(Equal(1), "Correct page number")
			})
		})
		Context("and we request issueCounts", func() {
			It("returns correct issueCounts", func() {
				severityCounts := map[string]model.SeverityCounts{}
				// Setup severityCounts for all componentVersions
				for _, cv := range seedCollection.ComponentVersionRows {
					cvId := fmt.Sprint(cv.Id.Int64)
					severityCounts[cvId] = model.SeverityCounts{}
					counts := severityCounts[cvId]
					for _, cvir := range seedCollection.ComponentVersionIssueRows {
						if cv.Id.Int64 == cvir.ComponentVersionId.Int64 {
							for _, iv := range seedCollection.IssueVariantRows {
								if cvir.IssueId.Int64 == iv.IssueId.Int64 {
									switch iv.Rating.String {
									case "Critical":
										counts.Critical++
									case "High":
										counts.High++
									case "Medium":
										counts.Medium++
									case "Low":
										counts.Low++
									case "None":
										counts.None++
									}
								}
							}
						}
					}
					severityCounts[cvId] = counts
				}

				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentVersion/withIssueCounts.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ComponentVersions model.ComponentVersionConnection `json:"ComponentVersions"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				for _, cvEdge := range respData.ComponentVersions.Edges {
					sc := severityCounts[cvEdge.Node.ID]
					Expect(cvEdge.Node.IssueCounts.Critical).To(Equal(sc.Critical))
					Expect(cvEdge.Node.IssueCounts.High).To(Equal(sc.High))
					Expect(cvEdge.Node.IssueCounts.Medium).To(Equal(sc.Medium))
					Expect(cvEdge.Node.IssueCounts.Low).To(Equal(sc.Low))
					Expect(cvEdge.Node.IssueCounts.None).To(Equal(sc.None))
				}
			})
		})
	})
})

var _ = Describe("Ordering ComponentVersion via API", Label("e2e", "ComponentVersions"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var respData struct {
		ComponentVersions model.ComponentVersionConnection `json:"ComponentVersions"`
	}

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

	var loadTestData = func() ([]mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, error) {
		issueVariants, err := test.LoadIssueVariants(test.GetTestDataPath("../database/mariadb/testdata/component_version_order/issue_variant.json"))
		if err != nil {
			return nil, nil, err
		}
		cvIssues, err := test.LoadComponentVersionIssues(test.GetTestDataPath("../database/mariadb/testdata/component_version_order/component_version_issue.json"))
		if err != nil {
			return nil, nil, err
		}
		return issueVariants, cvIssues, nil
	}

	When("ordering by severity", func() {
		BeforeEach(func() {
			seeder.SeedIssueRepositories()
			seeder.SeedIssues(10)
			components := seeder.SeedComponents(1)
			seeder.SeedComponentVersions(10, components)
			issueVariants, componentVersionIssues, err := loadTestData()
			Expect(err).To(BeNil())
			// Important: the order need to be preserved
			for _, iv := range issueVariants {
				_, err := seeder.InsertFakeIssueVariant(iv)
				Expect(err).To(BeNil())
			}
			for _, cvi := range componentVersionIssues {
				_, err := seeder.InsertFakeComponentVersionIssue(cvi)
				Expect(err).To(BeNil())
			}
		})

		var runOrderTest = func(orderDirection string, expectedOrder []string) {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentVersion/withOrder.graphql")
			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)
			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "")
			req.Var("orderBy", []map[string]string{
				{"by": "severity", "direction": orderDirection},
			})
			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()
			err = client.Run(ctx, req, &respData)
			Expect(err).To(BeNil(), "Error while unmarshaling")
			Expect(respData.ComponentVersions.TotalCount).To(Equal(10))
			Expect(len(respData.ComponentVersions.Edges)).To(Equal(10))
			for i, id := range expectedOrder {
				Expect(respData.ComponentVersions.Edges[i].Node.ID).To(BeEquivalentTo(id))
			}
		}

		It("can order descending by severity", func() {
			runOrderTest("desc", []string{"3", "8", "2", "7", "1", "6", "5", "4", "10", "9"})
		})

		It("can order ascending by severity", func() {
			runOrderTest("asc", []string{"9", "10", "4", "5", "6", "1", "7", "2", "8", "3"})
		})
	})
})

var _ = Describe("Creating ComponentVersion via API", Label("e2e", "ComponentVersions"), func() {

	var seeder *test.DatabaseSeeder
	var seedCollection *test.SeedCollection
	var s *server.Server
	var cfg util.Config
	var componentVersion entity.ComponentVersion
	var componentId int64

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
			seedCollection = seeder.SeedDbWithNFakeData(10)
			componentVersion = testentity.NewFakeComponentVersionEntity()
			componentId = seedCollection.ComponentRows[0].Id.Int64
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new componentVersion", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentVersion/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				testTag := "test-tag-e2e"

				req.Var("input", map[string]string{
					"version":     componentVersion.Version,
					"componentId": fmt.Sprintf("%d", componentId),
					"tag":         testTag,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ComponentVersion model.ComponentVersion `json:"createComponentVersion"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.ComponentVersion.Version).To(Equal(componentVersion.Version))
				Expect(*respData.ComponentVersion.ComponentID).To(Equal(fmt.Sprintf("%d", componentId)))
				Expect(*respData.ComponentVersion.Tag).To(Equal(testTag))
			})
		})
	})
})

var _ = Describe("Updating ComponentVersion via API", Label("e2e", "ComponentVersions"), func() {

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
			It("updates componentVersion", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentVersion/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				componentVersion := seedCollection.ComponentVersionRows[0].AsComponentVersion()
				componentVersion.Version = "4.2.0"
				componentVersion.Tag = "updated-tag-e2e"

				req.Var("id", fmt.Sprintf("%d", componentVersion.Id))
				req.Var("input", map[string]string{
					"version": componentVersion.Version,
					"tag":     componentVersion.Tag,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ComponentVersion model.ComponentVersion `json:"updateComponentVersion"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.ComponentVersion.Version).To(Equal(componentVersion.Version))
				Expect(*respData.ComponentVersion.Tag).To(Equal(componentVersion.Tag))
			})
		})
	})
})

var _ = Describe("Deleting ComponentVersion via API", Label("e2e", "ComponentVersions"), func() {

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
			It("deletes component", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentVersion/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.ComponentVersionRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteComponentVersion"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
