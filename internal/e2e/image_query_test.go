// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/util"
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
					},
				)

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
		It("returns images even when no service filter is provided", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFile[struct {
				Images model.ImageConnection `json:"Images"`
			}](
				imgTest.port,
				"../api/graphql/graph/queryCollection/image/query.graphql",
				map[string]any{
					"filter": map[string]any{},
					"first":  3,
					"after":  "",
				},
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.Images.TotalCount).To(BeNumerically(">=", 0))
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
					},
				)

				Expect(err).ToNot(HaveOccurred())
				expectRespDataCounts(respData.Images, 1, 1)
				expectRespImagesFilledAndInOrder(&respData.Images)
				Expect(*respData.Images.Counts).To(Equal(counts))
			},
		)
		It(
			"returns vulnerabilities for an image when using service filter and vulnerability severity filter",
			func() {
				// This test exercises the VulnerabilityBaseResolver fallthrough path
				// (not the batch preload) by passing a vulnerability filter, which sets
				// hasUserFilter=true in the image resolver and bypasses pre-loaded data.
				// Regression test: service filter must not break the CVI→CV join chain
				// needed for ComponentId-scoped vulnerability queries.
				service := imgTest.services[0]

				query := `query ($imgFilter: ImageFilter, $vulFilter: VulnerabilityFilter, $first: Int) {
					Images(first: $first, filter: $imgFilter) {
						edges {
							node {
								id
								vulnerabilities(filter: $vulFilter) {
									totalCount
									edges {
										node {
											id
											severity
											name
										}
									}
								}
							}
						}
					}
				}`

				respData, err := e2e_common.ExecuteGqlQuery[struct {
					Images model.ImageConnection `json:"Images"`
				}](
					imgTest.port,
					query,
					map[string]any{
						"imgFilter": map[string]any{
							"service": []string{service.CCRN.String},
						},
						"vulFilter": map[string]any{
							"severity": []string{"Critical"},
						},
						"first": 5,
					},
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(respData.Images.Edges)).To(BeNumerically(">", 0), "should return at least one image")

				// At least one image should have vulnerabilities with Critical severity
				hasVulns := false

				for _, edge := range respData.Images.Edges {
					if edge.Node.Vulnerabilities != nil && len(edge.Node.Vulnerabilities.Edges) > 0 {
						hasVulns = true

						for _, vuln := range edge.Node.Vulnerabilities.Edges {
							Expect(vuln.Node.Severity).ToNot(BeNil())
							Expect(vuln.Node.Severity.String()).To(Equal("Critical"))
						}
					}
				}

				Expect(hasVulns).To(BeTrue(), "at least one image should have Critical vulnerabilities")
			},
		)

		It(
			"vulnerability counts match detail view totals when filtered by service",
			func() {
				// Regression test: ensures that vulnerabilityCounts (from the MV) match
				// the actual vulnerabilities returned by the detail view (fallthrough path)
				// when both are scoped to the same service. Before the fix, clearing the
				// service filter in VulnerabilityBaseResolver caused vulnerabilities from
				// other services to leak into the detail view.
				svcACCRN := imgTest.seedServiceScopedVulnData()

				type vulnCountNode struct {
					ID                  string                         `json:"id"`
					Repository          *string                        `json:"repository"`
					VulnerabilityCounts *model.SeverityCounts          `json:"vulnerabilityCounts"`
					AllVulns            *model.VulnerabilityConnection `json:"allVulns"`
					CriticalVulns       *model.VulnerabilityConnection `json:"criticalVulns"`
					HighVulns           *model.VulnerabilityConnection `json:"highVulns"`
				}

				type vulnCountEdge struct {
					Node *vulnCountNode `json:"node"`
				}

				type vulnCountConnection struct {
					Edges []*vulnCountEdge `json:"edges"`
				}

				query := `query ($imgFilter: ImageFilter, $first: Int) {
					Images(first: $first, filter: $imgFilter) {
						edges {
							node {
								id
								repository
								vulnerabilityCounts {
									critical
									high
									medium
									low
									none
								}
								allVulns: vulnerabilities {
									totalCount
								}
								criticalVulns: vulnerabilities(filter: { severity: [Critical] }) {
									totalCount
								}
								highVulns: vulnerabilities(filter: { severity: [High] }) {
									totalCount
								}
							}
						}
					}
				}`

				respData, err := e2e_common.ExecuteGqlQuery[struct {
					Images vulnCountConnection `json:"Images"`
				}](
					imgTest.port,
					query,
					map[string]any{
						"imgFilter": map[string]any{
							"service": []string{svcACCRN},
						},
						"first": 20,
					},
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Images.Edges).ToNot(BeEmpty())

				for _, edge := range respData.Images.Edges {
					node := edge.Node
					Expect(node.VulnerabilityCounts).ToNot(BeNil(),
						"image %s must have vulnerability counts", node.ID)
					Expect(node.CriticalVulns).ToNot(BeNil())
					Expect(node.HighVulns).ToNot(BeNil())
					Expect(node.AllVulns).ToNot(BeNil())

					// Core regression assertion: badge count == filtered detail count
					Expect(node.VulnerabilityCounts.Critical).To(
						Equal(node.CriticalVulns.TotalCount),
						"image %s: vulnerabilityCounts.critical (%d) must equal vulnerabilities(severity:Critical).totalCount (%d)",
						node.ID, node.VulnerabilityCounts.Critical, node.CriticalVulns.TotalCount,
					)

					Expect(node.VulnerabilityCounts.High).To(
						Equal(node.HighVulns.TotalCount),
						"image %s: vulnerabilityCounts.high (%d) must equal vulnerabilities(severity:High).totalCount (%d)",
						node.ID, node.VulnerabilityCounts.High, node.HighVulns.TotalCount,
					)
				}

				// Sanity: find the seeded image (1 Critical, 1 High for service A)
				foundCompX := false

				for _, edge := range respData.Images.Edges {
					node := edge.Node
					if node.VulnerabilityCounts.Critical == 1 && node.VulnerabilityCounts.High == 1 {
						foundCompX = true
						// Verify detail counts match
						Expect(node.CriticalVulns.TotalCount).To(Equal(1))
						Expect(node.HighVulns.TotalCount).To(Equal(1))
					}
				}

				Expect(foundCompX).To(BeTrue(),
					"expected to find the seeded image with exactly 1 Critical and 1 High vulnerability "+
						"scoped to service A; if counts differ, service B's IssueMatch is leaking through")
			},
		)

		It(
			"soft-deleted IssueMatches and rating mismatches do not cause count discrepancies",
			func() {
				// Regression test: ensures that:
				// 1. Soft-deleted IssueMatches are excluded from the detail view
				// 2. The MV counts use issuematch_rating (not issuevariant_rating),
				//    so when they differ the counts still match the detail view.
				svcCCRN := imgTest.seedRatingMismatchData()

				type vulnNode struct {
					ID                  string                         `json:"id"`
					VulnerabilityCounts *model.SeverityCounts          `json:"vulnerabilityCounts"`
					AllVulns            *model.VulnerabilityConnection `json:"allVulns"`
					HighVulns           *model.VulnerabilityConnection `json:"highVulns"`
					CriticalVulns       *model.VulnerabilityConnection `json:"criticalVulns"`
				}

				type vulnEdge struct {
					Node *vulnNode `json:"node"`
				}

				type vulnConn struct {
					Edges []*vulnEdge `json:"edges"`
				}

				query := `query ($imgFilter: ImageFilter, $first: Int) {
					Images(first: $first, filter: $imgFilter) {
						edges {
							node {
								id
								vulnerabilityCounts {
									critical
									high
								}
								allVulns: vulnerabilities {
									totalCount
								}
								criticalVulns: vulnerabilities(filter: { severity: [Critical] }) {
									totalCount
								}
								highVulns: vulnerabilities(filter: { severity: [High] }) {
									totalCount
								}
							}
						}
					}
				}`

				respData, err := e2e_common.ExecuteGqlQuery[struct {
					Images vulnConn `json:"Images"`
				}](
					imgTest.port,
					query,
					map[string]any{
						"imgFilter": map[string]any{
							"service": []string{svcCCRN},
						},
						"first": 10,
					},
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Images.Edges).ToNot(BeEmpty())

				// Find the seeded image (should have exactly 1 Critical, 1 High)
				// V1: IssueVariant=Critical, IssueMatch=Critical → counted as Critical
				// V2: IssueVariant=High, IssueMatch=Critical (overridden) → counted as Critical
				// V3: IssueVariant=High, IssueMatch=High → counted as High
				// V4: IssueVariant=Medium, IssueMatch=Medium, soft-deleted → NOT counted
				// Expected: critical=2, high=1
				foundTarget := false

				for _, edge := range respData.Images.Edges {
					node := edge.Node
					if node.VulnerabilityCounts.Critical == 2 && node.VulnerabilityCounts.High == 1 {
						foundTarget = true

						Expect(node.CriticalVulns.TotalCount).To(Equal(2),
							"detail view critical count must match badge (both using issuematch_rating)")
						Expect(node.HighVulns.TotalCount).To(Equal(1),
							"detail view high count must match badge")
						Expect(node.AllVulns.TotalCount).To(Equal(3),
							"total should be 3 (V4 is soft-deleted and must not appear)")
					}
				}

				Expect(foundTarget).To(BeTrue(),
					"expected to find the seeded image with 2 Critical + 1 High; "+
						"if not found, either the rating override (issuematch_rating) is not being "+
						"used for counting, or soft-deleted IssueMatches are leaking through")
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

	for i := range it.componentVersions {
		it.componentVersions[i].ComponentId.Int64 = int64(len(it.components)) - it.componentVersions[i].ComponentId.Int64 + 1
		id, err := it.seeder.InsertFakeComponentVersion(it.componentVersions[i])
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

	err = it.seeder.RefreshMvComponentService()
	Expect(err).To(BeNil())

	err = it.seeder.RefreshMvVulnerabilityList()
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

	err = it.seeder.RefreshMvComponentService()
	Expect(err).To(BeNil())
}

// seedServiceScopedVulnData seeds a controlled scenario to test that the detail view
// correctly scopes vulnerabilities to the filtered service:
// - Component X deployed in service A AND service B
// - V1 (Critical) with IssueMatch only in service A
// - V2 (High) with IssueMatch in both services
// Returns service A's CCRN.
func (it *imageTest) seedServiceScopedVulnData() string {
	// Service A — the one to filter on
	svcA := test.NewFakeBaseService()
	svcAId, err := it.seeder.InsertFakeBaseService(svcA)
	Expect(err).To(BeNil())

	svcA.Id = sql.NullInt64{Int64: svcAId, Valid: true}

	// Service B — must not bleed through
	svcB := test.NewFakeBaseService()
	svcBId, err := it.seeder.InsertFakeBaseService(svcB)
	Expect(err).To(BeNil())

	svcB.Id = sql.NullInt64{Int64: svcBId, Valid: true}

	// Component X
	compX := test.NewFakeComponent()
	compXId, err := it.seeder.InsertFakeComponent(compX)
	Expect(err).To(BeNil())

	cvX := test.NewFakeComponentVersion()
	cvX.ComponentId = sql.NullInt64{Int64: compXId, Valid: true}
	cvXId, err := it.seeder.InsertFakeComponentVersion(cvX)
	Expect(err).To(BeNil())

	// V1 — Critical — IssueMatch only in service A
	issueV1 := test.NewFakeIssue()
	issueV1.Type = sql.NullString{String: entity.IssueTypeVulnerability.String(), Valid: true}
	v1Id, err := it.seeder.InsertFakeIssue(issueV1)
	Expect(err).To(BeNil())

	_, err = it.seeder.InsertFakeIssueVariant(mariadb.IssueVariantRow{
		SecondaryName:     sql.NullString{String: "CVE-REGRESSION-V1-CRITICAL", Valid: true},
		Description:       sql.NullString{String: "service-A-only critical vuln", Valid: true},
		IssueId:           sql.NullInt64{Int64: v1Id, Valid: true},
		IssueRepositoryId: sql.NullInt64{Int64: 1, Valid: true},
		Rating:            sql.NullString{String: entity.SeverityValuesCritical.String(), Valid: true},
	})
	Expect(err).To(BeNil())

	cviV1 := test.NewFakeComponentVersionIssue()
	cviV1.ComponentVersionId = sql.NullInt64{Int64: cvXId, Valid: true}
	cviV1.IssueId = sql.NullInt64{Int64: v1Id, Valid: true}
	_, err = it.seeder.InsertFakeComponentVersionIssue(cviV1)
	Expect(err).To(BeNil())

	// V2 — High — IssueMatch in both services
	issueV2 := test.NewFakeIssue()
	issueV2.Type = sql.NullString{String: entity.IssueTypeVulnerability.String(), Valid: true}
	v2Id, err := it.seeder.InsertFakeIssue(issueV2)
	Expect(err).To(BeNil())

	_, err = it.seeder.InsertFakeIssueVariant(mariadb.IssueVariantRow{
		SecondaryName:     sql.NullString{String: "CVE-REGRESSION-V2-HIGH", Valid: true},
		Description:       sql.NullString{String: "both-services high vuln", Valid: true},
		IssueId:           sql.NullInt64{Int64: v2Id, Valid: true},
		IssueRepositoryId: sql.NullInt64{Int64: 1, Valid: true},
		Rating:            sql.NullString{String: entity.SeverityValuesHigh.String(), Valid: true},
	})
	Expect(err).To(BeNil())

	cviV2 := test.NewFakeComponentVersionIssue()
	cviV2.ComponentVersionId = sql.NullInt64{Int64: cvXId, Valid: true}
	cviV2.IssueId = sql.NullInt64{Int64: v2Id, Valid: true}
	_, err = it.seeder.InsertFakeComponentVersionIssue(cviV2)
	Expect(err).To(BeNil())

	// ComponentInstance in service A
	ciA := test.NewFakeComponentInstance()
	ciA.ServiceId = sql.NullInt64{Int64: svcAId, Valid: true}
	ciA.ComponentVersionId = sql.NullInt64{Int64: cvXId, Valid: true}
	ciAId, err := it.seeder.InsertFakeComponentInstance(ciA)
	Expect(err).To(BeNil())

	// ComponentInstance in service B
	ciB := test.NewFakeComponentInstance()
	ciB.ServiceId = sql.NullInt64{Int64: svcBId, Valid: true}
	ciB.ComponentVersionId = sql.NullInt64{Int64: cvXId, Valid: true}
	ciBId, err := it.seeder.InsertFakeComponentInstance(ciB)
	Expect(err).To(BeNil())

	// IssueMatch: V1 only in service A
	imV1A := test.NewFakeIssueMatch()
	imV1A.Status = sql.NullString{String: entity.IssueMatchStatusValuesNew.String(), Valid: true}
	imV1A.Rating = sql.NullString{String: entity.SeverityValuesCritical.String(), Valid: true}
	imV1A.IssueId = sql.NullInt64{Int64: v1Id, Valid: true}
	imV1A.ComponentInstanceId = sql.NullInt64{Int64: ciAId, Valid: true}
	imV1A.UserId = sql.NullInt64{Int64: util.SystemUserId, Valid: true}
	imV1A.RemediationDate = sql.NullTime{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true}
	imV1A.TargetRemediationDate = sql.NullTime{Time: time.Now().Add(60 * 24 * time.Hour), Valid: true}
	_, err = it.seeder.InsertFakeIssueMatch(imV1A)
	Expect(err).To(BeNil())

	// IssueMatch: V2 in service A
	imV2A := test.NewFakeIssueMatch()
	imV2A.Status = sql.NullString{String: entity.IssueMatchStatusValuesNew.String(), Valid: true}
	imV2A.Rating = sql.NullString{String: entity.SeverityValuesHigh.String(), Valid: true}
	imV2A.IssueId = sql.NullInt64{Int64: v2Id, Valid: true}
	imV2A.ComponentInstanceId = sql.NullInt64{Int64: ciAId, Valid: true}
	imV2A.UserId = sql.NullInt64{Int64: util.SystemUserId, Valid: true}
	imV2A.RemediationDate = sql.NullTime{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true}
	imV2A.TargetRemediationDate = sql.NullTime{Time: time.Now().Add(60 * 24 * time.Hour), Valid: true}
	_, err = it.seeder.InsertFakeIssueMatch(imV2A)
	Expect(err).To(BeNil())

	// IssueMatch: V2 in service B — must NOT appear when filtering by service A
	imV2B := test.NewFakeIssueMatch()
	imV2B.Status = sql.NullString{String: entity.IssueMatchStatusValuesNew.String(), Valid: true}
	imV2B.Rating = sql.NullString{String: entity.SeverityValuesHigh.String(), Valid: true}
	imV2B.IssueId = sql.NullInt64{Int64: v2Id, Valid: true}
	imV2B.ComponentInstanceId = sql.NullInt64{Int64: ciBId, Valid: true}
	imV2B.UserId = sql.NullInt64{Int64: util.SystemUserId, Valid: true}
	imV2B.RemediationDate = sql.NullTime{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true}
	imV2B.TargetRemediationDate = sql.NullTime{Time: time.Now().Add(60 * 24 * time.Hour), Valid: true}
	_, err = it.seeder.InsertFakeIssueMatch(imV2B)
	Expect(err).To(BeNil())

	// Refresh all MVs
	err = it.seeder.RefreshComponentVulnerabilityCounts()
	Expect(err).To(BeNil())

	err = it.seeder.RefreshMvComponentService()
	Expect(err).To(BeNil())

	err = it.seeder.RefreshMvVulnerabilityList()
	Expect(err).To(BeNil())

	return svcA.CCRN.String
}

// seedRatingMismatchData seeds a scenario to test:
// 1. issuematch_rating is used for severity bucketing (not issuevariant_rating)
// 2. Soft-deleted IssueMatches are excluded from both counts and detail view
//
// Setup:
// - Component Y in service C
// - V1: IssueVariant=Critical, IssueMatch=Critical (normal case)
// - V2: IssueVariant=High, IssueMatch=Critical (overridden severity)
// - V3: IssueVariant=High, IssueMatch=High (normal case)
// - V4: IssueVariant=Medium, IssueMatch=Medium, soft-deleted
//
// Expected when using issuematch_rating: Critical=2, High=1, Medium=0
// If issuevariant_rating were used: Critical=1, High=2, Medium=1
func (it *imageTest) seedRatingMismatchData() string {
	svc := test.NewFakeBaseService()
	svcId, err := it.seeder.InsertFakeBaseService(svc)
	Expect(err).To(BeNil())

	comp := test.NewFakeComponent()
	compId, err := it.seeder.InsertFakeComponent(comp)
	Expect(err).To(BeNil())

	cv := test.NewFakeComponentVersion()
	cv.ComponentId = sql.NullInt64{Int64: compId, Valid: true}
	cvId, err := it.seeder.InsertFakeComponentVersion(cv)
	Expect(err).To(BeNil())

	ci := test.NewFakeComponentInstance()
	ci.ServiceId = sql.NullInt64{Int64: svcId, Valid: true}
	ci.ComponentVersionId = sql.NullInt64{Int64: cvId, Valid: true}
	ciId, err := it.seeder.InsertFakeComponentInstance(ci)
	Expect(err).To(BeNil())

	// V1: IssueVariant=Critical, IssueMatch=Critical
	v1Id := it.seedIssueWithMatch(cvId, ciId, "Critical", "Critical", false)
	_ = v1Id

	// V2: IssueVariant=High, IssueMatch=Critical (severity overridden)
	v2Id := it.seedIssueWithMatch(cvId, ciId, "High", "Critical", false)
	_ = v2Id

	// V3: IssueVariant=High, IssueMatch=High
	v3Id := it.seedIssueWithMatch(cvId, ciId, "High", "High", false)
	_ = v3Id

	// V4: IssueVariant=Medium, IssueMatch=Medium, SOFT-DELETED
	v4Id := it.seedIssueWithMatch(cvId, ciId, "Medium", "Medium", true)
	_ = v4Id

	// Refresh all MVs
	err = it.seeder.RefreshComponentVulnerabilityCounts()
	Expect(err).To(BeNil())

	err = it.seeder.RefreshMvComponentService()
	Expect(err).To(BeNil())

	err = it.seeder.RefreshMvVulnerabilityList()
	Expect(err).To(BeNil())

	return svc.CCRN.String
}

// seedIssueWithMatch creates an Issue + IssueVariant + CVI + IssueMatch with configurable
// variant rating, match rating, and soft-delete state.
func (it *imageTest) seedIssueWithMatch(cvId, ciId int64, variantRating, matchRating string, deleted bool) int64 {
	issue := test.NewFakeIssue()
	issue.Type = sql.NullString{String: entity.IssueTypeVulnerability.String(), Valid: true}
	issueId, err := it.seeder.InsertFakeIssue(issue)
	Expect(err).To(BeNil())

	_, err = it.seeder.InsertFakeIssueVariant(mariadb.IssueVariantRow{
		SecondaryName:     sql.NullString{String: gofakeit.UUID(), Valid: true},
		Description:       sql.NullString{String: "test variant", Valid: true},
		IssueId:           sql.NullInt64{Int64: issueId, Valid: true},
		IssueRepositoryId: sql.NullInt64{Int64: 1, Valid: true},
		Rating:            sql.NullString{String: variantRating, Valid: true},
	})
	Expect(err).To(BeNil())

	cvi := test.NewFakeComponentVersionIssue()
	cvi.ComponentVersionId = sql.NullInt64{Int64: cvId, Valid: true}
	cvi.IssueId = sql.NullInt64{Int64: issueId, Valid: true}
	_, err = it.seeder.InsertFakeComponentVersionIssue(cvi)
	Expect(err).To(BeNil())

	im := test.NewFakeIssueMatch()
	im.Status = sql.NullString{String: entity.IssueMatchStatusValuesNew.String(), Valid: true}
	im.Rating = sql.NullString{String: matchRating, Valid: true}
	im.IssueId = sql.NullInt64{Int64: issueId, Valid: true}
	im.ComponentInstanceId = sql.NullInt64{Int64: ciId, Valid: true}
	im.UserId = sql.NullInt64{Int64: util.SystemUserId, Valid: true}
	im.RemediationDate = sql.NullTime{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true}
	im.TargetRemediationDate = sql.NullTime{Time: time.Now().Add(60 * 24 * time.Hour), Valid: true}

	if deleted {
		im.DeletedAt = sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true}
	}

	_, err = it.seeder.InsertFakeIssueMatch(im)
	Expect(err).To(BeNil())

	return issueId
}

func (it *imageTest) testImageSortingWithTieBreaker() {
	it.seedTieBreakerData()

	respData, err := e2e_common.ExecuteGqlQueryFromFile[struct {
		Images model.ImageConnection `json:"Images"`
	}](
		it.port,
		"../api/graphql/graph/queryCollection/image/query.graphql",
		map[string]any{
			"filter": map[string]any{
				"service": lo.Map(
					it.services,
					func(item mariadb.BaseServiceRow, index int) string { return item.CCRN.String },
				),
			},
			"first": 20,
			"after": "",
		},
	)

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
