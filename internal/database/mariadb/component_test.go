// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"sort"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"math/rand"

	pkg_util "github.com/cloudoperators/heureka/pkg/util"
)

var _ = Describe("Component", Label("database", "Component"), func() {

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

	When("Getting All Component IDs", Label("GetAllComponentIds"), func() {

		var testGetComponentIds = func(filter *entity.ComponentFilter, expectedIds []int64, check func(entries []int64)) {
			res, err := db.GetAllComponentIds(filter)
			Expect(err).To(BeNil(), "GetAllComponentIds should not error")
			Expect(len(res)).To(BeEquivalentTo(len(expectedIds)), "GetAllComponentIds should return expected number of ids")
			check(res)
		}

		Context("and the database is empty", func() {
			It("can perform the query", func() {
				testGetComponentIds(nil, []int64{}, func(entries []int64) {})
			})
		})
		Context("and we have 20 Components in the database", func() {
			var seedCollection *test.SeedCollection
			var ids []int64
			var randomId int64

			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				ids = lo.Map(seedCollection.ComponentRows, func(c mariadb.ComponentRow, _ int) int64 {
					return c.Id.Int64
				})
				randomId = ids[rand.Intn(len(ids))]
			})
			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					testGetComponentIds(nil, ids, func(entries []int64) {
						var prev int64 = 0
						for _, r := range entries {
							Expect(r > prev).Should(BeTrue())
							prev = r
						}
						for _, r := range entries {
							Expect(lo.Contains(ids, r)).To(BeTrue())
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single component id that does exist", func() {
					filter := &entity.ComponentFilter{
						Id: []*int64{&randomId},
					}

					testGetComponentIds(filter, []int64{randomId}, func(entries []int64) {
						Expect(entries[0]).To(BeEquivalentTo(randomId))
					})
				})
			})
		})
	})

	When("Getting Components", Label("GetComponents"), func() {
		var testGetComponents = func(filter *entity.ComponentFilter, order []entity.Order, expectedComponents []mariadb.ComponentRow, check func(entries []entity.ComponentResult)) {
			res, err := db.GetComponents(filter, order)
			Expect(err).To(BeNil(), "GetComponents should not error")
			Expect(len(res)).To(BeEquivalentTo(len(expectedComponents)), "GetComponents should return expected number of components")
			check(res)
		}
		Context("and the database is empty", func() {
			It("can perform the list query", func() {
				testGetComponents(nil, []entity.Order{}, []mariadb.ComponentRow{}, func(entries []entity.ComponentResult) {})
			})
		})
		Context("and we have 10 components in the database", func() {
			var seedCollection *test.SeedCollection
			var randomComponent mariadb.ComponentRow
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(10)
				randomComponent = seedCollection.ComponentRows[rand.Intn(len(seedCollection.ComponentRows))]
			})

			Context("and using no filter", func() {
				It("can fetch the items correctly", func() {
					testGetComponents(nil, []entity.Order{}, seedCollection.ComponentRows, func(entries []entity.ComponentResult) {
						var prev int64 = 0
						for _, r := range entries {
							Expect(r.Id > prev).Should(BeTrue())
							prev = r.Id
							for _, row := range seedCollection.ComponentRows {
								if r.Id == row.Id.Int64 {
									Expect(r.CCRN).Should(BeEquivalentTo(row.CCRN.String), "CCRN should match")
									Expect(r.Repository).Should(BeEquivalentTo(row.Repository.String), "Repository should match")
									Expect(r.Organization).Should(BeEquivalentTo(row.Organization.String), "Organization should match")
									Expect(r.Url).Should(BeEquivalentTo(row.Url.String), "Url should match")
									Expect(r.Type).Should(BeEquivalentTo(row.Type.String), "Type should match")
								}
							}
						}
					})
				})
			})
			Context("and using a filter", func() {
				It("can filter by a single id", func() {
					filter := &entity.ComponentFilter{Id: []*int64{&randomComponent.Id.Int64}}
					testGetComponents(filter, []entity.Order{}, []mariadb.ComponentRow{randomComponent}, func(entries []entity.ComponentResult) {
						for _, entry := range entries {
							Expect(entry.Id).To(BeEquivalentTo(randomComponent.Id.Int64))
						}
					})
				})
				It("can filter by a single repository", func() {
					filter := &entity.ComponentFilter{Repository: []*string{&randomComponent.Repository.String}}
					testGetComponents(filter, []entity.Order{}, []mariadb.ComponentRow{randomComponent}, func(entries []entity.ComponentResult) {
						for _, entry := range entries {
							Expect(entry.Repository).To(BeEquivalentTo(randomComponent.Repository.String))
						}
					})
				})
				It("can filter by a single organization", func() {
					filter := &entity.ComponentFilter{Organization: []*string{&randomComponent.Organization.String}}
					testGetComponents(filter, []entity.Order{}, []mariadb.ComponentRow{randomComponent}, func(entries []entity.ComponentResult) {
						for _, entry := range entries {
							Expect(entry.Organization).To(BeEquivalentTo(randomComponent.Organization.String))
						}
					})
				})
				It("can filter by a single ccrn", func() {
					filter := &entity.ComponentFilter{CCRN: []*string{&randomComponent.CCRN.String}}
					testGetComponents(filter, []entity.Order{}, []mariadb.ComponentRow{randomComponent}, func(entries []entity.ComponentResult) {
						for _, entry := range entries {
							Expect(entry.CCRN).To(BeEquivalentTo(randomComponent.CCRN.String))
						}
					})
				})
				It("can filter by a random non existing component ccrn", func() {
					nonExistingCCRN := pkg_util.GenerateRandomString(40, nil)
					filter := &entity.ComponentFilter{CCRN: []*string{&nonExistingCCRN}}
					testGetComponents(filter, []entity.Order{}, []mariadb.ComponentRow{}, func(entries []entity.ComponentResult) {})
				})
				It("can filter by all existing component ccrns", func() {
					allCCRNs := lo.Map(seedCollection.ComponentRows, func(c mariadb.ComponentRow, _ int) *string {
						return &c.CCRN.String
					})
					filter := &entity.ComponentFilter{CCRN: allCCRNs}
					testGetComponents(filter, []entity.Order{}, seedCollection.ComponentRows, func(entries []entity.ComponentResult) {})
				})
				It("can filter by a single service ccrn", func() {
					serviceRow := seedCollection.ServiceRows[rand.Intn(len(seedCollection.ServiceRows))]

					cvIds := lo.FilterMap(seedCollection.ComponentInstanceRows, func(cir mariadb.ComponentInstanceRow, _ int) (int64, bool) {
						return cir.ComponentVersionId.Int64, cir.ServiceId.Int64 == serviceRow.Id.Int64
					})

					componentIds := lo.FilterMap(seedCollection.ComponentVersionRows, func(cvr mariadb.ComponentVersionRow, _ int) (int64, bool) {
						return cvr.ComponentId.Int64, lo.Contains(cvIds, cvr.Id.Int64)
					})

					components := lo.Filter(seedCollection.ComponentRows, func(cr mariadb.ComponentRow, _ int) bool {
						return lo.Contains(componentIds, cr.Id.Int64)
					})

					filter := &entity.ComponentFilter{ServiceCCRN: []*string{&serviceRow.CCRN.String}}

					testGetComponents(filter, []entity.Order{}, components, func(entries []entity.ComponentResult) {
						for _, entry := range entries {
							Expect(lo.Contains(componentIds, entry.Id)).To(BeTrue())
						}
					})
				})
				It("can filter by a single component repository", func() {
					filter := &entity.ComponentFilter{Repository: []*string{&randomComponent.Repository.String}}

					testGetComponents(filter, []entity.Order{}, []mariadb.ComponentRow{randomComponent}, func(entries []entity.ComponentResult) {
						for _, entry := range entries {
							Expect(entry.Id).To(BeEquivalentTo(randomComponent.Id.Int64))
						}
					})
				})
			})
			Context("and using pagination", func() {
				DescribeTable("can correctly paginate with x elements", func(pageSize int) {
					test.TestPaginationOfListWithOrder(
						db.GetComponents,
						func(first *int, after *int64, afterX *string) *entity.ComponentFilter {
							return &entity.ComponentFilter{
								PaginatedX: entity.PaginatedX{First: first, After: afterX},
							}
						},
						[]entity.Order{},
						func(entries []entity.ComponentResult) string {
							after, _ := mariadb.EncodeCursor(mariadb.WithComponent([]entity.Order{}, *entries[len(entries)-1].Component, entity.IssueSeverityCounts{}))
							return after
						},
						len(seedCollection.ComponentRows),
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
	When("Counting Components", Label("CountComponents"), func() {
		var testCountComponents = func(filter *entity.ComponentFilter, expectedCount int) {
			c, err := db.CountComponents(filter)
			Expect(err).To(BeNil(), "CountComponents should not error")
			Expect(c).To(BeEquivalentTo(expectedCount), "CountComponents should return expected count")
		}

		Context("and the database is empty", func() {
			It("can count correctly", func() {
				testCountComponents(nil, 0)
			})
		})
		Context("and the database has 100 entries", func() {
			var seedCollection *test.SeedCollection
			var componentRows []mariadb.ComponentRow
			var randomComponent mariadb.ComponentRow
			var count int
			BeforeEach(func() {
				seedCollection = seeder.SeedDbWithNFakeData(100)
				componentRows = seedCollection.ComponentRows
				randomComponent = componentRows[rand.Intn(len(componentRows))]
				count = len(componentRows)

			})
			Context("and using no filter", func() {
				It("can count", func() {
					testCountComponents(nil, count)
				})
			})
			Context("and using pagination", func() {
				It("can count", func() {
					after := ""
					filter := &entity.ComponentFilter{
						PaginatedX: entity.PaginatedX{
							First: &count,
							After: &after,
						},
					}
					testCountComponents(filter, count)
				})
			})

			Context("and using a filter", func() {
				DescribeTable("can count with a filter", func(pageSize int, filterMatches int) {
					components := lo.Filter(seedCollection.ComponentRows, func(cr mariadb.ComponentRow, _ int) bool {
						return cr.CCRN.String == randomComponent.CCRN.String
					})

					filter := &entity.ComponentFilter{
						PaginatedX: entity.PaginatedX{
							First: &pageSize,
							After: nil,
						},
						CCRN: []*string{&randomComponent.CCRN.String},
					}

					testCountComponents(filter, len(components))
				},
					Entry("and pageSize is 1 and it has 13 elements", 1, 13),
					Entry("and  pageSize is 20 and it has 5 elements", 20, 5),
					Entry("and  pageSize is 100 and it has 100 elements", 100, 100),
				)
			})
		})
		When("Insert Component", Label("InsertComponent"), func() {
			Context("and we have 10 Components in the database", func() {
				var newComponentRow mariadb.ComponentRow
				var newComponent entity.Component
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
					newComponentRow = test.NewFakeComponent()
					newComponent = newComponentRow.AsComponent()
				})
				It("can insert correctly", func() {
					component, err := db.CreateComponent(&newComponent)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("sets component id", func() {
						Expect(component).NotTo(BeEquivalentTo(0))
					})

					componentFilter := &entity.ComponentFilter{
						Id: []*int64{&component.Id},
					}

					c, err := db.GetComponents(componentFilter, []entity.Order{})
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning component", func() {
						Expect(len(c)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(c[0].Id).To(BeEquivalentTo(component.Id))
						Expect(c[0].CCRN).To(BeEquivalentTo(component.CCRN))
						Expect(c[0].Repository).To(BeEquivalentTo(component.Repository))
						Expect(c[0].Organization).To(BeEquivalentTo(component.Organization))
						Expect(c[0].Url).To(BeEquivalentTo(component.Url))
						Expect(c[0].Type).To(BeEquivalentTo(component.Type))
					})
				})
				It("does not insert component with existing ccrn", func() {
					componentRow := seedCollection.ComponentRows[0]
					component := componentRow.AsComponent()
					newComponent, err := db.CreateComponent(&component)

					By("throwing error", func() {
						Expect(err).ToNot(BeNil())
					})
					By("no component returned", func() {
						Expect(newComponent).To(BeNil())
					})

				})
			})
		})
		When("Update Component", Label("UpdateComponent"), func() {
			Context("and we have 10 Components in the database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
				})
				It("can update ccrn correctly", func() {
					component := seedCollection.ComponentRows[0].AsComponent()

					component.CCRN = "NewCCRN"
					err := db.UpdateComponent(&component)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					componentFilter := &entity.ComponentFilter{
						Id: []*int64{&component.Id},
					}

					c, err := db.GetComponents(componentFilter, []entity.Order{})
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning component", func() {
						Expect(len(c)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(c[0].Id).To(BeEquivalentTo(component.Id))
						Expect(c[0].CCRN).To(BeEquivalentTo(component.CCRN))
						Expect(c[0].Repository).To(BeEquivalentTo(component.Repository))
						Expect(c[0].Organization).To(BeEquivalentTo(component.Organization))
						Expect(c[0].Url).To(BeEquivalentTo(component.Url))
						Expect(c[0].Type).To(BeEquivalentTo(component.Type))
					})
				})
				It("can update type correctly", func() {
					component := seedCollection.ComponentRows[0].AsComponent()

					component.Type = "NewType"
					err := db.UpdateComponent(&component)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					componentFilter := &entity.ComponentFilter{
						Id: []*int64{&component.Id},
					}

					c, err := db.GetComponents(componentFilter, []entity.Order{})
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning component", func() {
						Expect(len(c)).To(BeEquivalentTo(1))
					})
					By("setting fields", func() {
						Expect(c[0].Id).To(BeEquivalentTo(component.Id))
						Expect(c[0].CCRN).To(BeEquivalentTo(component.CCRN))
						Expect(c[0].Repository).To(BeEquivalentTo(component.Repository))
						Expect(c[0].Organization).To(BeEquivalentTo(component.Organization))
						Expect(c[0].Url).To(BeEquivalentTo(component.Url))
						Expect(c[0].Type).To(BeEquivalentTo(component.Type))
					})
				})
			})
		})
		When("Delete Component", Label("DeleteComponent"), func() {
			Context("and we have 10 Components in the database", func() {
				var seedCollection *test.SeedCollection
				BeforeEach(func() {
					seedCollection = seeder.SeedDbWithNFakeData(10)
				})
				It("can delete component correctly", func() {
					component := seedCollection.ComponentRows[0].AsComponent()

					err := db.DeleteComponent(component.Id, util.SystemUserId)

					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})

					componentFilter := &entity.ComponentFilter{
						Id: []*int64{&component.Id},
					}

					c, err := db.GetComponents(componentFilter, []entity.Order{})
					By("throwing no error", func() {
						Expect(err).To(BeNil())
					})
					By("returning no component", func() {
						Expect(len(c)).To(BeEquivalentTo(0))
					})
				})
			})
		})
	})
})

var _ = Describe("Ordering Components", Label("ComponentOrdering"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	var seedCollection *test.SeedCollection
	var c *collate.Collator

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		c = collate.New(language.English)
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})
	AfterEach(func() {
		seeder.CloseDbConnection()
		dbm.TestTearDown(db)
	})

	var testOrder = func(
		order []entity.Order,
		verifyFunc func(res []entity.ComponentResult),
	) {
		res, err := db.GetComponents(nil, order)

		By("throwing no error", func() {
			Expect(err).Should(BeNil())
		})

		By("returning the correct number of results", func() {
			Expect(len(res)).Should(BeIdenticalTo(len(seedCollection.ComponentRows)))
		})

		By("returning the correct order", func() {
			verifyFunc(res)
		})
	}

	var loadTestData = func() ([]mariadb.ComponentVersionRow, []mariadb.ComponentInstanceRow, []mariadb.IssueVariantRow, []mariadb.ComponentVersionIssueRow, []mariadb.IssueMatchRow, error) {
		issueVariants, err := test.LoadIssueVariants(test.GetTestDataPath("testdata/component_version_order/issue_variant.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		cvIssues, err := test.LoadComponentVersionIssues(test.GetTestDataPath("testdata/component_order/component_version_issue.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		componentInstances, err := test.LoadComponentInstances(test.GetTestDataPath("testdata/service_order/component_instance.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		issueMatches, err := test.LoadIssueMatches(test.GetTestDataPath("testdata/component_order/issue_match.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		componentVersions, err := test.LoadComponentVersions(test.GetTestDataPath("testdata/component_order/component_version.json"))
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		return componentVersions, componentInstances, issueVariants, cvIssues, issueMatches, nil
	}

	When("order by count is used", func() {
		var componentFilter *entity.ComponentFilter
		BeforeEach(func() {
			seeder.SeedIssueRepositories()
			for i := 0; i < 10; i++ {
				issue := test.NewFakeIssue()
				issue.Type.String = entity.IssueTypeVulnerability.String()
				seeder.InsertFakeIssue(issue)
			}
			seeder.SeedComponents(5)
			var serviceCcrns []*string
			for _, s := range seeder.SeedServices(5) {
				serviceCcrns = append(serviceCcrns, &s.CCRN.String)
			}
			componentFilter = &entity.ComponentFilter{
				ServiceCCRN: serviceCcrns,
			}
			componentVersions, componentInstances, issueVariants, componentVersionIssues, issueMatches, err := loadTestData()
			Expect(err).To(BeNil())
			// Important: the order need to be preserved
			for _, iv := range issueVariants {
				_, err := seeder.InsertFakeIssueVariant(iv)
				Expect(err).To(BeNil())
			}
			for _, cv := range componentVersions {
				_, err := seeder.InsertFakeComponentVersion(cv)
				Expect(err).To(BeNil())
			}
			for _, cvi := range componentVersionIssues {
				_, err := seeder.InsertFakeComponentVersionIssue(cvi)
				Expect(err).To(BeNil())
			}
			for _, ci := range componentInstances {
				_, err := seeder.InsertFakeComponentInstance(ci)
				Expect(err).To(BeNil())
			}
			for _, im := range issueMatches {
				_, err := seeder.InsertFakeIssueMatch(im)
				Expect(err).To(BeNil())
			}
			err = seeder.RefreshComponentVulnerabilityCounts()
			Expect(err).To(BeNil())
		})
		It("can order desc by critical, high, medium, low and none", func() {
			order := []entity.Order{
				{By: entity.CriticalCount, Direction: entity.OrderDirectionDesc},
				{By: entity.HighCount, Direction: entity.OrderDirectionDesc},
				{By: entity.MediumCount, Direction: entity.OrderDirectionDesc},
				{By: entity.LowCount, Direction: entity.OrderDirectionDesc},
				{By: entity.NoneCount, Direction: entity.OrderDirectionDesc},
			}
			components, err := db.GetComponents(componentFilter, order)
			Expect(err).To(BeNil())
			Expect(components[0].Id).To(BeEquivalentTo(1))
			Expect(components[1].Id).To(BeEquivalentTo(2))
			Expect(components[2].Id).To(BeEquivalentTo(3))
			Expect(components[3].Id).To(BeEquivalentTo(4))
			Expect(components[4].Id).To(BeEquivalentTo(5))
		})
		It("can order asc by critical, high, medium, low and none", func() {
			order := []entity.Order{
				{By: entity.CriticalCount, Direction: entity.OrderDirectionAsc},
				{By: entity.HighCount, Direction: entity.OrderDirectionAsc},
				{By: entity.MediumCount, Direction: entity.OrderDirectionAsc},
				{By: entity.LowCount, Direction: entity.OrderDirectionAsc},
				{By: entity.NoneCount, Direction: entity.OrderDirectionAsc},
			}
			components, err := db.GetComponents(componentFilter, order)
			Expect(err).To(BeNil())
			Expect(components[0].Id).To(BeEquivalentTo(5))
			Expect(components[1].Id).To(BeEquivalentTo(4))
			Expect(components[2].Id).To(BeEquivalentTo(3))
			Expect(components[3].Id).To(BeEquivalentTo(2))
			Expect(components[4].Id).To(BeEquivalentTo(1))
		})
	})

	When("with ASC order", Label("ComponentASCOrder"), func() {

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		It("can order by id", func() {
			sort.Slice(seedCollection.ComponentRows, func(i, j int) bool {
				return seedCollection.ComponentRows[i].Id.Int64 < seedCollection.ComponentRows[j].Id.Int64
			})

			order := []entity.Order{
				{By: entity.ComponentId, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.ComponentResult) {
				for i, r := range res {
					Expect(r.Id).Should(BeEquivalentTo(seedCollection.ComponentRows[i].Id.Int64))
				}
			})
		})

		It("can order by component version repository", func() {
			order := []entity.Order{
				{By: entity.ComponentRepository, Direction: entity.OrderDirectionAsc},
			}

			testOrder(order, func(res []entity.ComponentResult) {
				var prev string = ""
				for _, r := range res {
					Expect(c.CompareString(r.Repository, prev)).Should(BeNumerically(">=", 0))
					prev = r.Repository
				}
			})
		})

	})

	When("with DESC order", Label("ComponentDESCOrder"), func() {

		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})

		It("can order by id", func() {
			sort.Slice(seedCollection.ComponentRows, func(i, j int) bool {
				return seedCollection.ComponentRows[i].Id.Int64 > seedCollection.ComponentRows[j].Id.Int64
			})

			order := []entity.Order{
				{By: entity.ComponentId, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.ComponentResult) {
				for i, r := range res {
					Expect(r.Id).Should(BeEquivalentTo(seedCollection.ComponentRows[i].Id.Int64))
				}
			})
		})

		It("can order by component version repository", func() {
			order := []entity.Order{
				{By: entity.ComponentRepository, Direction: entity.OrderDirectionDesc},
			}

			testOrder(order, func(res []entity.ComponentResult) {
				var prev string = "\U0010FFFF"
				for _, r := range res {
					Expect(c.CompareString(r.Repository, prev)).Should(BeNumerically("<=", 0))
					prev = r.Repository
				}
			})
		})

	})
})
