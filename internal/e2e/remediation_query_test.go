// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"time"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Getting Remediations via API", Label("e2e", "Remediations"), func() {
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

	When("the database is empty", func() {
		It("returns empty result set", func() {

			respData := e2e_common.ExecuteGqlQueryFromFile[struct {
				Remediations model.RemediationConnection `json:"Remediations"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/remediation/query.graphql",
				map[string]interface{}{
					"filter": map[string]string{},
					"first":  10,
					"after":  "",
				})
			Expect(respData.Remediations.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		type remediationRespDataType struct {
			Remediations model.RemediationConnection `json:"Remediations"`
		}
		var respData remediationRespDataType
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct result count", func() {
				respData = e2e_common.ExecuteGqlQueryFromFile[remediationRespDataType](
					cfg.Port,
					"../api/graphql/graph/queryCollection/remediation/query.graphql",
					map[string]interface{}{
						"filter": map[string]string{},
						"first":  5,
						"after":  ""})
				Expect(respData.Remediations.TotalCount).To(Equal(len(seedCollection.RemediationRows)))
				Expect(len(respData.Remediations.Edges)).To(Equal(5))
				//- returns the expected PageInfo
				Expect(*respData.Remediations.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.Remediations.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.Remediations.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.Remediations.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.Remediations.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				//- returns the expected content
				for _, remediation := range respData.Remediations.Edges {
					Expect(remediation.Node.ID).ToNot(BeNil(), "remediation has ID set")
					Expect(remediation.Node.Description).ToNot(BeNil(), "remediation has description set")
					Expect(remediation.Node.Severity).ToNot(BeNil(), "remediation has severity set")
					Expect(remediation.Node.RemediatedBy).ToNot(BeNil(), "remediation has remediatedBy set")
					Expect(remediation.Node.RemediationDate).ToNot(BeNil(), "remediation has remediationDate set")
					Expect(remediation.Node.Service).ToNot(BeNil(), "remediation has service set")
					Expect(remediation.Node.Vulnerability).ToNot(BeNil(), "remediation has vulnerability set")
					Expect(remediation.Node.ExpirationDate).ToNot(BeNil(), "remediation has expirationDate set")

					_, remediationFound := lo.Find(seedCollection.RemediationRows, func(row mariadb.RemediationRow) bool {
						return fmt.Sprintf("%d", row.Id.Int64) == remediation.Node.ID
					})
					Expect(remediationFound).To(BeTrue(), "remediation exists in seeded data")
				}
			})
		})
	})
})

var _ = Describe("Creating Remediation via API", Label("e2e", "Remediations"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var remediation entity.Remediation
	var db *mariadb.SqlDatabase
	var seedCollection *test.SeedCollection

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

	When("the database has 10 entries", func() {

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			remediation = testentity.NewFakeRemediationEntity()
			remediation.Service = seedCollection.ServiceRows[0].CCRN.String
			remediation.Component = seedCollection.ComponentRows[0].Repository.String
			remediation.Issue = seedCollection.IssueRows[0].PrimaryName.String
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new remediation", func() {
				respData := e2e_common.ExecuteGqlQueryFromFile[struct {
					Remediation model.Remediation `json:"createRemediation"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/remediation/create.graphql",
					map[string]interface{}{
						"input": map[string]string{
							"description":     remediation.Description,
							"type":            remediation.Type.String(),
							"severity":        remediation.Severity.String(),
							"service":         remediation.Service,
							"image":           remediation.Component,
							"vulnerability":   remediation.Issue,
							"remediationDate": remediation.RemediationDate.Format(time.RFC3339),
							"expirationDate":  remediation.ExpirationDate.Format(time.RFC3339),
							"remediatedBy":    remediation.RemediatedBy,
						}})
				Expect(*respData.Remediation.Description).To(Equal(remediation.Description))
				Expect(respData.Remediation.Severity.String()).To(Equal(remediation.Severity.String()))
				Expect(*respData.Remediation.Service).To(Equal(remediation.Service))
				Expect(*respData.Remediation.Vulnerability).To(Equal(remediation.Issue))
				Expect(*respData.Remediation.Image).To(Equal(remediation.Component))
				Expect(*respData.Remediation.RemediatedBy).To(Equal(remediation.RemediatedBy))
				Expect(*respData.Remediation.RemediationDate).To(Equal(remediation.RemediationDate.Format(time.RFC3339)))
				Expect(*respData.Remediation.ExpirationDate).To(Equal(remediation.ExpirationDate.Format(time.RFC3339)))
			})
		})
	})
})

var _ = Describe("Updating remediation via API", Label("e2e", "Remediations"), func() {
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

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		type remediationUpdateRespDataType struct {
			Remediation model.Remediation `json:"updateRemediation"`
		}
		var respData remediationUpdateRespDataType

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates remediation", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()
				description := "NewDescription"

				respData = e2e_common.ExecuteGqlQueryFromFile[remediationUpdateRespDataType](
					cfg.Port,
					"../api/graphql/graph/queryCollection/remediation/update.graphql",
					map[string]interface{}{
						"id": fmt.Sprintf("%d", remediation.Id),
						"input": map[string]string{
							"description": description,
						}})

				Expect(*respData.Remediation.Description).To(Equal(description))
				Expect(respData.Remediation.Severity.String()).To(Equal(remediation.Severity.String()))
				Expect(*respData.Remediation.Service).To(Equal(remediation.Service))
				Expect(*respData.Remediation.Vulnerability).To(Equal(remediation.Issue))
				Expect(*respData.Remediation.Image).To(Equal(remediation.Component))
				Expect(*respData.Remediation.RemediatedBy).To(Equal(remediation.RemediatedBy))
				Expect(*respData.Remediation.RemediationDate).To(Equal(remediation.RemediationDate.UTC().Format(time.RFC3339)))
				Expect(*respData.Remediation.ExpirationDate).To(Equal(remediation.ExpirationDate.UTC().Format(time.RFC3339)))
			})
			It("updates service id", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()
				service := seedCollection.ServiceRows[1]
				respData = e2e_common.ExecuteGqlQueryFromFile[remediationUpdateRespDataType](
					cfg.Port,
					"../api/graphql/graph/queryCollection/remediation/update.graphql",
					map[string]interface{}{
						"id": fmt.Sprintf("%d", remediation.Id),
						"input": map[string]string{
							"service": service.CCRN.String,
						}})

				Expect(*respData.Remediation.Service).To(Equal(service.CCRN.String))
				Expect(*respData.Remediation.ServiceID).To(Equal(fmt.Sprintf("%d", service.Id.Int64)))
			})
			It("updates component id", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()
				component := seedCollection.ComponentRows[1]

				respData = e2e_common.ExecuteGqlQueryFromFile[remediationUpdateRespDataType](
					cfg.Port,
					"../api/graphql/graph/queryCollection/remediation/update.graphql",
					map[string]interface{}{
						"id": fmt.Sprintf("%d", remediation.Id),
						"input": map[string]string{
							"image": component.Repository.String,
						}})

				Expect(*respData.Remediation.Image).To(Equal(component.Repository.String))
				Expect(*respData.Remediation.ImageID).To(Equal(fmt.Sprintf("%d", component.Id.Int64)))
			})
			It("updates issue id", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()
				issue := seedCollection.IssueRows[1]

				respData = e2e_common.ExecuteGqlQueryFromFile[remediationUpdateRespDataType](
					cfg.Port,
					"../api/graphql/graph/queryCollection/remediation/update.graphql",
					map[string]interface{}{
						"id": fmt.Sprintf("%d", remediation.Id),
						"input": map[string]string{
							"vulnerability": issue.PrimaryName.String,
						}})

				Expect(*respData.Remediation.Vulnerability).To(Equal(issue.PrimaryName.String))
				Expect(*respData.Remediation.VulnerabilityID).To(Equal(fmt.Sprintf("%d", issue.Id.Int64)))
			})
		})
	})
})

var _ = Describe("Deleting Remediation via API", Label("e2e", "Remediations"), func() {

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

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes remediation", func() {
				id := fmt.Sprintf("%d", seedCollection.RemediationRows[0].Id.Int64)
				respData := e2e_common.ExecuteGqlQueryFromFile[struct {
					Id string `json:"deleteRemediation"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/remediation/delete.graphql",
					map[string]interface{}{
						"id": id,
					})
				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
