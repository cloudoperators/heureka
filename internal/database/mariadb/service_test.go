// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"math/rand"

	"github.com/cloudoperators/heureka/pkg/util"
)

var _ = Describe("Service", Label("database", "Service"), func() {

	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Getting All Service IDs", Label("GetAllServiceIds"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetAllServiceIds(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 20 Services in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)

				for _, s := range seedCollection.ServiceRows {
					ids = append(ids, s.Id.Int64)
				}
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetAllServiceIds(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ServiceRows)))
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
				It("can filter by a single service id that does exist", func() {
					sId := ids[rand.Intn(len(ids))]
					filter := &entity.ServiceFilter{
						Id: []*int64{&sId},
					}

					entries, err := db.GetAllServiceIds(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(1))
					})

					By("returning expected elements", func() {
						Expect(entries[0]).To(BeEquivalentTo(sId))
					})
				})
			})
		})
	})

	When("Getting Services", Label("GetServices"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetServices(nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 services in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetServices(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ServiceRows)))
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
							for _, row := range seedCollection.ServiceRows {
								if r.Id == row.Id.Int64 {
									Expect(r.Name).Should(BeEquivalentTo(row.Name.String), "Name should match")
									Expect(r.BaseService.CreatedAt).ShouldNot(BeEquivalentTo(row.CreatedAt.Time), "CreatedAt matches")
									Expect(r.BaseService.UpdatedAt).ShouldNot(BeEquivalentTo(row.UpdatedAt.Time), "UpdatedAt matches")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single name", func() {
					row := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]
					filter := &entity.ServiceFilter{Name: []*string{&row.Name.String}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning some results", func() {
						Expect(entries).NotTo(BeEmpty())
					})
					By("returning entries include the service name", func() {
						for _, entry := range entries {
							Expect(entry.Name).To(BeEquivalentTo(row.Name.String))
						}
					})
				})
				It("can filter by a random non existing service name", func() {
					nonExistingName := util.GenerateRandomString(40, nil)
					filter := &entity.ServiceFilter{Name: []*string{&nonExistingName}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning no results", func() {
						Expect(entries).To(BeEmpty())
					})
				})
				It("can filter by all existing service names", func() {
					serviceNames := make([]*string, len(seedCollection.ServiceRows))
					for i, row := range seedCollection.ServiceRows {
						x := row.Name.String
						serviceNames[i] = &x
					}
					filter := &entity.ServiceFilter{Name: serviceNames}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning expected number of results", func() {
						Expect(len(entries)).To(BeEquivalentTo(len(seedCollection.ServiceRows)))
					})
				})
				It("can filter by a single service Id", func() {
					row := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]
					filter := &entity.ServiceFilter{Id: []*int64{&row.Id.Int64}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entry", func() {
						Expect(entries[0].Id).To(BeEquivalentTo(row.Id.Int64))
					})
				})
				It("can filter by a single support group name", func() {
					// select a support group
					sgRow := seedCollection.SupportGroupRows[rand.Intn(len(seedCollection.SupportGroupRows))]

					// collect all service ids that belong to the support group
					serviceIds := []int64{}
					for _, sgsRow := range seedCollection.SupportGroupServiceRows {
						if sgsRow.SupportGroupId.Int64 == sgRow.Id.Int64 {
							serviceIds = append(serviceIds, sgsRow.ServiceId.Int64)
						}
					}

					filter := &entity.ServiceFilter{SupportGroupName: []*string{&sgRow.Name.String}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entries", func() {
						for _, entry := range entries {
							Expect(serviceIds).To(ContainElement(entry.Id))
						}
					})

				})
				It("can filter by a single owner name", func() {
					// select a user
					userRow := seedCollection.UserRows[rand.Intn(len(seedCollection.UserRows))]

					// collect all service ids that belong to the owner
					serviceIds := []int64{}
					for _, ownerRow := range seedCollection.OwnerRows {
						if ownerRow.UserId.Int64 == userRow.Id.Int64 {
							serviceIds = append(serviceIds, ownerRow.ServiceId.Int64)
						}
					}

					filter := &entity.ServiceFilter{OwnerName: []*string{&userRow.Name.String}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entries", func() {
						for _, entry := range entries {
							Expect(serviceIds).To(ContainElement(entry.Id), "Returns correct entry")
						}
					})
				})
				It("can filter by a single owner id", func() {
					// select a owner
					ownerRow := seedCollection.OwnerRows[rand.Intn(len(seedCollection.OwnerRows))]

					// collect all service ids that belong to the owner
					serviceIds := []int64{}
					for _, oRow := range seedCollection.OwnerRows {
						if oRow.UserId.Int64 == ownerRow.UserId.Int64 {
							serviceIds = append(serviceIds, oRow.ServiceId.Int64)
						}
					}

					filter := &entity.ServiceFilter{OwnerId: []*int64{&ownerRow.UserId.Int64}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entries", func() {
						for _, entry := range entries {
							Expect(serviceIds).To(ContainElement(entry.Id), "Returns correct entry")
						}
					})
				})
				It("can filter by a single activity id", func() {
					// select a activity
					activityRow := seedCollection.ActivityRows[rand.Intn(len(seedCollection.ActivityRows))]

					// collect all service ids that belong to the activity
					serviceIds := []int64{}
					for _, ahsRow := range seedCollection.ActivityHasServiceRows {
						if ahsRow.ActivityId.Int64 == activityRow.Id.Int64 {
							serviceIds = append(serviceIds, ahsRow.ServiceId.Int64)
						}
					}

					filter := &entity.ServiceFilter{ActivityId: []*int64{&activityRow.Id.Int64}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entries", func() {
						for _, entry := range entries {
							Expect(serviceIds).To(ContainElement(entry.Id), "Returns correct entry")
						}
					})
				})
				It("can filter by a single issue repository id", func() {
					// select a issue repository
					irRow := seedCollection.IssueRepositoryRows[rand.Intn(len(seedCollection.IssueRepositoryRows))]

					// collect all service ids that belong to the issue repository
					serviceIds := []int64{}
					for _, irsRow := range seedCollection.IssueRepositoryServiceRows {
						if irsRow.IssueRepositoryId.Int64 == irRow.Id.Int64 {
							serviceIds = append(serviceIds, irsRow.ServiceId.Int64)
						}
					}

					filter := &entity.ServiceFilter{IssueRepositoryId: []*int64{&irRow.Id.Int64}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entries", func() {
						for _, entry := range entries {
							Expect(serviceIds).To(ContainElement(entry.Id), "Returns correct entry")
						}
					})
				})
				It("can filter by a single support group id", func() {
					// select a support group
					sgRow := seedCollection.SupportGroupRows[rand.Intn(len(seedCollection.SupportGroupRows))]

					// collect all service ids that belong to the support group
					serviceIds := []int64{}
					for _, sssRow := range seedCollection.SupportGroupServiceRows {
						if sssRow.SupportGroupId.Int64 == sgRow.Id.Int64 {
							serviceIds = append(serviceIds, sssRow.ServiceId.Int64)
						}
					}

					filter := &entity.ServiceFilter{SupportGroupId: []*int64{&sgRow.Id.Int64}}

					entries, err := db.GetServices(filter)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct entries", func() {
						for _, entry := range entries {
							Expect(serviceIds).To(ContainElement(entry.Id), "Returns correct entry")
						}
					})
				})
				It("can filter service ServiceName using wild card search", func() {
					row := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					const charactersToRemoveFromBeginning = 2
					const charactersToRemoveFromEnd = 2
					const minimalCharactersToKeep = 2

					start := charactersToRemoveFromBeginning
					end := len(row.Name.String) - charactersToRemoveFromEnd

					Expect(start+minimalCharactersToKeep < end).To(BeTrue())

					searchStr := row.Name.String[start:end]
					filter := &entity.ServiceFilter{Search: []*string{&searchStr}}

					entries, err := db.GetServices(filter)

					names := []string{}
					for _, entry := range entries {
						names = append(names, entry.Name)
					}

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("at least one element was discarded (filtered)", func() {
						Expect(len(seedCollection.ServiceRows) > len(names)).To(BeTrue())
					})

					By("returning the expected elements", func() {
						Expect(names).To(ContainElement(row.Name.String))
					})
				})
			})
			Context("and using pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfList(
						db.GetServices,
						func(first *int, after *int64) *entity.ServiceFilter {
							return &entity.ServiceFilter{
								Paginated: entity.Paginated{First: first, After: after},
							}
						},
						func(entries []entity.Service) *int64 { return &entries[len(entries)-1].Id },
						len(seedCollection.ServiceRows),
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
	When("Counting Services", Label("CountServices"), func() {
		Context("and the database is empty", func() {
			It("can count correctly", func() {
				c, err := db.CountServices(nil)

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
			var serviceRows []mariadb.BaseServiceRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				serviceRows = seedCollection.ServiceRows
				count = len(serviceRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					c, err := db.CountServices(nil)

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
					filter := &entity.ServiceFilter{
						Paginated: entity.Paginated{
							First: &f,
							After: nil,
						},
					}
					c, err := db.CountServices(filter)

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
					// select a support group
					sgRow := seedCollection.SupportGroupRows[rand.Intn(len(seedCollection.SupportGroupRows))]

					// collect all service ids that belong to the support group
					serviceIds := []int64{}
					for _, sgsRow := range seedCollection.SupportGroupServiceRows {
						if sgsRow.SupportGroupId.Int64 == sgRow.Id.Int64 {
							serviceIds = append(serviceIds, sgsRow.ServiceId.Int64)
						}
					}

					filter := &entity.ServiceFilter{
						Paginated: entity.Paginated{
							First: &pageSize,
							After: nil,
						},
						SupportGroupName: []*string{&sgRow.Name.String},
					}
					entries, err := db.CountServices(filter)
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					By("returning the correct count", func() {
						Expect(entries).To(BeEquivalentTo(len(serviceIds)))
					})
				},
					Entry("and pageSize is 1 and it has 13 elements", 1, 13),
					Entry("and  pageSize is 20 and it has 5 elements", 20, 5),
					Entry("and  pageSize is 100 and it has 100 elements", 100, 100),
				)
			})
		})

	})
	When("Insert Service", Label("InsertService"), func() {
		Context("and we have 10 Services in the database", func() {
			var newServiceRow mariadb.ServiceRow
			var newService entity.Service
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newServiceRow = test.NewFakeService()
				newService = newServiceRow.AsService()
			})
			It("can insert correctly", func() {
				service, err := db.CreateService(&newService)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("sets service id", func() {
					Expect(service).NotTo(BeEquivalentTo(0))
				})

				serviceFilter := &entity.ServiceFilter{
					Id: []*int64{&service.Id},
				}

				s, err := db.GetServices(serviceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning service", func() {
					Expect(len(s)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(s[0].Name).To(BeEquivalentTo(service.Name))
				})
			})
			It("does not insert service with existing name", func() {
				serviceRow := seedCollection.ServiceRows[0]
				service := serviceRow.AsService()
				newService, err := db.CreateService(&service)

				By("throwing error", func() {
					Expect(err).ToNot(BeNil())
				})
				By("no service returned", func() {
					Expect(newService).To(BeNil())
				})

			})
		})
	})
	When("Update Service", Label("UpdateService"), func() {
		Context("and we have 10 Services in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can update service name correctly", func() {
				service := seedCollection.ServiceRows[0].AsService()

				service.Name = "SecretService"
				err := db.UpdateService(&service)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				serviceFilter := &entity.ServiceFilter{
					Id: []*int64{&service.Id},
				}

				s, err := db.GetServices(serviceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning service", func() {
					Expect(len(s)).To(BeEquivalentTo(1))
				})
				By("setting fields", func() {
					Expect(s[0].Name).To(BeEquivalentTo(service.Name))
				})
			})
		})
	})
	When("Delete Service", Label("DeleteService"), func() {
		Context("and we have 10 Services in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})
			It("can delete service correctly", func() {
				service := seedCollection.ServiceRows[0].AsService()

				err := db.DeleteService(service.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				serviceFilter := &entity.ServiceFilter{
					Id: []*int64{&service.Id},
				}

				s, err := db.GetServices(serviceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning no service", func() {
					Expect(len(s)).To(BeEquivalentTo(0))
				})
			})
		})
	})
	When("Add Owner To Service", Label("AddOwnerToService"), func() {
		Context("and we have 10 Services in the database", func() {
			var seedCollection *test.SeedCollection
			var newOwnerRow mariadb.UserRow
			var newOwner entity.User
			var owner *entity.User
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newOwnerRow = test.NewFakeUser()
				newOwner = newOwnerRow.AsUser()
				owner, _ = db.CreateUser(&newOwner)
			})
			It("can add owner correctly", func() {
				service := seedCollection.ServiceRows[0].AsService()

				err := db.AddOwnerToService(service.Id, owner.Id)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				serviceFilter := &entity.ServiceFilter{
					OwnerId: []*int64{&owner.Id},
				}

				s, err := db.GetServices(serviceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning service", func() {
					Expect(len(s)).To(BeEquivalentTo(1))
				})
			})
		})
	})
	When("Remove Owner From Service", Label("RemoveOwnerFromService"), func() {
		Context("and we have 10 Services in the database", func() {
			var seedCollection *test.SeedCollection
			var ownerRow mariadb.OwnerRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				ownerRow = seedCollection.OwnerRows[0]
			})
			It("can remove owner correctly", func() {
				err := db.RemoveOwnerFromService(ownerRow.ServiceId.Int64, ownerRow.UserId.Int64)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				serviceFilter := &entity.ServiceFilter{
					OwnerId: []*int64{&ownerRow.UserId.Int64},
				}

				services, err := db.GetServices(serviceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				for _, s := range services {
					Expect(s.Id).ToNot(BeEquivalentTo(ownerRow.ServiceId.Int64))
				}
			})
		})
	})
	When("Add Issue Repository To Service", Label("AddIssueRepositoryToService"), func() {
		Context("and we have 10 Services in the database", func() {
			var seedCollection *test.SeedCollection
			var newIssueRepositoryRow mariadb.IssueRepositoryRow
			var newIssueRepository entity.IssueRepository
			var issueRepository *entity.IssueRepository
			var priority int64 = 1
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				newIssueRepositoryRow = test.NewFakeIssueRepository()
				newIssueRepository = newIssueRepositoryRow.AsIssueRepository()
				issueRepository, _ = db.CreateIssueRepository(&newIssueRepository)
			})
			It("can add issue repository correctly", func() {
				service := seedCollection.ServiceRows[0].AsService()

				err := db.AddIssueRepositoryToService(service.Id, issueRepository.Id, priority)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				serviceFilter := &entity.ServiceFilter{
					IssueRepositoryId: []*int64{&issueRepository.Id},
				}

				s, err := db.GetServices(serviceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning service", func() {
					Expect(len(s)).To(BeEquivalentTo(1))
				})
			})
		})
	})
	When("Remove Issue Repository From Service", Label("RemoveIssueRepositoryFromService"), func() {
		Context("and we have 10 Services in the database", func() {
			var seedCollection *test.SeedCollection
			var issueRepositoryServiceRow mariadb.IssueRepositoryServiceRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				issueRepositoryServiceRow = seedCollection.IssueRepositoryServiceRows[0]
			})
			It("can remove issue repository correctly", func() {
				err := db.RemoveIssueRepositoryFromService(issueRepositoryServiceRow.ServiceId.Int64, issueRepositoryServiceRow.IssueRepositoryId.Int64)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				serviceFilter := &entity.ServiceFilter{
					IssueRepositoryId: []*int64{&issueRepositoryServiceRow.IssueRepositoryId.Int64},
				}

				services, err := db.GetServices(serviceFilter)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})

				for _, s := range services {
					Expect(s.Id).ToNot(BeEquivalentTo(issueRepositoryServiceRow.ServiceId.Int64))
				}
			})
		})
	})
	When("Getting ServiceNames", Label("GetServiceNames"), func() {
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				res, err := db.GetServiceNames(nil)
				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 services in the database", func() {
			var seedCollection *test.SeedCollection
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					res, err := db.GetServiceNames(nil)

					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})

					By("returning the correct number of results", func() {
						Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ServiceRows)))
					})

					existingServiceNames := lo.Map(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, index int) string {
						return s.Name.String
					})

					By("returning the correct names", func() {
						left, right := lo.Difference(res, existingServiceNames)
						Expect(left).Should(BeEmpty())
						Expect(right).Should(BeEmpty())
					})
				})
			})
			Context("and using a ServiceName filter", func() {

				var filter *entity.ServiceFilter
				var expectedServiceNames []string
				BeforeEach(func() {
					namePointers := []*string{}

					name := "f1"
					namePointers = append(namePointers, &name)

					filter = &entity.ServiceFilter{
						Name: namePointers,
					}

					It("can fetch the filtered items correctly", func() {
						res, err := db.GetServiceNames(filter)

						By("throwing no error", func() {
							Expect(err).Should(BeNil())
						})

						By("returning the correct number of results", func() {
							Expect(len(res)).Should(BeIdenticalTo(len(expectedServiceNames)))
						})

						By("returning the correct names", func() {
							left, right := lo.Difference(res, expectedServiceNames)
							Expect(left).Should(BeEmpty())
							Expect(right).Should(BeEmpty())
						})
					})
					It("and using another filter", func() {

						var anotherFilter *entity.ServiceFilter
						BeforeEach(func() {

							nonExistentServiceName := "NonexistentService"

							nonExistentServiceNames := []*string{&nonExistentServiceName}

							anotherFilter = &entity.ServiceFilter{
								Name: nonExistentServiceNames,
							}

							It("returns an empty list when no services match the filter", func() {
								res, err := db.GetServiceNames(anotherFilter)
								Expect(err).Should(BeNil())
								Expect(res).Should(BeEmpty())

								By("throwing no error", func() {
									Expect(err).Should(BeNil())
								})

								By("returning an empty list", func() {
									Expect(res).Should(BeEmpty())
								})
							})
						})
					})
				})
			})
		})
	})
})
