// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Getting IssueCounts via API", Label("e2e", "IssueCounts"), func() {
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

	When("the database has entries", func() {

		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			var err error
			seedCollection, err = seeder.SeedForIssueCounts()
			Expect(err).To(BeNil(), "Seeding should work")
		})
		Context("and a filter is used", func() {
			It("correct filters by support group", func() {
				severityCounts, err := test.LoadSupportGroupIssueCounts(test.GetTestDataPath("../database/mariadb/testdata/issue_counts/issue_counts_per_support_group.json"))
				Expect(err).To(BeNil())
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueCounts/query.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)
				sg := seedCollection.SupportGroupRows[0]
				req.Var("filter", map[string]string{
					"supportGroupCcrn": sg.CCRN.String,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueCounts model.SeverityCounts `json:"IssueCounts"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				strId := fmt.Sprintf("%d", sg.Id.Int64)

				Expect(int64(respData.IssueCounts.Critical)).To(Equal(severityCounts[strId].Critical))
				Expect(int64(respData.IssueCounts.High)).To(Equal(severityCounts[strId].High))
				Expect(int64(respData.IssueCounts.Medium)).To(Equal(severityCounts[strId].Medium))
				Expect(int64(respData.IssueCounts.Low)).To(Equal(severityCounts[strId].Low))
				Expect(int64(respData.IssueCounts.None)).To(Equal(severityCounts[strId].None))
				Expect(int64(respData.IssueCounts.Total)).To(Equal(severityCounts[strId].Total))
			})
			It("it can filter by service in services query", func() {
				severityCounts, err := test.LoadServiceIssueCounts(test.GetTestDataPath("../database/mariadb/testdata/issue_counts/issue_counts_per_service.json"))

				Expect(err).To(BeNil())

				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/withIssueCounts.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Services model.ServiceConnection `json:"Services"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				for _, sEdge := range respData.Services.Edges {
					sc := severityCounts[sEdge.Node.ID]
					Expect(int64(sEdge.Node.IssueCounts.Critical)).To(Equal(sc.Critical), "Critical count is correct")
					Expect(int64(sEdge.Node.IssueCounts.High)).To(Equal(sc.High), "High count is correct")
					Expect(int64(sEdge.Node.IssueCounts.Medium)).To(Equal(sc.Medium), "Medium count is correct")
					Expect(int64(sEdge.Node.IssueCounts.Low)).To(Equal(sc.Low), "Low count is correct")
					Expect(int64(sEdge.Node.IssueCounts.None)).To(Equal(sc.None), "None count is correct")
					Expect(int64(sEdge.Node.IssueCounts.Total)).To(Equal(sc.Total), "Total count is correct")
				}

			})
		})
	})
})
