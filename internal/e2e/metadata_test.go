// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testIssuePrimaryName        = "PN-001"
	testCreatedIssueDescription = "Created Issue"
	testUpdatedIssueDescription = "Updated Issue"
	dbDateLayout                = "2006-01-02 15:04:05 -0700 MST"
)

var (
	testCreatedIssueType = entity.IssueTypeVulnerability.String()
	testUpdatedIssueType = entity.IssueTypePolicyViolation.String()
)

func createTestIssue(port string) string {
	return createIssue(port, testIssuePrimaryName, testCreatedIssueDescription, testCreatedIssueType)
}

func createIssue(port string, primaryName string, description string, itype string) string {
	issue := e2e_common.QueryCreateIssue(port, e2e_common.Issue{PrimaryName: primaryName, Description: description, Type: itype})
	Expect(*issue.PrimaryName).To(Equal(primaryName))
	Expect(*issue.Description).To(Equal(description))
	Expect(issue.Type.String()).To(Equal(itype))
	return issue.ID
}

func updateTestIssue(port string, iid string) {
	issue := e2e_common.QueryUpdateIssue(port, e2e_common.Issue{PrimaryName: testIssuePrimaryName, Description: testUpdatedIssueDescription, Type: testUpdatedIssueType}, iid)
	Expect(*issue.PrimaryName).To(Equal(testIssuePrimaryName))
	Expect(*issue.Description).To(Equal(testUpdatedIssueDescription))
	Expect(issue.Type.String()).To(Equal(testUpdatedIssueType))
}

func deleteTestIssue(port string, iid string) {
	issueId := e2e_common.QueryDeleteIssue(port, iid)
	Expect(issueId).To(Equal(iid))
}

func getTestIssue(port string) model.Issue {
	issues := e2e_common.QueryGetIssue(port, testIssuePrimaryName)
	Expect(issues.TotalCount).To(Equal(1))
	return *issues.Edges[0].Node
}

func getTestIssuesWithoutFiltering(port string) []model.Issue {
	issues := e2e_common.QueryGetIssuesWithoutFiltering(port)
	iss := []model.Issue{}
	for i := range issues.Edges {
		iss = append(iss, *issues.Edges[i].Node)
	}
	return iss
}

func getTestIssuesFilteringByState(port string, state []string) []model.Issue {
	issues := e2e_common.QueryGetIssuesFilteringByState(port, state)
	iss := []model.Issue{}
	for i := range issues.Edges {
		iss = append(iss, *issues.Edges[i].Node)
	}
	return iss
}

func parseTimeExpectNoError(t string) time.Time {
	tt, err := time.Parse(dbDateLayout, t)
	Expect(err).Should(BeNil())
	return tt
}

var _ = Describe("Creating, updating and state filtering of entity via API", Label("e2e", "Entities"), func() {
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		db = dbm.NewTestSchemaWithoutMigration()

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("New issue is created via API", func() {
		var issue model.Issue
		BeforeEach(func() {
			createTestIssue(cfg.Port)
			issue = getTestIssue(cfg.Port)
		})
		It("shall assign CreatedBy, CreatedAt, UpdatedBy and UpdatedAt metadata fields and shall keep zero value in DeltedAt metadata fields", func() {
			Expect(*issue.Description).To(Equal(testCreatedIssueDescription))
			Expect(issue.Type.String()).To(Equal(testCreatedIssueType))

			Expect(issue.Metadata).To(Not(BeNil()))
			Expect(*issue.Metadata.CreatedBy).To(Equal(fmt.Sprintf("%d", util.SystemUserId)))

			createdAt := parseTimeExpectNoError(*issue.Metadata.CreatedAt)
			updatedAt := parseTimeExpectNoError(*issue.Metadata.UpdatedAt)

			Expect(createdAt).Should(BeTemporally("~", time.Now().UTC(), 3*time.Second))
			Expect(*issue.Metadata.UpdatedBy).To(Equal(fmt.Sprintf("%d", util.SystemUserId)))
			Expect(updatedAt).To(Equal(createdAt))
			Expect(*issue.Metadata.DeletedAt).To(Equal(time.Unix(0, 0).Local().Format(dbDateLayout)))
		})
	})
	When("Issue is updated via API", func() {
		var issue model.Issue
		BeforeEach(func() {
			iid := createTestIssue(cfg.Port)
			time.Sleep(1100 * time.Millisecond)
			updateTestIssue(cfg.Port, iid)
			issue = getTestIssue(cfg.Port)
		})
		It("shall assign UpdatedBy and UpdatedAt metadata fields and shall keep zero value in DeletedAt metadata field", func() {
			Expect(*issue.Description).To(Equal(testUpdatedIssueDescription))
			Expect(issue.Type.String()).To(Equal(testUpdatedIssueType))

			Expect(issue.Metadata).To(Not(BeNil()))
			Expect(*issue.Metadata.CreatedBy).To(Equal(fmt.Sprintf("%d", util.SystemUserId)))

			createdAt := parseTimeExpectNoError(*issue.Metadata.CreatedAt)
			updatedAt := parseTimeExpectNoError(*issue.Metadata.UpdatedAt)

			Expect(createdAt).Should(BeTemporally("~", time.Now().UTC(), 3*time.Second))
			Expect(*issue.Metadata.UpdatedBy).To(Equal(fmt.Sprintf("%d", util.SystemUserId)))
			Expect(updatedAt).Should(BeTemporally("~", time.Now().UTC(), 2*time.Second))
			Expect(updatedAt).Should(BeTemporally(">", createdAt))
			Expect(*issue.Metadata.DeletedAt).To(Equal(time.Unix(0, 0).Local().Format(dbDateLayout)))
		})
	})
	When("Issue is deleted via API", func() {
		var issue model.Issue
		BeforeEach(func() {
			iid := createTestIssue(cfg.Port)
			time.Sleep(1100 * time.Millisecond)
			deleteTestIssue(cfg.Port, iid)
			issue = getTestIssue(cfg.Port)
		})
		It("shall assign UpdatedBy, DeletedAt and UpdatedAt metadata fields on delete", func() {
			Expect(*issue.Description).To(Equal(testCreatedIssueDescription))
			Expect(issue.Type.String()).To(Equal(testCreatedIssueType))

			Expect(issue.Metadata).To(Not(BeNil()))
			Expect(*issue.Metadata.CreatedBy).To(Equal(fmt.Sprintf("%d", util.SystemUserId)))

			createdAt := parseTimeExpectNoError(*issue.Metadata.CreatedAt)
			deletedAt := parseTimeExpectNoError(*issue.Metadata.DeletedAt)
			updatedAt := parseTimeExpectNoError(*issue.Metadata.UpdatedAt)

			Expect(createdAt).Should(BeTemporally("~", time.Now().UTC(), 3*time.Second))
			Expect(*issue.Metadata.UpdatedBy).To(Equal(fmt.Sprintf("%d", util.SystemUserId)))
			Expect(deletedAt).Should(BeTemporally("~", time.Now().UTC(), 2*time.Second))
			Expect(deletedAt).Should(BeTemporally(">", createdAt))
			Expect(deletedAt).To(Equal(updatedAt))
		})
	})

	When("Two issues are created and one of them is deleted", func() {
		var activeIssueId string
		var deletedIssueId string
		BeforeEach(func() {
			activeIssueId = createIssue(cfg.Port, "dummyPrimaryName1", "dummyDescription1", entity.IssueTypeVulnerability.String())
			deletedIssueId = createIssue(cfg.Port, "dummyPrimaryName2", "dummyDescription2", entity.IssueTypeVulnerability.String())
			deleteTestIssue(cfg.Port, deletedIssueId)
		})
		It("shall get one active issue when state filter is not specified", func() {
			issues := getTestIssuesWithoutFiltering(cfg.Port)
			Expect(len(issues)).To(Equal(1))
			Expect(issues[0].ID).To(Equal(activeIssueId))
		})
		It("shall get one active issue when state filter is specified as 'active'", func() {
			issues := getTestIssuesFilteringByState(cfg.Port, []string{model.StateFilterActive.String()})
			Expect(len(issues)).To(Equal(1))
			Expect(issues[0].ID).To(Equal(activeIssueId))
		})
		It("shall get one deleted issue when state filter is specified as 'deleted'", func() {
			issues := getTestIssuesFilteringByState(cfg.Port, []string{model.StateFilterDeleted.String()})
			Expect(len(issues)).To(Equal(1))
			Expect(issues[0].ID).To(Equal(deletedIssueId))
		})
		It("shall get both active and deleted issues when state filter is specified as 'active' and 'deleted'", func() {
			issues := getTestIssuesFilteringByState(cfg.Port, []string{model.StateFilterActive.String(), model.StateFilterDeleted.String()})
			Expect(len(issues)).To(Equal(2))
		})
	})
})
