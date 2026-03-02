// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Getting Services via API", Label("e2e", "Services"), func() {
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				Services model.ServiceConnection `json:"Services"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/service/minimal.graphql",
				map[string]any{
					"filter": map[string]string{},
					"first":  10,
					"after":  "",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.Services.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct result count", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Services model.ServiceConnection `json:"Services"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/minimal.graphql",
					map[string]any{
						"filter": map[string]string{},
						"first":  5,
						"after":  "",
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Services.TotalCount).To(Equal(len(seedCollection.ServiceRows)))
				Expect(len(respData.Services.Edges)).To(Equal(5))
			})
		})
		Context("and we request metadata", func() {
			It("returns correct metadata counts", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Services model.ServiceConnection `json:"Services"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/withObjectMetadata.graphql",
					map[string]any{
						"filter": map[string]string{},
						"first":  5,
						"after":  "",
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				for _, serviceEdge := range respData.Services.Edges {
					imCount := 0
					ciCount := 0
					for _, ciEdge := range serviceEdge.Node.ComponentInstances.Edges {
						imCount += ciEdge.Node.IssueMatches.TotalCount
						ciCount += *ciEdge.Node.Count
					}
					Expect(serviceEdge.Node.ObjectMetadata.IssueMatchCount).To(Equal(imCount))
					Expect(serviceEdge.Node.ObjectMetadata.ComponentInstanceCount).To(Equal(ciCount))
				}
			})
		})
		Context("and we query to resolve levels of relations", Label("directRelations.graphql"), func() {
			respData := struct {
				Services model.ServiceConnection `json:"Services"`
			}{}
			BeforeEach(func() {
				resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Services model.ServiceConnection `json:"Services"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/directRelations.graphql",
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
				Expect(respData.Services.TotalCount).To(Equal(len(seedCollection.ServiceRows)))
				Expect(len(respData.Services.Edges)).To(Equal(5))
			})

			It("- returns the expected content", func() {
				// this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
				// a complete verification
				// additional checks are added based on bugs discovered during usage

				for _, service := range respData.Services.Edges {
					for _, owner := range service.Node.Owners.Edges {
						Expect(owner.Node.ID).ToNot(BeNil(), "owner has a ID set")
						Expect(owner.Node.Name).ToNot(BeNil(), "owner has a name set")
						Expect(owner.Node.UniqueUserID).ToNot(BeNil(), "owner has a uniqueUserId set")
						Expect(owner.Node.Type).ToNot(BeNil(), "owner has a type set")

						_, ownerFound := lo.Find(seedCollection.OwnerRows, func(row mariadb.OwnerRow) bool {
							return fmt.Sprintf("%d", row.UserId.Int64) == owner.Node.ID && // correct owner
								fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // belongs actually to the service
						})
						Expect(ownerFound).To(BeTrue(), "attached owner does exist and belongs to service")
					}

					for _, sg := range service.Node.SupportGroups.Edges {
						Expect(sg.Node.ID).ToNot(BeNil(), "supportGroup has a ID set")
						Expect(sg.Node.Ccrn).ToNot(BeNil(), "supportGroup has a ccrn set")

						_, sgFound := lo.Find(seedCollection.SupportGroupServiceRows, func(row mariadb.SupportGroupServiceRow) bool {
							return fmt.Sprintf("%d", row.SupportGroupId.Int64) == sg.Node.ID && // correct sg
								fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // belongs actually to the service
						})
						Expect(sgFound).To(BeTrue(), "attached supportGroup does exist and belongs to service")
					}

					for _, ir := range service.Node.IssueRepositories.Edges {
						Expect(ir.Node.ID).ToNot(BeNil(), "issueRepository has a ID set")
						Expect(ir.Node.Name).ToNot(BeNil(), "issueRepository has a name set")
						Expect(ir.Node.URL).ToNot(BeNil(), "issueRepository has a url set")
						Expect(ir.Priority).ToNot(BeNil(), "issueRepository has a priority set")

						_, irFound := lo.Find(seedCollection.IssueRepositoryServiceRows, func(row mariadb.IssueRepositoryServiceRow) bool {
							return fmt.Sprintf("%d", row.IssueRepositoryId.Int64) == ir.Node.ID && // correct ar
								fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // belongs actually to the service
						})
						Expect(irFound).To(BeTrue(), "attached issueRepository does exist and belongs to service")
					}

					for _, ci := range service.Node.ComponentInstances.Edges {
						Expect(ci.Node.ID).ToNot(BeNil(), "componentInstance has a ID set")
						Expect(ci.Node.Ccrn).ToNot(BeNil(), "componentInstance has a ccrn set")
						Expect(ci.Node.Count).ToNot(BeNil(), "componentInstance has a count set")

						_, ciFound := lo.Find(seedCollection.ComponentInstanceRows, func(row mariadb.ComponentInstanceRow) bool {
							return fmt.Sprintf("%d", row.Id.Int64) == ci.Node.ID &&
								row.CCRN.String == *ci.Node.Ccrn &&
								int(row.Count.Int16) == *ci.Node.Count
						})
						Expect(ciFound).To(BeTrue(), "attached componentInstance does exist and belongs to service")
					}

					for _, im := range service.Node.IssueMatches.Edges {
						Expect(im.Node.ID).ToNot(BeNil(), "issueMatch has a ID set")
						Expect(im.Node.Status).ToNot(BeNil(), "issueMatch has a status set")
						Expect(im.Node.RemediationDate).ToNot(BeNil(), "issueMatch has a remediationDate set")
						Expect(im.Node.DiscoveryDate).ToNot(BeNil(), "issueMatch has a discoveryDate set")
						Expect(im.Node.TargetRemediationDate).ToNot(BeNil(), "issueMatch has a targetRemediationDate set")
					}

					for _, remediation := range service.Node.Remediations.Edges {
						Expect(remediation.Node.ID).ToNot(BeNil(), "remediation has a ID set")
						Expect(remediation.Node.Description).ToNot(BeNil(), "remediation has a description set")
						Expect(remediation.Node.Service).ToNot(BeNil(), "remediation has a service set")
						Expect(remediation.Node.Vulnerability).ToNot(BeNil(), "remediation has a vulnerability set")
						Expect(remediation.Node.RemediationDate).ToNot(BeNil(), "remediation has a remediation date set")
						Expect(remediation.Node.ExpirationDate).ToNot(BeNil(), "remediation has a expiration date set")
						Expect(remediation.Node.RemediatedBy).ToNot(BeNil(), "remediation has a remediated bby set")
					}
				}
			})
			It("- returns the expected PageInfo", func() {
				Expect(*respData.Services.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.Services.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.Services.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.Services.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.Services.PageInfo.PageNumber).To(Equal(1), "Correct page number")
			})
		})
		Context("and we use order", Label("withOrder.graphql"), func() {
			c := collate.New(language.English)

			It("can order by ccrn", Label("withOrder.graphql"), func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Services model.ServiceConnection `json:"Services"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/withOrder.graphql",
					map[string]any{
						"filter": map[string]string{},
						"first":  10,
						"after":  "",
						"orderBy": []map[string]string{
							{"by": "ccrn", "direction": "asc"},
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				By("- returns the correct result count", func() {
					Expect(respData.Services.TotalCount).To(Equal(len(seedCollection.ServiceRows)))
					Expect(len(respData.Services.Edges)).To(Equal(10))
				})

				By("- returns the expected content in order", func() {
					var prev string = ""
					for _, im := range respData.Services.Edges {
						Expect(c.CompareString(*im.Node.Ccrn, prev)).Should(BeNumerically(">=", 0))
						prev = *im.Node.Ccrn
					}
				})
			})
		})
	})
	loadTestData := func() ([]mariadb.ComponentInstanceRow, []mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, []mariadb.IssueMatchRow, error) {
		issueVariants, err := test.LoadIssueVariants(test.GetTestDataPath("../database/mariadb/testdata/component_version_order/issue_variant.json"))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		cvIssues, err := test.LoadComponentVersionIssues(test.GetTestDataPath("../database/mariadb/testdata/service_order/component_version_issue.json"))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		componentInstances, err := test.LoadComponentInstances(test.GetTestDataPath("../database/mariadb/testdata/service_order/component_instance.json"))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		issueMatches, err := test.LoadIssueMatches(test.GetTestDataPath("../database/mariadb/testdata/service_order/issue_match.json"))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return componentInstances, issueVariants, cvIssues, issueMatches, nil
	}
	When("issueCounts are involved", func() {
		BeforeEach(func() {
			seeder.SeedIssueRepositories()
			seeder.SeedIssues(10)
			components := seeder.SeedComponents(1)
			seeder.SeedComponentVersions(10, components)
			seeder.SeedServices(5)
			componentInstances, issueVariants, componentVersionIssues, issueMatches, err := loadTestData()
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
			for _, ci := range componentInstances {
				_, err := seeder.InsertFakeComponentInstance(ci)
				Expect(err).To(BeNil())
			}
			for _, im := range issueMatches {
				_, err := seeder.InsertFakeIssueMatch(im)
				Expect(err).To(BeNil())
			}
			err = seeder.RefreshServiceIssueCounters()
			Expect(err).To(BeNil())
		})

		runOrderTest := func(orderDirection string, expectedOrder []string) {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				Services model.ServiceConnection `json:"Services"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/service/withOrder.graphql",
				map[string]any{
					"filter": map[string]string{},
					"first":  10,
					"after":  "",
					"orderBy": []map[string]string{
						{"by": "severity", "direction": orderDirection},
					},
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.Services.TotalCount).To(Equal(5))
			Expect(len(respData.Services.Edges)).To(Equal(5))
			for i, id := range expectedOrder {
				Expect(respData.Services.Edges[i].Node.ID).To(BeEquivalentTo(id))
			}
		}

		It("can order descending by severity", func() {
			runOrderTest("desc", []string{"1", "3", "4", "5", "2"})
		})

		It("can order ascending by severity", func() {
			runOrderTest("asc", []string{"2", "5", "4", "3", "1"})
		})

		It("can count issues", Label("issueCount"), func() {
		})
	})
})

var _ = Describe("Creating Service via API", Label("e2e", "Services"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var service entity.Service
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		BeforeEach(func() {
			seeder.SeedDbWithNFakeData(10)
			service = testentity.NewFakeServiceEntity()
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new service", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Service model.Service `json:"createService"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/create.graphql",
					map[string]any{
						"input": map[string]string{
							"ccrn": service.CCRN,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.Service.Ccrn).To(Equal(service.CCRN))
			})
		})
	})
})

var _ = Describe("Updating service via API", Label("e2e", "Services"), func() {
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates service", func() {
				service := seedCollection.ServiceRows[0].AsService()
				service.CCRN = "SecretService"

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Service model.Service `json:"updateService"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/update.graphql",
					map[string]any{
						"id": fmt.Sprintf("%d", service.Id),
						"input": map[string]string{
							"ccrn": service.CCRN,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(*respData.Service.Ccrn).To(Equal(service.CCRN))
			})
		})
	})
})

var _ = Describe("Deleting Service via API", Label("e2e", "Services"), func() {
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes service", func() {
				id := fmt.Sprintf("%d", seedCollection.ServiceRows[0].Id.Int64)

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Id string `json:"deleteService"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/delete.graphql",
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

var _ = Describe("Modifying Owner of Service via API", Label("e2e", "Services"), func() {
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", func() {
			It("adds owner to service", Label("addOwner.graphql"), func() {
				service := seedCollection.ServiceRows[0].AsService()
				ownerIds := lo.FilterMap(seedCollection.OwnerRows, func(row mariadb.OwnerRow, _ int) (int64, bool) {
					if row.ServiceId.Int64 == service.Id {
						return row.UserId.Int64, true
					}
					return 0, false
				})

				ownerRow, _ := lo.Find(seedCollection.UserRows, func(row mariadb.UserRow) bool {
					return !lo.Contains(ownerIds, row.Id.Int64)
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Service model.Service `json:"addOwnerToService"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/addOwner.graphql",
					map[string]any{
						"serviceId": fmt.Sprintf("%d", service.Id),
						"userId":    fmt.Sprintf("%d", ownerRow.Id.Int64),
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				_, found := lo.Find(respData.Service.Owners.Edges, func(edge *model.UserEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", ownerRow.Id.Int64)
				})

				Expect(respData.Service.ID).To(Equal(fmt.Sprintf("%d", service.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes owner from service", Label("removeOwner.graphql"), func() {
				service := seedCollection.ServiceRows[0].AsService()

				ownerRow, _ := lo.Find(seedCollection.OwnerRows, func(row mariadb.OwnerRow) bool {
					return row.ServiceId.Int64 == service.Id
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Service model.Service `json:"removeOwnerFromService"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/removeOwner.graphql",
					map[string]any{
						"serviceId": fmt.Sprintf("%d", service.Id),
						"userId":    fmt.Sprintf("%d", ownerRow.UserId.Int64),
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				_, found := lo.Find(respData.Service.Owners.Edges, func(edge *model.UserEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", ownerRow.UserId.Int64)
				})

				Expect(respData.Service.ID).To(Equal(fmt.Sprintf("%d", service.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})

var _ = Describe("Modifying IssueRepository of Service via API", Label("e2e", "Services"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var priority int64
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			priority = 1
		})

		Context("and a mutation query is performed", func() {
			It("adds issueRepository to service", Label("addIssueRepository.graphql"), func() {
				service := seedCollection.ServiceRows[0].AsService()
				// find all issueRepositories that are attached to the service
				issueRepositoryIds := lo.FilterMap(seedCollection.IssueRepositoryServiceRows, func(row mariadb.IssueRepositoryServiceRow, _ int) (int64, bool) {
					if row.ServiceId.Int64 == service.Id {
						return row.IssueRepositoryId.Int64, true
					}
					return 0, false
				})
				// find an issueRepository that is not attached to the service
				issueRepositoryRow, _ := lo.Find(seedCollection.IssueRepositoryRows, func(row mariadb.BaseIssueRepositoryRow) bool {
					return !lo.Contains(issueRepositoryIds, row.Id.Int64)
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Service model.Service `json:"addIssueRepositoryToService"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/addIssueRepository.graphql",
					map[string]any{
						"serviceId":         fmt.Sprintf("%d", service.Id),
						"issueRepositoryId": fmt.Sprintf("%d", issueRepositoryRow.Id.Int64),
						"priority":          fmt.Sprintf("%d", priority),
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				_, found := lo.Find(respData.Service.IssueRepositories.Edges, func(edge *model.IssueRepositoryEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", issueRepositoryRow.Id.Int64)
				})

				Expect(respData.Service.ID).To(Equal(fmt.Sprintf("%d", service.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes issueRepository from service", Label("removeIssueRepository.graphql"), func() {
				service := seedCollection.ServiceRows[0].AsService()

				// find an issueRepository that is attached to the service
				issueRepositoryRow, _ := lo.Find(seedCollection.IssueRepositoryServiceRows, func(row mariadb.IssueRepositoryServiceRow) bool {
					return row.ServiceId.Int64 == service.Id
				})

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Service model.Service `json:"removeIssueRepositoryFromService"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/removeIssueRepository.graphql",
					map[string]any{
						"serviceId":         fmt.Sprintf("%d", service.Id),
						"issueRepositoryId": fmt.Sprintf("%d", issueRepositoryRow.IssueRepositoryId.Int64),
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				_, found := lo.Find(respData.Service.IssueRepositories.Edges, func(edge *model.IssueRepositoryEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", issueRepositoryRow.IssueRepositoryId.Int64)
				})

				Expect(respData.Service.ID).To(Equal(fmt.Sprintf("%d", service.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})
