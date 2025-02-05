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
	When("Creating a new ScannerRun", Label("ByUUID"), func() {
		Context("and the database is empty", func() {
			It("should be marked as failed correctly", func() {
				sr := &entity.ScannerRun{
					UUID: "6809de35-9716-4914-b090-15273f82e8ab",
					Tag:  "tag",
				}
				res, err := db.CreateScannerRun(sr)

				Expect(err).To(BeNil())
				Expect(res).To(BeTrue())

				success, err := db.FailScannerRun(sr.UUID, "All your base are belong to us")

				Expect(err).To(BeNil())
				Expect(success).To(BeTrue())
			})
		})
	})

	When("No ScannerRun Was Created", Label("None"), func() {
		Context("and the database is empty", func() {
			It("GetScannerRuns should return an empty list", func() {
				res, err := db.GetScannerRuns(nil)

				Expect(err).To(BeNil())
				Expect(len(res)).To(Equal(0))
			})
		})
	})

	When("One ScannerRun was created", Label("None"), func() {
		Context("and the database is empty", func() {
			It("GetScannerRuns should return one ScannerRun", func() {
				{
					sr := &entity.ScannerRun{
						UUID: "6809de35-9716-4914-b090-15273f82e8ab",
						Tag:  "tag",
					}
					res, err := db.CreateScannerRun(sr)

					Expect(err).To(BeNil())
					Expect(res).To(BeTrue())
				}

				res, err := db.GetScannerRuns(nil)

				Expect(err).To(BeNil())
				Expect(len(res)).To(Equal(1))
			})
		})
	})

	When("Two ScannerRuns where created", Label("None"), func() {
		Context("and the database is empty", func() {
			It("GetScannerRuns should return one ScannerRun", func() {
				{
					sr := &entity.ScannerRun{
						UUID: "6809de35-9716-4914-b090-15273f82e8ab",
						Tag:  "tag",
					}
					res, err := db.CreateScannerRun(sr)

					Expect(err).To(BeNil())
					Expect(res).To(BeTrue())
				}

				{
					sr := &entity.ScannerRun{
						UUID: "0af596d5-091c-4446-92aa-741f63f13dda",
						Tag:  "otherTag",
					}
					res, err := db.CreateScannerRun(sr)

					Expect(err).To(BeNil())
					Expect(res).To(BeTrue())
				}

				res, err := db.GetScannerRuns(nil)

				Expect(err).To(BeNil())
				Expect(len(res)).To(Equal(2))
			})
		})
	})

	When("Two ScannerRuns where created", Label("None"), func() {
		Context("and the database is empty", func() {
			It("GetScannerRuns should find one ScannerRun by tag", func() {
				{
					sr := &entity.ScannerRun{
						UUID: "6809de35-9716-4914-b090-15273f82e8ab",
						Tag:  "tag",
					}
					res, err := db.CreateScannerRun(sr)

					Expect(err).To(BeNil())
					Expect(res).To(BeTrue())
				}

				{
					sr := &entity.ScannerRun{
						UUID: "0af596d5-091c-4446-92aa-741f63f13dda",
						Tag:  "otherTag",
					}
					res, err := db.CreateScannerRun(sr)

					Expect(err).To(BeNil())
					Expect(res).To(BeTrue())
				}

				res, err := db.GetScannerRuns(&entity.ScannerRunFilter{
					Tag: []string{"tag"},
				})

				Expect(err).To(BeNil())
				Expect(len(res)).To(Equal(1))
			})
		})
	})
})
