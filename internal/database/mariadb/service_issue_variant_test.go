// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"database/sql"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/goark/go-cvss/v3/metric"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("ServiceIssueVariant - ", Label("database", "IssueVariant"), func() {
	var db *mariadb.SqlDatabase
	var seeder *test.DatabaseSeeder
	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Getting ServiceIssueVariants", Label("GetServiceIssueVariants"), func() {
		Context("and the database is empty", func() {
			It("can perform the query", func() {
				res, err := db.GetServiceIssueVariants(nil)

				By("throwing no error", func() {
					Expect(err).To(BeNil())
				})
				By("returning an empty list", func() {
					Expect(res).To(BeEmpty())
				})
			})
		})
		Context("and we have 10 issue variants in the database", func() {
			BeforeEach(func() {
				seeder.SeedDbWithNFakeData(10)
			})
			//this should work and give me all combinations back
			Context("and using no filter", func() {
				It("Should work", func() {
					_, err := db.GetServiceIssueVariants(nil)
					By("throwing no error", func() {
						Expect(err).Should(BeNil())
					})
				})
			})
		})

		//This is a testcase with a custom complex setup:
		// We need to setup a whole bunch of data to test the filtering for service issue variants based by component instances
		// The idea is to create
		DescribeTable("and filtering for component instances", func(filterForInstances int, totalInstances int, totalIssues int) {
			//Complex Setup
			var allCI []mariadb.ComponentInstanceRow
			var issueVariants []mariadb.IssueVariantRow
			var issueRepositories []mariadb.BaseIssueRepositoryRow
			issue_count := totalIssues / totalInstances
			// create issue repository
			issueRepositories = seeder.SeedIssueRepositories()

			for i := 0; i < totalInstances; i++ {
				components := make([]mariadb.ComponentRow, 0)
				services := make([]mariadb.BaseServiceRow, 0)

				// create the component
				// we do this until it got successfully created to avoid failures through unique constraint violations
				// this happens on bigger datasets due to the limited randomness of the fixtures
				for len(components) == 0 {
					components = seeder.SeedComponents(1)
				}

				// create the service
				// we do this until it got successfully created to avoid failures through unique constraint violations
				// this happens on bigger datasets due to the limited randomness of the fixtures
				for len(services) == 0 {
					services = seeder.SeedServices(1)
				}

				//create issues
				issues := seeder.SeedIssues(issue_count)

				// create component version and adding each issue to the component version
				componentVersions := seeder.SeedComponentVersions(1, components)
				cvirows := make([]mariadb.ComponentVersionIssueRow, issue_count)
				for idx, issue := range issues {
					cvi := mariadb.ComponentVersionIssueRow{
						ComponentVersionId: componentVersions[0].Id,
						IssueId:            issue.Id,
					}

					_, err := seeder.InsertFakeComponentVersionIssue(cvi)
					Expect(err).To(BeNil())
					cvirows[idx] = cvi
				}

				// create component instance
				componentInstances := seeder.SeedComponentInstances(1, componentVersions, services)
				allCI = append(allCI, componentInstances...)

				// create an issue variant per repo and issue (5 repos 10 issues)
				variantList := make([]mariadb.IssueVariantRow, issue_count*5)
				for idx, issue := range issues {
					for irdx, ir := range issueRepositories {
						variants := []string{fmt.Sprintf("GHSA-%d", i), fmt.Sprintf("RSHA-%d", i), fmt.Sprintf("VMSA-%d", i)}
						v := test.GenerateRandomCVSS31Vector()
						cvss, _ := metric.NewEnvironmental().Decode(v)
						rating := cvss.Severity().String()
						iv := mariadb.IssueVariantRow{
							IssueId:           issue.Id,
							IssueRepositoryId: ir.Id,
							SecondaryName:     sql.NullString{String: fmt.Sprintf("%s-%d-%d", gofakeit.RandomString(variants), gofakeit.Year(), gofakeit.Number(1000, 9999)), Valid: true},
							Description:       issue.Description,
							Vector:            sql.NullString{String: v, Valid: true},
							Rating:            sql.NullString{String: rating, Valid: true},
						}
						id, err := seeder.InsertFakeIssueVariant(iv)
						Expect(err).To(BeNil())
						iv.IssueId = sql.NullInt64{Int64: id, Valid: true}
						variantList[(idx*5)+irdx] = iv
					}
				}
				issueVariants = append(issueVariants, variantList...)

				// add to each repository the service with a increasing priority
				// this means the last repository is always the highest priority one
				for idx, ir := range issueRepositories {
					irs := mariadb.IssueRepositoryServiceRow{
						IssueRepositoryId: ir.Id,
						ServiceId:         services[0].Id,
						Priority:          sql.NullInt64{Int64: int64(idx + 1), Valid: true},
					}
					_, err := seeder.InsertFakeIssueRepositoryService(irs)
					Expect(err).To(BeNil())
				}
			}

			//Setup end

			//Except
			By(fmt.Sprintf("having in total %d component instances with each %d issues across the repositories", filterForInstances, totalIssues), func() {
				By("and filtering for this component instance", func() {
					By("it can perform the query correctly", func() {

						//get instance ids to filter for based on count of instances that we want to filter for
						cids := lo.Map(allCI, func(item mariadb.ComponentInstanceRow, _ int) *int64 { return lo.ToPtr(item.Id.Int64) })
						if len(cids) > filterForInstances {
							cids = cids[:filterForInstances]
						}
						filter := &entity.ServiceIssueVariantFilter{
							Paginated:           entity.Paginated{},
							ComponentInstanceId: cids,
						}
						res, err := db.GetServiceIssueVariants(filter)

						By("throwing no error", func() {
							Expect(err).To(BeNil())
						})

						By("and returning all the issue variants", func() {
							Expect(len(res)).To(BeIdenticalTo((len(issueVariants) / totalInstances) * filterForInstances))
						})
					})

				})
			})
		},
			Entry("1 of 1 component instance, with 10 issues", 1, 1, 10),
			Entry("1 of 2 component instance, with 10 issues", 1, 2, 10),
			Entry("1 of 1 component instance, with 100 issues", 1, 1, 100),
			Entry("2 of 2 component instance, with 10 issues", 2, 2, 10),
			Entry("4 of 100 component instance, with 50 issues", 4, 100, 50),
			Entry("4 of 4 component instance, with 4 issues", 4, 4, 4),
		)
	})
})
