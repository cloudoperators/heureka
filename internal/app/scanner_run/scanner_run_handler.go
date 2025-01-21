// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_run

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
)

type scannerRunHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewScannerRunHandler(db database.Database, er event.EventRegistry) ScannerRunHandler {
	return &scannerRunHandler{
		database:      db,
		eventRegistry: er,
	}
}

type ScannerRunHandlerError struct {
	msg string
}

func (srh *scannerRunHandler) CreateScannerRun(sr *entity.ScannerRun) (bool, error) {
	_, err := srh.database.CreateScannerRun(sr)

	if err != nil {
		return false, &ScannerRunHandlerError{msg: "Error creating scanner run"}
	}

	srh.eventRegistry.PushEvent(&CreateScannerRunEvent{sr})
	return true, nil
}

func (srh *scannerRunHandler) CompleteScannerRun(uuid string) (bool, error) {
	_, err := srh.database.CompleteScannerRun(uuid)

	if err != nil {
		return false, &ScannerRunHandlerError{msg: "Error updating scanner run"}
	}

	srh.eventRegistry.PushEvent(&UpdateScannerRunEvent{})
	return true, nil
}

func (srhe *ScannerRunHandlerError) Error() string {
	return srhe.msg
}
