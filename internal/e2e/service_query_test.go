// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"

	"os"

	"github.com/cloudoperators/heureka/internal/entity"
	testentity "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"

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

var _ = Describe("Getting Services via API", Label("e2e", "Services"), func() {
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			// create a queryCollection (safe to share across requests)
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				Services model.ServiceConnection `json:"Services"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.Services.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {

		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct result count", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/minimal.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "")

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Services model.ServiceConnection `json:"Services"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Services.TotalCount).To(Equal(len(seedCollection.ServiceRows)))
				Expect(len(respData.Services.Edges)).To(Equal(5))
			})

		})
		Context("and we request metadata", func() {
			It("returns correct metadata counts", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/withObjectMetadata.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "")

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Services model.ServiceConnection `json:"Services"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				for _, serviceEdge := range respData.Services.Edges {
					imCount := 0
					ciCount := 0
					for _, ciEdge := range serviceEdge.Node.ComponentInstances.Edges {
						imCount += ciEdge.Node.IssueMatches.TotalCount
						ciCount += *ciEdge.Node.Count
					}
					Expect(serviceEdge.Node.ObjectMetadata.IssueMatchCount).To(Equal(imCount))
					Expect(serviceEdge.Node.ObjectMetadata.ComponentInstanceCount).To(Equal(ciCount))
				}
			})

		})
		Context("and we query to resolve levels of relations", Label("directRelations.graphql"), func() {

			var respData struct {
				Services model.ServiceConnection `json:"Services"`
			}
			BeforeEach(func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/directRelations.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 5)
				req.Var("after", "")

				req.Header.Set("Cache-Control", "no-cache")

				ctx := context.Background()

				err = client.Run(ctx, req, &respData)

				Expect(err).To(BeNil(), "Error while unmarshaling")
			})

			It("- returns the correct result count", func() {
				Expect(respData.Services.TotalCount).To(Equal(len(seedCollection.ServiceRows)))
				Expect(len(respData.Services.Edges)).To(Equal(5))
			})

			It("- returns the expected content", func() {
				//this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
				// a complete verification
				// additional checks are added based on bugs discovered during usage

				for _, service := range respData.Services.Edges {
					for _, owner := range service.Node.Owners.Edges {
						Expect(owner.Node.ID).ToNot(BeNil(), "owner has a ID set")
						Expect(owner.Node.Name).ToNot(BeNil(), "owner has a name set")
						Expect(owner.Node.UniqueUserID).ToNot(BeNil(), "owner has a uniqueUserId set")
						Expect(owner.Node.Type).ToNot(BeNil(), "owner has a type set")

						_, ownerFound := lo.Find(seedCollection.OwnerRows, func(row mariadb.OwnerRow) bool {
							return fmt.Sprintf("%d", row.UserId.Int64) == owner.Node.ID && // correct owner
								fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // belongs actually to the service
						})
						Expect(ownerFound).To(BeTrue(), "attached owner does exist and belongs to service")
					}

					for _, sg := range service.Node.SupportGroups.Edges {
						Expect(sg.Node.ID).ToNot(BeNil(), "supportGroup has a ID set")
						Expect(sg.Node.Ccrn).ToNot(BeNil(), "supportGroup has a ccrn set")

						_, sgFound := lo.Find(seedCollection.SupportGroupServiceRows, func(row mariadb.SupportGroupServiceRow) bool {
							return fmt.Sprintf("%d", row.SupportGroupId.Int64) == sg.Node.ID && // correct sg
								fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // belongs actually to the service
						})
						Expect(sgFound).To(BeTrue(), "attached supportGroup does exist and belongs to service")
					}

					for _, activity := range service.Node.Activities.Edges {
						Expect(activity.Node.ID).ToNot(BeNil(), "activity has a ID set")

						_, activityFound := lo.Find(seedCollection.ActivityHasServiceRows, func(row mariadb.ActivityHasServiceRow) bool {
							return fmt.Sprintf("%d", row.ActivityId.Int64) == activity.Node.ID && // correct activity
								fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // belongs actually to the service
						})
						Expect(activityFound).To(BeTrue(), "attached activity does exist and belongs to service")
					}

					for _, ir := range service.Node.IssueRepositories.Edges {
						Expect(ir.Node.ID).ToNot(BeNil(), "issueRepository has a ID set")
						Expect(ir.Node.Name).ToNot(BeNil(), "issueRepository has a name set")
						Expect(ir.Node.URL).ToNot(BeNil(), "issueRepository has a url set")
						Expect(ir.Priority).ToNot(BeNil(), "issueRepository has a priority set")

						_, irFound := lo.Find(seedCollection.IssueRepositoryServiceRows, func(row mariadb.IssueRepositoryServiceRow) bool {
							return fmt.Sprintf("%d", row.IssueRepositoryId.Int64) == ir.Node.ID && // correct ar
								fmt.Sprintf("%d", row.ServiceId.Int64) == service.Node.ID // belongs actually to the service
						})
						Expect(irFound).To(BeTrue(), "attached issueRepository does exist and belongs to service")
					}

					for _, ci := range service.Node.ComponentInstances.Edges {
						Expect(ci.Node.ID).ToNot(BeNil(), "componentInstance has a ID set")
						Expect(ci.Node.Ccrn).ToNot(BeNil(), "componentInstance has a ccrn set")
						Expect(ci.Node.Count).ToNot(BeNil(), "componentInstance has a count set")

						_, ciFound := lo.Find(seedCollection.ComponentInstanceRows, func(row mariadb.ComponentInstanceRow) bool {
							return fmt.Sprintf("%d", row.Id.Int64) == ci.Node.ID &&
								row.CCRN.String == *ci.Node.Ccrn &&
								int(row.Count.Int16) == *ci.Node.Count
						})
						Expect(ciFound).To(BeTrue(), "attached componentInstance does exist and belongs to service")
					}

					for _, im := range service.Node.IssueMatches.Edges {
						Expect(im.Node.ID).ToNot(BeNil(), "issueMatch has a ID set")
						Expect(im.Node.Status).ToNot(BeNil(), "issueMatch has a status set")
						Expect(im.Node.RemediationDate).ToNot(BeNil(), "issueMatch has a remediationDate set")
						Expect(im.Node.DiscoveryDate).ToNot(BeNil(), "issueMatch has a discoveryDate set")
						Expect(im.Node.TargetRemediationDate).ToNot(BeNil(), "issueMatch has a targetRemediationDate set")
					}
				}
			})
			It("- returns the expected PageInfo", func() {
				Expect(*respData.Services.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.Services.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.Services.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.Services.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.Services.PageInfo.PageNumber).To(Equal(1), "Correct page number")
			})
		})
		Context("and we use order", Label("withOrder.graphql"), func() {
			var respData struct {
				Services model.ServiceConnection `json:"Services"`
			}
			c := collate.New(language.English)

			It("can order by ccrn", Label("withOrder.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/withOrder.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("filter", map[string]string{})
				req.Var("first", 10)
				req.Var("after", "")
				req.Var("orderBy", []map[string]string{
					{"by": "ccrn", "direction": "asc"},
				})

				req.Header.Set("Cache-Control", "no-cache")

				ctx := context.Background()

				err = client.Run(ctx, req, &respData)

				Expect(err).To(BeNil(), "Error while unmarshaling")

				By("- returns the correct result count", func() {
					Expect(respData.Services.TotalCount).To(Equal(len(seedCollection.ServiceRows)))
					Expect(len(respData.Services.Edges)).To(Equal(10))
				})

				By("- returns the expected content in order", func() {
					var prev string = ""
					for _, im := range respData.Services.Edges {
						Expect(c.CompareString(*im.Node.Ccrn, prev)).Should(BeNumerically(">=", 0))
						prev = *im.Node.Ccrn
					}
				})
			})

		})
	})
	var loadTestData = func() ([]mariadb.ComponentInstanceRow, []mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, []mariadb.IssueMatchRow, error) {
		issueVariants, err := test.LoadIssueVariants(test.GetTestDataPath("../database/mariadb/testdata/component_version_order/issue_variant.json"))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		cvIssues, err := test.LoadComponentVersionIssues(test.GetTestDataPath("../database/mariadb/testdata/service_order/component_version_issue.json"))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		componentInstances, err := test.LoadComponentInstances(test.GetTestDataPath("../database/mariadb/testdata/service_order/component_instance.json"))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		issueMatches, err := test.LoadIssueMatches(test.GetTestDataPath("../database/mariadb/testdata/service_order/issue_match.json"))
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return componentInstances, issueVariants, cvIssues, issueMatches, nil
	}
	When("issueCounts are involved", func() {
		BeforeEach(func() {
			seeder.SeedIssueRepositories()
			seeder.SeedIssues(10)
			components := seeder.SeedComponents(1)
			seeder.SeedComponentVersions(10, components)
			seeder.SeedServices(5)
			componentInstances, issueVariants, componentVersionIssues, issueMatches, err := loadTestData()
			Expect(err).To(BeNil())
			// Important: the order need to be preserved
			for _, iv := range issueVariants {
				_, err := seeder.InsertFakeIssueVariant(iv)
				Expect(err).To(BeNil())
			}
			for _, cvi := range componentVersionIssues {
				_, err := seeder.InsertFakeComponentVersionIssue(cvi)
				Expect(err).To(BeNil())
			}
			for _, ci := range componentInstances {
				_, err := seeder.InsertFakeComponentInstance(ci)
				Expect(err).To(BeNil())
			}
			for _, im := range issueMatches {
				_, err := seeder.InsertFakeIssueMatch(im)
				Expect(err).To(BeNil())
			}
			err = seeder.RefreshServiceIssueCounters()
			Expect(err).To(BeNil())
		})

		var runOrderTest = func(orderDirection string, expectedOrder []string) {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/withOrder.graphql")
			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)
			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "")
			req.Var("orderBy", []map[string]string{
				{"by": "severity", "direction": orderDirection},
			})
			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()
			var respData struct {
				Services model.ServiceConnection `json:"Services"`
			}
			err = client.Run(ctx, req, &respData)
			Expect(err).To(BeNil(), "Error while unmarshaling")
			Expect(respData.Services.TotalCount).To(Equal(5))
			Expect(len(respData.Services.Edges)).To(Equal(5))
			for i, id := range expectedOrder {
				Expect(respData.Services.Edges[i].Node.ID).To(BeEquivalentTo(id))
			}
		}

		It("can order descending by severity", func() {
			runOrderTest("desc", []string{"1", "3", "4", "5", "2"})
		})

		It("can order ascending by severity", func() {
			runOrderTest("asc", []string{"2", "5", "4", "3", "1"})
		})

		It("can count issues", Label("issueCount"), func() {

		})
	})
})

var _ = Describe("Creating Service via API", Label("e2e", "Services"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var service entity.Service
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {

		BeforeEach(func() {
			seeder.SeedDbWithNFakeData(10)
			service = testentity.NewFakeServiceEntity()
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new service", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"ccrn": service.CCRN,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Service model.Service `json:"createService"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Service.Ccrn).To(Equal(service.CCRN))
			})
		})
	})
})

var _ = Describe("Updating service via API", Label("e2e", "Services"), func() {

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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates service", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				service := seedCollection.ServiceRows[0].AsService()
				service.CCRN = "SecretService"

				req.Var("id", fmt.Sprintf("%d", service.Id))
				req.Var("input", map[string]string{
					"ccrn": service.CCRN,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Service model.Service `json:"updateService"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(*respData.Service.Ccrn).To(Equal(service.CCRN))
			})
		})
	})
})

var _ = Describe("Deleting Service via API", Label("e2e", "Services"), func() {

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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes service", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.ServiceRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteService"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})

var _ = Describe("Modifying Owner of Service via API", Label("e2e", "Services"), func() {

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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", func() {
			It("adds owner to service", Label("addOwner.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/addOwner.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				service := seedCollection.ServiceRows[0].AsService()
				ownerIds := lo.FilterMap(seedCollection.OwnerRows, func(row mariadb.OwnerRow, _ int) (int64, bool) {
					if row.ServiceId.Int64 == service.Id {
						return row.UserId.Int64, true
					}
					return 0, false
				})

				ownerRow, _ := lo.Find(seedCollection.UserRows, func(row mariadb.UserRow) bool {
					return !lo.Contains(ownerIds, row.Id.Int64)
				})

				req.Var("serviceId", fmt.Sprintf("%d", service.Id))
				req.Var("userId", fmt.Sprintf("%d", ownerRow.Id.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Service model.Service `json:"addOwnerToService"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Service.Owners.Edges, func(edge *model.UserEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", ownerRow.Id.Int64)
				})

				Expect(respData.Service.ID).To(Equal(fmt.Sprintf("%d", service.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes owner from service", Label("removeOwner.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/removeOwner.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				service := seedCollection.ServiceRows[0].AsService()

				ownerRow, _ := lo.Find(seedCollection.OwnerRows, func(row mariadb.OwnerRow) bool {
					return row.ServiceId.Int64 == service.Id
				})

				req.Var("serviceId", fmt.Sprintf("%d", service.Id))
				req.Var("userId", fmt.Sprintf("%d", ownerRow.UserId.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Service model.Service `json:"removeOwnerFromService"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Service.Owners.Edges, func(edge *model.UserEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", ownerRow.UserId.Int64)
				})

				Expect(respData.Service.ID).To(Equal(fmt.Sprintf("%d", service.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})

var _ = Describe("Modifying IssueRepository of Service via API", Label("e2e", "Services"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var priority int64
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
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
			priority = 1
		})

		Context("and a mutation query is performed", func() {
			It("adds issueRepository to service", Label("addIssueRepository.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/addIssueRepository.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				service := seedCollection.ServiceRows[0].AsService()
				// find all issueRepositories that are attached to the service
				issueRepositoryIds := lo.FilterMap(seedCollection.IssueRepositoryServiceRows, func(row mariadb.IssueRepositoryServiceRow, _ int) (int64, bool) {
					if row.ServiceId.Int64 == service.Id {
						return row.IssueRepositoryId.Int64, true
					}
					return 0, false
				})
				// find an issueRepository that is not attached to the service
				issueRepositoryRow, _ := lo.Find(seedCollection.IssueRepositoryRows, func(row mariadb.BaseIssueRepositoryRow) bool {
					return !lo.Contains(issueRepositoryIds, row.Id.Int64)
				})

				req.Var("serviceId", fmt.Sprintf("%d", service.Id))
				req.Var("issueRepositoryId", fmt.Sprintf("%d", issueRepositoryRow.Id.Int64))
				req.Var("priority", fmt.Sprintf("%d", priority))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Service model.Service `json:"addIssueRepositoryToService"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Service.IssueRepositories.Edges, func(edge *model.IssueRepositoryEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", issueRepositoryRow.Id.Int64)
				})

				Expect(respData.Service.ID).To(Equal(fmt.Sprintf("%d", service.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes issueRepository from service", Label("removeIssueRepository.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/removeIssueRepository.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				service := seedCollection.ServiceRows[0].AsService()

				// find an issueRepository that is attached to the service
				issueRepositoryRow, _ := lo.Find(seedCollection.IssueRepositoryServiceRows, func(row mariadb.IssueRepositoryServiceRow) bool {
					return row.ServiceId.Int64 == service.Id
				})

				req.Var("serviceId", fmt.Sprintf("%d", service.Id))
				req.Var("issueRepositoryId", fmt.Sprintf("%d", issueRepositoryRow.IssueRepositoryId.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Service model.Service `json:"removeIssueRepositoryFromService"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Service.IssueRepositories.Edges, func(edge *model.IssueRepositoryEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", issueRepositoryRow.IssueRepositoryId.Int64)
				})

				Expect(respData.Service.ID).To(Equal(fmt.Sprintf("%d", service.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})
