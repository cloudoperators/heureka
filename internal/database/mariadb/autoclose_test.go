// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	When("Running autoclose", Label("Autoclose"), func() {
		Context("and the database is empty", func() {
			It("Autoclose should return false and no error", func() {
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})
		Context("and the database has one empty scannerrun", func() {
			It("Autoclose should return false and no error", func() {
				databaseSeeder.SeedScannerRuns(test.ScannerRunDef{
					Tag:         "ScannerRunTag1",
					IsCompleted: true,
					Timestamp:   time.Now(),
				})
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})
		Context("and the database has two empty scannerruns", func() {
			It("Autoclose should return false and no error", func() {
				databaseSeeder.SeedScannerRuns(test.ScannerRunDef{
					Tag:         "ScannerRunTag1",
					IsCompleted: true,
					Timestamp:   time.Now(),
				}, test.ScannerRunDef{
					Tag:         "ScannerRunTag1",
					IsCompleted: true,
					Timestamp:   time.Now().Add(time.Minute),
				})
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})
		Context("and the database has three empty scannerruns", func() {
			It("Autoclose should return false and no error", func() {
				databaseSeeder.SeedScannerRuns(test.ScannerRunDef{
					Tag:         "ScannerRunTag1",
					IsCompleted: true,
					Timestamp:   time.Now(),
				}, test.ScannerRunDef{
					Tag:         "ScannerRunTag1",
					IsCompleted: true,
					Timestamp:   time.Now().Add(time.Minute),
				}, test.ScannerRunDef{
					Tag:         "ScannerRunTag1",
					IsCompleted: true,
					Timestamp:   time.Now().Add(time.Hour),
				})
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})

		Context("and the database has one scannerrun with one issue", func() {
			It("Autoclose should return false and no error", func() {
				databaseSeeder.SeedScannerRuns(test.ScannerRunDef{
					Tag:         "ScannerRunTag1",
					IsCompleted: true,
					Timestamp:   time.Now(),
					Issues:      []string{"Issue1"},
				})
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})

		Context("and the database has two scannerruns where the first run found an issue and the second one does not", func() {
			It("Autoclose should return true and no error", func() {
				err := databaseSeeder.SeedScannerRuns(
					test.ScannerRunDef{
						Tag:         "ScannerRunTag1",
						IsCompleted: true,
						Timestamp:   time.Now(),
						Issues:      []string{"Issue1"},
					},
					test.ScannerRunDef{
						Tag:         "ScannerRunTag1",
						IsCompleted: true,
						Timestamp:   time.Now(),
					})
				Expect(err).To(BeNil())
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())
			})
		})
	})
})
