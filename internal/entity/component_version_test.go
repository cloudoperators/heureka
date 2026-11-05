// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity_test

import (
	"testing"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
)

func TestIsVersionOutdated(t *testing.T) {
	cases := []struct {
		name     string
		tag      string
		latest   *string
		expected *bool
	}{
		{"running tag behind latest -> outdated", "v1.2.3", lo.ToPtr("v1.3.0"), lo.ToPtr(true)},
		{"running tag equals latest -> not outdated", "v1.3.0", lo.ToPtr("v1.3.0"), lo.ToPtr(false)},
		{"running tag ahead of latest -> not outdated", "v1.3.0", lo.ToPtr("v1.2.0"), lo.ToPtr(false)},
		{"patch bump -> outdated", "1.2.3", lo.ToPtr("1.2.4"), lo.ToPtr(true)},
		{"non-semver running tag -> unknown", "latest", lo.ToPtr("v1.3.0"), nil},
		{"non-semver latest -> unknown", "v1.2.3", lo.ToPtr("rolling"), nil},
		{"no latest version known -> unknown", "v1.2.3", nil, nil},
		{"empty latest version -> unknown", "v1.2.3", lo.ToPtr(""), nil},
		{"empty running tag -> unknown", "", lo.ToPtr("v1.3.0"), nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := entity.IsVersionOutdated(tc.tag, tc.latest)
			switch {
			case tc.expected == nil && got != nil:
				t.Fatalf("expected nil, got %v", *got)
			case tc.expected != nil && got == nil:
				t.Fatalf("expected %v, got nil", *tc.expected)
			case tc.expected != nil && *got != *tc.expected:
				t.Fatalf("expected %v, got %v", *tc.expected, *got)
			}
		})
	}
}
