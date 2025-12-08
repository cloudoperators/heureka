// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Autoclose", Label("database", "Autoclose"), func() {
	var db *mariadb.SqlDatabase
	var databaseSeeder *test.DatabaseSeeder

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		databaseSeeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	type scannerRun struct {
		isNotCompleted       bool
		issues               []string
		components           []string
		issueMatchComponents []string
	}

	var autocloseTests = []struct {
		description    string
		scannerRuns    []scannerRun
		expectedResult bool
	}{
		{
			"WHEN db has no scans",
			[]scannerRun{},
			false,
		},
		{
			"WHEN database has one not completed scan",
			[]scannerRun{
				scannerRun{isNotCompleted: true},
			},
			false,
		},
		{
			"WHEN database has one completed empty scan",
			[]scannerRun{
				scannerRun{},
			},
			false,
		},
		{
			"WHEN database has one completed scan with one issue",
			[]scannerRun{
				scannerRun{issues: []string{"Issue1"}},
			},
			false,
		},
		{
			"WHEN database has two completed empty scans",
			[]scannerRun{
				scannerRun{},
				scannerRun{},
			},
			false,
		},
		{
			"WHEN database has one completed scan with an issue and the second one not-completed without the issue",
			[]scannerRun{
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{isNotCompleted: true},
			},
			false,
		},
		{
			"WHEN database has two completed scans where the first run found an issue and the second one has no longer the issue from previous run",
			[]scannerRun{
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{},
			},
			true,
		},
		{
			"WHEN database has two completed scans where the first run found an issue and the second one has the same issue",
			[]scannerRun{
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{issues: []string{"Issue1"}},
			},
			false,
		},
		{
			"WHEN database has two completed scans where the first run found an issue and the second one has different issue",
			[]scannerRun{
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{issues: []string{"Issue2"}},
			},
			true,
		},
		{
			"WHEN database has two completed scans where the first run found a componentInstance issue and the second one has no longer the issue from previous run",
			[]scannerRun{
				scannerRun{
					issues:               []string{"Issue1"},
					components:           []string{"Component1"},
					issueMatchComponents: []string{"Issue1", "Component1"}},
				scannerRun{components: []string{"Component2"}},
			},
			true,
		},
		{
			"WHEN database has two completed scans where the first run found a componentInstance issue and the second one has the same issue",
			[]scannerRun{
				scannerRun{
					issues:               []string{"Issue1"},
					components:           []string{"Component1"},
					issueMatchComponents: []string{"Issue1", "Component1"}},
				scannerRun{
					issues:               []string{"Issue1"},
					components:           []string{"Component1"},
					issueMatchComponents: []string{"Issue1", "Component1"}},
			},
			false,
		},
		{
			"WHEN database has three completed empty scans",
			[]scannerRun{
				scannerRun{},
				scannerRun{},
				scannerRun{},
			},
			false,
		},
		{
			"WHEN 3 scans: <issue>, <no issue>, <no issue>",
			[]scannerRun{
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{},
				scannerRun{},
			},
			false,
		},
		{
			"WHEN 3 scans: <issue>, <issue>, <issue>",
			[]scannerRun{
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{issues: []string{"Issue1"}},
			},
			false,
		},
		{
			"WHEN 3 scans: <issue>, <no issue>, <issue>",
			[]scannerRun{
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{},
				scannerRun{issues: []string{"Issue1"}},
			},
			false,
		},
		{
			"WHEN 3 scans: <no issue>, <issue>, <no issue>",
			[]scannerRun{
				scannerRun{},
				scannerRun{issues: []string{"Issue1"}},
				scannerRun{},
			},
			true,
		},
	}

	When("Running autoclose", Label("Autoclose"), func() {
		Context("all scenarios DDT", func() {
			It("Returns no error and expected result per scenario", func() {
				for it, t := range autocloseTests {
					tag := fmt.Sprintf("ScannerRunTag_%d", it)
					scannerRuns := lo.Map(t.scannerRuns, func(r scannerRun, i int) test.ScannerRunDef {
						return test.ScannerRunDef{
							Tag:         tag,
							IsCompleted: !r.isNotCompleted,
							Timestamp:   time.Now().Add(time.Duration(i) * time.Minute),
							Issues:      r.issues,
						}
					})
					databaseSeeder.SeedScannerRuns(scannerRuns...)
					res, err := db.Autoclose()
					Expect(err).To(BeNil(), fmt.Sprintf("%s THEN autocolose should return no error", t.description))
					Expect(res).To(BeEquivalentTo(t.expectedResult), fmt.Sprintf("%s THEN autocolose should return %t", t.description, t.expectedResult))
					err = databaseSeeder.CleanupScannerRuns()
					Expect(err).To(BeNil())
				}
			})
		})
	})
})
