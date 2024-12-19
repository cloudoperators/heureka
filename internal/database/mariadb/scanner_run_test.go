// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("ScannerRun", Label("database", "ScannerRun"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Creating a new ScannerRun", Label("Create"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllUserIds(nil)
				res = e2e_common.SubtractSystemUserId(res)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list of non-system users", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Users in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, u := range seedCollection.UserRows {
					ids = append(ids, u.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllUserIds(nil)
					res = e2e_common.SubtractSystemUserId(res)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.UserRows)))
					})

					By("returning the correct order", func() {
						var prev int64 = 0
						for _, r := range res {

							Expect(r > prev).Should(BeTrue())
							prev = r

						}
					})

					By("returning the correct fields", func() {
						for _, r := range res {
							Expect(lo.Contains(ids, r)).To(BeTrue())
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single user id that does exist", func() {
					uId := ids[rand.Intn(len(ids))]
					filter := &entity.UserFilter{
						Id: []*int64{&uId},
					}

					entries, err := db.GetAllUserIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(uId))
					})
				})
			})
		})
	})
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
