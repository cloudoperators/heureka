// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/server"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Creating ScannerRun via API", Label("e2e", "ScannerRun"), func() {
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		db = dbm.NewTestSchemaWithoutMigration()

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Result bool `json:"createScannerRun"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/scannerRun/create.graphql",
					map[string]any{
						"input": map[string]string{
							"tag":  sampleTag,
							"uuid": sampleUUID,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Result).To(BeTrue())
			})
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun and completes the scanner run", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Result bool `json:"createScannerRun"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/scannerRun/create.graphql",
					map[string]any{
						"input": map[string]string{
							"tag":  sampleTag,
							"uuid": sampleUUID,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Result).To(BeTrue())

				newRespData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Result bool `json:"completeScannerRun"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/scannerRun/complete.graphql",
					map[string]any{
						"uuid": sampleUUID,
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(newRespData.Result).To(BeTrue())
			})
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun and fails the scanner run", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()
				sampleMessage := gofakeit.Sentence()

				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Result bool `json:"createScannerRun"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/scannerRun/create.graphql",
					map[string]any{
						"input": map[string]string{
							"tag":  sampleTag,
							"uuid": sampleUUID,
						},
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Result).To(BeTrue())

				newRespData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Result bool `json:"failScannerRun"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/scannerRun/fail.graphql",
					map[string]any{
						"message": sampleMessage,
						"uuid":    sampleUUID,
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(newRespData.Result).To(BeTrue())
			})
		})
	})
})

var _ = Describe("Querying ScannerRun via API", Label("e2e", "ScannerRun"), func() {
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		db = dbm.NewTestSchemaWithoutMigration()

		cfg = dbm.DbConfig()
		cfg.Port = e2e_common.GetRandomFreePort()
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("the database is empty", func() {
		Context("and a query for scannerruns is performed", Label("create.graphql"), func() {
			It("returns an empty list", func() {
				respData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
					Result int `json:"totalCount"`
				}](
					cfg.Port,
					"../api/graphql/graph/queryCollection/scannerRun/scannerruns.graphql",
					map[string]any{
						"filter": map[string]any{
							"tag":       []string{},
							"completed": false,
						},
						"first": 10,
						"after": 0,
					},
					nil,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(respData.Result).To(Equal(0))
			})
		})
	})
})
