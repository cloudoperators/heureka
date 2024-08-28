// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/samber/lo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb/test"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

var _ = Describe("IssueMatch", Label("database", "IssueMatch"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	When("Getting All IssueMatch IDs", Label("GetAllIssueMatchIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllIssueMatchIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 IssueMatches in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				issueMatches := seedCollection.GetValidIssueMatchRows()
				for _, im := range issueMatches {
					ids = append(ids, im.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllIssueMatchIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueMatchRows)))
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
				It("can filter by a single issue match id that does exist", func() {
					vmId := ids[rand.Intn(len(ids))]
					filter := &entity.IssueMatchFilter{
						Id: []*int64{&vmId},
					}

					entries, err := db.GetAllIssueMatchIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(vmId))
					})
				})
				It("can filter by a single issue id that does exist", func() {
					issueMatch := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					filter := &entity.IssueMatchFilter{
						Paginated: entity.Paginated{},
						IssueId:   []*int64{&issueMatch.IssueId.Int64},
					}

					var imIds []int64
					for _, e := range seedCollection.IssueMatchRows {
						if e.IssueId.Int64 == issueMatch.IssueId.Int64 {
							imIds = append(imIds, e.Id.Int64)
						}
					}

					entries, err := db.GetAllIssueMatchIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(imIds)))
					})

					By("returning expected elements", func() {
						for _, id := range entries {
							Expect(lo.Contains(imIds, id)).To(BeTrue())
						}
					})
				})
			})
		})
	})

	When("Getting IssueMatches", Label("GetIssueMatches"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetIssueMatches(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 IssueMatches in the database", func() {
			var seedCollection *test.SeedCollection
			var issueMatches []mariadb.IssueMatchRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				issueMatches = seedCollection.GetValidIssueMatchRows()
			})
			Context("and using no filter", func() {

				It("can fetch the items correctly", func() {
					res, err := db.GetIssueMatches(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.IssueMatchRows)))
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
							for _, row := range seedCollection.IssueMatchRows {
								if r.Id == row.Id.Int64 {
									Expect(r.RemediationDate.Unix()).Should(BeEquivalentTo(row.RemediationDate.Time.Unix()), "Remediation Date matches")
									Expect(r.CreatedAt.Unix()).ShouldNot(BeEquivalentTo(row.CreatedAt.Time.Unix()), "CreatedAt got set")
									Expect(r.UpdatedAt.Unix()).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time.Unix()), "UpdatedAt got set")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single issue match id that does exist", func() {
					im := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					filter := &entity.IssueMatchFilter{
						Id: []*int64{&im.Id.Int64},
					}

					entries, err := db.GetIssueMatches(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(im.Id.Int64))
					})
				})
				It("can filter by a single issue id that does exist", func() {
					issueMatch := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					filter := &entity.IssueMatchFilter{
						Paginated: entity.Paginated{},
						IssueId:   []*int64{&issueMatch.IssueId.Int64},
					}

					var imIds []int64
					for _, e := range seedCollection.IssueMatchRows {
						if e.IssueId.Int64 == issueMatch.IssueId.Int64 {
							imIds = append(imIds, e.Id.Int64)
						}
					}

					entries, err := db.GetIssueMatches(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(imIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(imIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single component instance id that does exist", func() {
					issueMatch := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					filter := &entity.IssueMatchFilter{
						Paginated:           entity.Paginated{},
						ComponentInstanceId: []*int64{&issueMatch.ComponentInstanceId.Int64},
					}

					var imIds []int64
					for _, e := range seedCollection.IssueMatchRows {
						if e.ComponentInstanceId.Int64 == issueMatch.ComponentInstanceId.Int64 {
							imIds = append(imIds, e.Id.Int64)
						}
					}

					entries, err := db.GetIssueMatches(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(imIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(imIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single evidence id that does exist", func() {
					issueMatch := seedCollection.IssueMatchEvidenceRows[rand.Intn(len(seedCollection.IssueMatchEvidenceRows))]
					filter := &entity.IssueMatchFilter{
						Paginated:  entity.Paginated{},
						EvidenceId: []*int64{&issueMatch.EvidenceId.Int64},
					}

					var imIds []int64
					for _, e := range seedCollection.IssueMatchEvidenceRows {
						if e.EvidenceId.Int64 == issueMatch.EvidenceId.Int64 {
							imIds = append(imIds, e.IssueMatchId.Int64)
						}
					}

					entries, err := db.GetIssueMatches(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(imIds)))
					})

					By("returning expected elements", func() {
						for _, entry := range entries {
							Expect(lo.Contains(imIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single support group name that does exist", func() {
					issueMatch := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					componentInstance, _ := lo.Find(seedCollection.ComponentInstanceRows, func(c mariadb.ComponentInstanceRow) bool {
						return c.Id.Int64 == issueMatch.ComponentInstanceId.Int64
					})
					service, _ := lo.Find(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow) bool {
						return s.Id.Int64 == componentInstance.ServiceId.Int64
					})
					supportGroupService, _ := lo.Find(seedCollection.SupportGroupServiceRows, func(s mariadb.SupportGroupServiceRow) bool {
						return s.ServiceId.Int64 == service.Id.Int64
					})
					supportGroup, sgFound := lo.Find(seedCollection.SupportGroupRows, func(s mariadb.SupportGroupRow) bool {
						return s.Id.Int64 == supportGroupService.SupportGroupId.Int64
					})

					filter := &entity.IssueMatchFilter{
						Paginated:        entity.Paginated{},
						SupportGroupName: []*string{&supportGroup.Name.String},
					}

					// fixture creation does not guarantee that a support group is always present
					if sgFound {
						entries, err := db.GetIssueMatches(filter)

						By("throwing no error", func() {
							Expect(err).To(BeNil())
						})

						By("returning expected number of results", func() {
							Expect(entries).ToNot(BeEmpty())
						})

						By("entries contain vm", func() {
							_, found := lo.Find(entries, func(e entity.IssueMatch) bool {
								return e.Id == issueMatch.Id.Int64
							})
							Expect(found).To(BeTrue())
						})
					}
				})
				Context("and and we use Pagination", func() {
					DescribeTable("can correctly paginate ", func(pageSize int) {
						test.TestPaginationOfList(
							db.GetIssueMatches,
							func(first *int, after *int64) *entity.IssueMatchFilter {
								return &entity.IssueMatchFilter{
									Paginated: entity.Paginated{
										First: first,
										After: after,
									},
								}
							},
							func(entries []entity.IssueMatch) *int64 { return &entries[len(entries)-1].Id },
							len(issueMatches),
							pageSize,
						)
					},
						Entry("when pageSize is 1", 1),
						Entry("when pageSize is 3", 3),
						Entry("when pageSize is 5", 5),
						Entry("when pageSize is 11", 11),
						Entry("when pageSize is 100", 100),
					)
				})
			})
		})
	})
	When("Counting Issue Matches", Label("CountIssueMatches"), func() {
		Context("and using no filter", func() {
			DescribeTable("it returns correct count", func(x int) {
				_ = seeder.SeedDbWithNFakeData(x)
				res, err := db.CountIssueMatches(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				By("returning the correct count", func() {
					Expect(res).To(BeEquivalentTo(x))
				})

			},
				Entry("when page size is 0", 0),
				Entry("when page size is 1", 1),
				Entry("when page size is 11", 11),
				Entry("when page size is 100", 100))
		})
		Context("and using a filter", func() {
			Context("and having 20 elements in the Database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(20)
				})
				It("does not influence the count when pagination is applied", func() {
					var first = 1
					var after int64 = 0
					filter := &entity.IssueMatchFilter{
						Paginated: entity.Paginated{
							First: &first,
							After: &after,
						},
					}
					res, err := db.CountIssueMatches(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(res).To(BeEquivalentTo(20))
					})
				})
				It("does show the correct amount when filtering for an issue", func() {
					issueMatch := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					filter := &entity.IssueMatchFilter{
						Paginated: entity.Paginated{},
						IssueId:   []*int64{&issueMatch.IssueId.Int64},
					}

					var imIds []int64
					for _, e := range seedCollection.IssueMatchRows {
						if e.IssueId.Int64 == issueMatch.IssueId.Int64 {
							imIds = append(imIds, e.Id.Int64)
						}
					}
					count, err := db.CountIssueMatches(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(count).To(BeEquivalentTo(len(imIds)))
					})
				})
			})
		})
	})
	When("Insert IssueMatch", Label("InsertIssueMatch"), func() {
		Context("and we have 10 IssueMatches in the database", func() {
			var newIssueMatchRow mariadb.IssueMatchRow
			var newIssueMatch entity.IssueMatch
			var seedCollection *test.SeedCollection
			var user entity.User
			var issue entity.Issue
			var componentInstance entity.ComponentInstance
			BeforeEach(func() {
				seeder.SeedDbWithNFakeData(10)
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newIssueMatchRow = test.NewFakeIssueMatch()
				newIssueMatch = newIssueMatchRow.AsIssueMatch()
				user = seedCollection.UserRows[rand.Intn(len(seedCollection.UserRows))].AsUser()
				issue = seedCollection.IssueRows[rand.Intn(len(seedCollection.IssueRows))].AsIssue()
				componentInstance = seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))].AsComponentInstance()
				newIssueMatch.UserId = user.Id
				newIssueMatch.IssueId = issue.Id
				newIssueMatch.ComponentInstanceId = componentInstance.Id
			})
			It("can insert correctly", func() {
				issueMatch, err := db.CreateIssueMatch(&newIssueMatch)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets issueMatch id", func() {
					Expect(issueMatch).NotTo(BeEquivalentTo(0))
				})

				issueMatchFilter := &entity.IssueMatchFilter{
					Id: []*int64{&issueMatch.Id},
				}

				im, err := db.GetIssueMatches(issueMatchFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issueMatch", func() {
					Expect(len(im)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(im[0].Severity.Cvss.Vector).To(BeEquivalentTo(issueMatch.Severity.Cvss.Vector))
					Expect(im[0].Severity.Value).To(BeEquivalentTo(issueMatch.Severity.Value))
					Expect(im[0].Status.String()).To(BeEquivalentTo(issueMatch.Status.String()))
					Expect(im[0].UserId).To(BeEquivalentTo(issueMatch.UserId))
					Expect(im[0].IssueId).To(BeEquivalentTo(issueMatch.IssueId))
					Expect(im[0].ComponentInstanceId).To(BeEquivalentTo(issueMatch.ComponentInstanceId))
					Expect(im[0].TargetRemediationDate.Unix()).To(BeEquivalentTo(issueMatch.TargetRemediationDate.Unix()))
					Expect(im[0].RemediationDate.Unix()).To(BeEquivalentTo(issueMatch.RemediationDate.Unix()))
				})
			})
		})
	})
	When("Update IssueMatch", Label("UpdateIssueMatch"), func() {
		Context("and we have 10 IssueMatches in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update issueMatch rating correctly", func() {
				issueMatch := seedCollection.IssueMatchRows[0].AsIssueMatch()

				if issueMatch.Status == entity.NewIssueMatchStatusValue("new") {
					issueMatch.Status = entity.NewIssueMatchStatusValue("risk_accepted")
				} else {
					issueMatch.Status = entity.NewIssueMatchStatusValue("new")
				}

				err := db.UpdateIssueMatch(&issueMatch)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueMatchFilter := &entity.IssueMatchFilter{
					Id: []*int64{&issueMatch.Id},
				}

				im, err := db.GetIssueMatches(issueMatchFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issueMatch", func() {
					Expect(len(im)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(im[0].Severity.Cvss.Vector).To(BeEquivalentTo(issueMatch.Severity.Cvss.Vector))
					Expect(im[0].Severity.Value).To(BeEquivalentTo(issueMatch.Severity.Value))
					Expect(im[0].Status.String()).To(BeEquivalentTo(issueMatch.Status.String()))
					Expect(im[0].UserId).To(BeEquivalentTo(issueMatch.UserId))
					Expect(im[0].IssueId).To(BeEquivalentTo(issueMatch.IssueId))
					Expect(im[0].ComponentInstanceId).To(BeEquivalentTo(issueMatch.ComponentInstanceId))
					Expect(im[0].TargetRemediationDate.Unix()).To(BeEquivalentTo(issueMatch.TargetRemediationDate.Unix()))
					Expect(im[0].RemediationDate.Unix()).To(BeEquivalentTo(issueMatch.RemediationDate.Unix()))
				})
			})
		})
	})
	When("Delete IssueMatch", Label("DeleteIssueMatch"), func() {
		Context("and we have 10 IssueMatches in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete issueMatch correctly", func() {
				issueMatch := seedCollection.IssueMatchRows[0].AsIssueMatch()

				err := db.DeleteIssueMatch(issueMatch.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueMatchFilter := &entity.IssueMatchFilter{
					Id: []*int64{&issueMatch.Id},
				}

				im, err := db.GetIssueMatches(issueMatchFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no issueMatch", func() {
					Expect(len(im)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Add Evidence To IssueMatch", Label("AddEvidenceToIssueMatch"), func() {
		Context("and we have 10 IssueMatches in the database", func() {
			var seedCollection *test.SeedCollection
			var newEvidenceRow mariadb.EvidenceRow
			var newEvidence entity.Evidence
			var evidence *entity.Evidence
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
				evidence, _ = db.CreateEvidence(&newEvidence)
			})
			It("can add evidence correctly", func() {
				issueMatch := seedCollection.IssueMatchRows[0].AsIssueMatch()

				err := db.AddEvidenceToIssueMatch(issueMatch.Id, evidence.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueMatchFilter := &entity.IssueMatchFilter{
					EvidenceId: []*int64{&evidence.Id},
				}

				im, err := db.GetIssueMatches(issueMatchFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning issueMatch", func() {
					Expect(len(im)).To(BeEquivalentTo(1))
				})
			})
		})
	})
	When("Remove Evidence From IssueMatch", Label("RemoveEvidenceFromIssueMatch"), func() {
		Context("and we have 10 IssueMatches in the database", func() {
			var seedCollection *test.SeedCollection
			var issueMatchEvidenceRow mariadb.IssueMatchEvidenceRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				issueMatchEvidenceRow = seedCollection.IssueMatchEvidenceRows[0]
			})
			It("can remove evidence correctly", func() {
				err := db.RemoveEvidenceFromIssueMatch(issueMatchEvidenceRow.IssueMatchId.Int64, issueMatchEvidenceRow.EvidenceId.Int64)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				issueMatchFilter := &entity.IssueMatchFilter{
					EvidenceId: []*int64{&issueMatchEvidenceRow.EvidenceId.Int64},
				}

				issueMatches, err := db.GetIssueMatches(issueMatchFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				for _, im := range issueMatches {
					Expect(im.Id).ToNot(BeEquivalentTo(issueMatchEvidenceRow.IssueMatchId.Int64))
				}
			})
		})
	})
})
