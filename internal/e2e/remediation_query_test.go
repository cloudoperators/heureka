// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Getting Remediations via API", Label("e2e", "Remediations"), func() {
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

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			b, err := os.ReadFile("../api/graphql/graph/queryCollection/remediation/query.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				Remediations model.RemediationConnection `json:"Remediations"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.Remediations.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {

		var seedCollection *test.SeedCollection
		var respData struct {
			Remediations model.RemediationConnection `json:"Remediations"`
		}
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct result count", func() {
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				b, err := os.ReadFile("../api/graphql/graph/queryCollection/remediation/query.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "")

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Remediations.TotalCount).To(Equal(len(seedCollection.RemediationRows)))
				Expect(len(respData.Remediations.Edges)).To(Equal(5))
			})
			It("- returns the expected PageInfo", func() {
				Expect(*respData.Remediations.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.Remediations.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.Remediations.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.Remediations.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.Remediations.PageInfo.PageNumber).To(Equal(1), "Correct page number")
			})
			It("- returns the expected content", func() {
				for _, remediation := range respData.Remediations.Edges {
					Expect(remediation.Node.ID).ToNot(BeNil(), "remediation has ID set")
					Expect(remediation.Node.Description).ToNot(BeNil(), "remediation has description set")
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

	When("the database has 10 entries", func() {

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			remediation = testentity.NewFakeRemediationEntity()
			remediation.Service = seedCollection.ServiceRows[0].CCRN.String
			remediation.Component = seedCollection.ComponentRows[0].CCRN.String
			remediation.Issue = seedCollection.IssueRows[0].PrimaryName.String
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new remediation", func() {
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				b, err := os.ReadFile("../api/graphql/graph/queryCollection/remediation/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"description":     remediation.Description,
					"type":            string(remediation.Type),
					"service":         remediation.Service,
					"image":           remediation.Component,
					"vulnerability":   remediation.Issue,
					"remediationDate": remediation.RemediationDate.Format(time.RFC3339),
					"expirationDate":  remediation.ExpirationDate.Format(time.RFC3339),
					"remediatedBy":    remediation.RemediatedBy,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Remediation model.Remediation `json:"createRemediation"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Remediation.Description).To(Equal(remediation.Description))
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

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		var client *graphql.Client
		var req *graphql.Request
		var ctx context.Context
		var respData struct {
			Remediation model.Remediation `json:"updateRemediation"`
		}

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			client = graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			b, err := os.ReadFile("../api/graphql/graph/queryCollection/remediation/update.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req = graphql.NewRequest(str)
			req.Header.Set("Cache-Control", "no-cache")
			ctx = context.Background()
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates remediation", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()

				description := "NewDescription"

				req.Var("id", fmt.Sprintf("%d", remediation.Id))
				req.Var("input", map[string]string{
					"description": description,
				})

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Remediation.Description).To(Equal(description))
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

				req.Var("id", fmt.Sprintf("%d", remediation.Id))
				req.Var("input", map[string]string{
					"service": service.CCRN.String,
				})

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Remediation.Service).To(Equal(service.CCRN.String))
				Expect(*respData.Remediation.ServiceID).To(Equal(fmt.Sprintf("%d", service.Id.Int64)))
			})
			It("updates component id", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()
				component := seedCollection.ComponentRows[1]

				req.Var("id", fmt.Sprintf("%d", remediation.Id))
				req.Var("input", map[string]string{
					"image": component.CCRN.String,
				})

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Remediation.Image).To(Equal(component.CCRN.String))
				Expect(*respData.Remediation.ImageID).To(Equal(fmt.Sprintf("%d", component.Id.Int64)))
			})
			It("updates issue id", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()
				issue := seedCollection.IssueRows[1]

				req.Var("id", fmt.Sprintf("%d", remediation.Id))
				req.Var("input", map[string]string{
					"vulnerability": issue.PrimaryName.String,
				})

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

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

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes remediation", func() {
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				b, err := os.ReadFile("../api/graphql/graph/queryCollection/remediation/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.RemediationRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteRemediation"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
