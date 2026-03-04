// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Getting ComponentFilterValues via API", Label("e2e", "ComponentFilterValues"), func() {
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
		It("returns empty for componentCcrns", func() {
			respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
				ComponentFilterValues model.ComponentFilterValue `json:"ComponentFilterValues"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/componentFilter/componentCcrn.graphqls",
				nil,
				nil,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(respData.ComponentFilterValues.ComponentCcrn.Values).To(BeEmpty())
		})
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct componentCcrns", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					ComponentFilterValues model.ComponentFilterValue `json:"ComponentFilterValues"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentFilter/componentCcrn.graphqls",
					nil,
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(respData.ComponentFilterValues.ComponentCcrn.Values)).To(Equal(len(seedCollection.ComponentRows)))

				existingComponentCcrns := lo.Map(seedCollection.ComponentRows, func(s mariadb.ComponentRow, index int) string {
					return s.CCRN.String
				})

				for _, ccrn := range respData.ComponentFilterValues.ComponentCcrn.Values {
					Expect(lo.Contains(existingComponentCcrns, *ccrn)).To(BeTrue())
				}
			})
		})
	})
})
