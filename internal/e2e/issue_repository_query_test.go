// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Getting IssueRepositories via API", Label("e2e", "IssueRepositories"), func() {
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
				IssueRepositories model.IssueRepositoryConnection `json:"IssueRepositories"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/issueRepository/minimal.graphql",
				map[string]any{
					"filter": map[string]string{},
					"first":  10,
					"after":  "0",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.IssueRepositories.TotalCount).To(Equal(0))
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
						IssueRepositories model.IssueRepositoryConnection `json:"IssueRepositories"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/issueRepository/minimal.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  5,
							"after":  "0",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())
					Expect(respData.IssueRepositories.TotalCount).To(Equal(len(seedCollection.IssueRepositoryRows)))
					Expect(len(respData.IssueRepositories.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {
				respData := struct {
					IssueRepositories model.IssueRepositoryConnection `json:"IssueRepositories"`
				}{}
				BeforeEach(func() {
					resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						IssueRepositories model.IssueRepositoryConnection `json:"IssueRepositories"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/issueRepository/directRelations.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  3,
							"after":  "0",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())

					respData = resp
				})

				It("- returns the correct result count", func() {
					Expect(respData.IssueRepositories.TotalCount).To(Equal(len(seedCollection.IssueRepositoryRows)))
					Expect(len(respData.IssueRepositories.Edges)).To(Equal(3))
				})

				It("- returns the expected content", func() {
					// this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage

					for _, ir := range respData.IssueRepositories.Edges {
						Expect(ir.Node.ID).ToNot(BeNil(), "issueRepository has a ID set")
						Expect(ir.Node.Name).ToNot(BeNil(), "issueRepository has a name set")
						Expect(ir.Node.URL).ToNot(BeNil(), "issueRepository has a url set")

						for _, iv := range ir.Node.IssueVariants.Edges {
							Expect(iv.Node.ID).ToNot(BeNil(), "IssueVariant has a ID set")
							Expect(iv.Node.SecondaryName).ToNot(BeNil(), "IssueVariant has a name set")
							Expect(iv.Node.Description).ToNot(BeNil(), "IssueVariant has a description set")

							_, ivFound := lo.Find(seedCollection.IssueVariantRows, func(row mariadb.IssueVariantRow) bool {
								return fmt.Sprintf("%d", row.Id.Int64) == iv.Node.ID && // correct issueVariant
									fmt.Sprintf("%d", row.IssueRepositoryId.Int64) == *iv.Node.IssueRepositoryID // references correct repository
							})
							Expect(ivFound).To(BeTrue(), "attached issueVariant does exist and belongs to repository")
						}

						for _, service := range ir.Node.Services.Edges {
							Expect(service.Node.ID).ToNot(BeNil(), "Service has a ID set")
							Expect(service.Node.Ccrn).ToNot(BeNil(), "Service has a name set")
							Expect(service.Priority).ToNot(BeNil(), "Service has a priority set")

							_, serviceFound := lo.Find(seedCollection.IssueRepositoryServiceRows, func(row mariadb.IssueRepositoryServiceRow) bool {
								return fmt.Sprintf("%d", row.IssueRepositoryId.Int64) == ir.Node.ID && // correct issue repository
									fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // references correct service
							})
							Expect(serviceFound).To(BeTrue(), "attached service does exist and belongs to repository")
						}
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.IssueRepositories.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.IssueRepositories.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.IssueRepositories.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.IssueRepositories.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
					Expect(*respData.IssueRepositories.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
		})
	})
})

var _ = Describe("Creating IssueRepository via API", Label("e2e", "IssueRepositories"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var issueRepository entity.IssueRepository
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
			issueRepository = testentity.NewFakeIssueRepositoryEntity()
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new issueRepository", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					IssueRepository model.IssueRepository `json:"createIssueRepository"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issueRepository/create.graphql",
					map[string]any{
						"input": map[string]string{
							"name": issueRepository.Name,
							"url":  issueRepository.Url,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.IssueRepository.Name).To(Equal(issueRepository.Name))
				Expect(*respData.IssueRepository.URL).To(Equal(issueRepository.Url))
			})
		})
	})
})

var _ = Describe("Updating issueRepository via API", Label("e2e", "IssueRepositories"), func() {
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
			It("updates issueRepository", func() {
				issueRepository := seedCollection.IssueRepositoryRows[0].AsIssueRepository()
				issueRepository.Name = "SecretRepository"
				issueRepository.Url = "https://google.com"

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					IssueRepository model.IssueRepository `json:"updateIssueRepository"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issueRepository/update.graphql",
					map[string]any{
						"id": fmt.Sprintf("%d", issueRepository.Id),
						"input": map[string]string{
							"name": issueRepository.Name,
							"url":  issueRepository.Url,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.IssueRepository.Name).To(Equal(issueRepository.Name))
				Expect(*respData.IssueRepository.URL).To(Equal(issueRepository.Url))
			})
		})
	})
})

var _ = Describe("Deleting IssueRepository via API", Label("e2e", "IssueRepositories"), func() {
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
			It("deletes issueRepository", func() {
				id := fmt.Sprintf("%d", seedCollection.IssueRepositoryRows[0].Id.Int64)

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Id string `json:"deleteIssueRepository"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issueRepository/delete.graphql",
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
