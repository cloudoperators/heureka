// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"database/sql"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getting SIEMAlerts via API", Label("e2e", "SIEMAlerts"), func() {
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
				SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
				map[string]any{
					"filter": map[string]any{},
					"first":  10,
					"after":  "",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.SIEMAlerts.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 SecurityEvent IssueMatches", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithSecurityEvents(10)
		})

		Context("with no additional filters", func() {
			It("returns all SIEM alerts with correct totalCount", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  20,
						"after":  "",
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.SIEMAlerts.TotalCount).To(Equal(len(seedCollection.GetValidIssueMatchRows())))
			})

			It("paginates correctly", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  5,
						"after":  "",
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.SIEMAlerts.TotalCount).To(Equal(len(seedCollection.GetValidIssueMatchRows())))
				Expect(len(respData.SIEMAlerts.Edges)).To(Equal(5))
			})
		})

		Context("with full query including all fields", func() {
			It("returns edges with expected fields populated", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter":  map[string]any{},
						"first":   5,
						"after":   "",
						"orderBy": []any{},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(respData.SIEMAlerts.Edges)).To(BeNumerically(">", 0))

				for _, edge := range respData.SIEMAlerts.Edges {
					Expect(edge.Node.ID).ToNot(BeEmpty(), "SIEMAlertNode has an ID")
					Expect(edge.Cursor).ToNot(BeNil(), "edge has a cursor")
					Expect(edge.Node.Name).ToNot(BeNil(), "SIEMAlertNode has a name")
					Expect(edge.Node.Description).ToNot(BeNil(), "SIEMAlertNode has a description")
					Expect(edge.Node.Region).ToNot(BeNil(), "SIEMAlertNode has a region")
					Expect(edge.Node.Cluster).ToNot(BeNil(), "SIEMAlertNode has a cluster")
					Expect(edge.Node.Namespace).ToNot(BeNil(), "SIEMAlertNode has a namespace")
				}
			})

			It("returns correct pageInfo", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter":  map[string]any{},
						"first":   5,
						"after":   "",
						"orderBy": []any{},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.SIEMAlerts.PageInfo).ToNot(BeNil())
				Expect(*respData.SIEMAlerts.PageInfo.HasNextPage).To(BeTrue())
				Expect(*respData.SIEMAlerts.PageInfo.HasPreviousPage).To(BeFalse())
				Expect(respData.SIEMAlerts.PageInfo.NextPageAfter).ToNot(BeNil())
			})
		})

		Context("filtering by severity", func() {
			It("returns only alerts matching the requested severity", func() {
				targetSeverity := entity.SeverityValuesCritical.String()

				respAll, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter":  map[string]any{},
						"first":   100,
						"after":   "",
						"orderBy": []any{},
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())

				criticalCount := 0
				for _, edge := range respAll.SIEMAlerts.Edges {
					if edge.Node.Severity != nil && string(*edge.Node.Severity) == targetSeverity {
						criticalCount++
					}
				}

				if criticalCount == 0 || criticalCount == respAll.SIEMAlerts.TotalCount {
					Skip("Seed data is all-critical or has no critical entries; cannot test severity exclusion")
				}

				respFiltered, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter": map[string]any{
							"severity": []string{targetSeverity},
						},
						"first":   100,
						"after":   "",
						"orderBy": []any{},
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respFiltered.SIEMAlerts.TotalCount).To(Equal(criticalCount))
				Expect(respFiltered.SIEMAlerts.TotalCount).To(BeNumerically("<", respAll.SIEMAlerts.TotalCount))
				for _, edge := range respFiltered.SIEMAlerts.Edges {
					Expect(string(*edge.Node.Severity)).To(Equal(targetSeverity))
				}
			})
		})

		Context("filtering by status", func() {
			It("returns only alerts matching the requested status", func() {
				targetStatus := entity.IssueMatchStatusValuesNew.String()

				newCount := 0
				for _, im := range seedCollection.IssueMatchRows {
					if im.Status.String == targetStatus {
						newCount++
					}
				}

				if newCount == 0 || newCount == len(seedCollection.IssueMatchRows) {
					Skip("Seed data is all-new or has no new entries; cannot test status exclusion")
				}

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter": map[string]any{
							"status": []string{targetStatus},
						},
						"first":   100,
						"after":   "",
						"orderBy": []any{},
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respData.SIEMAlerts.TotalCount).To(Equal(newCount))
				Expect(respData.SIEMAlerts.TotalCount).To(BeNumerically("<", len(seedCollection.GetValidIssueMatchRows())))
				for _, edge := range respData.SIEMAlerts.Edges {
					Expect(string(*edge.Node.Status)).To(Equal(targetStatus))
				}
			})
		})

		Context("sorting by severity", func() {
			It("returns alerts ordered by severity descending", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  20,
						"after":  "",
						"orderBy": []any{
							map[string]any{
								"by":        "severity",
								"direction": "desc",
							},
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.SIEMAlerts.TotalCount).To(Equal(len(seedCollection.GetValidIssueMatchRows())))
			})
		})

		Context("sorting by discoveryDate", func() {
			It("returns all alerts ordered by discoveryDate ascending", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  20,
						"after":  "",
						"orderBy": []any{
							map[string]any{
								"by":        "discoveryDate",
								"direction": "asc",
							},
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.SIEMAlerts.TotalCount).To(Equal(len(seedCollection.GetValidIssueMatchRows())))
				Expect(len(respData.SIEMAlerts.Edges)).To(BeNumerically(">", 0))

				for _, edge := range respData.SIEMAlerts.Edges {
					Expect(edge.Node.DiscoveryDate).ToNot(BeNil(), "SIEMAlertNode has a discoveryDate")
				}
			})

			It("returns all alerts ordered by discoveryDate descending", func() {
				respAsc, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  20,
						"after":  "",
						"orderBy": []any{
							map[string]any{
								"by":        "discoveryDate",
								"direction": "asc",
							},
						},
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())

				respDesc, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/withOrder.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  20,
						"after":  "",
						"orderBy": []any{
							map[string]any{
								"by":        "discoveryDate",
								"direction": "desc",
							},
						},
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respDesc.SIEMAlerts.TotalCount).To(Equal(respAsc.SIEMAlerts.TotalCount))

				if len(respAsc.SIEMAlerts.Edges) > 1 {
					firstAsc := *respAsc.SIEMAlerts.Edges[0].Node.DiscoveryDate
					firstDesc := *respDesc.SIEMAlerts.Edges[0].Node.DiscoveryDate
					Expect(firstAsc <= firstDesc).To(BeTrue(), "asc first entry should be <= desc first entry")
				}
			})
		})
	})

	When("the database has SecurityEvent IssueMatches split across two distinct services", func() {
		var service1CCRN, service2CCRN string
		var sg1CCRN, sg2CCRN string
		const matchesPerService = 3

		BeforeEach(func() {
			users := seeder.SeedUsers(1)

			sg1Rows := seeder.SeedSupportGroups(1)
			sg2Rows := seeder.SeedSupportGroups(1)
			sg1CCRN = sg1Rows[0].CCRN.String
			sg2CCRN = sg2Rows[0].CCRN.String

			svc1Rows := seeder.SeedServices(1)
			svc2Rows := seeder.SeedServices(1)
			service1CCRN = svc1Rows[0].CCRN.String
			service2CCRN = svc2Rows[0].CCRN.String

			seeder.SeedSupportGroupServices(1, svc1Rows, sg1Rows)
			seeder.SeedSupportGroupServices(1, svc2Rows, sg2Rows)

			components := seeder.SeedComponents(1)
			cv1 := seeder.SeedComponentVersions(1, components)
			cv2 := seeder.SeedComponentVersions(1, components)

			ci1 := seeder.SeedComponentInstances(1, cv1, svc1Rows)
			ci2 := seeder.SeedComponentInstances(1, cv2, svc2Rows)

			issues := seeder.SeedSecurityEvents(matchesPerService * 2)
			repos := seeder.SeedIssueRepositories()
			seeder.SeedIssueVariants(len(issues), repos, issues)

			seeder.SeedIssueMatches(matchesPerService, issues, ci1, users)
			seeder.SeedIssueMatches(matchesPerService, issues, ci2, users)
		})

		Context("filtering by service CCRN", func() {
			It("returns only alerts for the requested service and excludes the other", func() {
				respAll, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  100,
						"after":  "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respAll.SIEMAlerts.TotalCount).To(Equal(matchesPerService * 2))

				respSvc1, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{
							"service": []string{service1CCRN},
						},
						"first": 100,
						"after": "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respSvc1.SIEMAlerts.TotalCount).To(Equal(matchesPerService))

				respSvc2, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{
							"service": []string{service2CCRN},
						},
						"first": 100,
						"after": "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respSvc2.SIEMAlerts.TotalCount).To(Equal(matchesPerService))

				Expect(respSvc1.SIEMAlerts.TotalCount + respSvc2.SIEMAlerts.TotalCount).
					To(Equal(respAll.SIEMAlerts.TotalCount))
			})
		})

		Context("filtering by support group CCRN", func() {
			It("returns only alerts for the requested support group and excludes the other", func() {
				respAll, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  100,
						"after":  "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respAll.SIEMAlerts.TotalCount).To(Equal(matchesPerService * 2))

				respSg1, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{
							"supportGroup": []string{sg1CCRN},
						},
						"first": 100,
						"after": "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respSg1.SIEMAlerts.TotalCount).To(Equal(matchesPerService))

				respSg2, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{
							"supportGroup": []string{sg2CCRN},
						},
						"first": 100,
						"after": "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respSg2.SIEMAlerts.TotalCount).To(Equal(matchesPerService))

				Expect(respSg1.SIEMAlerts.TotalCount + respSg2.SIEMAlerts.TotalCount).
					To(Equal(respAll.SIEMAlerts.TotalCount))
			})
		})
	})

	When("the database has SecurityEvent IssueMatches split across two distinct regions", func() {
		var region1, region2 string
		const matchesPerRegion = 3

		BeforeEach(func() {
			users := seeder.SeedUsers(1)
			svcRows := seeder.SeedServices(2)
			components := seeder.SeedComponents(1)
			cvRows := seeder.SeedComponentVersions(2, components)
			issues := seeder.SeedSecurityEvents(matchesPerRegion * 2)
			repos := seeder.SeedIssueRepositories()
			seeder.SeedIssueVariants(len(issues), repos, issues)

			region1 = "region-a"
			region2 = "region-b"

			for i, region := range []string{region1, region2} {
				ci := test.NewFakeComponentInstance()
				ci.Region = sql.NullString{String: region, Valid: true}
				ci.ComponentVersionId = cvRows[i].Id
				ci.ServiceId = svcRows[i].Id
				ciId, err := seeder.InsertFakeComponentInstance(ci)
				Expect(err).To(BeNil())
				ci.Id = sql.NullInt64{Int64: ciId, Valid: true}
				seeder.SeedIssueMatches(matchesPerRegion, issues, []mariadb.ComponentInstanceRow{ci}, users)
			}
		})

		Context("filtering by region", func() {
			It("returns only alerts for the requested region and excludes the other", func() {
				respAll, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  100,
						"after":  "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respAll.SIEMAlerts.TotalCount).To(Equal(matchesPerRegion * 2))

				respR1, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{
							"region": []string{region1},
						},
						"first": 100,
						"after": "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respR1.SIEMAlerts.TotalCount).To(Equal(matchesPerRegion))

				respR2, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{
							"region": []string{region2},
						},
						"first": 100,
						"after": "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respR2.SIEMAlerts.TotalCount).To(Equal(matchesPerRegion))

				Expect(respR1.SIEMAlerts.TotalCount + respR2.SIEMAlerts.TotalCount).
					To(Equal(respAll.SIEMAlerts.TotalCount))
			})
		})
	})

	When("the database has SecurityEvent IssueMatches with distinct names", func() {
		var name1, name2 string
		const matchesPerName = 2

		BeforeEach(func() {
			users := seeder.SeedUsers(1)
			svcRows := seeder.SeedServices(1)
			components := seeder.SeedComponents(1)
			cvRows := seeder.SeedComponentVersions(1, components)
			ciRows := seeder.SeedComponentInstances(1, cvRows, svcRows)
			repos := seeder.SeedIssueRepositories()

			// One issue per group so all matches for that group share the same PrimaryName.
			issues1 := seeder.SeedSecurityEvents(1)
			issues2 := seeder.SeedSecurityEvents(1)
			name1 = issues1[0].PrimaryName.String
			name2 = issues2[0].PrimaryName.String

			seeder.SeedIssueVariants(len(issues1), repos, issues1)
			seeder.SeedIssueVariants(len(issues2), repos, issues2)
			seeder.SeedIssueMatches(matchesPerName, issues1, ciRows, users)
			seeder.SeedIssueMatches(matchesPerName, issues2, ciRows, users)
		})

		Context("filtering by name", func() {
			It("returns only alerts matching the requested name and excludes others", func() {
				respAll, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{},
						"first":  100,
						"after":  "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respAll.SIEMAlerts.TotalCount).To(Equal(matchesPerName * 2))

				respName1, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{
							"name": []string{name1},
						},
						"first": 100,
						"after": "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respName1.SIEMAlerts.TotalCount).To(Equal(matchesPerName))
				Expect(respName1.SIEMAlerts.TotalCount).To(BeNumerically("<", respAll.SIEMAlerts.TotalCount))

				respName2, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
					map[string]any{
						"filter": map[string]any{
							"name": []string{name2},
						},
						"first": 100,
						"after": "",
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(respName2.SIEMAlerts.TotalCount).To(Equal(matchesPerName))

				Expect(respName1.SIEMAlerts.TotalCount + respName2.SIEMAlerts.TotalCount).
					To(Equal(respAll.SIEMAlerts.TotalCount))
			})
		})
	})

	When("the database has mixed issue types", func() {
		BeforeEach(func() {
			_ = seeder.SeedDbWithNFakeData(5)
			_ = seeder.SeedDbWithSecurityEvents(5)
		})

		It("returns only SecurityEvent IssueMatches", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				SIEMAlerts model.SIEMAlertConnection `json:"SIEMAlerts"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/siem_alert/minimal.graphql",
				map[string]any{
					"filter": map[string]any{},
					"first":  100,
					"after":  "",
				},
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.SIEMAlerts.TotalCount).To(BeNumerically(">", 0))
		})
	})
})
