// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"
	"time"

	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// nolint due to weak random number generator for test reason
//
//nolint:gosec
var _ = Describe("Evidence", Label("database", "Evidence"), func() {
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

	When("Getting All Evidence IDs", Label("GetAllEvidenceIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllEvidenceIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Evidences in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, e := range seedCollection.EvidenceRows {
					ids = append(ids, e.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllEvidenceIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.EvidenceRows)))
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
				It("can filter by a single evidence id that does exist", func() {
					eId := ids[rand.Intn(len(ids))]
					filter := &entity.EvidenceFilter{
						Id: []*int64{&eId},
					}

					entries, err := db.GetAllEvidenceIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(eId))
					})
				})
			})
		})
	})

	When("Getting Evidences", Label("GetEvidences"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetEvidences(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 evidences in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetEvidences(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.EvidenceRows)))
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
							for _, row := range seedCollection.EvidenceRows {
								if r.Id == row.Id.Int64 {
									Expect(r.Description).Should(BeEquivalentTo(row.Description.String), "Description matches")
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
					evidence := seedCollection.EvidenceRows[rand.Intn(len(seedCollection.EvidenceRows))]
					filter := &entity.EvidenceFilter{
						Paginated:  entity.Paginated{},
						ActivityId: []*int64{&evidence.ActivityId.Int64},
					}

					var evidences []mariadb.EvidenceRow
					for _, e := range seedCollection.EvidenceRows {
						if e.ActivityId.Int64 == evidence.ActivityId.Int64 {
							evidences = append(evidences, e)
						}
					}

					entries, err := db.GetEvidences(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(evidences)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(entry.ActivityId).To(BeEquivalentTo(evidence.ActivityId.Int64))
						}
					})
				})
				It("can filter by a single id that does exist", func() {
					evidence := seedCollection.EvidenceRows[rand.Intn(len(seedCollection.EvidenceRows))]
					filter := &entity.EvidenceFilter{
						Paginated: entity.Paginated{},
						Id:        []*int64{&evidence.Id.Int64},
					}

					entries, err := db.GetEvidences(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(entry.Id).To(BeEquivalentTo(evidence.Id.Int64))
						}
					})
				})
				It("can filter by a single issue match id that does exist", func() {
					issueMatchEvidenceRow := seedCollection.IssueMatchEvidenceRows[rand.Intn(len(seedCollection.IssueMatchEvidenceRows))]

					var evidenceIds []int64
					for _, ime := range seedCollection.IssueMatchEvidenceRows {
						if ime.IssueMatchId.Int64 == issueMatchEvidenceRow.IssueMatchId.Int64 &&
							!lo.Contains(evidenceIds, ime.EvidenceId.Int64) {
							evidenceIds = append(evidenceIds, ime.EvidenceId.Int64)
						}
					}
					filter := &entity.EvidenceFilter{
						Paginated:    entity.Paginated{},
						IssueMatchId: []*int64{&issueMatchEvidenceRow.IssueMatchId.Int64},
					}

					entries, err := db.GetEvidences(filter)

					GinkgoLogr.Info("Test")

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(evidenceIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(evidenceIds, entry.Id)).To(BeTrue())
						}
					})
				})
				Context("and and we use Pagination", func() {
					DescribeTable("can correctly paginate with x elements", func(pageSize int) {
						test.TestPaginationOfList(
							db.GetEvidences,
							func(first *int, after *int64) *entity.EvidenceFilter {
								return &entity.EvidenceFilter{
									Paginated: entity.Paginated{
										First: first,
										After: after,
									},
								}
							},
							func(entries []entity.Evidence) *int64 { return &entries[len(entries)-1].Id },
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
		When("Counting Evidences", Label("CountEvidence"), func() {
			Context("and the database is empty", func() {
				It("returns a correct totalCount without an error", func() {
					c, err := db.CountEvidences(nil)
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
				var evidenceRows []mariadb.EvidenceRow
				var count int
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(100)
					evidenceRows = seedCollection.GetValidEvidenceRows()
					count = len(evidenceRows)
				})
				It("works when providing no filter", func() {
					c, err := db.CountEvidences(nil)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
				It("works when providing a filter that does paginate", func() {
					f := 10
					filter := &entity.EvidenceFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountEvidences(filter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})
		})
	})
	When("Insert Evidence", Label("InsertEvidence"), func() {
		Context("and we have 10 Evidences in the database", func() {
			var seedCollection *test.SeedCollection
			var newEvidenceRow mariadb.EvidenceRow
			var newEvidence entity.Evidence
			var activity entity.Activity
			var user entity.User
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newEvidenceRow = test.NewFakeEvidence()
				newEvidence = newEvidenceRow.AsEvidence()
				activity = seedCollection.ActivityRows[0].AsActivity()
				user = seedCollection.UserRows[0].AsUser()
				newEvidence.ActivityId = activity.Id
				newEvidence.UserId = user.Id
			})
			It("can insert correctly", func() {
				evidence, err := db.CreateEvidence(&newEvidence)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets evidence id", func() {
					Expect(evidence).NotTo(BeEquivalentTo(0))
				})

				evidenceFilter := &entity.EvidenceFilter{
					Id: []*int64{&evidence.Id},
				}

				e, err := db.GetEvidences(evidenceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning evidence", func() {
					Expect(len(e)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(e[0].Description).To(BeEquivalentTo(evidence.Description))
					Expect(e[0].ActivityId).To(BeEquivalentTo(evidence.ActivityId))
					Expect(e[0].UserId).To(BeEquivalentTo(evidence.UserId))
					Expect(e[0].Severity.Cvss.Vector).To(BeEquivalentTo(evidence.Severity.Cvss.Vector))
					Expect(e[0].RaaEnd.Unix()).To(BeEquivalentTo(evidence.RaaEnd.Unix()))
					Expect(e[0].Type).To(BeEquivalentTo(evidence.Type))
				})
			})
		})
	})
	When("Update Evidence", Label("UpdateEvidence"), func() {
		Context("and we have 10 Evidences in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update evidence correctly", func() {
				evidence := seedCollection.EvidenceRows[0].AsEvidence()

				evidence.Description = "New Description"
				evidence.RaaEnd = evidence.RaaEnd.Add(24 * time.Hour)
				err := db.UpdateEvidence(&evidence)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				evidenceFilter := &entity.EvidenceFilter{
					Id: []*int64{&evidence.Id},
				}

				e, err := db.GetEvidences(evidenceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning evidence", func() {
					Expect(len(e)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(e[0].Description).To(BeEquivalentTo(evidence.Description))
					Expect(e[0].ActivityId).To(BeEquivalentTo(evidence.ActivityId))
					Expect(e[0].UserId).To(BeEquivalentTo(evidence.UserId))
					Expect(e[0].Severity.Cvss.Vector).To(BeEquivalentTo(evidence.Severity.Cvss.Vector))
					Expect(e[0].RaaEnd.Unix()).To(BeEquivalentTo(evidence.RaaEnd.Unix()))
					Expect(e[0].Type).To(BeEquivalentTo(evidence.Type))
				})
			})
		})
	})
	When("Delete Evidence", Label("DeleteEvidence"), func() {
		Context("and we have 10 Evidences in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete evidence correctly", func() {
				evidence := seedCollection.EvidenceRows[0].AsEvidence()

				err := db.DeleteEvidence(evidence.Id, util.SystemUserId)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				evidenceFilter := &entity.EvidenceFilter{
					Id: []*int64{&evidence.Id},
				}

				e, err := db.GetEvidences(evidenceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no evidence", func() {
					Expect(len(e)).To(BeEquivalentTo(0))
				})
			})
		})
	})
})
