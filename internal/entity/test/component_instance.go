// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeComponentInstanceEntity() entity.ComponentInstance {
	region := gofakeit.RandomString([]string{"test-de-1", "test-de-2", "test-us-1", "test-jp-2", "test-jp-1"})
	cluster := gofakeit.RandomString([]string{"test-de-1", "test-de-2", "test-us-1", "test-jp-2", "test-jp-1", "a-test-de-1", "a-test-de-2", "a-test-us-1", "a-test-jp-2", "a-test-jp-1", "v-test-de-1", "v-test-de-2", "v-test-us-1", "v-test-jp-2", "v-test-jp-1", "s-test-de-1", "s-test-de-2", "s-test-us-1", "s-test-jp-2", "s-test-jp-1"})
	//make lower case to avoid conflicts in different lexicographical ordering between sql and golang due to collation
	namespace := strings.ToLower(gofakeit.ProductName())
	domain := strings.ToLower(gofakeit.SongName())
	project := strings.ToLower(gofakeit.BeerName())
	pod := strings.ToLower(gofakeit.UUID())
	container := strings.ToLower(gofakeit.UUID())
	t := gofakeit.RandomString(entity.AllComponentInstanceType)
	context := entity.Json{
		"timeout_nbd":               gofakeit.Number(1, 60),
		"remove_unused_base_images": gofakeit.Bool(),
		"my_ip":                     gofakeit.IPv4Address(),
	}
	return entity.ComponentInstance{
		Id:                 int64(gofakeit.Number(1, 10000000)),
		CCRN:               test.GenerateFakeCcrn(cluster, namespace),
		Region:             region,
		Cluster:            cluster,
		Namespace:          namespace,
		Domain:             domain,
		Project:            project,
		Pod:                pod,
		Container:          container,
		Type:               entity.NewComponentInstanceType(t),
		ParentId:           int64(gofakeit.Number(1, 10000000)),
		Context:            &context,
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
