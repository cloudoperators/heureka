// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"

	"github.com/cloudoperators/heureka/internal/entity"
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

	When("the database has 100 entries", func() {

		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(100)
		})
		Context("and a filter is used", func() {
			It("correct filters by support group", func() {
				sg := seedCollection.SupportGroupRows[rand.Intn(len(seedCollection.SupportGroupRows))]
				serviceIds := lo.FilterMap(seedCollection.SupportGroupServiceRows, func(sgs mariadb.SupportGroupServiceRow, _ int) (int64, bool) {
					return sgs.ServiceId.Int64, sg.Id.Int64 == sgs.SupportGroupId.Int64
				})

				ciIds := lo.FilterMap(seedCollection.ComponentInstanceRows, func(c mariadb.ComponentInstanceRow, _ int) (int64, bool) {
					return c.Id.Int64, lo.Contains(serviceIds, c.ServiceId.Int64)
				})

				issueIds := lo.FilterMap(seedCollection.IssueMatchRows, func(im mariadb.IssueMatchRow, _ int) (int64, bool) {
					return im.IssueId.Int64, lo.Contains(ciIds, im.ComponentInstanceId.Int64)
				})

				counts := model.SeverityCounts{}

				// avoid counting duplicates
				ratingIssueIds := map[string]bool{}
				for _, iv := range seedCollection.IssueVariantRows {
					key := fmt.Sprintf("%d-%s", iv.IssueId.Int64, iv.Rating.String)
					if _, ok := ratingIssueIds[key]; ok || !iv.Id.Valid {
						continue
					}
					if lo.Contains(issueIds, iv.IssueId.Int64) {
						switch iv.Rating.String {
						case entity.SeverityValuesCritical.String():
							counts.Critical++
						case entity.SeverityValuesHigh.String():
							counts.High++
						case entity.SeverityValuesMedium.String():
							counts.Medium++
						case entity.SeverityValuesLow.String():
							counts.Low++
						case entity.SeverityValuesNone.String():
							counts.None++
						}
					}
					ratingIssueIds[key] = true
				}

				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/service/withIssueCounts.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)
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

				Expect(respData.IssueCounts.Critical).To(Equal(counts.Critical))
				Expect(respData.IssueCounts.High).To(Equal(counts.High))
				Expect(respData.IssueCounts.Medium).To(Equal(counts.Medium))
				Expect(respData.IssueCounts.Low).To(Equal(counts.Low))
				Expect(respData.IssueCounts.None).To(Equal(counts.None))
			})
		})
	})
})
