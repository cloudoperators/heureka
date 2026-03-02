// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/samber/lo"
)

var _ = Describe("Getting Issues via API", Label("e2e", "Issues"), func() {
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
				Issues model.IssueConnection `json:"Issues"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/issue/minimal.graphql",
				map[string]any{
					"filter": map[string]string{},
					"first":  10,
					"after":  "",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
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
					respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						Issues model.IssueConnection `json:"Issues"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/issue/minimal.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  5,
							"after":  "",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())
					Expect(respData.Issues.TotalCount).To(Equal(len(seedCollection.IssueRows)))
					Expect(len(respData.Issues.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {
				respData := struct {
					Issues model.IssueConnection `json:"Issues"`
				}{}
				BeforeEach(func() {
					resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						Issues model.IssueConnection `json:"Issues"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/issue/directRelations.graphql",
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
					Expect(respData.Issues.TotalCount).To(Equal(len(seedCollection.IssueRows)))
					Expect(len(respData.Issues.Edges)).To(Equal(5))
				})

				It("- returns the expected content", func() {
					// this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
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
									fmt.Sprintf("%d", row.IssueRepositoryId.Int64) == iv.Node.IssueRepository.ID // references correct repository
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
			Context("and we request metadata", Label("withObjectMetadata.graphql"), func() {
				It("returns correct metadata counts", func() {
					respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						Issues model.IssueConnection `json:"Issues"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/issue/withObjectMetadata.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  5,
							"after":  "",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())

					for _, issueEdge := range respData.Issues.Edges {
						ciCount := 0
						serviceIdSet := map[string]bool{}
						for _, imEdge := range issueEdge.Node.IssueMatches.Edges {
							ciCount += *imEdge.Node.ComponentInstance.Count
							serviceIdSet[imEdge.Node.ComponentInstance.Service.ID] = true
						}
						Expect(issueEdge.Node.ObjectMetadata.IssueMatchCount).To(Equal(issueEdge.Node.IssueMatches.TotalCount), "IssueMatchCount is correct")
						Expect(issueEdge.Node.ObjectMetadata.ComponentInstanceCount).To(Equal(ciCount), "ComponentInstanceCount is correct")
						Expect(issueEdge.Node.ObjectMetadata.ServiceCount).To(Equal(len(serviceIdSet)), "ServiceCount is correct")
					}
				})
			})
			Context("and we use order", Label("withOrder.graphql"), func() {
				sendOrderRequest := func(orderBy []map[string]string) (*model.IssueConnection, error) {
					respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						Issues model.IssueConnection `json:"Issues"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/issue/withOrder.graphql",
						map[string]any{
							"orderBy": orderBy,
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())

					return &respData.Issues, nil
				}

				It("can order by primaryName", Label("withOrder.graphql"), func() {
					issues, err := sendOrderRequest([]map[string]string{
						{"by": "primaryName", "direction": "asc"},
					})

					Expect(err).To(BeNil(), "Error while unmarshaling")

					By("- returns the expected content in order", func() {
						var prev string = ""
						for _, i := range issues.Edges {
							Expect(*i.Node.PrimaryName >= prev).Should(BeTrue())
							prev = *i.Node.PrimaryName
						}
					})
				})
				It("can order by severity", Label("withOrder.graphql"), func() {
					issues, err := sendOrderRequest([]map[string]string{
						{"by": "severity", "direction": "asc"},
					})

					Expect(err).To(BeNil(), "Error while unmarshaling")

					By("- returns the expected content in order", func() {
						prev := -10
						for _, i := range issues.Edges {
							if len(i.Node.IssueVariants.Edges) > 0 {
								id, err := strconv.ParseInt(i.Node.ID, 10, 64)
								Expect(err).To(BeNil(), "Error while parsing ID")
								variants := seedCollection.GetIssueVariantsByIssueId(id)
								ratings := lo.Map(variants, func(iv mariadb.IssueVariantRow, _ int) int {
									return test.SeverityToNumerical(iv.Rating.String)
								})
								highestRating := lo.Max(ratings)
								Expect(highestRating >= prev).Should(BeTrue())
							}
						}
					})
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
			issue = testentity.NewFakeIssueEntity()
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new issue", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Issue model.Issue `json:"createIssue"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issue/create.graphql",
					map[string]any{
						"input": map[string]interface{}{
							"primaryName": issue.PrimaryName,
							"description": issue.Description,
							"type":        issue.Type.String(),
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
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
			It("updates issue", func() {
				issue := seedCollection.IssueRows[0].AsIssue()
				issue.Description = "New Description"

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Issue model.Issue `json:"updateIssue"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issue/update.graphql",
					map[string]any{
						"id": fmt.Sprintf("%d", issue.Id),
						"input": map[string]string{
							"description": issue.Description,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.Issue.Description).To(Equal(issue.Description))
			})
		})
	})
})

var _ = Describe("Deleting Issue via API", Label("e2e", "Issues"), func() {
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
			It("deletes issue", func() {
				id := fmt.Sprintf("%d", seedCollection.ServiceRows[0].Id.Int64)

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Id string `json:"deleteIssue"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issue/delete.graphql",
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

var _ = Describe("Modifying relationship of ComponentVersion of Issue via API", Label("e2e", "ComponentVersionIssueRelationship"), func() {
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
			It("adds componentVersion to issue", Label("addComponentVersion.graphql"), func() {
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

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Issue model.Issue `json:"addComponentVersionToIssue"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issue/addComponentVersion.graphql",
					map[string]any{
						"issueId":            fmt.Sprintf("%d", issue.Id),
						"componentVersionId": fmt.Sprintf("%d", componentVersionRow.Id.Int64),
					},
					nil,
				)

				_, found := lo.Find(respData.Issue.ComponentVersions.Edges, func(edge *model.ComponentVersionEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", componentVersionRow.Id.Int64)
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Issue.ID).To(Equal(fmt.Sprintf("%d", issue.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes componentVersion from issue", Label("removeComponentVersion.graphql"), func() {
				issue := seedCollection.IssueRows[0].AsIssue()

				// find a componentVersion that is assigned to the issue
				componentVersionRow, _ := lo.Find(seedCollection.ComponentVersionIssueRows, func(row mariadb.ComponentVersionIssueRow) bool {
					return row.IssueId.Int64 == issue.Id
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Issue model.Issue `json:"removeComponentVersionFromIssue"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issue/removeComponentVersion.graphql",
					map[string]any{
						"issueId":            fmt.Sprintf("%d", issue.Id),
						"componentVersionId": fmt.Sprintf("%d", componentVersionRow.ComponentVersionId.Int64),
					},
					nil,
				)

				_, found := lo.Find(respData.Issue.ComponentVersions.Edges, func(edge *model.ComponentVersionEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", componentVersionRow.ComponentVersionId.Int64)
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Issue.ID).To(Equal(fmt.Sprintf("%d", issue.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})
