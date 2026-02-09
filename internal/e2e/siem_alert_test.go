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
				input := map[string]interface{}{
					"name":        "Root or admin action - VAULT",
					"description": "some description",
					"severity":    "High",
					"url":         "https://example.test/alert/123",
					"region":      "eu-de-1",
					"cluster":     "eu-de-1",
					"namespace":   "vault",
					"pod":         "vault-1",
					"container":   "audit",
					"service":     "vault",
					"supportGroup": "src",
				}

				respData := e2e_common.ExecuteGqlQueryFromFile[struct {
					SIEM model.SIEMAlert `json:"createSIEMAlert"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/siem_alert/create.graphql",
					map[string]interface{}{
						"input": input,
					})

				Expect(*respData.SIEM.Name).To(Equal("Root or admin action - VAULT"))
				Expect(*respData.SIEM.Severity).To(Equal(model.SeverityValues("High")))
				Expect(*respData.SIEM.URL).To(Equal("https://example.test/alert/123"))

				name := input["name"].(string)
				issues, err := db.GetIssues(&entity.IssueFilter{PrimaryName: []*string{&name}}, nil)
				Expect(err).To(BeNil())
				Expect(len(issues)).To(Equal(1))
				issueId := issues[0].Issue.Id

				ivs, err := db.GetIssueVariants(&entity.IssueVariantFilter{IssueId: []*int64{&issueId}})
				Expect(err).To(BeNil())
				Expect(len(ivs)).To(BeNumerically(">=", 1))

				foundVariant := false
				for _, v := range ivs {
					if v.ExternalUrl == input["url"].(string) {
						Expect(v.Severity.Value).To(Equal("High"))
						foundVariant = true
						break
					}
				}
				Expect(foundVariant).To(BeTrue())

				serviceFilter := &entity.ServiceFilter{CCRN: []*string{func() *string { s := input["service"].(string); return &s }()}}
				services, err := db.GetServices(serviceFilter, nil)
				Expect(err).To(BeNil())
				Expect(len(services)).To(BeNumerically(">=", 1))

				sg := input["supportGroup"].(string)
				sgFilter := &entity.SupportGroupFilter{CCRN: []*string{&sg}}
				sgs, err := db.GetSupportGroups(sgFilter, nil)
				Expect(err).To(BeNil())
				Expect(len(sgs)).To(BeNumerically(">=", 1))

				region := input["region"].(string)
				namespace := input["namespace"].(string)
				pod := input["pod"].(string)
				ccrn := fmt.Sprintf("%s/%s/%s/%s/%s/%s", input["service"].(string), region, input["cluster"].(string), namespace, pod, input["container"].(string))

				cis, err := db.GetComponentInstances(&entity.ComponentInstanceFilter{CCRN: []*string{&ccrn}}, nil)
				Expect(err).To(BeNil())
				Expect(len(cis)).To(BeNumerically(">=", 1))
				ciId := cis[0].ComponentInstance.Id

				ims, err := db.GetIssueMatches(&entity.IssueMatchFilter{IssueId: []*int64{&issueId}}, nil)
				Expect(err).To(BeNil())
				Expect(len(ims)).To(BeNumerically(">=", 1))

				matchFound := false
				for _, m := range ims {
					if m.IssueMatch.ComponentInstanceId == ciId {
						matchFound = true
						break
					}
				}
				Expect(matchFound).To(BeTrue())
			})
		})
	})
})
