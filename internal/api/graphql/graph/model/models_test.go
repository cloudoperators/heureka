// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"testing"

	"github.com/cloudoperators/heureka/internal/entity"
)

func ptrServiceOrderByField(field ServiceOrderByField) *ServiceOrderByField {
	return &field
}

func ptrOrderDirection(direction OrderDirection) *OrderDirection {
	return &direction
}

func TestServiceOrderBy_ToOrderEntity_Severity(t *testing.T) {
	orderBy := &ServiceOrderBy{
		By:        ptrServiceOrderByField(ServiceOrderByFieldSeverity),
		Direction: ptrOrderDirection(OrderDirectionDesc),
	}

	order := orderBy.ToOrderEntity()
	if order.By != entity.CriticalCount {
		t.Fatalf("expected severity order to map to entity.CriticalCount, got %v", order.By)
	}

	if order.Direction != entity.OrderDirectionDesc {
		t.Fatalf("expected order direction desc, got %v", order.Direction)
	}
}
