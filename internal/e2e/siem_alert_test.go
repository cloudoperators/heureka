// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Creating SIEMAlert via API", Label("e2e", "SIEMAlert"), func() {
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

	When("the database has a user and we create a SIEM alert", func() {
		BeforeEach(func() {
			_ = seeder.SeedDbWithNFakeData(1)
		})

		Context("and a mutation query is performed", func() {
			It("creates issue/service/support group/component instance/issuevariant/issuematch", func() {
				alertName := "Root or admin action - VAULT"
				alertDescription := "some description"
				alertSeverity := "High"
				alertURL := "https://example.test/alert/123"
				region := "eu-de-1"
				clusterName := "eu-de-1"
				namespace := "vault"
				pod := "vault-1"
				container := "audit"
				service := "vault"
				supportGroup := "src"
				graphqlPath := "../api/graphql/graph/queryCollection/siem_alert/create.graphql"

				input := map[string]interface{}{
					"name":         alertName,
					"description":  alertDescription,
					"severity":     alertSeverity,
					"url":          alertURL,
					"region":       region,
					"cluster":      clusterName,
					"namespace":    namespace,
					"pod":          pod,
					"container":    container,
					"service":      service,
					"supportGroup": supportGroup,
				}

				respData, err := e2e_common.ExecuteGqlQueryFromFile[struct {
					SIEM model.SIEMAlert `json:"createSIEMAlert"`
				}](
					cfg.Port,
					graphqlPath,
					map[string]interface{}{
						"input": input,
					})
				Expect(err).To(BeNil())

				Expect(*respData.SIEM.Name).To(Equal(alertName))
				Expect(*respData.SIEM.Severity).To(Equal(model.SeverityValues(alertSeverity)))
				Expect(*respData.SIEM.URL).To(Equal(alertURL))

				issues, err := db.GetIssues(&entity.IssueFilter{PrimaryName: []*string{&alertName}}, nil)
				Expect(err).To(BeNil())
				Expect(len(issues)).To(Equal(1))
				issueId := issues[0].Issue.Id

				ivs, err := db.GetIssueVariants(&entity.IssueVariantFilter{IssueId: []*int64{&issueId}}, []entity.Order{})
				Expect(err).To(BeNil())
				Expect(len(ivs)).To(BeNumerically(">=", 1))

				Expect(ivs).To(ContainElement(
					HaveField("ExternalUrl", Equal(alertURL)),
				))

				issueVariantWithSeverity := false
				for _, v := range ivs {
					if v.ExternalUrl == alertURL && v.Severity.Value == alertSeverity {
						issueVariantWithSeverity = true
						break
					}
				}
				Expect(issueVariantWithSeverity).To(BeTrue())

				serviceFilter := &entity.ServiceFilter{CCRN: []*string{&service}}
				services, err := db.GetServices(serviceFilter, nil)
				Expect(err).To(BeNil())
				Expect(len(services)).To(BeNumerically(">=", 1))

				sgFilter := &entity.SupportGroupFilter{CCRN: []*string{&supportGroup}}
				sgs, err := db.GetSupportGroups(sgFilter, nil)
				Expect(err).To(BeNil())
				Expect(len(sgs)).To(BeNumerically(">=", 1))

				ccrn := fmt.Sprintf("%s/%s/%s/%s/%s/%s", service, region, clusterName, namespace, pod, container)

				cis, err := db.GetComponentInstances(&entity.ComponentInstanceFilter{CCRN: []*string{&ccrn}}, nil)
				Expect(err).To(BeNil())
				Expect(len(cis)).To(BeNumerically(">=", 1))
				ciId := cis[0].ComponentInstance.Id

				ims, err := db.GetIssueMatches(&entity.IssueMatchFilter{IssueId: []*int64{&issueId}}, nil)
				Expect(err).To(BeNil())
				Expect(len(ims)).To(BeNumerically(">=", 1))

				Expect(ims).To(ContainElement(
					HaveField("IssueMatch.ComponentInstanceId", Equal(ciId)),
				))
			})

			It("creates SIEM alert with a specific source", func() {
				alertName := "Custom Source Alert"
				alertDescription := "description for custom source"
				alertSeverity := "Medium"
				alertURL := "https://example.test/alert/789"
				region := "eu-de-1"
				clusterName := "eu-de-1"
				namespace := "vault"
				pod := "vault-2"
				container := "audit"
				service := "vault"
				supportGroup := "src"
				source := "custom-siem-source"
				graphqlPath := "../api/graphql/graph/queryCollection/siem_alert/create.graphql"

				input := map[string]interface{}{
					"name":         alertName,
					"description":  alertDescription,
					"severity":     alertSeverity,
					"url":          alertURL,
					"region":       region,
					"cluster":      clusterName,
					"namespace":    namespace,
					"pod":          pod,
					"container":    container,
					"service":      service,
					"supportGroup": supportGroup,
					"source":       source,
				}

				respData, err := e2e_common.ExecuteGqlQueryFromFile[struct {
					SIEM model.SIEMAlert `json:"createSIEMAlert"`
				}](
					cfg.Port,
					graphqlPath,
					map[string]interface{}{
						"input": input,
					})
				Expect(err).To(BeNil())

				Expect(*respData.SIEM.Name).To(Equal(alertName))
				Expect(*respData.SIEM.Source).To(Equal(source))

				// Verify the IssueRepository in the database
				repoFilter := &entity.IssueRepositoryFilter{Name: []*string{&source}}
				repos, err := db.GetIssueRepositories(repoFilter, []entity.Order{})
				Expect(err).To(BeNil())
				Expect(len(repos)).To(Equal(1))
				Expect(repos[0].Name).To(Equal(source))

				// Verify IssueVariant is linked to this repository
				issues, err := db.GetIssues(&entity.IssueFilter{PrimaryName: []*string{&alertName}}, nil)
				Expect(err).To(BeNil())
				issueId := issues[0].Issue.Id

				ivs, err := db.GetIssueVariants(&entity.IssueVariantFilter{
					IssueId:           []*int64{&issueId},
					IssueRepositoryId: []*int64{&repos[0].Id},
				}, []entity.Order{})
				Expect(err).To(BeNil())
				Expect(len(ivs)).To(Equal(1))
			})
		})

		Context("and a mutation query is performed without ComponentInstance data", func() {
			It("should not create the SIEM alert", func() {
				alertName := "Missing Container Alert"
				alertDescription := "alert without container"
				alertSeverity := "High"
				alertURL := "https://example.test/alert/456"
				region := "eu-de-1"
				clusterName := "eu-de-1"
				namespace := "vault"
				pod := "vault-1"
				service := "vault"
				supportGroup := "src"
				graphqlPath := "../api/graphql/graph/queryCollection/siem_alert/create.graphql"

				input := map[string]interface{}{
					"name":         alertName,
					"description":  alertDescription,
					"severity":     alertSeverity,
					"url":          alertURL,
					"region":       region,
					"cluster":      clusterName,
					"namespace":    namespace,
					"pod":          pod,
					"container":    "", // Empty container
					"service":      service,
					"supportGroup": supportGroup,
				}

				_, err := e2e_common.ExecuteGqlQueryFromFile[struct {
					SIEM model.SIEMAlert `json:"createSIEMAlert"`
				}](
					cfg.Port,
					graphqlPath,
					map[string]interface{}{
						"input": input,
					})
				Expect(err).NotTo(BeNil(), "Expected mutation to fail with missing ComponentInstance data")

				// Verify the alert was not created in the database
				issues, err := db.GetIssues(&entity.IssueFilter{PrimaryName: []*string{&alertName}}, nil)
				Expect(err).To(BeNil())
				Expect(len(issues)).To(Equal(0), "Alert should not be created when ComponentInstance data is missing")
			})
		})
	})
})
