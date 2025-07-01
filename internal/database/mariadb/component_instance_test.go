// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math"
	"math/rand"
	"sort"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	entityTest "github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("ComponentInstance - ", Label("database", "ComponentInstance"), func() {
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

	When("Getting All ComponentInstance IDs", Label("GetAllComponentInstanceIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				canPerformComponentInstanceQuery(db.GetAllComponentInstanceIds)
			})
		})
		Context("and we have 20 ComponentInstances in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, ci := range seedCollection.ComponentInstanceRows {
					ids = append(ids, ci.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllComponentInstanceIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentInstanceRows)))
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
				It("can filter by a single componentInstance id that does exist", func() {
					ciId := ids[rand.Intn(len(ids))]
					filter := &entity.ComponentInstanceFilter{
						Id: []*int64{&ciId},
					}

					entries, err := db.GetAllComponentInstanceIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(ciId))
					})
				})
				It("can filter by a single service ccrn that does exist", func() {
					ciRow := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]
					serviceRow, _ := lo.Find(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow) bool {
						return s.Id.Int64 == ciRow.ServiceId.Int64
					})
					filter := &entity.ComponentInstanceFilter{
						ServiceCcrn: []*string{&serviceRow.CCRN.String},
					}

					entries, err := db.GetComponentInstances(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(Not(BeZero()))
					})

					for _, entry := range entries {
						Expect(entry.ServiceId).To(BeEquivalentTo(ciRow.ServiceId.Int64))
					}
				})
				It("can filter by a single ccrn that does exist", func() {
					ciRow := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]
					filter := &entity.ComponentInstanceFilter{
						CCRN: []*string{&ciRow.CCRN.String},
					}

					entries, err := db.GetComponentInstances(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(entries).To(HaveLen(1))
					})
					By("returning expected ccrn", func() {
						Expect(entries[0].CCRN).To(BeEquivalentTo(ciRow.CCRN.String))
						Expect(entries[0].Region).To(BeEquivalentTo(ciRow.Region.String))
						Expect(entries[0].Cluster).To(BeEquivalentTo(ciRow.Cluster.String))
						Expect(entries[0].Namespace).To(BeEquivalentTo(ciRow.Namespace.String))
						Expect(entries[0].Domain).To(BeEquivalentTo(ciRow.Domain.String))
						Expect(entries[0].Project).To(BeEquivalentTo(ciRow.Project.String))
						Expect(entries[0].Pod).To(BeEquivalentTo(ciRow.Pod.String))
						Expect(entries[0].Container).To(BeEquivalentTo(ciRow.Container.String))
						Expect(entries[0].Type.String()).To(BeEquivalentTo(ciRow.Type.String))
						Expect((*map[string]interface{})(entries[0].Context)).To(BeEquivalentTo(util.ConvertStrToJsonNoError(&ciRow.Context.String)))
					})

				})
				It("can filter by a single componentVersion id that does exist", func() {
					// select a component version
					cvRow := seedCollection.ComponentVersionRows[rand.Intn(len(seedCollection.ComponentVersionRows))]

					// collect all componentInstance ids that belong to the component version
					ciIds := []int64{}
					for _, ciRow := range seedCollection.ComponentInstanceRows {
						if ciRow.ComponentVersionId.Int64 == cvRow.Id.Int64 {
							ciIds = append(ciIds, ciRow.ServiceId.Int64)
						}
					}

					filter := &entity.ComponentInstanceFilter{
						ComponentVersionId: []*int64{&cvRow.Id.Int64},
					}

					entries, err := db.GetAllComponentInstanceIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected elements", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(ciIds)))
					})
				})
				It("can filter Component Instance Ccrn using wild card search", func() {
					row := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]

					const charactersToRemoveFromBeginning = 2
					const charactersToRemoveFromEnd = 2
					const minimalCharactersToKeep = 2

					start := charactersToRemoveFromBeginning
					end := len(row.CCRN.String) - charactersToRemoveFromEnd

					Expect(start+minimalCharactersToKeep < end).To(BeTrue())

					searchStr := row.CCRN.String[start:end]
					filter := &entity.ComponentInstanceFilter{Search: []*string{&searchStr}}

					entries, err := db.GetComponentInstances(filter, nil)

					ccrn := []string{}
					for _, entry := range entries {
						ccrn = append(ccrn, entry.CCRN)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("at least one element was discarded (filtered)", func() {
						Expect(len(seedCollection.ServiceRows) > len(ccrn)).To(BeTrue())
					})

					By("returning the expected elements", func() {
						Expect(ccrn).To(ContainElement(row.CCRN.String))
					})
				})
			})
		})
	})

	When("Getting ComponentInstances", Label("GetComponentInstance"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetComponentInstances(nil, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 component instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			Context("and using no filter", func() {

				It("can fetch the items correctly", func() {
					res, err := db.GetComponentInstances(nil, nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentInstanceRows)))
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
							for _, row := range seedCollection.ComponentInstanceRows {
								if r.Id == row.Id.Int64 {
									Expect(r.CCRN).Should(BeEquivalentTo(row.CCRN.String), "CCRN matches")
									Expect(r.Region).Should(BeEquivalentTo(row.Region.String), "Region matches")
									Expect(r.Cluster).Should(BeEquivalentTo(row.Cluster.String), "Cluster matches")
									Expect(r.Namespace).Should(BeEquivalentTo(row.Namespace.String), "Namespace matches")
									Expect(r.Domain).Should(BeEquivalentTo(row.Domain.String), "Domain matches")
									Expect(r.Project).Should(BeEquivalentTo(row.Project.String), "Project matches")
									Expect(r.Pod).Should(BeEquivalentTo(row.Pod.String), "Pod matches")
									Expect(r.Container).Should(BeEquivalentTo(row.Container.String), "Container matches")
									Expect(r.Type.String()).Should(BeEquivalentTo(row.Type.String), "Type matches")
									Expect((*map[string]interface{})(r.Context)).To(BeEquivalentTo(util.ConvertStrToJsonNoError(&row.Context.String)), "Context matches")
									Expect(r.Count).Should(BeEquivalentTo(row.Count.Int16), "Count matches")
									Expect(r.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single component instance id that does exist", func() {
					ci := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]
					filter := &entity.ComponentInstanceFilter{
						Id: []*int64{&ci.Id.Int64},
					}

					entries, err := db.GetComponentInstances(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(ci.Id.Int64))
					})
				})
				It("can filter by a single issue match id that does exist", func() {
					//get a service that should return at least one issue
					rnd := seedCollection.IssueMatchRows[rand.Intn(len(seedCollection.IssueMatchRows))]
					ciId := rnd.ComponentInstanceId.Int64
					filter := &entity.ComponentInstanceFilter{
						PaginatedX:   entity.PaginatedX{},
						IssueMatchId: []*int64{&rnd.Id.Int64},
					}

					entries, err := db.GetComponentInstances(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(ciId))
					})

				})
				It("can filter by a single service id that does exist", func() {
					cir := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]
					filter := &entity.ComponentInstanceFilter{
						PaginatedX: entity.PaginatedX{},
						ServiceId:  []*int64{&cir.ServiceId.Int64},
					}

					entries, err := db.GetComponentInstances(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(Not(BeZero()))
					})

					By("returning expected elements", func() {
						for i := range entries {
							Expect(entries[i].ServiceId).To(BeEquivalentTo(cir.ServiceId.Int64))
						}
					})

				})
				It("can filter by a single component version id that does exist", func() {
					cir := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]
					filter := &entity.ComponentInstanceFilter{
						PaginatedX:         entity.PaginatedX{},
						ComponentVersionId: []*int64{&cir.ComponentVersionId.Int64},
					}

					entries, err := db.GetComponentInstances(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(Not(BeZero()))
					})

					By("returning expected elements", func() {
						for i := range entries {
							Expect(entries[i].ComponentVersionId).To(BeEquivalentTo(cir.ComponentVersionId.Int64))
						}
					})

				})
				It("can filter by a single component version version that does exist", func() {
					cir := seedCollection.ComponentInstanceRows[rand.Intn(len(seedCollection.ComponentInstanceRows))]
					cvr, _ := lo.Find(seedCollection.ComponentVersionRows, func(cv mariadb.ComponentVersionRow) bool {
						return cv.Id.Int64 == cir.ComponentVersionId.Int64
					})

					filter := &entity.ComponentInstanceFilter{
						PaginatedX:              entity.PaginatedX{},
						ComponentVersionVersion: []*string{&cvr.Version.String},
					}

					entries, err := db.GetComponentInstances(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(Not(BeZero()))
					})

					By("returning expected elements", func() {
						for i := range entries {
							Expect(entries[i].ComponentVersionId).To(BeEquivalentTo(cir.ComponentVersionId.Int64))
						}
					})

				})
				It("can filter by all existing issue match ids ", func() {
					expectedComponentInstances, ids := seedCollection.GetComponentInstanceByIssueMatches(seedCollection.IssueMatchRows)
					filter := &entity.ComponentInstanceFilter{IssueMatchId: ids}

					entries, err := db.GetComponentInstances(filter, nil)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected result count", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(expectedComponentInstances)))
					})
				})
			})
			Context("and using Pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfListWithOrder(
						db.GetComponentInstances,
						func(first *int, after *int64, afterX *string) *entity.ComponentInstanceFilter {
							return &entity.ComponentInstanceFilter{
								PaginatedX: entity.PaginatedX{First: first, After: afterX},
							}
						},
						[]entity.Order{},
						func(entries []entity.ComponentInstanceResult) string {
							after, _ := mariadb.EncodeCursor(mariadb.WithComponentInstance([]entity.Order{}, *entries[len(entries)-1].ComponentInstance))
							return after
						},
						len(seedCollection.ComponentInstanceRows),
						pageSize,
					)
				},
					Entry("when x is 1", 1),
					Entry("when x is 3", 3),
					Entry("when x is 5", 5),
					Entry("when x is 11", 11),
					Entry("when x is 100", 100),
				)
			})
		})
	})
	When("Counting ComponentInstances", Label("CountComponentInstance"), func() {
		Context("and the database is empty", func() {
			It("returns a correct totalCount without an error", func() {
				c, err := db.CountComponentInstances(nil)

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
			var componentInstanceRows []mariadb.ComponentInstanceRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				componentInstanceRows = seedCollection.GetValidComponentInstanceRows()
				count = len(componentInstanceRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountComponentInstances(nil)

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
					filter := &entity.ComponentInstanceFilter{
						PaginatedX: entity.PaginatedX{
							First: &f,
							After: nil,
						},
						IssueMatchId: nil,
					}
					c, err := db.CountComponentInstances(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(c).To(BeEquivalentTo(count))
					})
				})
			})
			Context("and using a filter", func() {
				DescribeTable("does return totalCount of applied filter", func(pageSize int, filterMatches int) {

					imCol := seedCollection.IssueMatchRows[:filterMatches]
					expectedComponentInstances, ids := seedCollection.GetComponentInstanceByIssueMatches(imCol)

					filter := &entity.ComponentInstanceFilter{
						PaginatedX: entity.PaginatedX{
							First: &pageSize,
							After: nil,
						},
						IssueMatchId: ids,
					}
					entries, err := db.CountComponentInstances(filter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning the correct count", func() {
						Expect(entries).To(BeEquivalentTo(len(expectedComponentInstances)))
					})
					Expect(err).To(BeNil(), "No error should be thrown")
				},
					Entry("when pageSize is 1 and it has 13 elements", 1, 13),
					Entry("when pageSize is 20 and it has 5 elements", 20, 5),
					Entry("when pageSize is 100 and it has 100 elements", 100, 100),
				)
			})
		})

	})
	When("Insert ComponentInstance", Label("InsertComponentInstance"), func() {
		Context("and we have 10 ComponentInstances in the database", func() {
			var newComponentInstanceRow mariadb.ComponentInstanceRow
			var newComponentInstance entity.ComponentInstance
			var componentVersion entity.ComponentVersion
			var service entity.Service
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newComponentInstanceRow = test.NewFakeComponentInstance()
				componentVersion = seedCollection.ComponentVersionRows[0].AsComponentVersion()
				service = seedCollection.ServiceRows[0].AsService()
				newComponentInstance = newComponentInstanceRow.AsComponentInstance()
				newComponentInstance.ComponentVersionId = componentVersion.Id
				newComponentInstance.ServiceId = service.Id
			})
			It("can insert correctly (with ParentID)", func() {
				newComponentInstance.ParentId = 2
				componentInstance, err := db.CreateComponentInstance(&newComponentInstance)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets componentInstance id", func() {
					Expect(componentInstance).NotTo(BeEquivalentTo(0))
				})

				By("setting fields", func() {
					Expect(componentInstance.CCRN).To(BeEquivalentTo(newComponentInstance.CCRN))
					Expect(componentInstance.Region).To(BeEquivalentTo(newComponentInstance.Region))
					Expect(componentInstance.Cluster).To(BeEquivalentTo(newComponentInstance.Cluster))
					Expect(componentInstance.Namespace).To(BeEquivalentTo(newComponentInstance.Namespace))
					Expect(componentInstance.Domain).To(BeEquivalentTo(newComponentInstance.Domain))
					Expect(componentInstance.Project).To(BeEquivalentTo(newComponentInstance.Project))
					Expect(componentInstance.Pod).To(BeEquivalentTo(newComponentInstance.Pod))
					Expect(componentInstance.Container).To(BeEquivalentTo(newComponentInstance.Container))
					Expect(componentInstance.Type.String()).To(BeEquivalentTo(newComponentInstance.Type.String()))
					Expect(componentInstance.Context).To(BeEquivalentTo(newComponentInstance.Context))
					Expect(componentInstance.Count).To(BeEquivalentTo(newComponentInstance.Count))
					Expect(componentInstance.ComponentVersionId).To(BeEquivalentTo(newComponentInstance.ComponentVersionId))
					Expect(componentInstance.ServiceId).To(BeEquivalentTo(newComponentInstance.ServiceId))
					Expect(componentInstance.ParentId).To(BeEquivalentTo(newComponentInstance.ParentId))
				})
			})
			It("can insert correctly (without ParentID)", func() {
				componentInstance, err := db.CreateComponentInstance(&newComponentInstance)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets componentInstance id", func() {
					Expect(componentInstance).NotTo(BeEquivalentTo(0))
				})

				By("setting fields", func() {
					Expect(componentInstance.CCRN).To(BeEquivalentTo(newComponentInstance.CCRN))
					Expect(componentInstance.Region).To(BeEquivalentTo(newComponentInstance.Region))
					Expect(componentInstance.Cluster).To(BeEquivalentTo(newComponentInstance.Cluster))
					Expect(componentInstance.Namespace).To(BeEquivalentTo(newComponentInstance.Namespace))
					Expect(componentInstance.Domain).To(BeEquivalentTo(newComponentInstance.Domain))
					Expect(componentInstance.Project).To(BeEquivalentTo(newComponentInstance.Project))
					Expect(componentInstance.Pod).To(BeEquivalentTo(newComponentInstance.Pod))
					Expect(componentInstance.Container).To(BeEquivalentTo(newComponentInstance.Container))
					Expect(componentInstance.Type.String()).To(BeEquivalentTo(newComponentInstance.Type.String()))
					Expect(componentInstance.Count).To(BeEquivalentTo(newComponentInstance.Count))
					Expect(componentInstance.ComponentVersionId).To(BeEquivalentTo(newComponentInstance.ComponentVersionId))
					Expect(componentInstance.ServiceId).To(BeEquivalentTo(newComponentInstance.ServiceId))
					Expect(componentInstance.ParentId).To(BeEquivalentTo(newComponentInstance.ParentId))
				})
			})
		})
	})
	When("Update ComponentInstance", Label("UpdateComponentInstance"), func() {
		Context("and we have 10 ComponentInstances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

			})
			It("can update componentInstance count correctly", func() {
				componentInstance := seedCollection.ComponentInstanceRows[0].AsComponentInstance()
				componentInstance.ParentId = 1
				componentInstance.Count = componentInstance.Count + 1
				err := db.UpdateComponentInstance(&componentInstance)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				componentInstanceFilter := &entity.ComponentInstanceFilter{
					Id: []*int64{&componentInstance.Id},
				}

				ci, err := db.GetComponentInstances(componentInstanceFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning componentInstance", func() {
					Expect(len(ci)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(ci[0].CCRN).To(BeEquivalentTo(componentInstance.CCRN))
					Expect(ci[0].Region).To(BeEquivalentTo(componentInstance.Region))
					Expect(ci[0].Cluster).To(BeEquivalentTo(componentInstance.Cluster))
					Expect(ci[0].Namespace).To(BeEquivalentTo(componentInstance.Namespace))
					Expect(ci[0].Domain).To(BeEquivalentTo(componentInstance.Domain))
					Expect(ci[0].Project).To(BeEquivalentTo(componentInstance.Project))
					Expect(ci[0].Pod).To(BeEquivalentTo(componentInstance.Pod))
					Expect(ci[0].Container).To(BeEquivalentTo(componentInstance.Container))
					Expect(ci[0].Type.String()).To(BeEquivalentTo(componentInstance.Type.String()))
					Expect(ci[0].Context).To(BeEquivalentTo(componentInstance.Context))
					Expect(ci[0].Count).To(BeEquivalentTo(componentInstance.Count))
					Expect(ci[0].ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
					Expect(ci[0].ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
					Expect(ci[0].ParentId).To(BeEquivalentTo(componentInstance.ParentId))
				})
			})
			It("can update componentInstance fields correctly", func() {
				componentInstance := seedCollection.ComponentInstanceRows[0].AsComponentInstance()

				newComponentInstanceValues := entityTest.NewFakeComponentInstanceEntity()
				newComponentInstanceValues.Id = componentInstance.Id
				newComponentInstanceValues.ComponentVersionId = componentInstance.ComponentVersionId
				newComponentInstanceValues.ServiceId = componentInstance.ServiceId
				newComponentInstanceValues.ParentId = componentInstance.ParentId
				err := db.UpdateComponentInstance(&newComponentInstanceValues)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				componentInstanceFilter := &entity.ComponentInstanceFilter{
					Id: []*int64{&componentInstance.Id},
				}

				ci, err := db.GetComponentInstances(componentInstanceFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning componentInstance", func() {
					Expect(len(ci)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(ci[0].CCRN).To(BeEquivalentTo(newComponentInstanceValues.CCRN))
					Expect(ci[0].Region).To(BeEquivalentTo(newComponentInstanceValues.Region))
					Expect(ci[0].Cluster).To(BeEquivalentTo(newComponentInstanceValues.Cluster))
					Expect(ci[0].Namespace).To(BeEquivalentTo(newComponentInstanceValues.Namespace))
					Expect(ci[0].Domain).To(BeEquivalentTo(newComponentInstanceValues.Domain))
					Expect(ci[0].Project).To(BeEquivalentTo(newComponentInstanceValues.Project))
					Expect(ci[0].Pod).To(BeEquivalentTo(newComponentInstanceValues.Pod))
					Expect(ci[0].Container).To(BeEquivalentTo(newComponentInstanceValues.Container))
					Expect(ci[0].Type.String()).To(BeEquivalentTo(newComponentInstanceValues.Type.String()))
					Expect(ci[0].Context.String()).To(BeEquivalentTo(newComponentInstanceValues.Context.String()))
					Expect(ci[0].Count).To(BeEquivalentTo(newComponentInstanceValues.Count))
					Expect(ci[0].ComponentVersionId).To(BeEquivalentTo(newComponentInstanceValues.ComponentVersionId))
					Expect(ci[0].ServiceId).To(BeEquivalentTo(newComponentInstanceValues.ServiceId))
					Expect(ci[0].ParentId).To(BeEquivalentTo(newComponentInstanceValues.ParentId))
				})
			})
		})
	})
	When("Delete ComponentInstance", Label("DeleteComponentInstance"), func() {
		Context("and we have 10 ComponentInstances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete componentInstance correctly", func() {
				componentInstance := seedCollection.ComponentInstanceRows[0].AsComponentInstance()

				err := db.DeleteComponentInstance(componentInstance.Id, systemUserId)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				componentInstanceFilter := &entity.ComponentInstanceFilter{
					Id: []*int64{&componentInstance.Id},
				}

				ci, err := db.GetComponentInstances(componentInstanceFilter, nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no service", func() {
					Expect(len(ci)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Getting CCRN", Label("GetCcrn"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetCcrn)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			var testOrder = func(
				order []entity.Order,
				verifyFunc func(res []entity.ComponentInstanceResult),
			) {
				res, err := db.GetComponentInstances(nil, order)

				By("throwing no error", func() {
					Expect(err).Should(BeNil())
				})

				By("returning the correct number of results", func() {
					Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentInstanceRows)))
				})

				By("returning the correct order", func() {
					verifyFunc(res)
				})
			}

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedCcrns := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.CCRN.String
					})
					canFetchComponentInstanceQueryItems(db.GetCcrn, expectedCcrns)
				})
			})
			Context("and using a CCRN filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir := seedCollection.GetComponentInstance()
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{CCRN: []*string{&cir.CCRN.String}},
						[]string{cir.CCRN.String},
					)
				})
				It("and using notexisting value returns an empty list when no CCRN match the filter", func() {
					notexistentCcrn := "NotexistentCCRN"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{CCRN: []*string{&notexistentCcrn}},
						[]string{},
					)
				})
			})
			Context("and using a Region filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Region.String == picked.Region.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Region: []*string{&cir.Region.String}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Region match the filter", func() {
					notexistentRegion := "NotexistentRegion"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{CCRN: []*string{&notexistentRegion}},
						[]string{},
					)
				})
			})
			Context("and using a Cluster filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Cluster.String == picked.Cluster.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Cluster: []*string{&cir.Cluster.String}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Cluster match the filter", func() {
					notexistentCluster := "NotexistentCluster"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Cluster: []*string{&notexistentCluster}},
						[]string{},
					)
				})
			})
			Context("and using a Namespace filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Namespace.String == picked.Namespace.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Namespace: []*string{&cir.Namespace.String}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Namespace match the filter", func() {
					notexistentNamespace := "NotexistentNamespace"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Namespace: []*string{&notexistentNamespace}},
						[]string{},
					)
				})
			})
			Context("and using a Domain filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Domain.String == picked.Domain.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Domain: []*string{&cir.Domain.String}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Domain match the filter", func() {
					notexistentDomain := "NotexistentDomain"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Domain: []*string{&notexistentDomain}},
						[]string{},
					)
				})
			})
			Context("and using a Project filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Project.String == picked.Project.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Project: []*string{&cir.Project.String}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Project match the filter", func() {
					notexistentProject := "NotexistentProject"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Project: []*string{&notexistentProject}},
						[]string{},
					)
				})
			})
			Context("and using a Pod filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Pod.String == picked.Pod.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Pod: []*string{&cir.Pod.String}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Pod match the filter", func() {
					notexistentPod := "NotexistentPod"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Pod: []*string{&notexistentPod}},
						[]string{},
					)
				})
			})
			Context("and using a Container filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Container.String == picked.Container.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Container: []*string{&cir.Container.String}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Container match the filter", func() {
					notexistentContainer := "NotexistentContainer"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Container: []*string{&notexistentContainer}},
						[]string{},
					)
				})
			})
			Context("and using a Type filter", func() {
				It("using existing value can fetch the filterd items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Type.String == picked.Type.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Type: []*string{&cir.Type.String}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Type match the filter", func() {
					notexistentType := "NotexistentType"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Type: []*string{&notexistentType}},
						[]string{},
					)
				})
			})
			Context("and using a Context filter", func() {
				It("using existing value can fetch the filterd items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, (*util.ConvertStrToJsonNoError(&iter.Context.String))["my_ip"] == (*util.ConvertStrToJsonNoError(&picked.Context.String))["my_ip"]
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Context: []*entity.Json{{
							"my_ip": (*util.ConvertStrToJsonNoError(&cir.Context.String))["my_ip"],
						}}},
						expectedCcrns,
					)
				})
				It("using multiple existing values can fetch the filterd items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							iterContext := *util.ConvertStrToJsonNoError(&iter.Context.String)
							pickedContext := *util.ConvertStrToJsonNoError(&picked.Context.String)
							return iter.CCRN.String, iterContext["my_ip"] == pickedContext["my_ip"] && iterContext["remove_unused_base_images"] == pickedContext["remove_unused_base_images"]
						},
					)
					cirContext := *util.ConvertStrToJsonNoError(&cir.Context.String)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Context: []*entity.Json{{
							"my_ip":                     cirContext["my_ip"],
							"remove_unused_base_images": cirContext["remove_unused_base_images"],
						}}},
						expectedCcrns,
					)
				})
				It("and using notexisting value returns an empty list when no Type match the filter", func() {
					notexistentContextAttr := map[string]interface{}{
						"not_real": "value",
					}
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{Context: []*entity.Json{(*entity.Json)(&notexistentContextAttr)}},
						[]string{},
					)
				})
			})
			Context("and using multiple filter attributes", func() {
				It("using existing values of CCRN attributes can fetch the filtered items correctly", func() {
					cir, expectedCcrns := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.CCRN.String, iter.Pod.String == picked.Pod.String &&
								iter.Container.String == picked.Container.String &&
								iter.Type.String == picked.Type.String &&
								iter.Context.String == picked.Context.String &&
								iter.Project.String == picked.Project.String &&
								iter.Domain.String == picked.Domain.String &&
								iter.Namespace.String == picked.Namespace.String &&
								iter.Cluster.String == picked.Cluster.String &&
								iter.Region.String == picked.Region.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{
							Region:    []*string{&cir.Region.String},
							Cluster:   []*string{&cir.Cluster.String},
							Namespace: []*string{&cir.Namespace.String},
							Domain:    []*string{&cir.Domain.String},
							Project:   []*string{&cir.Project.String},
							Pod:       []*string{&cir.Pod.String},
							Container: []*string{&cir.Container.String},
							Type:      []*string{&cir.Type.String},
							Context:   []*entity.Json{(*entity.Json)(util.ConvertStrToJsonNoError(&cir.Context.String))},
						},
						expectedCcrns,
					)
				})
				It("using one notexisting value of all CCRN attributes returns an empty list", func() {
					cir := seedCollection.GetComponentInstance()
					notexistentProject := "NotexistentProject"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCcrn,
						&entity.ComponentInstanceFilter{
							Region:    []*string{&cir.Region.String},
							Cluster:   []*string{&cir.Cluster.String},
							Namespace: []*string{&cir.Namespace.String},
							Domain:    []*string{&cir.Domain.String},
							Project:   []*string{&notexistentProject},
							Pod:       []*string{&cir.Pod.String},
							Container: []*string{&cir.Container.String},
							Type:      []*string{&cir.Type.String},
							Context:   []*entity.Json{(*entity.Json)(util.ConvertStrToJsonNoError(&cir.Context.String))},
						},
						[]string{},
					)
				})
			})
			Context("and using asc order", func() {
				It("can order by id", func() {
					sort.Slice(seedCollection.ComponentInstanceRows, func(i, j int) bool {
						return seedCollection.ComponentInstanceRows[i].Id.Int64 < seedCollection.ComponentInstanceRows[j].Id.Int64
					})

					order := []entity.Order{
						{By: entity.ComponentInstanceId, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						for i, r := range res {
							Expect(r.Id).Should(BeEquivalentTo(seedCollection.ComponentInstanceRows[i].Id.Int64))
						}
					})
				})
				It("can order by ccrn", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = ""
						for _, r := range res {
							Expect(r.ComponentInstance.CCRN >= prev).Should(BeTrue())
							prev = r.ComponentInstance.CCRN
						}
					})
				})
				It("can order by region", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceRegion, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = ""
						for _, r := range res {
							Expect(r.ComponentInstance.Region >= prev).Should(BeTrue())
							prev = r.ComponentInstance.Region
						}
					})
				})
				It("can order by namespace", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceNamespace, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = ""
						for _, r := range res {
							Expect(r.ComponentInstance.Namespace >= prev).Should(BeTrue())
							prev = r.ComponentInstance.Namespace
						}
					})
				})
				It("can order by cluster", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceCluster, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = ""
						for _, r := range res {
							Expect(r.ComponentInstance.Cluster >= prev).Should(BeTrue())
							prev = r.ComponentInstance.Cluster
						}
					})
				})
				It("can order by domain", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceDomain, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = ""
						for _, r := range res {
							Expect(r.ComponentInstance.Domain >= prev).Should(BeTrue())
							prev = r.ComponentInstance.Domain
						}
					})
				})
				It("can order by project", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceProject, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = ""
						for _, r := range res {
							Expect(r.ComponentInstance.Project >= prev).Should(BeTrue())
							prev = r.ComponentInstance.Project
						}
					})
				})
				It("can order by pod", func() {
					order := []entity.Order{
						{By: entity.ComponentInstancePod, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = ""
						for _, r := range res {
							Expect(r.ComponentInstance.Pod >= prev).Should(BeTrue())
							prev = r.ComponentInstance.Pod
						}
					})
				})
				It("can order by container", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceContainer, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = ""
						for _, r := range res {
							Expect(r.ComponentInstance.Container >= prev).Should(BeTrue())
							prev = r.ComponentInstance.Container
						}
					})
				})
				It("can order by type", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceTypeOrder, Direction: entity.OrderDirectionAsc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev int = -1
						for _, r := range res {
							Expect(r.ComponentInstance.Type.Index() >= prev).Should(BeTrue())
							prev = r.ComponentInstance.Type.Index()
						}
					})
				})
			})
			Context("and using desc order", func() {
				It("can order by id", func() {
					sort.Slice(seedCollection.ComponentInstanceRows, func(i, j int) bool {
						return seedCollection.ComponentInstanceRows[i].Id.Int64 > seedCollection.ComponentInstanceRows[j].Id.Int64
					})

					order := []entity.Order{
						{By: entity.ComponentInstanceId, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						for i, r := range res {
							Expect(r.Id).Should(BeEquivalentTo(seedCollection.ComponentInstanceRows[i].Id.Int64))
						}
					})
				})
				It("can order by ccrn", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = "\U0010FFFF"
						for _, r := range res {
							Expect(r.ComponentInstance.CCRN <= prev).Should(BeTrue())
							prev = r.ComponentInstance.CCRN
						}
					})
				})
				It("can order by region", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceRegion, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = "\U0010FFFF"
						for _, r := range res {
							Expect(r.ComponentInstance.Region <= prev).Should(BeTrue())
							prev = r.ComponentInstance.Region
						}
					})
				})
				It("can order by namespace", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceNamespace, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = "\U0010FFFF"
						for _, r := range res {
							Expect(r.ComponentInstance.Namespace <= prev).Should(BeTrue())
							prev = r.ComponentInstance.Namespace
						}
					})
				})
				It("can order by cluster", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceCluster, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = "\U0010FFFF"
						for _, r := range res {
							Expect(r.ComponentInstance.Cluster <= prev).Should(BeTrue())
							prev = r.ComponentInstance.Cluster
						}
					})
				})
				It("can order by domain", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceDomain, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = "\U0010FFFF"
						for _, r := range res {
							Expect(r.ComponentInstance.Domain <= prev).Should(BeTrue())
							prev = r.ComponentInstance.Domain
						}
					})
				})
				It("can order by project", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceProject, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = "\U0010FFFF"
						for _, r := range res {
							Expect(r.ComponentInstance.Project <= prev).Should(BeTrue())
							prev = r.ComponentInstance.Project
						}
					})
				})
				It("can order by pod", func() {
					order := []entity.Order{
						{By: entity.ComponentInstancePod, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = "\U0010FFFF"
						for _, r := range res {
							Expect(r.ComponentInstance.Pod <= prev).Should(BeTrue())
							prev = r.ComponentInstance.Pod
						}
					})
				})
				It("can order by container", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceContainer, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev string = "\U0010FFFF"
						for _, r := range res {
							Expect(r.ComponentInstance.Container <= prev).Should(BeTrue())
							prev = r.ComponentInstance.Container
						}
					})
				})
				It("can order by type", func() {
					order := []entity.Order{
						{By: entity.ComponentInstanceTypeOrder, Direction: entity.OrderDirectionDesc},
					}

					testOrder(order, func(res []entity.ComponentInstanceResult) {
						var prev int = math.MaxInt
						for _, r := range res {
							Expect(r.ComponentInstance.Type.Index() <= prev).Should(BeTrue())
							prev = r.ComponentInstance.Type.Index()
						}
					})
				})
			})
		})
	})
	When("Getting Region", Label("GetRegion"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetRegion)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedRegions := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Region.String
					})
					canFetchComponentInstanceQueryItems(db.GetRegion, expectedRegions)
				})
			})
			Context("and using a Cluster filter for Region", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedRegions := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Region.String, iter.Cluster.String == picked.Cluster.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetRegion,
						&entity.ComponentInstanceFilter{Cluster: []*string{&cir.Cluster.String}},
						expectedRegions,
					)
				})
				It("and using notexisting value returns an empty list when no Region match the filter", func() {
					notexistentRegion := "NotexistentRegion"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetRegion,
						&entity.ComponentInstanceFilter{Region: []*string{&notexistentRegion}},
						[]string{},
					)
				})
			})
		})
	})
	When("Getting Cluster", Label("GetCluster"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetCluster)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedClusters := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Cluster.String
					})
					canFetchComponentInstanceQueryItems(db.GetCluster, expectedClusters)
				})
			})
			Context("and using a Namespace filter for Cluster", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedClusters := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Cluster.String, iter.Namespace.String == picked.Namespace.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCluster,
						&entity.ComponentInstanceFilter{Namespace: []*string{&cir.Namespace.String}},
						expectedClusters,
					)
				})
				It("and using notexisting value returns an empty list when no Cluster match the filter", func() {
					notexistentCluster := "NotexistentCluster"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetCluster,
						&entity.ComponentInstanceFilter{Cluster: []*string{&notexistentCluster}},
						[]string{},
					)
				})
			})
		})
	})
	When("Getting Namespace", Label("GetNamespace"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetNamespace)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedNamespaces := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Namespace.String
					})
					canFetchComponentInstanceQueryItems(db.GetNamespace, expectedNamespaces)
				})
			})
			Context("and using a Domain filter for Namespace", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedNamespaces := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Namespace.String, iter.Domain.String == picked.Domain.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetNamespace,
						&entity.ComponentInstanceFilter{Domain: []*string{&cir.Domain.String}},
						expectedNamespaces,
					)
				})
				It("and using notexisting value returns an empty list when no Namespace match the filter", func() {
					notexistentNamespace := "NotexistentNamespace"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetNamespace,
						&entity.ComponentInstanceFilter{Namespace: []*string{&notexistentNamespace}},
						[]string{},
					)
				})
			})
		})
	})
	When("Getting Domain", Label("GetDomain"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetDomain)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedDomains := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Domain.String
					})
					canFetchComponentInstanceQueryItems(db.GetDomain, expectedDomains)
				})
			})
			Context("and using a Project filter for Domain", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedDomains := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Domain.String, iter.Project.String == picked.Project.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetDomain,
						&entity.ComponentInstanceFilter{Project: []*string{&cir.Project.String}},
						expectedDomains,
					)
				})
				It("and using notexisting value returns an empty list when no Domain match the filter", func() {
					notexistentDomain := "NotexistentDomain"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetDomain,
						&entity.ComponentInstanceFilter{Domain: []*string{&notexistentDomain}},
						[]string{},
					)
				})
			})
		})
	})
	When("Getting Project", Label("GetProject"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetProject)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedProjects := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Project.String
					})
					canFetchComponentInstanceQueryItems(db.GetProject, expectedProjects)
				})
			})
			Context("and using a Pod filter for Project", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedProjects := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Project.String, iter.Pod.String == picked.Pod.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetProject,
						&entity.ComponentInstanceFilter{Pod: []*string{&cir.Pod.String}},
						expectedProjects,
					)
				})
				It("and using notexisting value filter returns an empty list when no Project match the filter", func() {
					notexistentProject := "NotexistentProject"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetProject,
						&entity.ComponentInstanceFilter{Project: []*string{&notexistentProject}},
						[]string{},
					)
				})
			})
		})
	})
	When("Getting Pod", Label("GetPod"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetPod)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedPods := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Pod.String
					})
					canFetchComponentInstanceQueryItems(db.GetPod, expectedPods)
				})
			})
			Context("and using a Container filter for Pod", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedPods := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Pod.String, iter.Container.String == picked.Container.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetPod,
						&entity.ComponentInstanceFilter{Container: []*string{&cir.Container.String}},
						expectedPods,
					)
				})
				It("and using notexisting value filter returns an empty list when no Pod match the filter", func() {
					notexistentPod := "NotexistentPod"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetPod,
						&entity.ComponentInstanceFilter{Pod: []*string{&notexistentPod}},
						[]string{},
					)
				})
			})
		})
	})
	When("Getting Container", Label("GetContainer"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetContainer)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedContainers := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Container.String
					})
					canFetchComponentInstanceQueryItems(db.GetContainer, expectedContainers)
				})
			})
			Context("and using a Type filter for Container", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedContainers := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Container.String, iter.Type.String == picked.Type.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetContainer,
						&entity.ComponentInstanceFilter{Type: []*string{&cir.Type.String}},
						expectedContainers,
					)
				})
				It("and using notexisting value filter returns an empty list when no Container match the filter", func() {
					notexistentContainer := "NotexistentContainer"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetContainer,
						&entity.ComponentInstanceFilter{Container: []*string{&notexistentContainer}},
						[]string{},
					)
				})
			})
		})
	})
	When("Getting Type", Label("GetType"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetType)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedTypes := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Type.String
					})
					canFetchComponentInstanceQueryItems(db.GetType, expectedTypes)
				})
			})
			Context("and using a Context filter for Type", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedTypes := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Type.String, iter.Context.String == picked.Context.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetType,
						&entity.ComponentInstanceFilter{Context: []*entity.Json{(*entity.Json)(util.ConvertStrToJsonNoError(&cir.Context.String))}},
						expectedTypes,
					)
				})
				It("and using notexisting value filter returns an empty list when no Type match the filter", func() {
					notexistentType := "NotexistentType"
					issueComponentInstanceAttrFilterWithExpect(
						db.GetType,
						&entity.ComponentInstanceFilter{Type: []*string{&notexistentType}},
						[]string{},
					)
				})
			})
		})
	})

	When("Getting Context", Label("GetContext"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				canPerformComponentInstanceQuery(db.GetContext)
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					expectedContexts := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
						return cir.Context.String
					})
					canFetchComponentInstanceQueryItems(db.GetContext, expectedContexts)
				})
			})
			Context("and using a Region filter for Context", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir, expectedContexts := seedCollection.GetComponentInstanceWithPredicateVal(
						func(picked, iter mariadb.ComponentInstanceRow) (string, bool) {
							return iter.Context.String, iter.Region.String == picked.Region.String
						},
					)
					issueComponentInstanceAttrFilterWithExpect(
						db.GetContext,
						&entity.ComponentInstanceFilter{Region: []*string{&cir.Region.String}},
						expectedContexts,
					)
				})
				It("and using notexisting value filter returns an empty list when no Context match the filter", func() {
					notexistentContext := entity.Json{"not_real": "value"}
					issueComponentInstanceAttrFilterWithExpect(
						db.GetContext,
						&entity.ComponentInstanceFilter{Context: []*entity.Json{&notexistentContext}},
						[]string{},
					)
				})
			})
		})
	})
})

func canPerformComponentInstanceQuery[T any](getFunc func(filter *entity.ComponentInstanceFilter) ([]T, error)) {
	res, err := getFunc(nil)
	By("throwing no error", func() {
		Expect(err).To(BeNil())
	})
	By("returning an empty list", func() {
		Expect(res).To(BeEmpty())
	})
}

func canFetchComponentInstanceQueryItems(
	getFunc func(filter *entity.ComponentInstanceFilter) ([]string, error),
	expectedItems []string) {
	res, err := getFunc(nil)

	By("throwing no error", func() {
		Expect(err).Should(BeNil())
	})

	By("returning the correct number of results", func() {
		Expect(len(res)).Should(BeIdenticalTo(len(expectedItems)))
	})

	By("returning the correct items", func() {
		left, right := lo.Difference(res, expectedItems)
		Expect(left).Should(BeEmpty())
		Expect(right).Should(BeEmpty())
	})
}

func issueComponentInstanceAttrFilterWithExpect(
	getAttrFunc func(filter *entity.ComponentInstanceFilter) ([]string, error),
	cifilter *entity.ComponentInstanceFilter,
	expectedAttrVal []string) {
	res, err := getAttrFunc(cifilter)
	By("throwing no error", func() {
		Expect(err).Should(BeNil())
	})

	By("returning the correct number of results", func() {
		Expect(len(res)).Should(BeEquivalentTo(len(expectedAttrVal)))
	})

	By("returning the correct names", func() {
		left, right := lo.Difference(res, expectedAttrVal)
		Expect(left).Should(BeEmpty())
		Expect(right).Should(BeEmpty())
	})
}
