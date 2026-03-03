// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getting IssueCounts via API", Label("e2e", "IssueCounts"), func() {
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

	When("the database has entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			var err error
			seedCollection, err = seeder.SeedForIssueCounts()
			Expect(err).To(BeNil(), "Seeding should work")
			err = seeder.RefreshCountIssueRatings()
			Expect(err).To(BeNil(), "Refresh should work")
		})
		Context("and a filter is used", func() {
			It("correct filters by support group", func() {
				sg := seedCollection.SupportGroupRows[0]

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					IssueCounts model.SeverityCounts `json:"IssueCounts"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/issueCounts/query.graphql",
					map[string]any{
						"filter": map[string]string{
							"supportGroupCcrn": sg.CCRN.String,
						},
					},
					nil,
				)
				Expect(err).ToNot(HaveOccurred())

				severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../database/mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
				Expect(err).ToNot(HaveOccurred())

				strId := fmt.Sprintf("%d", sg.Id.Int64)

				Expect(int64(respData.IssueCounts.Critical)).To(Equal(severityCounts[strId].Critical))
				Expect(int64(respData.IssueCounts.High)).To(Equal(severityCounts[strId].High))
				Expect(int64(respData.IssueCounts.Medium)).To(Equal(severityCounts[strId].Medium))
				Expect(int64(respData.IssueCounts.Low)).To(Equal(severityCounts[strId].Low))
				Expect(int64(respData.IssueCounts.None)).To(Equal(severityCounts[strId].None))
				Expect(int64(respData.IssueCounts.Total)).To(Equal(severityCounts[strId].Total))
			})
			It("it can filter by service in services query", func() {
				severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../database/mariadb/testdata/issue_counts/issue_counts_per_service.json"))

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Services model.ServiceConnection `json:"Services"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/service/withIssueCounts.graphql",
					nil,
					nil,
				)
				Expect(err).ToNot(HaveOccurred())

				for _, sEdge := range respData.Services.Edges {
					sc := severityCounts[sEdge.Node.ID]
					Expect(int64(sEdge.Node.IssueCounts.Critical)).To(Equal(sc.Critical), "Critical count is correct")
					Expect(int64(sEdge.Node.IssueCounts.High)).To(Equal(sc.High), "High count is correct")
					Expect(int64(sEdge.Node.IssueCounts.Medium)).To(Equal(sc.Medium), "Medium count is correct")
					Expect(int64(sEdge.Node.IssueCounts.Low)).To(Equal(sc.Low), "Low count is correct")
					Expect(int64(sEdge.Node.IssueCounts.None)).To(Equal(sc.None), "None count is correct")
					Expect(int64(sEdge.Node.IssueCounts.Total)).To(Equal(sc.Total), "Total count is correct")
				}
			})
		})
		It("correct filters by component version id", func() {
			severityCounts, err := test.LoadComponentVersionIssueCounts(test.GetTestDataPath("../database/mariadb/testdata/issue_counts/issue_counts_per_component_version.json"))
			Expect(err).To(BeNil())

			cvId := fmt.Sprintf("%d", seedCollection.ComponentVersionIssueRows[0].ComponentVersionId.Int64)

			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				IssueCounts model.SeverityCounts `json:"IssueCounts"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/issueCounts/query.graphql",
				map[string]any{
					"filter": map[string]string{
						"componentVersionId": cvId,
					},
				},
				nil,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(int64(respData.IssueCounts.Critical)).To(Equal(severityCounts[cvId].Critical))
			Expect(int64(respData.IssueCounts.High)).To(Equal(severityCounts[cvId].High))
			Expect(int64(respData.IssueCounts.Medium)).To(Equal(severityCounts[cvId].Medium))
			Expect(int64(respData.IssueCounts.Low)).To(Equal(severityCounts[cvId].Low))
			Expect(int64(respData.IssueCounts.None)).To(Equal(severityCounts[cvId].None))
			Expect(int64(respData.IssueCounts.Total)).To(Equal(severityCounts[cvId].Total))
		})
	})
})
