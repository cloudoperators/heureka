// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"

	"github.wdf.sap.corp/cc/heureka/internal/entity"
	testentity "github.wdf.sap.corp/cc/heureka/internal/entity/test"
	"github.wdf.sap.corp/cc/heureka/internal/util"
	util2 "github.wdf.sap.corp/cc/heureka/pkg/util"

	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/model"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb/test"
	"github.wdf.sap.corp/cc/heureka/internal/server"
)

var _ = Describe("Getting Activities via API", Label("e2e", "Activity"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)
		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database is empty", func() {
		It("returns empty resultset", func() {
			// create a queryCollection (safe to share across requests)
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/withPageInfo.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "0")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				Activities model.ActivityConnection `json:"Activities"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.Activities.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {

		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context(", no additional filters are present", func() {
			Context("and  a query including PageInfo is performed", Label("withPageInfo.graphql"), func() {
				It("returns correct result count", func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/withPageInfo.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 5)
					req.Var("after", "0")

					req.Header.Set("Cache-Control", "no-cache")
					ctx := context.Background()

					var respData struct {
						Activities model.ActivityConnection `json:"Activities"`
					}
					if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
						logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
					}

					Expect(respData.Activities.TotalCount).To(Equal(len(seedCollection.ActivityRows)))
					Expect(len(respData.Activities.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve direct relations", Label("directRelations.graphql"), func() {

				var respData struct {
					Activities model.ActivityConnection `json:"Activities"`
				}
				BeforeEach(func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/directRelations.graphql")

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
					Expect(respData.Activities.TotalCount).To(Equal(len(seedCollection.ActivityRows)))
					Expect(len(respData.Activities.Edges)).To(Equal(5))
				})

				It("- returns the direct relations", func() {
					//this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage
					for _, activity := range respData.Activities.Edges {
						Expect(activity.Node.ID).ToNot(BeEmpty())
						for _, issue := range activity.Node.Issues.Edges {
							Expect(issue.Node.ID).ToNot(BeEmpty())
						}
						for _, s := range activity.Node.Services.Edges {
							Expect(s.Node.ID).ToNot(BeEmpty())
							Expect(*s.Node.Name).ToNot(BeEmpty())
						}
						for _, e := range activity.Node.Evidences.Edges {
							Expect(e.Node.ID).ToNot(BeEmpty())
						}
						for _, imc := range activity.Node.IssueMatchChanges.Edges {
							Expect(imc.Node.ID).ToNot(BeNil(), "issueMatchChange has a ID set")
							Expect(imc.Node.Action).ToNot(BeNil(), "issueMatchChange has an action set")
							Expect(imc.Node.ActivityID).ToNot(BeNil(), "issueMatchChange has an activityId set")
							Expect(imc.Node.IssueMatchID).ToNot(BeNil(), "issueMatchChange has a issueMatchId set")
						}
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.Activities.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.Activities.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.Activities.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.Activities.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
					Expect(*respData.Activities.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
		})
	})
})

var _ = Describe("Creating Activity via API", Label("e2e", "Activities"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var activity entity.Activity

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database has 10 entries", func() {

		BeforeEach(func() {
			seeder.SeedDbWithNFakeData(10)
			activity = testentity.NewFakeActivityEntity()
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new activity", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"status": activity.Status.String(),
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Activity model.Activity `json:"createActivity"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Activity.ID).To(Not(BeEmpty()))
				Expect(respData.Activity.Status.String()).To(Equal(activity.Status.String()))
			})
		})
	})
})

var _ = Describe("Updating activity via API", Label("e2e", "Activities"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("update.graphql"), func() {
			It("updates activity", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/update.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				activity := seedCollection.ActivityRows[0].AsActivity()

				req.Var("id", fmt.Sprintf("%d", activity.Id))
				req.Var("input", map[string]string{
					"status": activity.Status.String(),
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Activity model.Activity `json:"updateActivity"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Activity.Status.String()).To(Equal(activity.Status.String()))
			})
		})
	})
})

var _ = Describe("Deleting Activity via API", Label("e2e", "Activities"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", Label("delete.graphql"), func() {
			It("deletes activity", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/delete.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				id := fmt.Sprintf("%d", seedCollection.ActivityRows[0].Id.Int64)

				req.Var("id", id)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Id string `json:"deleteActivity"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Id).To(Equal(id))
			})
		})
	})
})

var _ = Describe("Modifying Services of Activity via API", Label("e2e", "ServiceActivity"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", func() {
			It("adds service to activity", Label("addService.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/addService.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				activity := seedCollection.ActivityRows[0].AsActivity()
				// find all services that are assigned to the activity
				serviceIds := lo.FilterMap(seedCollection.ActivityHasServiceRows, func(row mariadb.ActivityHasServiceRow, _ int) (int64, bool) {
					if row.ActivityId.Int64 == activity.Id {
						return row.ServiceId.Int64, true
					}
					return 0, false
				})

				// find a service that is not assigned to the activity
				serviceRow, _ := lo.Find(seedCollection.ServiceRows, func(row mariadb.BaseServiceRow) bool {
					return !lo.Contains(serviceIds, row.Id.Int64)
				})

				req.Var("activityId", fmt.Sprintf("%d", activity.Id))
				req.Var("serviceId", fmt.Sprintf("%d", serviceRow.Id.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Activity model.Activity `json:"addServiceToActivity"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Activity.Services.Edges, func(edge *model.ServiceEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", serviceRow.Id.Int64)
				})

				Expect(respData.Activity.ID).To(Equal(fmt.Sprintf("%d", activity.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes service from activity", Label("removeService.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/removeService.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				activity := seedCollection.ActivityRows[0].AsActivity()

				// find a service that is assigned to the activity
				serviceRow, _ := lo.Find(seedCollection.ActivityHasServiceRows, func(row mariadb.ActivityHasServiceRow) bool {
					return row.ActivityId.Int64 == activity.Id
				})

				req.Var("activityId", fmt.Sprintf("%d", activity.Id))
				req.Var("serviceId", fmt.Sprintf("%d", serviceRow.ServiceId.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Activity model.Activity `json:"removeServiceFromActivity"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Activity.Services.Edges, func(edge *model.ServiceEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", serviceRow.ServiceId.Int64)
				})

				Expect(respData.Activity.ID).To(Equal(fmt.Sprintf("%d", activity.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})

var _ = Describe("Modifying Issues of Activity via API", Label("e2e", "ServiceIssue"), func() {

	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		Context("and a mutation query is performed", func() {
			It("adds issue to activity", Label("addIssue.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/addIssue.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				activity := seedCollection.ActivityRows[0].AsActivity()
				issueIds := lo.FilterMap(seedCollection.ActivityHasIssueRows, func(row mariadb.ActivityHasIssueRow, _ int) (int64, bool) {
					if row.ActivityId.Int64 == activity.Id {
						return row.IssueId.Int64, true
					}
					return 0, false
				})

				issueRow, _ := lo.Find(seedCollection.IssueRows, func(row mariadb.IssueRow) bool {
					return !lo.Contains(issueIds, row.Id.Int64)
				})

				req.Var("activityId", fmt.Sprintf("%d", activity.Id))
				req.Var("issueId", fmt.Sprintf("%d", issueRow.Id.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Activity model.Activity `json:"addIssueToActivity"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Activity.Issues.Edges, func(edge *model.IssueEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", issueRow.Id.Int64)
				})

				Expect(respData.Activity.ID).To(Equal(fmt.Sprintf("%d", activity.Id)))
				Expect(found).To(BeTrue())
			})
			It("removes issue from activity", Label("removeIssue.graphql"), func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/activity/removeIssue.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				activity := seedCollection.ActivityRows[0].AsActivity()

				issueRow, _ := lo.Find(seedCollection.ActivityHasIssueRows, func(row mariadb.ActivityHasIssueRow) bool {
					return row.ActivityId.Int64 == activity.Id
				})

				req.Var("activityId", fmt.Sprintf("%d", activity.Id))
				req.Var("issueId", fmt.Sprintf("%d", issueRow.IssueId.Int64))

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Activity model.Activity `json:"removeIssueFromActivity"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				_, found := lo.Find(respData.Activity.Issues.Edges, func(edge *model.IssueEdge) bool {
					return edge.Node.ID == fmt.Sprintf("%d", issueRow.ActivityId.Int64)
				})

				Expect(respData.Activity.ID).To(Equal(fmt.Sprintf("%d", activity.Id)))
				Expect(found).To(BeFalse())
			})
		})
	})
})
