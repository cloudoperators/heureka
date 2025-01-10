// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ScannerRun", Label("database", "ScannerRun"), func() {
	var db *mariadb.SqlDatabase
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Creating a new ScannerRun", Label("Create"), func() {
		Context("and the database is empty", func() {
			It("should be initialized correctly", func() {
				sr := &entity.ScannerRun{
					UUID: "6809de35-9716-4914-b090-15273f82e8ab",
					Tag:  "tag",
				}
				_, err := db.CreateScannerRun(sr)
				Expect(err).To(BeNil())
				Expect(sr.RunID).To(BeNumerically(">=", 0))
				Expect(sr.IsCompleted()).To(BeFalse())
			})
		})
	})

	When("Creating a new ScannerRun and marking it as complete", Label("Create"), func() {
		Context("and the database is empty", func() {
			It("should be marked as completed correctly", func() {
				sr := &entity.ScannerRun{
					UUID: "6809de35-9716-4914-b090-15273f82e8ab",
					Tag:  "tag",
				}
				_, err := db.CreateScannerRun(sr)
				Expect(err).To(BeNil())

				err = db.CompleteScannerRun(sr)

				Expect(err).To(BeNil())
				Expect(sr.IsCompleted()).To(BeTrue())
				Expect(sr.EndRun.After(sr.StartRun)).To(BeTrue())
			})
		})
	})
})
