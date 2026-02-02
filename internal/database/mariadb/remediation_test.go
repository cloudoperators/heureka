// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"database/sql"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"math/rand"
)

var _ = Describe("Remediation", Label("database", "Remediation"), func() {

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

	When("Getting Remediations", Label("GetRemediations"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetRemediations(nil, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 remediations in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetRemediations(nil, nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.RemediationRows)))
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
							for _, row := range seedCollection.RemediationRows {
								if r.Id == row.Id.Int64 {
									Expect(r.Description).Should(BeEquivalentTo(row.Description.String), "Description should match")
									Expect(r.Service).Should(BeEquivalentTo(row.Service.String), "Service should match")
									Expect(r.ServiceId).Should(BeEquivalentTo(row.ServiceId.Int64), "ServiceId should match")
									Expect(r.Component).Should(BeEquivalentTo(row.Component.String), "Component should match")
									Expect(r.ComponentId).Should(BeEquivalentTo(row.ComponentId.Int64), "ComponentId should match")
									Expect(r.Issue).Should(BeEquivalentTo(row.Issue.String), "Issue should match")
									Expect(r.IssueId).Should(BeEquivalentTo(row.IssueId.Int64), "IssueId should match")
									Expect(r.RemediationDate).ShouldNot(BeEquivalentTo(row.RemediationDate.Time), "RemediationDate matches")
									Expect(r.ExpirationDate).ShouldNot(BeEquivalentTo(row.ExpirationDate.Time), "ExpirationDate matches")
									Expect(r.RemediatedBy).Should(BeEquivalentTo(row.RemediatedBy.String), "RemediatedBy should match")
									Expect(r.RemediatedById).Should(BeEquivalentTo(row.RemediatedById.Int64), "RemediatedBy should match")
									Expect(r.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single id", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]
					filter := &entity.RemediationFilter{Id: []*int64{&row.Id.Int64}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning one result", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})
					By("returned entry includes the id", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(row.Id.Int64))
					})
				})
				It("can filter by a single serverity", func() {
					severity := gofakeit.RandomString(entity.AllSeverityValuesString)
					filter := &entity.RemediationFilter{Severity: []*string{&severity}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returned entry includes the severity", func() {
						for _, entry := range entries {
							Expect(entry.Severity).To(BeEquivalentTo(severity))
						}
					})
				})
				It("can filter by a single service", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]
					filter := &entity.RemediationFilter{Service: []*string{&row.Service.String}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the service name", func() {
						for _, entry := range entries {
							Expect(entry.Service).To(BeEquivalentTo(row.Service.String))
						}
					})
				})
				It("can filter by a single service id", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]
					filter := &entity.RemediationFilter{ServiceId: []*int64{&row.ServiceId.Int64}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the service id", func() {
						for _, entry := range entries {
							Expect(entry.ServiceId).To(BeEquivalentTo(row.ServiceId.Int64))
						}
					})
				})
				It("can filter by a single component", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]
					filter := &entity.RemediationFilter{Component: []*string{&row.Component.String}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the component", func() {
						for _, entry := range entries {
							Expect(entry.Component).To(BeEquivalentTo(row.Component.String))
						}
					})
				})
				It("can filter by a single component id", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]
					filter := &entity.RemediationFilter{ComponentId: []*int64{&row.ComponentId.Int64}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the component id", func() {
						for _, entry := range entries {
							Expect(entry.ComponentId).To(BeEquivalentTo(row.ComponentId.Int64))
						}
					})
				})
				It("can filter by a single issue", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]
					filter := &entity.RemediationFilter{Issue: []*string{&row.Issue.String}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the issue", func() {
						for _, entry := range entries {
							Expect(entry.Issue).To(BeEquivalentTo(row.Issue.String))
						}
					})
				})
				It("can filter by a single issue id", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]
					filter := &entity.RemediationFilter{IssueId: []*int64{&row.IssueId.Int64}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the issue id", func() {
						for _, entry := range entries {
							Expect(entry.IssueId).To(BeEquivalentTo(row.IssueId.Int64))
						}
					})
				})
				It("can filter by a single type", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]
					filter := &entity.RemediationFilter{Type: []*string{&row.Type.String}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the type", func() {
						for _, entry := range entries {
							Expect(entry.Type).To(BeEquivalentTo(row.Type.String))
						}
					})
				})
				It("can filter by 'risk_accepted' type", func() {
					remediationType := entity.RemediationTypeRiskAccepted.String()
					filter := &entity.RemediationFilter{Type: []*string{&remediationType}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})

					By("returning entries include the type", func() {
						for _, entry := range entries {
							Expect(entry.Type).To(BeEquivalentTo(entity.RemediationTypeRiskAccepted.String()))
						}
					})
				})
				It("can filter by 'mitigation' type", func() {
					remediationType := entity.RemediationTypeMitigation.String()
					filter := &entity.RemediationFilter{Type: []*string{&remediationType}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})

					By("returning entries include the type", func() {
						for _, entry := range entries {
							Expect(entry.Type).To(BeEquivalentTo(entity.RemediationTypeMitigation.String()))
						}
					})
				})
				It("can filter by 'rescore' type", func() {
					remediationType := entity.RemediationTypeRescore.String()
					filter := &entity.RemediationFilter{Type: []*string{&remediationType}}

					entries, err := db.GetRemediations(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})

					By("returning entries include the type", func() {
						for _, entry := range entries {
							Expect(entry.Type).To(BeEquivalentTo(entity.RemediationTypeRescore.String()))
						}
					})
				})
				It("can filter by wildcard search", func() {
					row := seedCollection.RemediationRows[rand.Intn(len(seedCollection.RemediationRows))]

					const charactersToRemoveFromBeginning = 2
					const charactersToRemoveFromEnd = 2
					const minimalCharactersToKeep = 5

					start := charactersToRemoveFromBeginning
					end := len(row.Issue.String) - charactersToRemoveFromEnd

					Expect(start+minimalCharactersToKeep < end).To(BeTrue())

					searchStr := row.Issue.String[start:end]
					filter := &entity.RemediationFilter{Search: []*string{&searchStr}}

					entries, err := db.GetRemediations(filter, nil)

					ids := []int64{}
					for _, entry := range entries {
						ids = append(ids, entry.Remediation.Id)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("at least one element was discarded (filtered)", func() {
						Expect(len(seedCollection.IssueRows) > len(ids)).To(BeTrue())
					})

					By("returning the expected elements", func() {
						Expect(ids).To(ContainElement(row.Id.Int64))
					})

				})
			})
			Context("and using pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfListWithOrder(
						db.GetRemediations,
						func(first *int, after *int64, afterX *string) *entity.RemediationFilter {
							return &entity.RemediationFilter{
								PaginatedX: entity.PaginatedX{First: first, After: afterX},
							}
						},
						[]entity.Order{},
						func(entries []entity.RemediationResult) string {
							after, _ := mariadb.EncodeCursor(mariadb.WithRemediation([]entity.Order{}, *entries[len(entries)-1].Remediation))
							return after
						},
						len(seedCollection.RemediationRows),
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
	When("Counting Remediations", Label("CountRemediations"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountRemediations(nil)

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
			var remediationRows []mariadb.RemediationRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				remediationRows = seedCollection.RemediationRows
				count = len(remediationRows)
			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountRemediations(nil)

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
					after := ""
					filter := &entity.RemediationFilter{
						PaginatedX: entity.PaginatedX{
							First: &f,
							After: &after,
						},
					}
					c, err := db.CountRemediations(filter)

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
	When("Insert Remediation", Label("InsertRemediation"), func() {
		Context("and we have 10 Remediations in the database", func() {
			var newRemediationRow mariadb.RemediationRow
			var newRemediation entity.Remediation
			BeforeEach(func() {
				seeder.SeedDbWithNFakeData(10)
				newRemediationRow = mariadb.RemediationRow{
					Type:            sql.NullString{String: entity.RemediationTypeFalsePositive.String(), Valid: true},
					ExpirationDate:  sql.NullTime{Time: time.Now(), Valid: true},
					RemediationDate: sql.NullTime{Time: time.Now(), Valid: true},
					Severity:        sql.NullString{String: "Medium", Valid: true},
					Description:     sql.NullString{String: "New Remediation", Valid: true},
					Service:         sql.NullString{String: "Service", Valid: true},
					ServiceId:       sql.NullInt64{Int64: 1, Valid: true},
					Issue:           sql.NullString{String: "Issue", Valid: true},
					IssueId:         sql.NullInt64{Int64: 1, Valid: true},
					Component:       sql.NullString{String: "Component", Valid: true},
					ComponentId:     sql.NullInt64{Int64: 1, Valid: true},
					RemediatedBy:    sql.NullString{String: "User", Valid: true},
					RemediatedById:  sql.NullInt64{Int64: 1, Valid: true},
					CreatedBy:       sql.NullInt64{Int64: 1, Valid: true},
					UpdatedBy:       sql.NullInt64{Int64: 1, Valid: true},
				}
				newRemediation = newRemediationRow.AsRemediation()
			})
			It("can insert correctly", func() {
				remediation, err := db.CreateRemediation(&newRemediation)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets remediation id", func() {
					Expect(remediation).NotTo(BeEquivalentTo(0))
				})

				remediationFilter := &entity.RemediationFilter{
					Id: []*int64{&remediation.Id},
				}

				r, err := db.GetRemediations(remediationFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning remediation", func() {
					Expect(len(r)).To(BeEquivalentTo(1))
				})
			})
		})
	})
	When("Update Remediation", Label("UpdateRemediation"), func() {
		Context("and we have 10 Remediations in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update description correctly", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()
				remediation.Description = "Updated Description"

				err := db.UpdateRemediation(&remediation)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				remediationFilter := &entity.RemediationFilter{
					Id: []*int64{&remediation.Id},
				}

				r, err := db.GetRemediations(remediationFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning remediation", func() {
					Expect(len(r)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(r[0].Description).To(BeEquivalentTo(remediation.Description))
				})
			})
		})
	})
	When("Delete Remediation", Label("DeleteRemediation"), func() {
		Context("and we have 10 Remediations in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete remediation correctly", func() {
				remediation := seedCollection.RemediationRows[0].AsRemediation()

				err := db.DeleteRemediation(remediation.Id, util.SystemUserId)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				remediationFilter := &entity.RemediationFilter{
					Id: []*int64{&remediation.Id},
				}

				r, err := db.GetRemediations(remediationFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no remediation", func() {
					Expect(len(r)).To(BeEquivalentTo(0))
				})
			})
		})
	})
})
