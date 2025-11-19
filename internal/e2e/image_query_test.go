// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"reflect"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Getting Images via API", Label("e2e", "Images"), func() {
	var imgTest *imageTest

	BeforeEach(func() {
		imgTest = newImageTest()
	})

	AfterEach(func() {
		imgTest.teardown()
	})

	When("the database has 10 issues", func() {
		BeforeEach(func() {
			imgTest.seed10Entries()
		})
		It("returns the expected content and the expected PageInfo when filtered using service", func() {
			respData := e2e_common.ExecuteGqlQueryFromFile[struct {
				Images model.ImageConnection `json:"Images"`
			}](
				imgTest.getQueryUrl(),
				"../api/graphql/graph/queryCollection/image/query.graphql",
				map[string]interface{}{
					"filter": map[string]any{
						"service": lo.Map(imgTest.services, func(item mariadb.BaseServiceRow, index int) string { return item.CCRN.String }),
					},
					"first": 3,
					"after": "",
				},
				map[string]string{"Cache-Control": "no-cache"})

			expectRespDataCounts(respData.Images, 5, 3)
			expectRespImagesFilledAndInOrder(&respData.Images)
			expectPageInfoTwoPagesBeingOnFirst(respData.Images.PageInfo)
			Expect(*respData.Images.Counts).To(Equal(imgTest.counts))
		})
		It("returns the expected content and the expected PageInfo when filtered using repository", func() {
			respData := e2e_common.ExecuteGqlQueryFromFile[struct {
				Images model.ImageConnection `json:"Images"`
			}](
				imgTest.getQueryUrl(),
				"../api/graphql/graph/queryCollection/image/query.graphql",
				map[string]interface{}{
					"filter": map[string]any{
						"repository": lo.Map(imgTest.componentVersions, func(item mariadb.ComponentVersionRow, index int) string { return item.Repository.String }),
					},
					"first": 3,
					"after": "",
				},
				map[string]string{"Cache-Control": "no-cache"})

			expectRespDataCounts(respData.Images, 5, 3)
			expectRespImagesFilledAndInOrder(&respData.Images)
			expectPageInfoTwoPagesBeingOnFirst(respData.Images.PageInfo)
			Expect(*respData.Images.Counts).To(Equal(imgTest.counts))
		})
	})

})

type imageTest struct {
	cfg    util.Config
	counts model.SeverityCounts
	db     *mariadb.SqlDatabase
	seeder *test.DatabaseSeeder
	server *server.Server

	componentVersions []mariadb.ComponentVersionRow
	services          []mariadb.BaseServiceRow
}

func newImageTest() *imageTest {
	db := dbm.NewTestSchema()

	cfg := dbm.DbConfig()
	cfg.Port = util2.GetRandomFreePort()

	seeder, err := test.NewDatabaseSeeder(cfg)
	Expect(err).To(BeNil(), "Database Seeder Setup should work")

	server := server.NewServer(cfg)
	server.NonBlockingStart()

	return &imageTest{
		cfg:    cfg,
		db:     db,
		seeder: seeder,
		server: server,
	}
}

func (it *imageTest) teardown() {
	it.server.BlockingStop()
	dbm.TestTearDown(it.db)
}

func (it *imageTest) seed10Entries() {
	it.seeder.SeedIssueRepositories()
	for i := 0; i < 10; i++ {
		issue := test.NewFakeIssue()
		issue.Type.String = entity.IssueTypeVulnerability.String()
		it.seeder.InsertFakeIssue(issue)
	}
	it.seeder.SeedComponents(5)
	it.services = it.seeder.SeedServices(5)

	componentVersions, componentInstances, issueVariants, componentVersionIssues, issueMatches, err := loadTestData()
	it.componentVersions = componentVersions
	Expect(err).To(BeNil())
	// Important: the order need to be preserved
	for _, iv := range issueVariants {
		_, err := it.seeder.InsertFakeIssueVariant(iv)
		Expect(err).To(BeNil())
		switch iv.Rating.String {
		case entity.SeverityValuesCritical.String():
			it.counts.Critical++
		case entity.SeverityValuesHigh.String():
			it.counts.High++
		case entity.SeverityValuesNone.String():
			it.counts.Medium++
		case entity.SeverityValuesLow.String():
			it.counts.Low++
		case entity.SeverityValuesMedium.String():
			it.counts.None++
		}
		it.counts.Total++
	}
	for _, cv := range it.componentVersions {
		_, err := it.seeder.InsertFakeComponentVersion(cv)
		Expect(err).To(BeNil())
	}
	for _, cvi := range componentVersionIssues {
		_, err := it.seeder.InsertFakeComponentVersionIssue(cvi)
		Expect(err).To(BeNil())
	}
	for _, ci := range componentInstances {
		_, err := it.seeder.InsertFakeComponentInstance(ci)
		Expect(err).To(BeNil())
	}
	for _, im := range issueMatches {
		_, err := it.seeder.InsertFakeIssueMatch(im)
		Expect(err).To(BeNil())
	}
	err = it.seeder.RefreshComponentVulnerabilityCounts()
	Expect(err).To(BeNil())
}

func loadTestData() ([]mariadb.ComponentVersionRow, []mariadb.ComponentInstanceRow, []mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, []mariadb.IssueMatchRow, error) {
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

func (it *imageTest) getQueryUrl() string {
	return fmt.Sprintf("http://localhost:%s/query", it.cfg.Port)
}

func EqualCounts(expected model.SeverityCounts) types.GomegaMatcher {
	return &SeverityCountsMatcher{
		expected: expected,
	}
}

type SeverityCountsMatcher struct {
	expected model.SeverityCounts
}

func (m *SeverityCountsMatcher) Match(actual interface{}) (bool, error) {
	a, ok := actual.(model.SeverityCounts)
	if !ok {
		return false, fmt.Errorf("EqualCounts matcher expects a model.SeverityCounts, got %T", actual)
	}

	return a == m.expected, nil
}

func (m *SeverityCountsMatcher) FailureMessage(actual interface{}) string {
	a := actual.(model.SeverityCounts)

	return fmt.Sprintf(
		"Counts do not match:\n%s",
		format.Message(a, "to equal", m.expected),
	)
}

func (m *SeverityCountsMatcher) NegatedFailureMessage(actual interface{}) string {
	a := actual.(model.SeverityCounts)

	return fmt.Sprintf(
		"Counts unexpectedly match:\n%s",
		format.Message(a, "not to equal", m.expected),
	)
}

func expectPageInfoTwoPagesBeingOnFirst(pi *model.PageInfo) {
	Expect(*pi.HasNextPage).To(BeTrue(), "hasNextPage is set")
	Expect(*pi.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
	Expect(*pi.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
	Expect(len(pi.Pages)).To(Equal(2), "Correct amount of pages")
	Expect(*pi.PageNumber).To(Equal(1), "Correct page number")
}

func expectRespImagesFilledAndInOrder(images *model.ImageConnection) {
	prevSeverity := model.SeverityCounts{Critical: 9999, High: 9999, Medium: 9999, Low: 9999, None: 9999, Total: 9999}
	for _, image := range images.Edges {
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
}

func expectRespDataCounts[T any](data T, total int, edges int) {
	v := reflect.ValueOf(data)

	totalCount := v.FieldByName("TotalCount")
	Expect(totalCount.IsValid()).To(BeTrue(), "struct must have TotalCount field")
	Expect(int(totalCount.Int())).To(Equal(total))

	edgesField := v.FieldByName("Edges")
	Expect(edgesField.IsValid()).To(BeTrue(), "struct must have Edges field")
	Expect(edgesField.Len()).To(Equal(edges))
}
