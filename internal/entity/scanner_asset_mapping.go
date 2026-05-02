// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type ScannerAssetMapping struct {
	Id                  int64     `db:"id"`
	ScannerName         string    `db:"scanner_name"`
	ArtifactUri         string    `db:"artifact_uri"`
	ComponentInstanceId int64     `db:"component_instance_id"`
	ServiceId           int64     `db:"service_id"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
}
