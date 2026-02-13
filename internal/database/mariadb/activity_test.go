// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

// nolint due to weak random number generator for test reason
//
//nolint:gosec
var _ = Describe("Activity", Label("database", "Activity"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	AfterEach(func() {
		_ = dbm.TestTearDown(db)
	})

	When("Getting All Activity IDs", Label("GetAllActivityIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllActivityIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Activities in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, a := range seedCollection.ActivityRows {
					ids = append(ids, a.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllActivityIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ActivityRows)))
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
				It("can filter by a single activity id that does exist", func() {
					aId := ids[rand.Intn(len(ids))]
					filter := &entity.ActivityFilter{
						Id: []*int64{&aId},
					}

					entries, err := db.GetAllActivityIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(aId))
					})
				})
			})
		})
	})

	When("Getting Activity", Label("GetActivities"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetActivities(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 activities in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetActivities(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ActivityRows)))
					})

					By("returning the correct order", func() {
						var prev int64 = 0
						for _, r := range res {

							Expect(r.Id > prev).Should(BeTrue())
							prev = r.Id

						}
					})

					By("returning the correct fields", func() {
						for _, r := range res {
							for _, row := range seedCollection.ActivityRows {
								if r.Id == row.Id.Int64 {
									Expect(r.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single activity id that does exist", func() {
					activity := seedCollection.ActivityRows[rand.Intn(len(seedCollection.ActivityRows))]
					filter := &entity.ActivityFilter{
						Id: []*int64{&activity.Id.Int64},
					}

					entries, err := db.GetActivities(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(activity.Id.Int64))
					})
				})
				It("can filter by a single status", func() {
					activityRow := seedCollection.ActivityRows[rand.Intn(len(seedCollection.ActivityRows))]

					filter := &entity.ActivityFilter{Status: []*string{&activityRow.Status.String}}

					entries, err := db.GetActivities(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(entry.Status.String()).To(Equal(activityRow.Status.String))
						}
					})
				})
				It("can filter by a single service id", func() {
					// select a service
					serviceRow := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					// collect all activity ids that belong to the service
					activityIds := []int64{}
					for _, ahsRow := range seedCollection.ActivityHasServiceRows {
						if ahsRow.ServiceId.Int64 == serviceRow.Id.Int64 {
							activityIds = append(activityIds, ahsRow.ActivityId.Int64)
						}
					}

					filter := &entity.ActivityFilter{ServiceId: []*int64{&serviceRow.Id.Int64}}

					entries, err := db.GetActivities(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(activityIds).To(ContainElement(entry.Id))
						}
					})
				})
				It("can filter by a single service name", func() {
					// select a service
					serviceRow := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					// collect all activity ids that belong to the service
					activityIds := []int64{}
					for _, ahsRow := range seedCollection.ActivityHasServiceRows {
						if ahsRow.ServiceId.Int64 == serviceRow.Id.Int64 {
							activityIds = append(activityIds, ahsRow.ActivityId.Int64)
						}
					}

					filter := &entity.ActivityFilter{ServiceCCRN: []*string{&serviceRow.CCRN.String}}

					entries, err := db.GetActivities(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(activityIds).To(ContainElement(entry.Id))
						}
					})
				})
			})
			It("can filter by a single evidence id", func() {
				// select a service
				e := seedCollection.EvidenceRows[rand.Intn(len(seedCollection.EvidenceRows))]
				filter := &entity.ActivityFilter{EvidenceId: []*int64{&e.Id.Int64}}

				entries, err := db.GetActivities(filter)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning expected elements", func() {
					for _, entry := range entries {
						Expect(e.ActivityId.Int64).To(BeIdenticalTo(entry.Id))
					}
				})
			})
			It("can filter by a single issue id", func() {
				// select a service
				ahi := seedCollection.ActivityHasIssueRows[rand.Intn(len(seedCollection.ActivityHasIssueRows))]

				// collect all activity ids that belong to the service
				activityIds := []int64{}
				for _, ahiRow := range seedCollection.ActivityHasIssueRows {
					if ahiRow.IssueId.Int64 == ahi.IssueId.Int64 &&
						!lo.Contains(activityIds, ahiRow.ActivityId.Int64) {
						activityIds = append(activityIds, ahiRow.ActivityId.Int64)
					}
				}

				filter := &entity.ActivityFilter{IssueId: []*int64{&ahi.IssueId.Int64}}

				entries, err := db.GetActivities(filter)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning expected element count", func() {
					Expect(len(entries)).To(BeIdenticalTo(len(activityIds)))
				})

				By("returning expected elements", func() {
					for _, entry := range entries {
						Expect(activityIds).To(ContainElement(entry.Id))
					}
				})
			})
			Context("and and we use Pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfList(
						db.GetActivities,
						func(first *int, after *int64) *entity.ActivityFilter {
							return &entity.ActivityFilter{
								Paginated: entity.Paginated{
									First: first,
									After: after,
								},
							}
						},
						func(entries []entity.Activity) *int64 { return &entries[len(entries)-1].Id },
						10,
						pageSize,
					)
				},
					Entry("When x is 1", 1),
					Entry("When x is 3", 3),
					Entry("When x is 5", 5),
					Entry("When x is 11", 11),
					Entry("When x is 100", 100),
				)
			})
		})
	})
	When("Counting Activities", Label("CountActivities"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountActivities(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning the correct count", func() {
					Expect(c).To(BeEquivalentTo(0))
				})
			})
		})
		Context("and the database has 100 entries", func() {
			var seedCollection *test.SeedCollection
			var ActivityRows []mariadb.ActivityRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				ActivityRows = seedCollection.ActivityRows
				count = len(ActivityRows)
			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountActivities(nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})
			Context("and using pagination", func() {
				It("can count", func() {
					f := 10
					filter := &entity.ActivityFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountActivities(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})

			Context("and using a filter", func() {
				DescribeTable("can count with a filter", func(pageSize int, filterMatches int) {
					// select a service
					serviceRow := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					// collect all activity ids that belong to the service
					activityIds := []int64{}
					for _, ahsRow := range seedCollection.ActivityHasServiceRows {
						if ahsRow.ServiceId.Int64 == serviceRow.Id.Int64 {
							activityIds = append(activityIds, ahsRow.ActivityId.Int64)
						}
					}

					filter := &entity.ActivityFilter{ServiceCCRN: []*string{&serviceRow.CCRN.String}}

					entries, err := db.CountActivities(filter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(entries).To(BeEquivalentTo(len(activityIds)))
					})
				},
					Entry("and pageSize is 1 and it has 13 elements", 1, 13),
					Entry("and  pageSize is 20 and it has 5 elements", 20, 5),
					Entry("and  pageSize is 100 and it has 100 elements", 100, 100),
				)
			})
		})
	})
	When("Insert Activity", Label("InsertActivity"), func() {
		Context("and we have 10 Activities in the database", func() {
			var newActivityRow mariadb.ActivityRow
			var newActivity entity.Activity
			BeforeEach(func() {
				seeder.SeedDbWithNFakeData(10)
				newActivityRow = test.NewFakeActivity()
				newActivity = newActivityRow.AsActivity()
			})
			It("can insert correctly", func() {
				activity, err := db.CreateActivity(&newActivity)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets activity id", func() {
					Expect(activity).NotTo(BeEquivalentTo(0))
				})

				activityFilter := &entity.ActivityFilter{
					Id: []*int64{&activity.Id},
				}

				a, err := db.GetActivities(activityFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning activity", func() {
					Expect(len(a)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(a[0].Status).To(BeEquivalentTo(activity.Status))
				})
			})
		})
	})
	When("Update Activity", Label("UpdateActivity"), func() {
		Context("and we have 10 Activities in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update activity status correctly", func() {
				activity := seedCollection.ActivityRows[0].AsActivity()

				if activity.Status == "open" {
					activity.Status = "closed"
				} else {
					activity.Status = "open"
				}

				err := db.UpdateActivity(&activity)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				activityFilter := &entity.ActivityFilter{
					Id: []*int64{&activity.Id},
				}

				a, err := db.GetActivities(activityFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning activity", func() {
					Expect(len(a)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(a[0].Status).To(BeEquivalentTo(activity.Status))
				})
			})
		})
	})
	When("Delete Activity", Label("DeleteActivity"), func() {
		Context("and we have 10 Activities in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete activity correctly", func() {
				activity := seedCollection.ActivityRows[0].AsActivity()

				err := db.DeleteActivity(activity.Id, util.SystemUserId)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				activityFilter := &entity.ActivityFilter{
					Id: []*int64{&activity.Id},
				}

				a, err := db.GetActivities(activityFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no activity", func() {
					Expect(len(a)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Add Service To Activity", Label("AddServiceToActivity"), func() {
		Context("and we have 10 activities in the database", func() {
			var seedCollection *test.SeedCollection
			var newServiceRow mariadb.ServiceRow
			var newService entity.Service
			var service *entity.Service
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newServiceRow = test.NewFakeService()
				newService = newServiceRow.AsService()
				service, _ = db.CreateService(&newService)
			})
			It("can add service correctly", func() {
				activity := seedCollection.ActivityRows[0].AsActivity()

				err := db.AddServiceToActivity(activity.Id, service.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				activityFilter := &entity.ActivityFilter{
					ServiceId: []*int64{&service.Id},
				}

				a, err := db.GetActivities(activityFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning activity", func() {
					Expect(len(a)).To(BeEquivalentTo(1))
				})
			})
		})
	})
	When("Remove Service From Activity", Label("RemoveServiceFromActivity"), func() {
		Context("and we have 10 Activities in the database", func() {
			var seedCollection *test.SeedCollection
			var activityHasServiceRow mariadb.ActivityHasServiceRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				activityHasServiceRow = seedCollection.ActivityHasServiceRows[0]
			})
			It("can remove service correctly", func() {
				err := db.RemoveServiceFromActivity(activityHasServiceRow.ActivityId.Int64, activityHasServiceRow.ServiceId.Int64)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				activityFilter := &entity.ActivityFilter{
					ServiceId: []*int64{&activityHasServiceRow.ServiceId.Int64},
				}

				activities, err := db.GetActivities(activityFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				for _, a := range activities {
					Expect(a.Id).ToNot(BeEquivalentTo(activityHasServiceRow.ActivityId.Int64))
				}
			})
		})
	})
	When("Add Issue To Activity", Label("AddIssueToActivity"), func() {
		Context("and we have 10 Activities in the database", func() {
			var seedCollection *test.SeedCollection
			var newIssueRow mariadb.IssueRow
			var newIssue entity.Issue
			var issue *entity.Issue
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newIssueRow = test.NewFakeIssue()
				newIssue = newIssueRow.AsIssue()
				issue, _ = db.CreateIssue(&newIssue)
			})
			It("can add issue correctly", func() {
				activity := seedCollection.ActivityRows[0].AsActivity()

				err := db.AddIssueToActivity(activity.Id, issue.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				activityFilter := &entity.ActivityFilter{
					IssueId: []*int64{&issue.Id},
				}

				a, err := db.GetActivities(activityFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning activity", func() {
					Expect(len(a)).To(BeEquivalentTo(1))
				})
			})
		})
	})
	When("Remove Issue From Activity", Label("RemoveIssueFromActivity"), func() {
		Context("and we have 10 Activities in the database", func() {
			var seedCollection *test.SeedCollection
			var activityHasIssueRow mariadb.ActivityHasIssueRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				activityHasIssueRow = seedCollection.ActivityHasIssueRows[0]
			})
			It("can remove issue correctly", func() {
				err := db.RemoveIssueFromActivity(activityHasIssueRow.ActivityId.Int64, activityHasIssueRow.IssueId.Int64)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				activityFilter := &entity.ActivityFilter{
					IssueId: []*int64{&activityHasIssueRow.IssueId.Int64},
				}

				activities, err := db.GetActivities(activityFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				for _, a := range activities {
					Expect(a.Id).ToNot(BeEquivalentTo(activityHasIssueRow.ActivityId.Int64))
				}
			})
		})
	})
})
