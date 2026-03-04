// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Getting ServiceFilterValues via API", Label("e2e", "ServiceFilterValues"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchemaWithoutMigration()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty resultset for serviceFilter", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/serviceFilter/serviceCcrns.graphql",
				nil,
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.ServiceFilterValues.ServiceCcrn.Values).To(BeEmpty())
		})
		It("returns empty for supportGroupCcrns", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/serviceFilter/supportGroupCcrns.graphql",
				nil,
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.ServiceFilterValues.SupportGroupCcrn.Values).To(BeEmpty())
		})
		It("returns empty for userNames", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/serviceFilter/userNames.graphql",
				nil,
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			e2e_common.ExpectNonSystemUserNames(respData.ServiceFilterValues.UserName.Values, []*string{})
		})
		It("returns empty for uniqueUserID", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/serviceFilter/uniqueUserId.graphql",
				nil,
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			e2e_common.ExpectNonSystemUserUniqueUserIds(respData.ServiceFilterValues.UniqueUserID.Values, []*string{})
		})
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct serviceCcrns", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/serviceFilter/serviceCcrns.graphql",
					nil,
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(respData.ServiceFilterValues.ServiceCcrn.Values)).To(Equal(len(seedCollection.ServiceRows)))

				existingServiceCcrns := lo.Map(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, index int) string {
					return s.CCRN.String
				})

				for _, ccrn := range respData.ServiceFilterValues.ServiceCcrn.Values {
					Expect(lo.Contains(existingServiceCcrns, *ccrn)).To(BeTrue())
				}
			})
			It("returns correct supportGroupCcrns", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/serviceFilter/supportGroupCcrns.graphql",
					nil,
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				Expect(len(respData.ServiceFilterValues.SupportGroupCcrn.Values)).To(Equal(len(seedCollection.SupportGroupRows)))

				existingsupportGroupCcrns := lo.Map(seedCollection.SupportGroupRows, func(s mariadb.SupportGroupRow, index int) string {
					return s.CCRN.String
				})

				for _, ccrn := range respData.ServiceFilterValues.SupportGroupCcrn.Values {
					Expect(lo.Contains(existingsupportGroupCcrns, *ccrn)).To(BeTrue())
				}
			})
			It("returns correct userNames", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/serviceFilter/userNames.graphql",
					nil,
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				e2e_common.ExpectNonSystemUserCount(len(respData.ServiceFilterValues.UserName.Values), len(seedCollection.UserRows))

				existingUserNames := lo.Map(seedCollection.UserRows, func(s mariadb.UserRow, index int) string {
					return s.Name.String
				})

				for _, name := range e2e_common.SubtractSystemUserName(respData.ServiceFilterValues.UserName.Values) {
					Expect(lo.Contains(existingUserNames, *name)).To(BeTrue())
				}
			})
			It("returns correct UniqueUserID", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/serviceFilter/uniqueUserId.graphql",
					nil,
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				e2e_common.ExpectNonSystemUserCount(len(respData.ServiceFilterValues.UniqueUserID.Values), len(seedCollection.UserRows))

				existingUniqueUserIds := lo.Map(seedCollection.UserRows, func(s mariadb.UserRow, index int) string {
					return s.UniqueUserID.String
				})

				for _, name := range e2e_common.SubtractSystemUserUniqueUserId(respData.ServiceFilterValues.UniqueUserID.Values) {
					Expect(lo.Contains(existingUniqueUserIds, *name)).To(BeTrue())
				}
			})
			It("returns correct Name With Id", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/serviceFilter/userNamesWithIds.graphql",
					nil,
					nil,
				)

				Expect(err).ToNot(HaveOccurred())

				e2e_common.ExpectNonSystemUserCount(len(respData.ServiceFilterValues.User.Values), len(seedCollection.UserRows))

				existingDisplays := lo.Map(seedCollection.UserRows, func(s mariadb.UserRow, index int) string {
					return fmt.Sprintf("%s (%s)", s.Name.String, s.UniqueUserID.String)
				})
				existingValues := lo.Map(seedCollection.UserRows, func(s mariadb.UserRow, index int) string {
					return s.UniqueUserID.String
				})

				for _, valueItem := range e2e_common.SubtractSystemUserNameFromValueItems(respData.ServiceFilterValues.User.Values) {
					Expect(lo.Contains(existingDisplays, *valueItem.Display)).To(BeTrue())
					Expect(lo.Contains(existingValues, *valueItem.Value)).To(BeTrue())
				}
			})
		})
	})
})
