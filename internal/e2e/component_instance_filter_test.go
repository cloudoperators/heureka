// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Getting ComponentInstanceFilterValues via API", Label("e2e", "ComponentInstanceFilterValues"), func() {
	var seeder *test.DatabaseSeeder
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		seeder, err = test.NewDatabaseSeeder(dbm.DbConfig())
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("the database is empty", func() {
		It("returns empty resultset for ccrnFilter", func() {
			// create a queryCollection (safe to share across requests)
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstanceFilter/ccrn.graphqls")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ComponentInstanceFilterValues model.ComponentInstanceFilterValue `json:"ComponentInstanceFilterValues"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ComponentInstanceFilterValues.Ccrn.Values).To(BeEmpty())
		})
		It("returns empty for serviceCcrns", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstanceFilter/serviceCcrn.graphqls")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ComponentInstanceFilterValues model.ComponentInstanceFilterValue `json:"ComponentInstanceFilterValues"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ComponentInstanceFilterValues.ServiceCcrn.Values).To(BeEmpty())
		})
		It("returns empty for supportGroupCcrn", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			b, err := os.ReadFile("../api/graphql/graph/queryCollection/componentInstanceFilter/supportGroupCcrn.graphqls")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ComponentInstanceFilterValues model.ComponentInstanceFilterValue `json:"ComponentInstanceFilterValues"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ComponentInstanceFilterValues.SupportGroupCcrn.Values).To(BeEmpty())
		})
	})

	When("the database has 10 entries", func() {

		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct ccrn", func() {
				existingCcrns := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
					return cir.CCRN.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/ccrn.graphqls",
					existingCcrns,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.Ccrn.Values
					})
			})
			It("returns correct supportGroupCcrns", func() {
				existingSupportGroupCcrns := lo.Map(seedCollection.SupportGroupRows, func(s mariadb.SupportGroupRow, index int) string {
					return s.CCRN.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/supportGroupCcrn.graphqls",
					existingSupportGroupCcrns,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.SupportGroupCcrn.Values
					})
			})
			It("returns correct serviceCcrns", func() {
				existingServiceCcrns := lo.Map(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, index int) string {
					return s.CCRN.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/serviceCcrn.graphqls",
					existingServiceCcrns,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.ServiceCcrn.Values
					})
			})
			It("returns correct region", func() {
				existingRegions := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
					return cir.Region.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/region.graphqls",
					existingRegions,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.Region.Values
					})
			})
			It("returns correct cluster", func() {
				existingClusters := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
					return cir.Cluster.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/cluster.graphqls",
					existingClusters,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.Cluster.Values
					})
			})
			It("returns correct namespace", func() {
				existingNamespaces := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
					return cir.Namespace.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/namespace.graphqls",
					existingNamespaces,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.Namespace.Values
					})
			})
			It("returns correct domain", func() {
				existingDomains := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
					return cir.Domain.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/domain.graphqls",
					existingDomains,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.Domain.Values
					})
			})
			It("returns correct project", func() {
				existingProjects := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
					return cir.Project.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/project.graphqls",
					existingProjects,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.Project.Values
					})
			})
			It("returns correct pod", func() {
				existingPods := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
					return cir.Pod.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/pod.graphqls",
					existingPods,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.Pod.Values
					})
			})
			It("returns correct container", func() {
				existingContainers := seedCollection.GetComponentInstanceVal(func(cir mariadb.ComponentInstanceRow) string {
					return cir.Container.String
				})
				queryComponentInstanceFilterAndExpectVal(
					cfg.Port,
					"../api/graphql/graph/queryCollection/componentInstanceFilter/container.graphqls",
					existingContainers,
					func(cifv model.ComponentInstanceFilterValue) []*string {
						return cifv.Container.Values
					})
			})

		})
	})
})

func queryComponentInstanceFilterAndExpectVal(port string, gqlQueryFilePath string, expectedVal []string, getRespVal func(cifv model.ComponentInstanceFilterValue) []*string) {
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	b, err := os.ReadFile(gqlQueryFilePath)

	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		ComponentInstanceFilterValues model.ComponentInstanceFilterValue `json:"ComponentInstanceFilterValues"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}

	respVal := getRespVal(respData.ComponentInstanceFilterValues)
	Expect(len(respVal)).To(Equal(len(expectedVal)))

	for _, name := range respVal {
		Expect(lo.Contains(expectedVal, *name)).To(BeTrue())
	}
}
