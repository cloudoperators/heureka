// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"

	"os"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getting Images via API", Label("e2e", "Images"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)
		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
		dbm.TestTearDown(db)
	})

	var loadTestData = func() ([]mariadb.ComponentVersionRow, []mariadb.ComponentInstanceRow, []mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, []mariadb.IssueMatchRow, error) {
		issueVariants, err := test.LoadIssueVariants(test.GetTestDataPath("../database/mariadb/testdata/component_version_order/issue_variant.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		cvIssues, err := test.LoadComponentVersionIssues(test.GetTestDataPath("../database/mariadb/testdata/component_order/component_version_issue.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		componentInstances, err := test.LoadComponentInstances(test.GetTestDataPath("../database/mariadb/testdata/service_order/component_instance.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		issueMatches, err := test.LoadIssueMatches(test.GetTestDataPath("../database/mariadb/testdata/component_order/issue_match.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		componentVersions, err := test.LoadComponentVersions(test.GetTestDataPath("../database/mariadb/testdata/component_order/component_version.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		return componentVersions, componentInstances, issueVariants, cvIssues, issueMatches, nil
	}
	When("the database has 10 issues", func() {
		var respData struct {
			Images model.ImageConnection `json:"Images"`
		}
		var counts model.SeverityCounts
		var services []mariadb.BaseServiceRow
		BeforeEach(func() {
			seeder.SeedIssueRepositories()
			for i := 0; i < 10; i++ {
				issue := test.NewFakeIssue()
				issue.Type.String = entity.IssueTypeVulnerability.String()
				seeder.InsertFakeIssue(issue)
			}
			seeder.SeedComponents(5)
			services = seeder.SeedServices(5)
			componentVersions, componentInstances, issueVariants, componentVersionIssues, issueMatches, err := loadTestData()
			Expect(err).To(BeNil())
			// Important: the order need to be preserved
			for _, iv := range issueVariants {
				_, err := seeder.InsertFakeIssueVariant(iv)
				Expect(err).To(BeNil())
				switch iv.Rating.String {
				case entity.SeverityValuesCritical.String():
					counts.Critical++
				case entity.SeverityValuesHigh.String():
					counts.High++
				case entity.SeverityValuesNone.String():
					counts.Medium++
				case entity.SeverityValuesLow.String():
					counts.Low++
				case entity.SeverityValuesMedium.String():
					counts.None++
				}
				counts.Total++
			}
			for _, cv := range componentVersions {
				_, err := seeder.InsertFakeComponentVersion(cv)
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
			err = seeder.RefreshComponentVulnerabilityCounts()
			Expect(err).To(BeNil())
		})

		It("returns the expected content and the expected PageInfo", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/image/query.graphql")
			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)
			req.Var("filter", map[string]any{
				"service": lo.Map(services, func(item mariadb.BaseServiceRow, index int) string { return item.CCRN.String }),
			})
			req.Var("first", 3)
			req.Var("after", "")
			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()
			err = client.Run(ctx, req, &respData)
			Expect(err).To(BeNil(), "Error while unmarshaling")
			Expect(respData.Images.TotalCount).To(Equal(5))
			Expect(len(respData.Images.Edges)).To(Equal(3))
			prevSeverity := model.SeverityCounts{Critical: 9999, High: 9999, Medium: 9999, Low: 9999, None: 9999, Total: 9999}
			for _, image := range respData.Images.Edges {
				Expect(image.Node.Repository).ToNot(BeNil(), "image has a repository set")
				Expect(image.Node.ImageRegistryURL).ToNot(BeNil(), "image has a registry url set")

				Expect(e2e_common.CompareSeverityCounts(*image.Node.VulnerabilityCounts, prevSeverity)).To(BeNumerically("<=", 0), "severity is in descending order")
				prevSeverity = *image.Node.VulnerabilityCounts

				for _, version := range image.Node.Versions.Edges {
					Expect(version.Node.Version).ToNot(BeNil(), "version has a version set")
				}

				for _, vulnerability := range image.Node.Vulnerabilities.Edges {
					Expect(vulnerability.Node.Severity).ToNot(BeNil(), "vulnerability has a severity set")
					Expect(vulnerability.Node.Name).ToNot(BeNil(), "vulnerability has a name set")
					Expect(vulnerability.Node.SourceURL).ToNot(BeNil(), "vulnerability has a source url set")
					Expect(vulnerability.Node.EarliestTargetRemediationDate).ToNot(BeNil(), "vulnerability has a earliest target remediation date set")
					Expect(vulnerability.Node.Description).ToNot(BeNil(), "vulnerability has a description set")
				}
			}
			Expect(*respData.Images.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
			Expect(*respData.Images.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
			Expect(respData.Images.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
			Expect(len(respData.Images.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
			Expect(*respData.Images.PageInfo.PageNumber).To(Equal(1), "Correct page number")

			Expect(respData.Images.Counts.Critical).To(Equal(counts.Critical), "Critical count is correct")
			Expect(respData.Images.Counts.High).To(Equal(counts.High), "High count is correct")
			Expect(respData.Images.Counts.Medium).To(Equal(counts.Medium), "Medium count is correct")
			Expect(respData.Images.Counts.Low).To(Equal(counts.Low), "Low count is correct")
			Expect(respData.Images.Counts.None).To(Equal(counts.None), "None count is correct")
			Expect(respData.Images.Counts.Total).To(Equal(counts.Total), "Total count is correct")
		})
	})
})
