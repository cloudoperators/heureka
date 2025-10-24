// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_run

import (
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/openfga"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServiceHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Service Test Suite")
}

var er event.EventRegistry
var authz openfga.Authorization

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db, authz)
})

var sre *entity.ScannerRun

var _ = Describe("ScannerRun", Label("app", "CreateScannerRun"), func() {
	var (
		db                *mocks.MockDatabase
		scannerRunHandler ScannerRunHandler
		handlerContext    common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		sre = test.NewFakeScannerRunEntity()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
		}
	})

	It("creates a scannerrun", func() {
		db.On("CreateScannerRun", sre).Return(true, nil)

		scannerRunHandler = NewScannerRunHandler(handlerContext)
		_, err := scannerRunHandler.CreateScannerRun(sre)
		Expect(err).To(BeNil())
	})

	It("creates a scannerrun and completes it", func() {
		db.On("CreateScannerRun", sre).Return(true, nil)
		db.On("CompleteScannerRun", sre.UUID).Return(true, nil)
		db.On("Autoclose").Return(true, nil)

		scannerRunHandler = NewScannerRunHandler(handlerContext)
		scannerRunHandler.CreateScannerRun(sre)
		_, err := scannerRunHandler.CompleteScannerRun(sre.UUID)

		Expect(err).To(BeNil())
	})

	It("creates a scannerrun and fails it", func() {
		db.On("CreateScannerRun", sre).Return(true, nil)
		db.On("FailScannerRun", sre.UUID, "Booom!").Return(true, nil)

		scannerRunHandler = NewScannerRunHandler(handlerContext)
		scannerRunHandler.CreateScannerRun(sre)
		_, err := scannerRunHandler.FailScannerRun(sre.UUID, "Booom!")

		Expect(err).To(BeNil())
	})

	It("Retrieves list of scannerrun tags", func() {
		db.On("GetScannerRunTags").Return([]string{"tag1", "tag2"}, nil)

		scannerRunHandler = NewScannerRunHandler(handlerContext)
		tags, err := scannerRunHandler.GetScannerRunTags()

		Expect(err).To(BeNil())
		Expect(tags).To(HaveLen(2))
		Expect(tags).To(Equal([]string{"tag1", "tag2"}))
	})

	It("Retrieves list of scannerruns", func() {

		db.On("GetScannerRuns", &entity.ScannerRunFilter{}).Return([]entity.ScannerRun{*sre}, nil)

		scannerRunHandler = NewScannerRunHandler(handlerContext)
		scannerRuns, err := scannerRunHandler.GetScannerRuns(&entity.ScannerRunFilter{}, nil)

		Expect(err).To(BeNil())
		Expect(scannerRuns).To(HaveLen(1))
		Expect(scannerRuns).To(Equal([]entity.ScannerRun{*sre}))
	})
})
