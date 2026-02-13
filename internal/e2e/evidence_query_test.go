// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"
	"time"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

const (
	testDescription = "New Description"
)

var _ = Describe("Getting Evidences via API", Label("e2e", "Evidences"), func() {
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
		_ = dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			// create a queryCollection (safe to share across requests)
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/evidence/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "0")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				Evidences model.EvidenceConnection `json:"Evidences"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.Evidences.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context(", no additional filters are present", func() {
			Context("and  a minimal query is performed", Label("minimal.graphql"), func() {
				It("returns correct result count", func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/evidence/minimal.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 5)
					req.Var("after", "0")

					req.Header.Set("Cache-Control", "no-cache")
					ctx := context.Background()

					var respData struct {
						Evidences model.EvidenceConnection `json:"Evidences"`
					}
					if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
						logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
					}

					Expect(respData.Evidences.TotalCount).To(Equal(len(seedCollection.EvidenceRows)))
					Expect(len(respData.Evidences.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {
				var respData struct {
					Evidences model.EvidenceConnection `json:"Evidences"`
				}
				BeforeEach(func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/evidence/directRelations.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 5)
					req.Var("after", "0")

					req.Header.Set("Cache-Control", "no-cache")
					ctx := context.Background()

					err = client.Run(ctx, req, &respData)

					Expect(err).To(BeNil(), "Error while unmarshaling")
				})

				It("- returns the correct result count", func() {
					Expect(respData.Evidences.TotalCount).To(Equal(len(seedCollection.EvidenceRows)))
					Expect(len(respData.Evidences.Edges)).To(Equal(5))
				})

				It("- returns the expected content", func() {
					// this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage

					for _, evidence := range respData.Evidences.Edges {
						Expect(evidence.Node.ID).ToNot(BeNil(), "evidence has a ID set")
						Expect(evidence.Node.Description).ToNot(BeNil(), "evidence has a description set")

						for _, im := range evidence.Node.IssueMatches.Edges {
							Expect(im.Node.ID).ToNot(BeNil(), "evidence has a ID set")
							Expect(im.Node.Status).ToNot(BeNil(), "evidence has a status set")
							Expect(im.Node.RemediationDate).ToNot(BeNil(), "evidence has a remediationDate set")
							Expect(im.Node.DiscoveryDate).ToNot(BeNil(), "evidence has a discoveryDate set")
							Expect(im.Node.TargetRemediationDate).ToNot(BeNil(), "evidence has a targetRemediationDate set")

							_, evidenceFound := lo.Find(seedCollection.IssueMatchEvidenceRows, func(row mariadb.IssueMatchEvidenceRow) bool {
								return fmt.Sprintf("%d", row.EvidenceId.Int64) == evidence.Node.ID && // correct evidence
									fmt.Sprintf("%d", row.IssueMatchId.Int64) == im.Node.ID // belongs actually to the im
							})
							Expect(evidenceFound).To(BeTrue(), "attached vm does exist and belongs to evidence")
						}

						author := evidence.Node.Author
						Expect(author.ID).ToNot(BeNil(), "author has a ID set")
						Expect(author.UniqueUserID).ToNot(BeNil(), "author has a uniqueUserId set")
						Expect(author.Type).ToNot(BeNil(), "author has a type set")
						Expect(author.Name).ToNot(BeNil(), "author has a name set")

						activity := evidence.Node.Activity
						Expect(activity.ID).ToNot(BeNil(), "activity has a ID set")
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.Evidences.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.Evidences.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.Evidences.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.Evidences.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
					Expect(*respData.Evidences.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
		})
	})
})

var _ = Describe("Creating Evidence via API", Label("e2e", "Evidences"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var evidence entity.Evidence
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
		_ = dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			evidence = testentity.NewFakeEvidenceEntity()
			evidence.UserId = seedCollection.UserRows[0].Id.Int64
			evidence.ActivityId = seedCollection.ActivityRows[0].Id.Int64
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new evidence", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/evidence/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]interface{}{
					"description": evidence.Description,
					"authorId":    fmt.Sprintf("%d", evidence.UserId),
					"activityId":  fmt.Sprintf("%d", evidence.ActivityId),
					"type":        evidence.Type.String(),
					"severity": map[string]string{
						"vector": evidence.Severity.Cvss.Vector,
					},
					"raaEnd": evidence.RaaEnd.Format(time.RFC3339),
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Evidence model.Evidence `json:"createEvidence"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Evidence.Description).To(Equal(evidence.Description))
				Expect(*respData.Evidence.AuthorID).To(Equal(fmt.Sprintf("%d", evidence.UserId)))
				Expect(*respData.Evidence.ActivityID).To(Equal(fmt.Sprintf("%d", evidence.ActivityId)))
				Expect(*respData.Evidence.Vector).To(Equal(evidence.Severity.Cvss.Vector))
				Expect(*respData.Evidence.RaaEnd).To(Equal(evidence.RaaEnd.Format(time.RFC3339)))
			})
		})
	})
})

var _ = Describe("Updating evidence via API", Label("e2e", "Evidences"), func() {
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
		_ = dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates evidence", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/evidence/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				evidence := seedCollection.EvidenceRows[0].AsEvidence()
				evidence.Description = testDescription

				req.Var("id", fmt.Sprintf("%d", evidence.Id))
				req.Var("input", map[string]string{
					"description": evidence.Description,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Evidence model.Evidence `json:"updateEvidence"`
				}

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Evidence.Description).To(Equal(evidence.Description))
				Expect(*respData.Evidence.AuthorID).To(Equal(fmt.Sprintf("%d", evidence.UserId)))
				Expect(*respData.Evidence.ActivityID).To(Equal(fmt.Sprintf("%d", evidence.ActivityId)))
				Expect(*respData.Evidence.Vector).To(Equal(evidence.Severity.Cvss.Vector))
				Expect(*respData.Evidence.RaaEnd).To(Equal(evidence.RaaEnd.Format(time.RFC3339)))
			})
		})
	})
})

var _ = Describe("Deleting Evidence via API", Label("e2e", "Evidences"), func() {
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
		_ = dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes evidence", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				// @todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/evidence/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.EvidenceRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteEvidence"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})
