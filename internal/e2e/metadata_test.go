// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/e2e/common"
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
	issue := e2e_common.QueryCreateIssue(port, e2e_common.Issue{PrimaryName: testIssuePrimaryName, Description: testCreatedIssueDescription, Type: testCreatedIssueType})
	Expect(*issue.PrimaryName).To(Equal(testIssuePrimaryName))
	Expect(*issue.Description).To(Equal(testCreatedIssueDescription))
	Expect(issue.Type.String()).To(Equal(testCreatedIssueType))
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

func parseTimeExpectNoError(t string) time.Time {
	tt, err := time.Parse(dbDateLayout, t)
	Expect(err).Should(BeNil())
	return tt
}

var _ = Describe("Creating and updating entity via API", Label("e2e", "Entities"), func() {
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		_ = dbm.NewTestSchema()

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
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
			Expect(*issue.Metadata.CreatedBy).To(Equal(fmt.Sprintf("%d", e2e_common.SystemUserId)))

			createdAt := parseTimeExpectNoError(*issue.Metadata.CreatedAt)
			updatedAt := parseTimeExpectNoError(*issue.Metadata.UpdatedAt)

			Expect(createdAt).Should(BeTemporally("~", time.Now().UTC(), 3*time.Second))
			Expect(*issue.Metadata.UpdatedBy).To(Equal(fmt.Sprintf("%d", e2e_common.SystemUserId)))
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
			Expect(*issue.Metadata.CreatedBy).To(Equal(fmt.Sprintf("%d", e2e_common.SystemUserId)))

			createdAt := parseTimeExpectNoError(*issue.Metadata.CreatedAt)
			updatedAt := parseTimeExpectNoError(*issue.Metadata.UpdatedAt)

			Expect(createdAt).Should(BeTemporally("~", time.Now().UTC(), 3*time.Second))
			Expect(*issue.Metadata.UpdatedBy).To(Equal(fmt.Sprintf("%d", e2e_common.SystemUserId)))
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
			Expect(*issue.Metadata.CreatedBy).To(Equal(fmt.Sprintf("%d", e2e_common.SystemUserId)))

			createdAt := parseTimeExpectNoError(*issue.Metadata.CreatedAt)
			deletedAt := parseTimeExpectNoError(*issue.Metadata.DeletedAt)
			updatedAt := parseTimeExpectNoError(*issue.Metadata.UpdatedAt)

			Expect(createdAt).Should(BeTemporally("~", time.Now().UTC(), 3*time.Second))
			Expect(*issue.Metadata.UpdatedBy).To(Equal(fmt.Sprintf("%d", e2e_common.SystemUserId)))
			Expect(deletedAt).Should(BeTemporally("~", time.Now().UTC(), 2*time.Second))
			Expect(deletedAt).Should(BeTemporally(">", createdAt))
			Expect(deletedAt).To(Equal(updatedAt))
		})
	})
})
