// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"context"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Service Batch", Label("database", "ServiceBatch"), func() {
	var db *mariadb.SqlDatabase

	var seeder *test.DatabaseSeeder

	BeforeEach(func() {
		var err error

		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	AfterEach(func() {
		dbm.TestTearDown(db)
	})

	When("Getting Owners By Service IDs", Label("GetOwnersByServiceIDs"), func() {
		Context("and the database is empty", func() {
			It("returns empty map for empty input", func() {
				res, err := db.GetOwnersByServiceIDs(context.Background(), []int64{})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty map", func() {
					Expect(res).To(BeEmpty())
				})
			})

			It("returns empty map for non-existent IDs", func() {
				res, err := db.GetOwnersByServiceIDs(context.Background(), []int64{99999})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty map", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})

		Context("and we have seeded data", func() {
			var seedCollection *test.SeedCollection

			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			It("returns owners grouped by service ID correctly", func() {
				serviceIDs := lo.Map(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, _ int) int64 {
					return s.Id.Int64
				})

				res, err := db.GetOwnersByServiceIDs(context.Background(), serviceIDs)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning correct owner mappings", func() {
					for _, ownerRow := range seedCollection.OwnerRows {
						serviceID := ownerRow.ServiceId.Int64
						userID := ownerRow.UserId.Int64

						users, exists := res[serviceID]
						if !exists {
							continue
						}

						hasUser := false

						for _, u := range users {
							if u.Id == userID {
								hasUser = true
								break
							}
						}

						Expect(hasUser).To(BeTrue(), "Expected user %d to be owner of service %d", userID, serviceID)
					}
				})
			})

			It("returns data for a single service ID", func() {
				var targetServiceID int64
				for _, ownerRow := range seedCollection.OwnerRows {
					targetServiceID = ownerRow.ServiceId.Int64
					break
				}

				if targetServiceID == 0 {
					Skip("No owners found in seed data")
				}

				res, err := db.GetOwnersByServiceIDs(context.Background(), []int64{targetServiceID})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning owners for the requested service only", func() {
					Expect(res).To(HaveLen(1))
					Expect(res).To(HaveKey(targetServiceID))
					Expect(res[targetServiceID]).ToNot(BeEmpty())
				})
			})
		})
	})

	When("Getting Support Groups By Service IDs", Label("GetSupportGroupsByServiceIDs"), func() {
		Context("and the database is empty", func() {
			It("returns empty map for empty input", func() {
				res, err := db.GetSupportGroupsByServiceIDs(context.Background(), []int64{})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty map", func() {
					Expect(res).To(BeEmpty())
				})
			})

			It("returns empty map for non-existent IDs", func() {
				res, err := db.GetSupportGroupsByServiceIDs(context.Background(), []int64{99999})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty map", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})

		Context("and we have seeded data", func() {
			var seedCollection *test.SeedCollection

			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			It("returns support groups grouped by service ID correctly", func() {
				serviceIDs := lo.Map(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, _ int) int64 {
					return s.Id.Int64
				})

				res, err := db.GetSupportGroupsByServiceIDs(context.Background(), serviceIDs)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning correct support group mappings", func() {
					for _, sgsRow := range seedCollection.SupportGroupServiceRows {
						serviceID := sgsRow.ServiceId.Int64
						sgID := sgsRow.SupportGroupId.Int64

						sgs, exists := res[serviceID]
						if !exists {
							continue
						}

						hasSG := false

						for _, sg := range sgs {
							if sg.Id == sgID {
								hasSG = true
								break
							}
						}

						Expect(hasSG).To(BeTrue(), "Expected support group %d to be linked to service %d", sgID, serviceID)
					}
				})
			})

			It("returns data for a single service ID", func() {
				var targetServiceID int64
				for _, sgsRow := range seedCollection.SupportGroupServiceRows {
					targetServiceID = sgsRow.ServiceId.Int64
					break
				}

				if targetServiceID == 0 {
					Skip("No support group services found in seed data")
				}

				res, err := db.GetSupportGroupsByServiceIDs(context.Background(), []int64{targetServiceID})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning support groups for the requested service only", func() {
					Expect(res).To(HaveLen(1))
					Expect(res).To(HaveKey(targetServiceID))
					Expect(res[targetServiceID]).ToNot(BeEmpty())
				})
			})
		})
	})

	When("Getting Issue Counts By Service IDs", Label("GetIssueCountsByServiceIDs"), func() {
		Context("and the database is empty", func() {
			It("returns empty map for empty input", func() {
				res, err := db.GetIssueCountsByServiceIDs(context.Background(), []int64{})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty map", func() {
					Expect(res).To(BeEmpty())
				})
			})

			It("returns empty map for non-existent IDs", func() {
				res, err := db.GetIssueCountsByServiceIDs(context.Background(), []int64{99999})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty map", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})

		Context("and we have seeded data", func() {
			var seedCollection *test.SeedCollection

			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			It("returns issue counts grouped by service ID", func() {
				serviceIDs := lo.Map(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, _ int) int64 {
					return s.Id.Int64
				})

				res, err := db.GetIssueCountsByServiceIDs(context.Background(), serviceIDs)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning valid severity counts", func() {
					for _, counts := range res {
						Expect(counts.Critical).To(BeNumerically(">=", 0))
						Expect(counts.High).To(BeNumerically(">=", 0))
						Expect(counts.Medium).To(BeNumerically(">=", 0))
						Expect(counts.Low).To(BeNumerically(">=", 0))
						Expect(counts.None).To(BeNumerically(">=", 0))
						Expect(counts.Total).To(Equal(
							counts.Critical + counts.High + counts.Medium + counts.Low + counts.None,
						))
					}
				})
			})

			It("returns data for a single service ID", func() {
				if len(seedCollection.ServiceRows) == 0 {
					Skip("No services found in seed data")
				}

				targetServiceID := seedCollection.ServiceRows[0].Id.Int64
				res, err := db.GetIssueCountsByServiceIDs(context.Background(), []int64{targetServiceID})

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning at most one entry", func() {
					Expect(len(res)).To(BeNumerically("<=", 1))
				})
			})
		})
	})
})
