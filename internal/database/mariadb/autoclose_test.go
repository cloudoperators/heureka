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
				databaseSeeder.SeedScannerRuns("ScannerRunTag1", true, time.Now())
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})
		Context("and the database has two empty scannerruns", func() {
			It("Autoclose should return false and no error", func() {
				databaseSeeder.SeedScannerRuns("ScannerRunTag1", true, time.Now(), time.Now().Add(time.Hour))
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})
		Context("and the database has three empty scannerruns", func() {
			It("Autoclose should return false and no error", func() {
				databaseSeeder.SeedScannerRuns("ScannerRunTag1", true, time.Now(), time.Now().Add(time.Minute), time.Now().Add(time.Hour))
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})
	})
})
