// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeComponentInstanceEntity() entity.ComponentInstance {
	region := gofakeit.UUID()
	cluster := gofakeit.UUID()
	namespace := gofakeit.UUID()
	domain := gofakeit.UUID()
	project := gofakeit.UUID()
	return entity.ComponentInstance{
		Id:                 int64(gofakeit.Number(1, 10000000)),
		CCRN:               test.GenerateFakeCcrn(cluster, namespace),
		Region:             region,
		Cluster:            cluster,
		Namespace:          namespace,
		Domain:             domain,
		Project:            project,
		Count:              int16(gofakeit.Number(1, 100)),
		ComponentVersion:   nil,
		ComponentVersionId: int64(gofakeit.Number(1, 10000000)),
		Service:            nil,
		ServiceId:          int64(gofakeit.Number(1, 10000000)),
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakeComponentInstances(n int) []entity.ComponentInstance {
	r := make([]entity.ComponentInstance, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeComponentInstanceEntity()
	}
	return r
}

func NewFakeComponentInstanceResult() entity.ComponentInstanceResult {
	componentInstance := NewFakeComponentInstanceEntity()
	return entity.ComponentInstanceResult{
		ComponentInstance: &componentInstance,
	}
}
