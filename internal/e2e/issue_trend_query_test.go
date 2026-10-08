// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"time"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getting IssueTrends via API", Label("e2e", "IssueTrends"), func() {
	var (
		seeder *test.DatabaseSeeder
		s      *server.Server
		cfg    util.Config
		db     *mariadb.SqlDatabase
	)

	BeforeEach(func() {
		var err error

		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		cfg.AuthzOpenFgaApiUrl = ""
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	// All seed matches are on 2025-06-15; this window captures exactly that day.
	const (
		filterAfter  = "2025-06-14T00:00:00Z"
		filterBefore = "2025-06-16T00:00:00Z"
	)

	type issueTrendResponse struct {
		IssueTrends struct {
			Buckets []*model.IssueTrendBucket `json:"buckets"`
		} `json:"IssueTrends"`
	}

	queryTrendBuckets := func(after, before string, granularity string, serviceCcrns, sgCcrns []string) ([]*model.IssueTrendBucket, error) {
		filter := map[string]any{
			"after":       after,
			"before":      before,
			"granularity": granularity,
		}
		if len(serviceCcrns) > 0 {
			filter["serviceCcrn"] = serviceCcrns
		}

		if len(sgCcrns) > 0 {
			filter["supportGroupCcrn"] = sgCcrns
		}

		resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[issueTrendResponse](
			cfg.Port,
			"../api/graphql/graph/queryCollection/issueTrend/query.graphql",
			map[string]any{"filter": filter},
			nil,
		)
		if err != nil {
			return nil, err
		}

		return resp.IssueTrends.Buckets, nil
	}

	When("the database has trend seed data", func() {
		var seed *test.IssueTrendSeedData

		BeforeEach(func() {
			var err error

			seed, err = seeder.SeedForIssueTrend()
			Expect(err).To(BeNil(), "SeedForIssueTrend should succeed")
		})

		It("returns one bucket with correct totals when no filter is applied", func() {
			buckets, err := queryTrendBuckets(filterAfter, filterBefore, "DAILY", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(buckets).To(HaveLen(1), "one bucket for the seed day")

			b := buckets[0]
			Expect(b.Total).To(Equal(6), "6 issue matches total")
			Expect(b.Critical).To(Equal(3), "3 critical matches")
			Expect(b.High).To(Equal(1), "1 high match")
			Expect(b.Medium).To(Equal(1), "1 medium match")
			Expect(b.Low).To(Equal(1), "1 low match")
			Expect(b.Remediated).To(Equal(3), "only mitigated matches count as remediated")

			bucketTime, parseErr := time.Parse(time.RFC3339, b.Date)
			Expect(parseErr).ToNot(HaveOccurred())
			Expect(bucketTime.UTC().Format("2006-01-02")).To(Equal("2025-06-15"))
		})

		It("returns correct counts when filtering by serviceCcrn", func() {
			buckets, err := queryTrendBuckets(filterAfter, filterBefore, "DAILY",
				[]string{seed.Service1.CCRN.String}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(buckets).To(HaveLen(1))

			b := buckets[0]
			Expect(b.Total).To(Equal(3), "service1 has 3 matches")
			Expect(b.Critical).To(Equal(2))
			Expect(b.High).To(Equal(1))
			Expect(b.Remediated).To(Equal(2), "service1 has 2 mitigated matches")
		})

		Context("supportGroupCcrn filter", func() {
			It("returns all matches without fan-out for a group owning multiple services", func() {
				// sg1 owns service1 and service2. service2 also belongs to sg2, meaning
				// service2 has two SupportGroupService rows that sg1's filter touches.
				// Before the EXISTS fix this doubled service2's counts.
				buckets, err := queryTrendBuckets(filterAfter, filterBefore, "DAILY",
					nil, []string{seed.SupportGroup1.CCRN.String})
				Expect(err).ToNot(HaveOccurred())
				Expect(buckets).To(HaveLen(1))

				b := buckets[0]
				Expect(b.Total).To(Equal(6), "sg1 covers both services → 6 matches, no fan-out")
				Expect(b.Remediated).To(Equal(3))
			})

			It("returns only service2 matches when filtering by the single-service group", func() {
				buckets, err := queryTrendBuckets(filterAfter, filterBefore, "DAILY",
					nil, []string{seed.SupportGroup2.CCRN.String})
				Expect(err).ToNot(HaveOccurred())
				Expect(buckets).To(HaveLen(1))

				b := buckets[0]
				Expect(b.Total).To(Equal(3), "sg2 covers only service2 → 3 matches")
				Expect(b.Critical).To(Equal(1))
				Expect(b.Medium).To(Equal(1))
				Expect(b.Low).To(Equal(1))
				Expect(b.Remediated).To(Equal(1), "only the mitigated match counts")
			})
		})

		It("counts only mitigated status as remediated, not false_positive or risk_accepted", func() {
			// service2 has exactly: 1 mitigated, 1 false_positive, 1 risk_accepted.
			buckets, err := queryTrendBuckets(filterAfter, filterBefore, "DAILY",
				[]string{seed.Service2.CCRN.String}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(buckets).To(HaveLen(1))

			b := buckets[0]
			Expect(b.Total).To(Equal(3))
			Expect(b.Remediated).To(Equal(1), "false_positive and risk_accepted must not count as remediated")
		})

		It("returns no buckets when the time window is outside all matches", func() {
			buckets, err := queryTrendBuckets("2020-01-01T00:00:00Z", "2020-01-02T00:00:00Z",
				"DAILY", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(buckets).To(BeEmpty())
		})
	})
})
