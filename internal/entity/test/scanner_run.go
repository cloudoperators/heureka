// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeScannerRunEntity() *entity.ScannerRun {
	startRun := gofakeit.Date()
	endRun := startRun.Add(time.Hour)

	return &entity.ScannerRun{
		RunID: int64(gofakeit.Number(1, 10000000)),

		UUID:      gofakeit.UUID(),
		Tag:       gofakeit.Word(),
		StartRun:  startRun,
		EndRun:    endRun,
		Completed: false,
	}
}
