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
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var loadTestData = func() ([]mariadb.IssueVariantRow, []mariadb.IssueMatchRow, error) {
	issueVariants, err := test.LoadIssueVariants(test.GetTestDataPath("../database/mariadb/testdata/component_version_order/issue_variant.json"))
	if err != nil {
		return nil, nil, err
	}
	issueMatches, err := test.LoadIssueMatches(test.GetTestDataPath("../database/mariadb/testdata/issue_counts/issue_matches.json"))
	if err != nil {
		return nil, nil, err
	}
	return issueVariants, issueMatches, nil
}

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

		var supportGroups []mariadb.SupportGroupRow
		BeforeEach(func() {
			supportGroups = seeder.SeedSupportGroups(1)
			services := seeder.SeedServices(3)
			seeder.SeedIssueRepositories()
			seeder.SeedIssues(10)
			components := seeder.SeedComponents(1)
			componentVersions := seeder.SeedComponentVersions(10, components)
			seeder.SeedComponentInstances(10, componentVersions, services)
			issueVariants, issueMatches, err := loadTestData()
			Expect(err).To(BeNil())
			// Important: the order need to be preserved
			for _, iv := range issueVariants {
				_, err := seeder.InsertFakeIssueVariant(iv)
				Expect(err).To(BeNil())
			}
			for _, im := range issueMatches {
				_, err := seeder.InsertFakeIssueMatch(im)
				Expect(err).To(BeNil())
			}
			for _, s := range services {
				sgs := mariadb.SupportGroupServiceRow{
					SupportGroupId: supportGroups[0].Id,
					ServiceId:      s.Id,
				}
				_, err := seeder.InsertFakeSupportGroupService(sgs)
				Expect(err).To(BeNil())
			}
		})
		Context("and a filter is used", func() {
			It("correct filters by support group", func() {
				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/issueCounts/query.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)
				req.Var("filter", map[string]string{
					"supportGroupCcrn": supportGroups[0].CCRN.String,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					IssueCounts model.SeverityCounts `json:"IssueCounts"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.IssueCounts.Critical).To(Equal(2))
				Expect(respData.IssueCounts.High).To(Equal(2))
				Expect(respData.IssueCounts.Medium).To(Equal(2))
				Expect(respData.IssueCounts.Low).To(Equal(2))
				Expect(respData.IssueCounts.None).To(Equal(2))
			})
		})
	})
})
