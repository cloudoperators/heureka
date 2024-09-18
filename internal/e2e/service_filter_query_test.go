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

var _ = Describe("Getting ServiceFilterValues via API", Label("e2e", "ServiceFilterValues"), func() {
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
		It("returns empty resultset for serviceFilter", func() {
			// create a queryCollection (safe to share across requests)
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
			b, err := os.ReadFile("../api/graphql/graph/queryCollection/serviceFilter/serviceNames.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ServiceFilterValues.ServiceName.Values).To(BeEmpty())
		})
		It("returns empty for supportGroupNames", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			b, err := os.ReadFile("../api/graphql/graph/queryCollection/serviceFilter/supportGroupNames.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ServiceFilterValues.SupportGroupName.Values).To(BeEmpty())
		})
		It("returns empty for userNames", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			b, err := os.ReadFile("../api/graphql/graph/queryCollection/serviceFilter/userNames.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ServiceFilterValues.UserName.Values).To(BeEmpty())
		})
		It("returns empty for uniqueUserID", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

			b, err := os.ReadFile("../api/graphql/graph/queryCollection/serviceFilter/uniqueUserId.graphql")

			Expect(err).To(BeNil())
			str := string(b)
			req := graphql.NewRequest(str)

			req.Header.Set("Cache-Control", "no-cache")
			ctx := context.Background()

			var respData struct {
				ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
			}
			if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
				logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
			}

			Expect(respData.ServiceFilterValues.UniqueUserID.Values).To(BeEmpty())
		})
	})

	When("the database has 10 entries", func() {

		var seedCollection *test.SeedCollection
		BeforeEach(func() {
			seedCollection = seeder.SeedDbWithNFakeData(10)
		})
		Context("and no additional filters are present", func() {
			It("returns correct serviceNames", func() {
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				b, err := os.ReadFile("../api/graphql/graph/queryCollection/serviceFilter/serviceNames.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(len(respData.ServiceFilterValues.ServiceName.Values)).To(Equal(len(seedCollection.ServiceRows)))

				existingServiceNames := lo.Map(seedCollection.ServiceRows, func(s mariadb.BaseServiceRow, index int) string {
					return s.Name.String
				})

				for _, name := range respData.ServiceFilterValues.ServiceName.Values {
					Expect(lo.Contains(existingServiceNames, *name)).To(BeTrue())
				}
			})
			It("returns correct supportGroupNames", func() {
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				b, err := os.ReadFile("../api/graphql/graph/queryCollection/serviceFilter/supportGroupNames.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(len(respData.ServiceFilterValues.SupportGroupName.Values)).To(Equal(len(seedCollection.SupportGroupRows)))

				existingSupportGroupNames := lo.Map(seedCollection.SupportGroupRows, func(s mariadb.SupportGroupRow, index int) string {
					return s.Name.String
				})

				for _, name := range respData.ServiceFilterValues.SupportGroupName.Values {
					Expect(lo.Contains(existingSupportGroupNames, *name)).To(BeTrue())
				}
			})
			It("returns correct userNames", func() {
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				b, err := os.ReadFile("../api/graphql/graph/queryCollection/serviceFilter/userNames.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(len(respData.ServiceFilterValues.UserName.Values)).To(Equal(len(seedCollection.UserRows)))

				existingUserNames := lo.Map(seedCollection.UserRows, func(s mariadb.UserRow, index int) string {
					return s.Name.String
				})

				for _, name := range respData.ServiceFilterValues.UserName.Values {
					Expect(lo.Contains(existingUserNames, *name)).To(BeTrue())
				}
			})
			It("returns correct UniqueUserID", func() {
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				b, err := os.ReadFile("../api/graphql/graph/queryCollection/serviceFilter/uniqueUserId.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					ServiceFilterValues model.ServiceFilterValue `json:"ServiceFilterValues"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(len(respData.ServiceFilterValues.UniqueUserID.Values)).To(Equal(len(seedCollection.UserRows)))

				existingUniqueUserIds := lo.Map(seedCollection.UserRows, func(s mariadb.UserRow, index int) string {
					return s.UniqueUserID.String
				})

				for _, name := range respData.ServiceFilterValues.UniqueUserID.Values {
					Expect(lo.Contains(existingUniqueUserIds, *name)).To(BeTrue())
				}
			})
		})
	})
})
