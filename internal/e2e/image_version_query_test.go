// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"database/sql"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getting ImageVersions via API", Label("e2e", "ImageVersions"), func() {
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

	loadTestData := func() ([]mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, error) {
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

	When("the database has 10 entries", func() {
		BeforeEach(func() {
			seeder.SeedIssueRepositories()
			seeder.SeedVulnerabilities(10)
			services := seeder.SeedServices(1)
			components := seeder.SeedComponents(1)
			componentVersions := seeder.SeedComponentVersions(10, components)
			versionInstanceIds := make(map[int64]int64)
			issueVariantByIssueId := make(map[int64]mariadb.IssueVariantRow)
			issueVariants, componentVersionIssues, err := loadTestData()
			Expect(err).To(BeNil(), "Loading test data should work")
			// Important: the order need to be preserved
			for _, iv := range issueVariants {
				_, err := seeder.InsertFakeIssueVariant(iv)
				Expect(err).To(BeNil())
				issueVariantByIssueId[iv.IssueId.Int64] = iv
			}
			for _, cv := range componentVersions {
				ci := test.NewFakeComponentInstance()
				ci.ComponentVersionId = cv.Id
				ci.ServiceId = services[0].Id
				ciId, err := seeder.InsertFakeComponentInstance(ci)
				Expect(err).To(BeNil())
				versionInstanceIds[cv.Id.Int64] = ciId
			}
			for _, cvi := range componentVersionIssues {
				_, err := seeder.InsertFakeComponentVersionIssue(cvi)
				Expect(err).To(BeNil())
				im := test.NewFakeIssueMatch()
				im.IssueId = cvi.IssueId
				im.Status = sql.NullString{String: entity.IssueMatchStatusValuesNew.String(), Valid: true}
				im.UserId = sql.NullInt64{Int64: 1, Valid: true}
				// im.Rating = sql.NullString{String: entity.IssueMatchRatingValuesVulnerable.String(), Valid: true}
				issueVariant := issueVariantByIssueId[cvi.IssueId.Int64]
				im.Rating = sql.NullString{String: issueVariant.Rating.String, Valid: true}
				im.Vector = sql.NullString{String: "", Valid: true}
				ciId := versionInstanceIds[cvi.ComponentVersionId.Int64]
				im.ComponentInstanceId = sql.NullInt64{Int64: ciId, Valid: true}
				_, err = seeder.InsertFakeIssueMatch(im)
				Expect(err).To(BeNil())
			}
			seeder.RefreshCountIssueRatings()
		})

		It("can query image versions", func() {
			idsBySeverity := []string{"3", "8", "2", "7", "1", "6", "5", "4", "10", "9"}

			respData, err := e2e_common.ExecuteGqlQueryFromFile[struct {
				ImageVersions model.ImageVersionConnection `json:"ImageVersions"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/imageVersion/query.graphql",
				map[string]interface{}{"first": 10, "after": ""},
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(respData.ImageVersions.TotalCount).To(Equal(10))
			Expect(len(respData.ImageVersions.Edges)).To(Equal(10))

			totalVc := entity.IssueSeverityCounts{}

			for i, id := range idsBySeverity {
				iv := respData.ImageVersions.Edges[i].Node
				Expect(iv.ID).To(BeEquivalentTo(id), "ImageVersion has ID set")
				Expect(iv.Tag).ToNot(BeNil(), "ImageVersion has tag set")
				Expect(iv.Repository).ToNot(BeNil(), "ImageVersion has repository set")
				Expect(iv.Version).ToNot(BeNil(), "ImageVersion has version set")
				Expect(len(iv.Vulnerabilities.Edges)).To(BeNumerically(">", 0), "ImageVersion has vulnerabilities")
				Expect(len(iv.Occurences.Edges)).To(BeNumerically(">", 0), "ImageVersion has occurences")

				vc := entity.IssueSeverityCounts{}
				for _, vulnerability := range iv.Vulnerabilities.Edges {
					Expect(vulnerability.Node.ID).ToNot(BeNil(), "vulnerability has a ID set")
					Expect(vulnerability.Node.Name).ToNot(BeNil(), "vulnerability has name set")
					switch vulnerability.Node.Severity.String() {
					case entity.SeverityValuesCritical.String():
						vc.Critical++
						totalVc.Critical++
					case entity.SeverityValuesHigh.String():
						vc.High++
						totalVc.High++
					case entity.SeverityValuesMedium.String():
						vc.Medium++
						totalVc.Medium++
					case entity.SeverityValuesLow.String():
						vc.Low++
						totalVc.Low++
					case entity.SeverityValuesNone.String():
						vc.None++
						totalVc.None++
					}
					vc.Total++
					totalVc.Total++
				}
				Expect(iv.VulnerabilityCounts.Critical).To(BeEquivalentTo(vc.Critical), "Critical count matches")
				Expect(iv.VulnerabilityCounts.High).To(BeEquivalentTo(vc.High), "High count matches")
				Expect(iv.VulnerabilityCounts.Medium).To(BeEquivalentTo(vc.Medium), "Medium count matches")
				Expect(iv.VulnerabilityCounts.Low).To(BeEquivalentTo(vc.Low), "Low count matches")
				Expect(iv.VulnerabilityCounts.None).To(BeEquivalentTo(vc.None), "None count matches")
				Expect(iv.VulnerabilityCounts.Total).To(BeEquivalentTo(vc.Total), "Total count matches")

				for _, o := range iv.Occurences.Edges {
					Expect(o.Node.ID).ToNot(BeNil(), "occurence has a ID set")
					Expect(o.Node.Ccrn).ToNot(BeNil(), "occurence has ccrn set")

					Expect(*o.Node.ComponentVersionID).To(BeEquivalentTo(iv.ID))
				}
			}
			Expect(respData.ImageVersions.Counts.Critical).To(BeEquivalentTo(totalVc.Critical), "Total Critical count matches")
			Expect(respData.ImageVersions.Counts.High).To(BeEquivalentTo(totalVc.High), "Total High count matches")
			Expect(respData.ImageVersions.Counts.Medium).To(BeEquivalentTo(totalVc.Medium), "Total Medium count matches")
			Expect(respData.ImageVersions.Counts.Low).To(BeEquivalentTo(totalVc.Low), "Total Low count matches")
			Expect(respData.ImageVersions.Counts.None).To(BeEquivalentTo(totalVc.None), "Total None count matches")
			Expect(respData.ImageVersions.Counts.Total).To(BeEquivalentTo(totalVc.Total), "Total count matches")
		})
		Context("and end of life filter presents as true", func() {
			It("returns correct result", func() {
				resp, err := e2e_common.ExecuteGqlQueryFromFile[struct {
					ImageVersions model.ImageVersionConnection `json:"ImageVersions"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/imageVersion/query.graphql",
					map[string]any{
						"filter": map[string]bool{
							"endOfLife": true,
						},
					},
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(resp.ImageVersions.Edges)).To(Equal(5))

				for _, edge := range resp.ImageVersions.Edges {
					Expect(*edge.Node.EndOfLife).To(BeTrue())
				}
			})
		})
		Context("and end of life filter presents as false", func() {
			It("returns correct result", func() {
				resp, err := e2e_common.ExecuteGqlQueryFromFile[struct {
					ImageVersions model.ImageVersionConnection `json:"ImageVersions"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/imageVersion/query.graphql",
					map[string]any{
						"filter": map[string]bool{
							"endOfLife": false,
						},
					},
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(resp.ImageVersions.Edges)).To(Equal(5))

				for _, edge := range resp.ImageVersions.Edges {
					Expect(*edge.Node.EndOfLife).To(BeFalse())
				}
			})
		})
	})
})
