// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"database/sql"
	"fmt"
	"time"

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
)

var _ = Describe("Getting IssueMatches via API", Label("e2e", "IssueMatches"), func() {
	var (
		seeder *test.DatabaseSeeder
		s      *server.Server
		cfg    util.Config
		db     *mariadb.SqlDatabase
	)

	BeforeEach(func() {
		var err error

		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		cfg.AuthzOpenFgaApiUrl = ""
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/issueMatch/minimal.graphql",
				map[string]any{
					"filter": map[string]string{},
					"first":  10,
					"after":  "",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
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
					respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/issueMatch/minimal.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  5,
							"after":  "",
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())
					Expect(
						respData.IssueMatches.TotalCount,
					).To(Equal(len(seedCollection.IssueMatchRows)))
					Expect(len(respData.IssueMatches.Edges)).To(Equal(5))
				})
			})
			Context(
				"and  we query to resolve levels of relations",
				Label("directRelations.graphql"),
				func() {
					respData := struct {
						IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
					}{}

					BeforeEach(func() {
						resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
							IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
						}](
							cfg.Port,
							"../api/graphql/graph/queryCollection/issueMatch/directRelations.graphql",
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
						Expect(
							respData.IssueMatches.TotalCount,
						).To(Equal(len(seedCollection.IssueMatchRows)))
						Expect(len(respData.IssueMatches.Edges)).To(Equal(5))
					})

					It("- returns the expected content", func() {
						// this just checks partial attributes to check whatever every sub-relation
						// does resolve some reasonable data and is not doing
						// a complete verification
						// additional checks are added based on bugs discovered during usage
						for _, im := range respData.IssueMatches.Edges {
							Expect(im.Node.ID).ToNot(BeNil(), "issueMatch has a ID set")
							Expect(im.Node.Status).ToNot(BeNil(), "issueMatch has a status set")
							Expect(
								im.Node.RemediationDate,
							).ToNot(BeNil(), "issueMatch has a remediation date set")
							Expect(
								im.Node.TargetRemediationDate,
							).ToNot(BeNil(), "issueMatch has a target remediation date set")

							if im.Node.Severity != nil {
								Expect(
									im.Node.Severity.Value,
								).ToNot(BeNil(), "issueMatch has a severity value set")
								Expect(
									im.Node.Severity.Score,
								).ToNot(BeNil(), "issueMatch has a severity score set")
							}

							for _, eiv := range im.Node.EffectiveIssueVariants.Edges {
								Expect(
									eiv.Node.ID,
								).ToNot(BeNil(), "effectiveIssueVariant has a ID set")
								Expect(
									eiv.Node.Description,
								).ToNot(BeNil(), "effectiveIssueVariant has a description set")
								Expect(
									eiv.Node.SecondaryName,
								).ToNot(BeNil(), "effectiveIssueVariant has a name set")
							}

							issue := im.Node.Issue
							Expect(issue.ID).ToNot(BeNil(), "issue has a ID set")
							Expect(
								issue.LastModified,
							).ToNot(BeNil(), "issue has a lastModified set")
						}
					})
					It("- returns the expected PageInfo", func() {
						Expect(
							*respData.IssueMatches.PageInfo.HasNextPage,
						).To(BeTrue(), "hasNextPage is set")
						Expect(
							*respData.IssueMatches.PageInfo.HasPreviousPage,
						).To(BeFalse(), "hasPreviousPage is set")
						Expect(
							respData.IssueMatches.PageInfo.NextPageAfter,
						).ToNot(BeNil(), "nextPageAfter is set")
						Expect(
							len(respData.IssueMatches.PageInfo.Pages),
						).To(Equal(2), "Correct amount of pages")
						Expect(
							*respData.IssueMatches.PageInfo.PageNumber,
						).To(Equal(1), "Correct page number")
					})
				},
			)
			Context("we use ordering", Label("withOrder.graphql"), func() {
				It("can order by primaryName", Label("withOrder.graphql"), func() {
					respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
						IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
					}](
						cfg.Port,
						"../api/graphql/graph/queryCollection/issueMatch/withOrder.graphql",
						map[string]any{
							"filter": map[string]string{},
							"first":  10,
							"after":  "",
							"orderBy": []map[string]string{
								{"by": "primaryName", "direction": "asc"},
							},
						},
						nil,
					)

					Expect(err).ToNot(HaveOccurred())
					By("- returns the correct result count", func() {
						Expect(
							respData.IssueMatches.TotalCount,
						).To(Equal(len(seedCollection.IssueMatchRows)))
						Expect(len(respData.IssueMatches.Edges)).To(Equal(10))
					})

					By("- returns the expected content in order", func() {
						prev := ""
						for _, im := range respData.IssueMatches.Edges {
							Expect(*im.Node.Issue.PrimaryName >= prev).Should(BeTrue())
							prev = *im.Node.Issue.PrimaryName
						}
					})
				})

				It(
					"can order by primaryName and targetRemediationDate",
					Label("withOrder.graphql"),
					func() {
						respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
							IssueMatches model.IssueMatchConnection `json:"IssueMatches"`
						}](
							cfg.Port,
							"../api/graphql/graph/queryCollection/issueMatch/withOrder.graphql",
							map[string]any{
								"filter": map[string]string{},
								"first":  10,
								"after":  "",
								"orderBy": []map[string]string{
									{"by": "primaryName", "direction": "asc"},
									{"by": "targetRemediationDate", "direction": "desc"},
								},
							},
							nil,
						)

						Expect(err).ToNot(HaveOccurred())
						By("- returns the correct result count", func() {
							Expect(
								respData.IssueMatches.TotalCount,
							).To(Equal(len(seedCollection.IssueMatchRows)))
							Expect(len(respData.IssueMatches.Edges)).To(Equal(10))
						})

						By("- returns the expected content in order", func() {
							prevPn := ""
							prevTrd := time.Now()

							for _, im := range respData.IssueMatches.Edges {
								if *im.Node.Issue.PrimaryName == prevPn {
									trd, err := time.Parse(
										"2006-01-02T15:04:05Z",
										*im.Node.TargetRemediationDate,
									)
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
					},
				)
			})
		})
	})
})

var _ = Describe("Creating IssueMatch via API", Label("e2e", "IssueMatches"), func() {
	var (
		seeder     *test.DatabaseSeeder
		s          *server.Server
		cfg        util.Config
		issueMatch entity.IssueMatch
		db         *mariadb.SqlDatabase
	)

	BeforeEach(func() {
		var err error

		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		cfg.AuthzOpenFgaApiUrl = ""
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
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					IssueMatch model.IssueMatch `json:"createIssueMatch"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issueMatch/create.graphql",
					map[string]any{
						"input": map[string]any{
							"status":              issueMatch.Status,
							"userId":              issueMatch.UserId,
							"componentInstanceId": issueMatch.ComponentInstanceId,
							"issueId":             fmt.Sprintf("%d", issueMatch.IssueId),
							"remediationDate": issueMatch.RemediationDate.Format(
								time.RFC3339,
							),
							"targetRemediationDate": issueMatch.TargetRemediationDate.Format(
								time.RFC3339,
							),
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.IssueMatch.Status.String()).To(Equal(issueMatch.Status.String()))
				Expect(
					*respData.IssueMatch.IssueID,
				).To(Equal(fmt.Sprintf("%d", issueMatch.IssueId)))
				Expect(*respData.IssueMatch.UserID).To(Equal(fmt.Sprintf("%d", issueMatch.UserId)))
				Expect(
					*respData.IssueMatch.ComponentInstanceID,
				).To(Equal(fmt.Sprintf("%d", issueMatch.ComponentInstanceId)))
				Expect(
					*respData.IssueMatch.RemediationDate,
				).To(Equal(issueMatch.RemediationDate.Format(time.RFC3339)))
				Expect(
					*respData.IssueMatch.TargetRemediationDate,
				).To(Equal(issueMatch.TargetRemediationDate.Format(time.RFC3339)))
			})
		})
	})
})

var _ = Describe("Updating issueMatch via API", Label("e2e", "IssueMatches"), func() {
	var (
		seeder *test.DatabaseSeeder
		s      *server.Server
		cfg    util.Config
		db     *mariadb.SqlDatabase
	)

	BeforeEach(func() {
		var err error

		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		cfg.AuthzOpenFgaApiUrl = ""
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
				issueMatch := seedCollection.IssueMatchRows[0].AsIssueMatch()
				issueMatch.RemediationDate = issueMatch.RemediationDate.Add(time.Hour * 24 * 7)

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					IssueMatch model.IssueMatch `json:"updateIssueMatch"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issueMatch/update.graphql",
					map[string]any{
						"id": fmt.Sprintf("%d", issueMatch.Id),
						"input": map[string]string{
							"remediationDate": issueMatch.RemediationDate.Format(time.RFC3339),
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(
					*respData.IssueMatch.RemediationDate,
				).To(Equal(issueMatch.RemediationDate.Format(time.RFC3339)))
			})
		})
	})
})

var _ = Describe("Deleting IssueMatch via API", Label("e2e", "IssueMatches"), func() {
	var (
		seeder *test.DatabaseSeeder
		s      *server.Server
		cfg    util.Config
		db     *mariadb.SqlDatabase
	)

	BeforeEach(func() {
		var err error

		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		cfg.AuthzOpenFgaApiUrl = ""
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
				id := fmt.Sprintf("%d", seedCollection.IssueVariantRows[0].Id.Int64)

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Id string `json:"deleteIssueMatch"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issueMatch/delete.graphql",
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

var _ = Describe("Getting IssueMatchesOverdue via API", Label("e2e", "IssueMatchesOverdue"), func() {
	var (
		seeder *test.DatabaseSeeder
		s      *server.Server
		cfg    util.Config
		db     *mariadb.SqlDatabase
		// sc holds FK rows needed by insertOverdueMatch; it contains no IssueMatches.
		sc *test.SeedCollection
	)

	BeforeEach(func() {
		var err error

		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		cfg.AuthzOpenFgaApiUrl = ""
		s = e2e_common.NewRunningServer(cfg)

		// Seed only the FK dependencies (no IssueMatches), so each test
		// controls exactly which IssueMatches exist and their TRDs.
		users := seeder.SeedUsers(1)
		services := seeder.SeedServices(1)
		components := seeder.SeedComponents(1)
		componentVersions := seeder.SeedComponentVersions(1, components)
		componentInstances := seeder.SeedComponentInstances(1, componentVersions, services)
		issues := seeder.SeedIssues(1)
		sc = &test.SeedCollection{
			UserRows:              users,
			ServiceRows:           services,
			ComponentRows:         components,
			ComponentVersionRows:  componentVersions,
			ComponentInstanceRows: componentInstances,
			IssueRows:             issues,
		}
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	queryOverdue := func(port string, first int) (model.IssueMatchConnection, error) {
		type overdueResp struct {
			IssueMatchesOverdue model.IssueMatchConnection `json:"IssueMatchesOverdue"`
		}

		respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[overdueResp](
			port,
			"../api/graphql/graph/queryCollection/issueMatch/overdue.graphql",
			map[string]any{
				"filter": map[string]string{},
				"first":  first,
				"after":  "",
			},
			nil,
		)
		if err != nil {
			return model.IssueMatchConnection{}, err
		}

		return respData.IssueMatchesOverdue, nil
	}

	insertOverdueMatch := func(status entity.IssueMatchStatusValue, trd time.Time) {
		row := test.NewFakeIssueMatch()
		row.IssueId = sql.NullInt64{Int64: sc.IssueRows[0].Id.Int64, Valid: true}
		row.ComponentInstanceId = sql.NullInt64{Int64: sc.ComponentInstanceRows[0].Id.Int64, Valid: true}
		row.UserId = sql.NullInt64{Int64: sc.UserRows[0].Id.Int64, Valid: true}
		row.Status = sql.NullString{String: status.String(), Valid: true}
		row.TargetRemediationDate = sql.NullTime{Time: trd, Valid: true}
		_, err := seeder.InsertFakeIssueMatch(row)
		Expect(err).To(BeNil())
	}

	When("the database is empty", func() {
		It("returns an empty result set", func() {
			result, err := queryOverdue(cfg.Port, 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.TotalCount).To(Equal(0))
			Expect(result.Edges).To(BeEmpty())
		})
	})

	When("the database has only overdue IssueMatches with valid statuses", func() {
		BeforeEach(func() {
			pastDate := time.Now().UTC().Add(-7 * 24 * time.Hour)
			insertOverdueMatch(entity.IssueMatchStatusValuesNew, pastDate)
			insertOverdueMatch(entity.IssueMatchStatusValuesRiskAccepted, pastDate)
			insertOverdueMatch(entity.IssueMatchStatusValuesFalsePositive, pastDate)
		})

		It("returns all three overdue matches", func() {
			result, err := queryOverdue(cfg.Port, 10)
			Expect(err).ToNot(HaveOccurred())
			By("returning the correct total count", func() {
				Expect(result.TotalCount).To(Equal(3))
			})
			By("returning nodes with required fields populated", func() {
				for _, edge := range result.Edges {
					Expect(edge.Node.TargetRemediationDate).NotTo(BeNil())
					Expect(edge.Node.Status).NotTo(BeNil())
					Expect(edge.Node.Severity).NotTo(BeNil())
				}
			})
		})
	})

	When("the database has only future-dated IssueMatches with valid status", func() {
		BeforeEach(func() {
			futureDate := time.Now().UTC().Add(30 * 24 * time.Hour)
			insertOverdueMatch(entity.IssueMatchStatusValuesNew, futureDate)
			insertOverdueMatch(entity.IssueMatchStatusValuesNew, futureDate)
		})

		It("returns an empty result set (not yet overdue)", func() {
			result, err := queryOverdue(cfg.Port, 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.TotalCount).To(Equal(0))
		})
	})

	When("the database has only mitigated IssueMatches with past target date", func() {
		BeforeEach(func() {
			pastDate := time.Now().UTC().Add(-7 * 24 * time.Hour)
			insertOverdueMatch(entity.IssueMatchStatusValuesMitigated, pastDate)
			insertOverdueMatch(entity.IssueMatchStatusValuesMitigated, pastDate)
		})

		It("returns an empty result set (already mitigated)", func() {
			result, err := queryOverdue(cfg.Port, 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.TotalCount).To(Equal(0))
		})
	})

	When("the database has a mix of overdue, future, and mitigated IssueMatches", func() {
		BeforeEach(func() {
			pastDate := time.Now().UTC().Add(-7 * 24 * time.Hour)
			futureDate := time.Now().UTC().Add(30 * 24 * time.Hour)
			// 2 overdue (valid statuses)
			insertOverdueMatch(entity.IssueMatchStatusValuesNew, pastDate)
			insertOverdueMatch(entity.IssueMatchStatusValuesRiskAccepted, pastDate)
			// 2 future (valid status but not overdue)
			insertOverdueMatch(entity.IssueMatchStatusValuesNew, futureDate)
			insertOverdueMatch(entity.IssueMatchStatusValuesNew, futureDate)
			// 1 past but mitigated
			insertOverdueMatch(entity.IssueMatchStatusValuesMitigated, pastDate)
		})

		It("returns only the 2 overdue matches", func() {
			result, err := queryOverdue(cfg.Port, 10)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.TotalCount).To(Equal(2))
		})
	})

	When("the database has more overdue entries than the page size", func() {
		BeforeEach(func() {
			pastDate := time.Now().UTC().Add(-7 * 24 * time.Hour)
			for i := 0; i < 5; i++ {
				insertOverdueMatch(entity.IssueMatchStatusValuesNew, pastDate)
			}
		})

		It("honours the first parameter and sets hasNextPage", func() {
			result, err := queryOverdue(cfg.Port, 2)
			Expect(err).ToNot(HaveOccurred())
			By("returning only the requested page size", func() {
				Expect(len(result.Edges)).To(Equal(2))
			})
			By("indicating there are more results", func() {
				Expect(result.PageInfo).NotTo(BeNil())
				Expect(*result.PageInfo.HasNextPage).To(BeTrue())
			})
		})
	})
})
