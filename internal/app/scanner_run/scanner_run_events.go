// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_run

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	CreateScannerRunEventName event.EventName = "CreateScannerRun"
	UpdateScannerRunEventName event.EventName = "UpdateScannerRun"
)

type CreateScannerRunEvent struct {
	ScannerRun *entity.ScannerRun
}

func (e CreateScannerRunEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateScannerRunEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

type UpdateScannerRunEvent struct {
	successfulRun bool
}

func (e UpdateScannerRunEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateScannerRunEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (csr *CreateScannerRunEvent) Name() event.EventName {
	return CreateScannerRunEventName
}

func (csr *UpdateScannerRunEvent) Name() event.EventName {
	return UpdateScannerRunEventName
}
