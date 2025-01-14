// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_run

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	CreateScannerRunEventName event.EventName = "CreateScannerRun"
	UpdateScannerRunEventName event.EventName = "UpdateScannerRun"
)

type CreateScannerRunEvent struct {
	ScannerRun *entity.ScannerRun
}

type UpdateScannerRunEvent struct {
	ScannerRun *entity.ScannerRun
}

func (csr *CreateScannerRunEvent) Name() event.EventName {
	return CreateScannerRunEventName
}

func (csr *UpdateScannerRunEvent) Name() event.EventName {
	return UpdateScannerRunEventName
}
