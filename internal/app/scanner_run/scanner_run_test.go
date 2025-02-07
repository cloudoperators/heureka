// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_run

import (
	"testing"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity/test"

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

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

var sre *entity.ScannerRun

var _ = Describe("ScannerRun", Label("app", "CreateScannerRun"), func() {
	var (
		db                *mocks.MockDatabase
		scannerRunHandler ScannerRunHandler
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		sre = test.NewFakeScannerRunEntity()
	})

	It("creates a scannerrun", func() {
		db.On("CreateScannerRun", sre).Return(sre, nil)

		scannerRunHandler = NewScannerRunHandler(db, er)
		scannerRunHandler.CreateScannerRun(sre)
	})

	It("creates a scannerrun and completes it", func() {
		db.On("CreateScannerRun", sre).Return(sre, nil)
		db.On("CompleteScannerRun", sre.UUID).Return(true, nil)

		scannerRunHandler = NewScannerRunHandler(db, er)
		scannerRunHandler.CreateScannerRun(sre)
		scannerRunHandler.CompleteScannerRun(sre.UUID)
	})
})
