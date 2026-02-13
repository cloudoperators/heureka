// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Getting Patches via API", Label("e2e", "Patches"), func() {
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
		cfg.Port = util2.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		_ = dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		It("returns empty result set", func() {
			respData := e2e_common.ExecuteGqlQueryFromFile[struct {
				Patches model.PatchConnection `json:"Patches"`
			}](
				cfg.Port,
				"../api/graphql/graph/queryCollection/patch/query.graphql",
				map[string]interface{}{
					"filter": map[string]string{},
					"first":  10,
					"after":  "",
				})
			Expect(respData.Patches.TotalCount).To(Equal(0))
		})
	})

	When("the database has 10 entries", func() {
		var seedCollection *test.SeedCollection
		type patchRespDataType struct {
			Patches model.PatchConnection `json:"Patches"`
		}
		var respData patchRespDataType
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct result count", func() {
				respData = e2e_common.ExecuteGqlQueryFromFile[patchRespDataType](
					cfg.Port,
					"../api/graphql/graph/queryCollection/patch/query.graphql",
					map[string]interface{}{
						"filter": map[string]string{},
						"first":  5,
						"after":  "",
					})
				Expect(respData.Patches.TotalCount).To(Equal(len(seedCollection.PatchRows)))
				Expect(len(respData.Patches.Edges)).To(Equal(5))
				//- returns the expected PageInfo
				Expect(*respData.Patches.PageInfo.HasNextPage).To(BeTrue(), "hasNextPage is set")
				Expect(*respData.Patches.PageInfo.HasPreviousPage).To(BeFalse(), "hasPreviousPage is set")
				Expect(respData.Patches.PageInfo.NextPageAfter).ToNot(BeNil(), "nextPageAfter is set")
				Expect(len(respData.Patches.PageInfo.Pages)).To(Equal(2), "Correct amount of pages")
				Expect(*respData.Patches.PageInfo.PageNumber).To(Equal(1), "Correct page number")
				//- returns the expected content
				for _, patch := range respData.Patches.Edges {
					Expect(patch.Node.ID).ToNot(BeNil(), "patch has ID set")
					Expect(patch.Node.ServiceID).ToNot(BeNil(), "patch has Service ID set")
					Expect(patch.Node.ServiceName).ToNot(BeNil(), "patch has Service Name set")
					Expect(patch.Node.ComponentVersionID).ToNot(BeNil(), "patch has Component Version ID set")
					Expect(patch.Node.ComponentVersionName).ToNot(BeNil(), "patch has Component Version Name set")

					_, patchFound := lo.Find(seedCollection.PatchRows, func(row mariadb.PatchRow) bool {
						return fmt.Sprintf("%d", row.Id.Int64) == patch.Node.ID
					})
					Expect(patchFound).To(BeTrue(), "patch exists in seeded data")
				}
			})
		})
	})
})
