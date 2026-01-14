// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_run

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
)

type scannerRunHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewScannerRunHandler(handlerContext common.HandlerContext) ScannerRunHandler {
	return &scannerRunHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
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

	// Trigger autopatch whenever a scanner run has completed successfully
	if _, err := srh.database.Autopatch(); err != nil {
		return false, &ScannerRunHandlerError{msg: "Error executing autopatch in CompleteScannerRun"}
	}

	srh.eventRegistry.PushEvent(&UpdateScannerRunEvent{successfulRun: true})
	return true, nil
}

func (srh *scannerRunHandler) FailScannerRun(uuid string, message string) (bool, error) {
	_, err := srh.database.FailScannerRun(uuid, message)

	if err != nil {
		return false, &ScannerRunHandlerError{msg: fmt.Sprintf("Error updating scanner run: %v", err)}
	}

	srh.eventRegistry.PushEvent(&UpdateScannerRunEvent{successfulRun: false})
	return true, nil
}

func (srhe *ScannerRunHandlerError) Error() string {
	return srhe.msg
}

func (srh *scannerRunHandler) GetScannerRunTags() ([]string, error) {
	var res []string

	res, err := srh.database.GetScannerRunTags()

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (srh *scannerRunHandler) GetScannerRuns(filter *entity.ScannerRunFilter, listOptions *entity.ListOptions) ([]entity.ScannerRun, error) {
	var res []entity.ScannerRun

	res, err := srh.database.GetScannerRuns(filter)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (srh *scannerRunHandler) CountScannerRuns(filter *entity.ScannerRunFilter) (int, error) {
	var res int

	res, err := srh.database.CountScannerRuns(filter)

	if err != nil {
		return -1, err
	}

	return res, nil
}
