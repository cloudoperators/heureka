// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"database/sql"
	"fmt"
	"reflect"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
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
		It(
			"returns the expected content and the expected PageInfo when filtered using service",
			func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFile[struct {
					Images model.ImageConnection `json:"Images"`
				}](
					imgTest.port,
					"../api/graphql/graph/queryCollection/image/query.graphql",
					map[string]any{
						"filter": map[string]any{
							"service": lo.Map(
								imgTest.services,
								func(item mariadb.BaseServiceRow, index int) string { return item.CCRN.String },
							),
						},
						"first": 3,
						"after": "",
					})

				Expect(err).ToNot(HaveOccurred())
				expectRespDataCounts(respData.Images, 5, 3)
				expectRespImagesFilledAndInOrder(&respData.Images)
				expectPageInfoTwoPagesBeingOnFirst(respData.Images.PageInfo)
				Expect(*respData.Images.Counts).To(Equal(imgTest.counts))
			},
		)
		It("returns images sorted by vulnerability severity counts then by repository name", func() {
			imgTest.testImageSortingWithTieBreaker()
		})
		It(
			"returns the expected content and the expected PageInfo when filtered using repository",
			func() {
				service := imgTest.services[0]
				componentInstances := lo.Filter(
					imgTest.componentInstances,
					func(ci mariadb.ComponentInstanceRow, _ int) bool {
						return ci.ServiceId.Int64 == service.Id.Int64
					},
				)
				componentVersion, found := lo.Find(
					imgTest.componentVersions,
					func(cv mariadb.ComponentVersionRow) bool {
						return cv.Id.Int64 == componentInstances[0].ComponentVersionId.Int64
					},
				)
				Expect(found).To(BeTrue(), "ComponentVersion for ComponentInstance should be found")
				component, foundComp := lo.Find(
					imgTest.components,
					func(c mariadb.ComponentRow) bool {
						return c.Id.Int64 == componentVersion.ComponentId.Int64
					},
				)
				Expect(foundComp).To(BeTrue(), "Component for ComponentVersion should be found")
				// test data is setup so that first two component versions (having each one critical
				// vulnerability)
				// belong to first service and first component
				counts := model.SeverityCounts{Critical: 2, Total: 2}

				respData, err := e2e_common.ExecuteGqlQueryFromFile[struct {
					Images model.ImageConnection `json:"Images"`
				}](
					imgTest.port,
					"../api/graphql/graph/queryCollection/image/query.graphql",
					map[string]any{
						"filter": map[string]any{
							"repository": []string{component.Repository.String},
							"service":    []string{service.CCRN.String},
						},
						"first": 3,
						"after": "",
					})

				Expect(err).ToNot(HaveOccurred())
				expectRespDataCounts(respData.Images, 1, 1)
				expectRespImagesFilledAndInOrder(&respData.Images)
				Expect(*respData.Images.Counts).To(Equal(counts))
			},
		)
	})
})

type imageTest struct {
	port   string
	counts model.SeverityCounts
	db     *mariadb.SqlDatabase
	seeder *test.DatabaseSeeder
	server *server.Server

	componentVersions      []mariadb.ComponentVersionRow
	services               []mariadb.BaseServiceRow
	componentInstances     []mariadb.ComponentInstanceRow
	componentVersionIssues []mariadb.ComponentVersionIssueRow
	components             []mariadb.ComponentRow
}

func newImageTest() *imageTest {
	db := dbm.NewTestSchemaWithoutMigration()

	cfg := dbm.DbConfig()
	cfg.Port = e2e_common.GetRandomFreePort()

	seeder, err := test.NewDatabaseSeeder(cfg)
	Expect(err).To(BeNil(), "Database Seeder Setup should work")

	cfg.AuthzOpenFgaApiUrl = ""
	server := e2e_common.NewRunningServer(cfg)

	return &imageTest{
		port:   cfg.Port,
		db:     db,
		seeder: seeder,
		server: server,
	}
}

func (it *imageTest) teardown() {
	e2e_common.ServerTeardown(it.server)
	dbm.TestTearDown(it.db)
}

func (it *imageTest) seed10Entries() {
	it.seeder.SeedIssueRepositories()

	for range 10 {
		issue := test.NewFakeIssue()
		issue.Type.String = entity.IssueTypeVulnerability.String()
		it.seeder.InsertFakeIssue(issue)
	}

	it.components = it.seeder.SeedComponents(5)
	it.services = it.seeder.SeedServices(5)

	componentVersions, componentInstances, issueVariants, componentVersionIssues, issueMatches, err := loadTestData()
	it.componentVersions = componentVersions
	it.componentInstances = componentInstances
	it.componentVersionIssues = componentVersionIssues

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

	for i, cv := range it.componentVersions {
		id, err := it.seeder.InsertFakeComponentVersion(cv)
		it.componentVersions[i].Id.Int64 = id

		Expect(err).To(BeNil())
	}

	for _, cvi := range componentVersionIssues {
		_, err := it.seeder.InsertFakeComponentVersionIssue(cvi)
		Expect(err).To(BeNil())
	}

	for i, ci := range componentInstances {
		id, err := it.seeder.InsertFakeComponentInstance(ci)
		it.componentInstances[i].Id.Int64 = id

		Expect(err).To(BeNil())
	}

	for _, im := range issueMatches {
		_, err := it.seeder.InsertFakeIssueMatch(im)
		Expect(err).To(BeNil())
	}

	err = it.seeder.RefreshComponentVulnerabilityCounts()
	Expect(err).To(BeNil())
}

func (it *imageTest) seedTieBreakerData() {
	service := test.NewFakeBaseService()
	serviceId, err := it.seeder.InsertFakeBaseService(service)
	Expect(err).To(BeNil())

	service.Id = sql.NullInt64{Int64: serviceId, Valid: true}
	it.services = append(it.services, service)

	issue := test.NewFakeIssue()
	issue.Type = sql.NullString{String: entity.IssueTypeVulnerability.String(), Valid: true}
	issueId, err := it.seeder.InsertFakeIssue(issue)
	Expect(err).To(BeNil())

	iv := mariadb.IssueVariantRow{
		SecondaryName:     sql.NullString{String: "CVE-2025-TIE-BREAKER", Valid: true},
		Description:       sql.NullString{String: "A tie-breaker vulnerability", Valid: true},
		IssueId:           sql.NullInt64{Int64: issueId, Valid: true},
		IssueRepositoryId: sql.NullInt64{Int64: 1, Valid: true}, // From SeedIssueRepositories
		Rating:            sql.NullString{String: entity.SeverityValuesCritical.String(), Valid: true},
	}
	_, err = it.seeder.InsertFakeIssueVariant(iv)
	Expect(err).To(BeNil())

	for _, repoName := range []string{"B_tie_repo", "A_tie_repo"} {
		component := test.NewFakeComponent()
		component.Repository = sql.NullString{String: repoName, Valid: true}
		compId, err := it.seeder.InsertFakeComponent(component)
		Expect(err).To(BeNil())

		cv := test.NewFakeComponentVersion()
		cv.ComponentId = sql.NullInt64{Int64: compId, Valid: true}
		cvId, err := it.seeder.InsertFakeComponentVersion(cv)
		Expect(err).To(BeNil())

		cvi := test.NewFakeComponentVersionIssue()
		cvi.ComponentVersionId = sql.NullInt64{Int64: cvId, Valid: true}
		cvi.IssueId = sql.NullInt64{Int64: issueId, Valid: true}
		_, err = it.seeder.InsertFakeComponentVersionIssue(cvi)
		Expect(err).To(BeNil())

		ci := test.NewFakeComponentInstance()
		ci.ServiceId = sql.NullInt64{Int64: serviceId, Valid: true}
		ci.ComponentVersionId = sql.NullInt64{Int64: cvId, Valid: true}
		_, err = it.seeder.InsertFakeComponentInstance(ci)
		Expect(err).To(BeNil())
	}

	err = it.seeder.RefreshComponentVulnerabilityCounts()
	Expect(err).To(BeNil())
}

func (it *imageTest) testImageSortingWithTieBreaker() {
	it.seedTieBreakerData()

	respData, err := e2e_common.ExecuteGqlQueryFromFile[struct {
		Images model.ImageConnection `json:"Images"`
	}](
		it.port,
		"../api/graphql/graph/queryCollection/image/query.graphql",
		map[string]interface{}{
			"filter": map[string]any{
				"service": lo.Map(
					it.services,
					func(item mariadb.BaseServiceRow, index int) string { return item.CCRN.String },
				),
			},
			"first": 20,
			"after": "",
		})

	Expect(err).ToNot(HaveOccurred())
	Expect(respData.Images.Edges).To(HaveLen(7), "Should return all 7 images")

	var previousCounts model.SeverityCounts

	var previousRepository string

	hadEqualCountsPair := false

	for i, edge := range respData.Images.Edges {
		Expect(edge.Node.Repository).ToNot(BeNil(), "Image should have repository")
		Expect(edge.Node.VulnerabilityCounts).ToNot(BeNil(), "Image should have vulnerability counts")

		counts := *edge.Node.VulnerabilityCounts
		repository := *edge.Node.Repository

		if i > 0 {
			comparison := e2e_common.CompareSeverityCounts(counts, previousCounts)
			// Verify Primary Ordering: Severity Counts (Descending)
			Expect(comparison).To(BeNumerically("<=", 0),
				fmt.Sprintf("Image %d (%s) should have equal or lower severity than image %d (%s). Counts: %v vs %v",
					i, repository, i-1, previousRepository, counts, previousCounts))

			// Verify Secondary Ordering: Repository Name (Ascending)
			if comparison == 0 {
				hadEqualCountsPair = true

				Expect(previousRepository <= repository).To(BeTrue(),
					fmt.Sprintf("Image %d (%s) should sort after image %d (%s) when severity counts are equal. Counts: %v",
						i, repository, i-1, previousRepository, counts))
			}
		}

		previousCounts = counts
		previousRepository = repository
	}

	Expect(hadEqualCountsPair).To(BeTrue(),
		"Test setup should include at least one pair of images with equal vulnerability counts to verify secondary ordering")
}

func loadTestData() ([]mariadb.ComponentVersionRow, []mariadb.ComponentInstanceRow, []mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, []mariadb.IssueMatchRow, error) {
	issueVariants, err := test.LoadIssueVariants(
		test.GetTestDataPath(
			"../database/mariadb/testdata/component_version_order/issue_variant.json",
		),
	)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	cvIssues, err := test.LoadComponentVersionIssues(
		test.GetTestDataPath(
			"../database/mariadb/testdata/component_order/component_version_issue.json",
		),
	)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	componentInstances, err := test.LoadComponentInstances(
		test.GetTestDataPath("../database/mariadb/testdata/service_order/component_instance.json"),
	)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	issueMatches, err := test.LoadIssueMatches(
		test.GetTestDataPath("../database/mariadb/testdata/component_order/issue_match.json"),
	)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	componentVersions, err := test.LoadComponentVersions(
		test.GetTestDataPath("../database/mariadb/testdata/component_order/component_version.json"),
	)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return componentVersions, componentInstances, issueVariants, cvIssues, issueMatches, nil
}

func EqualCounts(expected model.SeverityCounts) types.GomegaMatcher {
	return &SeverityCountsMatcher{
		expected: expected,
	}
}

type SeverityCountsMatcher struct {
	expected model.SeverityCounts
}

func (m *SeverityCountsMatcher) Match(actual any) (bool, error) {
	a, ok := actual.(model.SeverityCounts)
	if !ok {
		return false, fmt.Errorf(
			"EqualCounts matcher expects a model.SeverityCounts, got %T",
			actual,
		)
	}

	return a == m.expected, nil
}

func (m *SeverityCountsMatcher) FailureMessage(actual any) string {
	a := actual.(model.SeverityCounts)

	return fmt.Sprintf(
		"Counts do not match:\n%s",
		format.Message(a, "to equal", m.expected),
	)
}

func (m *SeverityCountsMatcher) NegatedFailureMessage(actual any) string {
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
	prevSeverity := model.SeverityCounts{
		Critical: 9999,
		High:     9999,
		Medium:   9999,
		Low:      9999,
		None:     9999,
		Total:    9999,
	}

	for _, image := range images.Edges {
		Expect(image.Node.Repository).ToNot(BeNil(), "image has a repository set")
		Expect(image.Node.ImageRegistryURL).ToNot(BeNil(), "image has a registry url set")

		Expect(
			e2e_common.CompareSeverityCounts(*image.Node.VulnerabilityCounts, prevSeverity),
		).To(BeNumerically("<=", 0), "severity is in descending order")
		prevSeverity = *image.Node.VulnerabilityCounts

		for _, version := range image.Node.Versions.Edges {
			Expect(version.Node.Version).ToNot(BeNil(), "version has a version set")
		}

		for _, vulnerability := range image.Node.Vulnerabilities.Edges {
			Expect(vulnerability.Node.Severity).ToNot(BeNil(), "vulnerability has a severity set")
			Expect(vulnerability.Node.Name).ToNot(BeNil(), "vulnerability has a name set")
			Expect(
				vulnerability.Node.SourceURL,
			).ToNot(BeNil(), "vulnerability has a source url set")
			Expect(
				vulnerability.Node.EarliestTargetRemediationDate,
			).ToNot(BeNil(), "vulnerability has a earliest target remediation date set")
			Expect(
				vulnerability.Node.Description,
			).ToNot(BeNil(), "vulnerability has a description set")
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
