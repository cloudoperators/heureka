// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/server"

	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Creating ScannerRun via API", Label("e2e", "ScannerRun"), func() {

	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		db = dbm.NewTestSchema()

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()

				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/scannerRun/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"tag":  sampleTag,
					"uuid": sampleUUID,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Result bool `json:"createScannerRun"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Result).To(BeTrue())
			})
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun and completes the scanner run", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()

				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/scannerRun/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"tag":  sampleTag,
					"uuid": sampleUUID,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Result bool `json:"createScannerRun"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Result).To(BeTrue())

				b, err = os.ReadFile("../api/graphql/graph/queryCollection/scannerRun/complete.graphql")

				Expect(err).To(BeNil())
				str = string(b)
				new_req := graphql.NewRequest(str)

				new_req.Var("uuid", sampleUUID)

				new_req.Header.Set("Cache-Control", "no-cache")
				ctx = context.Background()

				var newRespData struct {
					Result bool `json:"completeScannerRun"`
				}

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, new_req, &newRespData) }); err != nil {
					logrus.WithError(err).WithField("request", new_req).Fatalln("Error while unmarshaling")
				}

				Expect(newRespData.Result).To(BeTrue())
			})
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun and fails the scanner run", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()
				sampleMessage := gofakeit.Sentence(10)

				// create a queryCollection (safe to share across requests)
				client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))

				//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/scannerRun/create.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				req := graphql.NewRequest(str)

				req.Var("input", map[string]string{
					"tag":  sampleTag,
					"uuid": sampleUUID,
				})

				req.Header.Set("Cache-Control", "no-cache")
				ctx := context.Background()

				var respData struct {
					Result bool `json:"createScannerRun"`
				}
				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
					logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
				}

				Expect(respData.Result).To(BeTrue())

				b, err = os.ReadFile("../api/graphql/graph/queryCollection/scannerRun/fail.graphql")

				Expect(err).To(BeNil())
				str = string(b)
				new_req := graphql.NewRequest(str)

				new_req.Var("message", sampleMessage)
				new_req.Var("uuid", sampleUUID)

				new_req.Header.Set("Cache-Control", "no-cache")
				ctx = context.Background()

				var newRespData struct {
					Result bool `json:"failScannerRun"`
				}

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, new_req, &newRespData) }); err != nil {
					logrus.WithError(err).WithField("request", new_req).Fatalln("Error while unmarshaling")
				}

				Expect(newRespData.Result).To(BeTrue())
			})
		})

	})

})

var _ = Describe("Querying ScannerRun via API", Label("e2e", "ScannerRun"), func() {

	var s *server.Server
	var cfg util.Config
	var client *graphql.Client
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		db = dbm.NewTestSchema()

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
		client = graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))
	})

	AfterEach(func() {
		s.BlockingStop()
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {

		Context("and a query for scannerruns is performed", Label("create.graphql"), func() {
			It("returns an empty list", func() {
				b, err := os.ReadFile("../api/graphql/graph/queryCollection/scannerRun/scannerruns.graphql")

				Expect(err).To(BeNil())
				str := string(b)
				graphql.NewRequest(str)

				new_req := graphql.NewRequest(str)

				new_req.Header.Set("Cache-Control", "no-cache")

				new_req.Var("filter", struct {
					Tag       []string `json:"tag"`
					Completed bool     `json:"completed"`
				}{Tag: []string{}, Completed: false})

				new_req.Var("first", 10)
				new_req.Var("after", 0)

				ctx := context.Background()

				var newRespData struct {
					Result int `json:"totalCount"`
				}

				if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, new_req, &newRespData) }); err != nil {
					logrus.WithError(err).WithField("request", new_req).Fatalln("Error while unmarshaling")
				}

				Expect(newRespData.Result).To(Equal(0))
			})
		})

	})

})
