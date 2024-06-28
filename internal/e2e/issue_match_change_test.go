// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"

	"github.wdf.sap.corp/cc/heureka/internal/util"
	util2 "github.wdf.sap.corp/cc/heureka/pkg/util"

	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/model"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb/test"
	"github.wdf.sap.corp/cc/heureka/internal/server"
)

var _ = Describe("Getting IssueMatchChanges via API", Label("e2e", "IssueMatchChanges"), func() {

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

			//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/minimal.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Var("filter", map[string]string{})
			req.Var("first", 10)
			req.Var("after", "0")

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				IssueMatchChanges model.IssueMatchChangeConnection `json:"IssueMatchChanges"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.IssueMatchChanges.TotalCount).To(Equal(0))
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

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/minimal.graphql")

					Expect(err).To(BeNil())
					str := string(b)
					req := graphql.NewRequest(str)

					req.Var("filter", map[string]string{})
					req.Var("first", 5)
					req.Var("after", "0")

					req.Header.Set("Cache-Control", "no-cache")
					ctx := context.Background()

					var respData struct {
						IssueMatchChanges model.IssueMatchChangeConnection `json:"IssueMatchChanges"`
					}
					if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
						logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
					}

					Expect(respData.IssueMatchChanges.TotalCount).To(Equal(len(seedCollection.IssueMatchChangeRows)))
					Expect(len(respData.IssueMatchChanges.Edges)).To(Equal(5))
				})
			})
			Context("and  we query to resolve levels of relations", Label("directRelations.graphql"), func() {

				var respData struct {
					IssueMatchChanges model.IssueMatchChangeConnection `json:"IssueMatchChanges"`
				}
				BeforeEach(func() {
					// create a queryCollection (safe to share across requests)
					client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

					//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
					b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueMatchChange/directRelations.graphql")

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
					Expect(respData.IssueMatchChanges.TotalCount).To(Equal(len(seedCollection.IssueMatchChangeRows)))
					Expect(len(respData.IssueMatchChanges.Edges)).To(Equal(5))
				})

				It("- returns the expected content", func() {
					//this just checks partial attributes to check whatever every sub-relation does resolve some reasonable data and is not doing
					// a complete verification
					// additional checks are added based on bugs discovered during usage

					for _, imc := range respData.IssueMatchChanges.Edges {
						Expect(imc.Node.ID).ToNot(BeNil(), "issueMatchChange has a ID set")
						Expect(imc.Node.Action).ToNot(BeNil(), "issueMatchChange has an action set")
						Expect(imc.Node.ActivityID).ToNot(BeNil(), "issueMatchChange has an activityId set")
						Expect(imc.Node.IssueMatchID).ToNot(BeNil(), "issueMatchChange has a issueMatchId set")

						activity := imc.Node.Activity
						Expect(activity.ID).ToNot(BeNil(), "activity has a ID set")

						im := imc.Node.IssueMatch
						Expect(im.ID).ToNot(BeNil(), "issueMatch has a ID set")
						Expect(im.Status).ToNot(BeNil(), "issueMatch has a status set")
						Expect(im.RemediationDate).ToNot(BeNil(), "issueMatch has a remediationDate set")
						Expect(im.DiscoveryDate).ToNot(BeNil(), "issueMatch has a discoveryDate set")
						Expect(im.TargetRemediationDate).ToNot(BeNil(), "issueMatch has a targetRemediationDate set")
						Expect(im.IssueID).ToNot(BeNil(), "issueMatch has a issueId set")
						Expect(im.ComponentInstanceID).ToNot(BeNil(), "issueMatch has a componentInstanceId set")
					}
				})
				It("- returns the expected PageInfo", func() {
					Expect(*respData.IssueMatchChanges.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
					Expect(*respData.IssueMatchChanges.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
					Expect(respData.IssueMatchChanges.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
					Expect(len(respData.IssueMatchChanges.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
					Expect(*respData.IssueMatchChanges.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				})
			})
		})
	})
})
