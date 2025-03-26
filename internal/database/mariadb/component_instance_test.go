// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"math/rand"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
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

	When("Getting All ComponentInstance IDs", Label("GetAllComponentInstanceIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllComponentInstanceIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
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

					entries, err := db.GetComponentInstances(filter)

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

					entries, err := db.GetComponentInstances(filter)

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

					entries, err := db.GetComponentInstances(filter)

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
				res, err := db.GetComponentInstances(nil)
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
					res, err := db.GetComponentInstances(nil)

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

					entries, err := db.GetComponentInstances(filter)

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
						Paginated:    entity.Paginated{},
						IssueMatchId: []*int64{&rnd.Id.Int64},
					}

					entries, err := db.GetComponentInstances(filter)

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
						Paginated: entity.Paginated{},
						ServiceId: []*int64{&cir.ServiceId.Int64},
					}

					entries, err := db.GetComponentInstances(filter)

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
				It("can filter by all existing issue match ids ", func() {
					expectedComponentInstances, ids := seedCollection.GetComponentInstanceByIssueMatches(seedCollection.IssueMatchRows)
					filter := &entity.ComponentInstanceFilter{IssueMatchId: ids}

					entries, err := db.GetComponentInstances(filter)

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
					test.TestPaginationOfList(
						db.GetComponentInstances,
						func(first *int, after *int64) *entity.ComponentInstanceFilter {
							return &entity.ComponentInstanceFilter{
								Paginated: entity.Paginated{
									First: first,
									After: after,
								},
							}
						},
						func(entries []entity.ComponentInstance) *int64 { return &entries[len(entries)-1].Id },
						10,
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
						Paginated: entity.Paginated{
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
						Paginated: entity.Paginated{
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
			It("can insert correctly", func() {
				componentInstance, err := db.CreateComponentInstance(&newComponentInstance)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets componentInstance id", func() {
					Expect(componentInstance).NotTo(BeEquivalentTo(0))
				})

				componentInstanceFilter := &entity.ComponentInstanceFilter{
					Id: []*int64{&componentInstance.Id},
				}

				ci, err := db.GetComponentInstances(componentInstanceFilter)
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
					Expect(ci[0].Count).To(BeEquivalentTo(componentInstance.Count))
					Expect(ci[0].ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
					Expect(ci[0].ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
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

				componentInstance.Count = componentInstance.Count + 1
				err := db.UpdateComponentInstance(&componentInstance)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				componentInstanceFilter := &entity.ComponentInstanceFilter{
					Id: []*int64{&componentInstance.Id},
				}

				ci, err := db.GetComponentInstances(componentInstanceFilter)
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
					Expect(ci[0].Count).To(BeEquivalentTo(componentInstance.Count))
					Expect(ci[0].ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
					Expect(ci[0].ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
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

				ci, err := db.GetComponentInstances(componentInstanceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no service", func() {
					Expect(len(ci)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Getting Component Instances", Label("GetComponentInstances"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetCcrn(nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 Component Instances in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetCcrn(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentInstanceRows)))
					})

					existingCCRN := lo.Map(seedCollection.ComponentInstanceRows, func(s mariadb.ComponentInstanceRow, index int) string {
						return s.CCRN.String
					})

					By("returning the correct CCRN", func() {
						left, right := lo.Difference(res, existingCCRN)
						Expect(left).Should(BeEmpty())
						Expect(right).Should(BeEmpty())
					})
				})
			})
			Context("and using a CCRN filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir := seedCollection.ComponentInstanceRows[0]
					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{CCRN: []*string{&cir.CCRN.String}},
						[]string{cir.CCRN.String},
					)
				})
				It("and using notexisting value returns an empty list when no CCRN match the filter", func() {
					notexistentCcrn := "NotexistentCCRN"
					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{CCRN: []*string{&notexistentCcrn}},
						[]string{},
					)
				})
			})
			Context("and using a Region filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir := seedCollection.ComponentInstanceRows[0]

					filteredSeed := lo.FilterMap(
						seedCollection.ComponentInstanceRows,
						func(s mariadb.ComponentInstanceRow, index int) (string, bool) {
							return s.CCRN.String, s.Region.String == cir.Region.String
						})

					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Region: []*string{&cir.Region.String}},
						filteredSeed,
					)
				})
				It("and using notexisting value returns an empty list when no Region match the filter", func() {
					notexistentRegion := "NotexistentRegion"
					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{CCRN: []*string{&notexistentRegion}},
						[]string{},
					)
				})
			})
			Context("and using a Cluster filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir := seedCollection.ComponentInstanceRows[0]

					filteredSeed := lo.FilterMap(
						seedCollection.ComponentInstanceRows,
						func(s mariadb.ComponentInstanceRow, index int) (string, bool) {
							return s.CCRN.String, s.Cluster.String == cir.Cluster.String
						})

					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Cluster: []*string{&cir.Cluster.String}},
						filteredSeed,
					)
				})
				It("and using notexisting value returns an empty list when no Cluster match the filter", func() {
					notexistentCluster := "NotexistentCluster"
					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Cluster: []*string{&notexistentCluster}},
						[]string{},
					)
				})
			})
			Context("and using a Namespace filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir := seedCollection.ComponentInstanceRows[0]

					filteredSeed := lo.FilterMap(
						seedCollection.ComponentInstanceRows,
						func(s mariadb.ComponentInstanceRow, index int) (string, bool) {
							return s.CCRN.String, s.Namespace.String == cir.Namespace.String
						})

					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Namespace: []*string{&cir.Namespace.String}},
						filteredSeed,
					)
				})
				It("and using notexisting value returns an empty list when no Namespace match the filter", func() {
					notexistentNamespace := "NotexistentNamespace"
					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Namespace: []*string{&notexistentNamespace}},
						[]string{},
					)
				})
			})
			Context("and using a Domain filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir := seedCollection.ComponentInstanceRows[0]

					filteredSeed := lo.FilterMap(
						seedCollection.ComponentInstanceRows,
						func(s mariadb.ComponentInstanceRow, index int) (string, bool) {
							return s.CCRN.String, s.Domain.String == cir.Domain.String
						})

					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Domain: []*string{&cir.Domain.String}},
						filteredSeed,
					)
				})
				It("and using notexisting value returns an empty list when no Domain match the filter", func() {
					notexistentDomain := "NotexistentDomain"
					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Domain: []*string{&notexistentDomain}},
						[]string{},
					)
				})
			})
			Context("and using a Project filter", func() {
				It("using existing value can fetch the filtered items correctly", func() {
					cir := seedCollection.ComponentInstanceRows[0]

					filteredSeed := lo.FilterMap(
						seedCollection.ComponentInstanceRows,
						func(s mariadb.ComponentInstanceRow, index int) (string, bool) {
							return s.CCRN.String, s.Project.String == cir.Project.String
						})

					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Project: []*string{&cir.Project.String}},
						filteredSeed,
					)
				})
				It("and using notexisting value returns an empty list when no Project match the filter", func() {
					notexistentProject := "NotexistentProject"
					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{Project: []*string{&notexistentProject}},
						[]string{},
					)
				})
			})
			Context("and using multiple filter attributes", func() {
				It("using existing values of CCRN attributes can fetch the filtered items correctly", func() {

					cir := seedCollection.ComponentInstanceRows[0]

					filteredSeed := lo.FilterMap(
						seedCollection.ComponentInstanceRows,
						func(s mariadb.ComponentInstanceRow, index int) (string, bool) {
							return s.CCRN.String,
								s.Project.String == cir.Project.String &&
									s.Domain.String == cir.Domain.String &&
									s.Namespace.String == cir.Namespace.String &&
									s.Cluster.String == cir.Cluster.String &&
									s.Region.String == cir.Region.String
						})

					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{
							Region:    []*string{&cir.Region.String},
							Cluster:   []*string{&cir.Cluster.String},
							Namespace: []*string{&cir.Namespace.String},
							Domain:    []*string{&cir.Domain.String},
							Project:   []*string{&cir.Project.String},
						},
						filteredSeed,
					)
				})
				It("using one notexisting value of all CCRN attributes returns an empty list", func() {
					cir := seedCollection.ComponentInstanceRows[0]
					notexistentProject := "NotexistentProject"
					issueComponentInstanceFilterWithExpectCcrn(
						db,
						&entity.ComponentInstanceFilter{
							Region:    []*string{&cir.Region.String},
							Cluster:   []*string{&cir.Cluster.String},
							Namespace: []*string{&cir.Namespace.String},
							Domain:    []*string{&cir.Domain.String},
							Project:   []*string{&notexistentProject},
						},
						[]string{},
					)
				})
			})
		})
	})
})

func issueComponentInstanceFilterWithExpectCcrn(db *mariadb.SqlDatabase, cifilter *entity.ComponentInstanceFilter, expectedCcrn []string) {
	res, err := db.GetCcrn(cifilter)
	By("throwing no error", func() {
		Expect(err).Should(BeNil())
	})

	By("returning the correct number of results", func() {
		Expect(len(res)).Should(BeEquivalentTo(len(expectedCcrn)))
	})

	By("returning the correct names", func() {
		left, right := lo.Difference(res, expectedCcrn)
		Expect(left).Should(BeEmpty())
		Expect(right).Should(BeEmpty())
	})
}
