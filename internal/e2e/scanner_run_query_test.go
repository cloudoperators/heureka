// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/server"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type scannerRunTest struct {
	s   *server.Server
	cfg util.Config
	db  *mariadb.SqlDatabase
}

func newScannerRunTest() *scannerRunTest {
	db := dbm.NewTestSchemaWithoutMigration()
	cfg := dbm.DbConfig()
	cfg.Port = e2e_common.GetRandomFreePort()
	cfg.AuthzOpenFgaApiUrl = ""
	s := e2e_common.NewRunningServer(cfg)

	return &scannerRunTest{s: s, cfg: cfg, db: db}
}

func (srt *scannerRunTest) tearDown() {
	e2e_common.ServerTeardown(srt.s)
	dbm.TestTearDown(srt.db)
}

func (srt *scannerRunTest) requestCreateAndExpectSuccess(uuid string, tag string) {
	resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Result bool `json:"createScannerRun"`
	}](
		srt.cfg.Port,
		"../api/graphql/graph/queryCollection/scannerRun/create.graphql",
		map[string]any{
			"input": map[string]string{
				"tag":  tag,
				"uuid": uuid,
			},
		},
		nil,
	)

	Expect(err).ToNot(HaveOccurred())
	Expect(resp.Result).To(BeTrue())
}

func (srt *scannerRunTest) requestGetCountAndExpectSuccess(tag string, completed bool) model.ScannerRunConnection {
	resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Result model.ScannerRunConnection `json:"ScannerRuns"`
	}](
		srt.cfg.Port,
		"../api/graphql/graph/queryCollection/scannerRun/scannerruns.graphql",
		map[string]any{
			"filter": map[string]any{
				"tag":       []string{tag},
				"completed": completed,
			},
			"first": 10,
			"after": 0,
		},
		nil,
	)

	Expect(err).ToNot(HaveOccurred())

	return resp.Result
}

func (srt *scannerRunTest) requestCompleteAndExpectSuccess(uuid string) {
	newRespData, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Result bool `json:"completeScannerRun"`
	}](
		srt.cfg.Port,
		"../api/graphql/graph/queryCollection/scannerRun/complete.graphql",
		map[string]any{
			"uuid": uuid,
		},
		nil,
	)

	Expect(err).ToNot(HaveOccurred())
	Expect(newRespData.Result).To(BeTrue())
}

func (srt *scannerRunTest) requestFailAndExpectSuccess(uuid string, errMsg string) {
	respDataFail, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Result bool `json:"failScannerRun"`
	}](
		srt.cfg.Port,
		"../api/graphql/graph/queryCollection/scannerRun/fail.graphql",
		map[string]any{
			"message": errMsg,
			"uuid":    uuid,
		},
		nil,
	)

	Expect(err).ToNot(HaveOccurred())
	Expect(respDataFail.Result).To(BeTrue())
}

func (srt *scannerRunTest) requestGetAndExpectSuccess(tag string, completed bool) model.ScannerRunConnection {
	resp, err := e2e_common.ExecuteGqlQueryFromFileWithHeaders[struct {
		Result model.ScannerRunConnection `json:"ScannerRuns"`
	}](
		srt.cfg.Port,
		"../api/graphql/graph/queryCollection/scannerRun/state.graphql",
		map[string]any{
			"filter": map[string]any{
				"tag":       []string{tag},
				"completed": completed,
			},
			"first": 10,
			"after": 0,
		},
		nil,
	)
	Expect(err).ToNot(HaveOccurred())

	return resp.Result
}

var _ = Describe("Creating ScannerRun via API", Label("e2e", "ScannerRun"), func() {
	var srt *scannerRunTest

	BeforeEach(func() {
		srt = newScannerRunTest()
	})

	AfterEach(func() {
		srt.tearDown()
	})

	When("the database is empty", func() {
		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()

				srt.requestCreateAndExpectSuccess(sampleUUID, sampleTag)
			})
		})

		Context("and a get query is performed", Label("create.graphql"), func() {
			It("responds with ScannerRun data", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()

				srt.requestCreateAndExpectSuccess(sampleUUID, sampleTag)
				result := srt.requestGetCountAndExpectSuccess(sampleTag, false)

				Expect(result.TotalCount).To(Equal(1))
			})
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun and completes the scanner run", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()

				srt.requestCreateAndExpectSuccess(sampleUUID, sampleTag)
				srt.requestCompleteAndExpectSuccess(sampleUUID)
				result := srt.requestGetAndExpectSuccess(sampleTag, false)

				Expect(result.TotalCount).To(Equal(1))
				Expect(result.Edges).To(HaveLen(1))
				Expect(result.Edges[0].Node.Completed).To(Equal(true))
				Expect(result.Edges[0].Node.Error).To(BeEmpty())
			})
		})

		Context("and a mutation query is performed", Label("create.graphql"), func() {
			It("creates new ScannerRun and fails the scanner run", func() {
				sampleTag := gofakeit.Word()
				sampleUUID := gofakeit.UUID()
				sampleErrorMessage := gofakeit.Sentence()

				srt.requestCreateAndExpectSuccess(sampleUUID, sampleTag)
				srt.requestCompleteAndExpectSuccess(sampleUUID)
				srt.requestFailAndExpectSuccess(sampleUUID, sampleErrorMessage)
				result := srt.requestGetAndExpectSuccess(sampleTag, false)

				Expect(result.TotalCount).To(Equal(1))
				Expect(result.Edges).To(HaveLen(1))
				Expect(result.Edges[0].Node.Completed).To(Equal(false))
				Expect(result.Edges[0].Node.Error).To(Equal(sampleErrorMessage))
			})
		})
	})
})
