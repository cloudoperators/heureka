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
				res, err := db.CreateScannerRun(sr)
				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())
			})
		})
	})

	When("Creating a new ScannerRun and marking it as complete", Label("Update"), func() {
		Context("and the database is empty", func() {
			It("should be marked as completed correctly", func() {
				sr := &entity.ScannerRun{
					UUID: "6809de35-9716-4914-b090-15273f82e8ab",
					Tag:  "tag",
				}
				res, err := db.CreateScannerRun(sr)

				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				success, err := db.CompleteScannerRun(sr.UUID)

				Expect(err).To(BeNil())
				Expect(success).To(BeTrue())
			})
		})
	})

	When("Creating a new ScannerRun and retrieving it by UUID should work", Label("ByUUID"), func() {
		Context("and the database is empty", func() {
			It("should be marked as completed correctly", func() {
				sr := &entity.ScannerRun{
					UUID: "6809de35-9716-4914-b090-15273f82e8ab",
					Tag:  "tag",
				}
				res, err := db.CreateScannerRun(sr)

				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				newsr, err := db.ScannerRunByUUID(sr.UUID)

				Expect(err).To(BeNil())
				Expect(newsr.Tag).To(Equal(sr.Tag))
			})
		})
	})
})
