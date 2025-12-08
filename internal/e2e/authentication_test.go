// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"strconv"
	"time"

	access_test "github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/oidc"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
)

const defaultTestFakeDataItems = 10

var _ = Describe("Creating issue value via API using OIDC authentication", Label("e2e", "Authentication"), func() {
	var authTest *authenticationTest
	var testUser entity.User
	var issue entity.Issue

	BeforeEach(func() {
		authTest = newAuthenticationTest()
		testUser = authTest.getTestUser()
		issue = testentity.NewFakeIssueEntity()
	})

	AfterEach(func() {
		authTest.teardown()
	})

	When("user creates issue", func() {
		It("assign authenticated user in CreatedBy and UpdatedBy fields", func() {
			issueResponse := authTest.createIssueByUser(issue, testUser)
			Expect(*issueResponse.Metadata.CreatedBy).To(Equal(strconv.FormatInt(testUser.Id, 10)))
			Expect(*issueResponse.Metadata.UpdatedBy).To(Equal(strconv.FormatInt(testUser.Id, 10)))
		})
	})
})

var _ = Describe("Updating issue value via API using OIDC authentication", Label("e2e", "Authentication"), func() {
	var authTest *authenticationTest
	var testUser entity.User
	var issue entity.Issue

	BeforeEach(func() {
		authTest = newAuthenticationTest()
		testUser = authTest.getTestUser()
		issue = authTest.getTestIssueCreatedByAndUpdatedBySystemUser()
	})

	AfterEach(func() {
		authTest.teardown()
	})

	When("user updates issue", func() {
		It("assign authenticated user in UpdatedBy field", func() {
			issueResponse := authTest.updateIssueByUser(issue, testUser)
			Expect(*issueResponse.Metadata.UpdatedBy).To(Equal(strconv.FormatInt(testUser.Id, 10)))
		})
	})
})

var _ = Describe("Deleting issue value via API using OIDC authentication", Label("e2e", "Authentication"), func() {
	var authTest *authenticationTest
	var testUser entity.User
	var issue entity.Issue

	BeforeEach(func() {
		authTest = newAuthenticationTest()
		testUser = authTest.getTestUser()
		issue = authTest.getTestIssueCreatedByAndUpdatedBySystemUser()
	})

	AfterEach(func() {
		authTest.teardown()
	})

	When("user deletes issue", func() {
		It("assign authenticated user in UpdatedBy field", func() {
			issueId := authTest.deleteIssueByUser(issue, testUser)
			issueResponse := authTest.getDeletedIssue(issueId, testUser)
			Expect(*issueResponse.Metadata.UpdatedBy).To(Equal(strconv.FormatInt(testUser.Id, 10)))
		})
	})
})

type authenticationTest struct {
	cfg            util.Config
	db             *mariadb.SqlDatabase
	oidcProvider   *oidc.Provider
	seeder         *test.DatabaseSeeder
	seedCollection *test.SeedCollection
	server         *server.Server
}

func newAuthenticationTest() *authenticationTest {
	db := dbm.NewTestSchemaWithoutMigration()

	cfg := dbm.DbConfig()
	cfg.Port = util2.GetRandomFreePort()
	cfg.AuthOidcClientId = "mock-client-id"
	cfg.AuthOidcUrl = fmt.Sprintf("http://localhost:%s", util2.GetRandomFreePort())
	oidcProvider := oidc.NewProvider(cfg.AuthOidcUrl, enableOidcProviderLog)
	oidcProvider.Start()

	seeder, err := test.NewDatabaseSeeder(cfg)
	Expect(err).To(BeNil(), "Database Seeder Setup should work")

	cfg.Port = util2.GetRandomFreePort()
	server := e2e_common.NewRunningServer(cfg)

	seedCollection := seeder.SeedDbWithNFakeData(defaultTestFakeDataItems)

	return &authenticationTest{
		cfg:            cfg,
		db:             db,
		oidcProvider:   oidcProvider,
		seeder:         seeder,
		seedCollection: seedCollection,
		server:         server}
}

func (at *authenticationTest) getTestUser() entity.User {
	user := at.seedCollection.UserRows[0].AsUser()
	Expect(user.Id).To(Not(Equal(util.SystemUserId)))
	return user
}

func (at *authenticationTest) getTestIssueCreatedByAndUpdatedBySystemUser() entity.Issue {
	issue := at.seedCollection.IssueRows[0].AsIssue()
	Expect(issue.Metadata.CreatedBy).To(Equal(util.SystemUserId))
	Expect(issue.Metadata.UpdatedBy).To(Equal(util.SystemUserId))
	return issue
}

func (at *authenticationTest) teardown() {
	e2e_common.ServerTeardown(at.server)
	at.oidcProvider.Stop()
	dbm.TestTearDown(at.db)
}

func (at *authenticationTest) createIssueByUser(issue entity.Issue, user entity.User) model.Issue {
	respData := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Issue model.Issue `json:"createIssue"`
	}](
		at.cfg.Port,
		"../api/graphql/graph/queryCollection/authentication/issue_create.graphql",
		map[string]interface{}{
			"input": map[string]interface{}{
				"primaryName": issue.PrimaryName,
				"description": issue.Description,
				"type":        issue.Type.String(),
			},
		},
		at.getHeaders(user))

	Expect(*respData.Issue.PrimaryName).To(Equal(issue.PrimaryName))
	Expect(*respData.Issue.Description).To(Equal(issue.Description))
	Expect(respData.Issue.Type.String()).To(Equal(issue.Type.String()))
	return respData.Issue
}

func (at *authenticationTest) updateIssueByUser(issue entity.Issue, user entity.User) model.Issue {
	issue.Description = "New Description"
	respData := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Issue model.Issue `json:"updateIssue"`
	}](
		at.cfg.Port,
		"../api/graphql/graph/queryCollection/authentication/issue_update.graphql",
		map[string]interface{}{
			"id":    strconv.FormatInt(issue.Id, 10),
			"input": map[string]string{"description": issue.Description},
		},
		at.getHeaders(user))

	Expect(*respData.Issue.Description).To(Equal(issue.Description))
	Expect(*respData.Issue.Metadata.CreatedBy).To(Equal(strconv.FormatInt(util.SystemUserId, 10)))
	return respData.Issue
}

func (at *authenticationTest) deleteIssueByUser(issue entity.Issue, user entity.User) string {
	respData := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Id string `json:"deleteIssue"`
	}](
		at.cfg.Port,
		"../api/graphql/graph/queryCollection/authentication/issue_delete.graphql",
		map[string]interface{}{
			"id": strconv.FormatInt(issue.Id, 10),
		},
		at.getHeaders(user))

	Expect(respData.Id).To(Equal(strconv.FormatInt(issue.Id, 10)))
	return respData.Id
}

func (at *authenticationTest) getDeletedIssue(issueId string, user entity.User) model.Issue {
	respData := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Issues model.IssueConnection `json:"Issues"`
	}](
		at.cfg.Port,
		"../api/graphql/graph/queryCollection/authentication/issue_get.graphql",
		map[string]interface{}{
			"filter": map[string]string{"state": "Deleted"},
			"first":  defaultTestFakeDataItems,
			"after":  "",
		},
		at.getHeaders(user))

	item, ok := lo.Find(respData.Issues.Edges, func(e *model.IssueEdge) bool { return e.Node.ID == issueId })
	Expect(ok).To(BeTrue(), "issue id '%s' not found in deleted items")
	return *item.Node
}

func (at *authenticationTest) getHeaders(user entity.User) map[string]string {
	headers := e2e_common.GqlStandardHeaders
	headers["Authorization"] = at.getAuthenticationHeaderForUser(user)
	return headers
}

func (at *authenticationTest) getAuthenticationHeaderForUser(user entity.User) string {
	oidcTokenStringHandler := access_test.CreateOidcTokenStringHandler(at.cfg.AuthOidcUrl, at.cfg.AuthOidcClientId, user.UniqueUserID)
	token := access_test.GenerateJwtWithRsaSignature(oidcTokenStringHandler, at.oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
	return access_test.WithBearer(token)
}
