// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_common

import "github.com/cloudoperators/heureka/internal/api/graphql/graph/model"

// CompareSeverityCounts compares two SeverityCounts structs.
// It returns 1 if a > b, -1 if a < b, and 0 if they are equal.
// The comparison is done in the order of Critical, High, Medium, Low, None.
func CompareSeverityCounts(a model.SeverityCounts, b model.SeverityCounts) int {
	aCounts := []int{a.Critical, a.High, a.Medium, a.Low, a.None}
	bCounts := []int{b.Critical, b.High, b.Medium, b.Low, b.None}

	for i := 0; i < len(aCounts); i++ {
		if aCounts[i] != bCounts[i] {
			if aCounts[i] > bCounts[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}
