// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

// unionAccessibleIds merges two GetListOfAccessibleObjectIds results, treating the [-1]
// "no access" sentinel as contributing no ids.
func unionAccessibleIds(a []*int64, b []*int64) []*int64 {
	// Check if either lists are empty or contain [-1]
	noAccess := func(ids []*int64) bool {
		return len(ids) == 0 || (len(ids) == 1 && ids[0] != nil && *ids[0] == -1)
	}

	aEmpty, bEmpty := noAccess(a), noAccess(b)
	if aEmpty && bEmpty {
		if len(a) > 0 {
			return a
		}
		return b
	}

	// merge otherwise
	seen := make(map[int64]struct{})
	var merged []*int64

	for _, ids := range [][]*int64{a, b} {
		if noAccess(ids) {
			continue
		}
		for _, id := range ids {
			if id == nil {
				continue
			}
			if _, ok := seen[*id]; !ok {
				seen[*id] = struct{}{}
				merged = append(merged, id)
			}
		}
	}

	return merged
}

// CombineFilterWithAccessibleIds combines filterIds and accessibleIds based on the following rules:
// - If accessibleIds is empty, return filterIds (full access)
// - If accessibleIds contains only -1, return [-1] (no access)
// - If filterIds is empty, return accessibleIds (use accessibleIds as filter)
// - Otherwise, calculate & return the intersection of filterIds and accessibleIds
func CombineFilterWithAccessibleIds(filterIds []*int64, accessibleIds []*int64) []*int64 {
	if len(accessibleIds) == 1 && accessibleIds[0] != nil && *accessibleIds[0] == -1 {
		filterIds = accessibleIds
	} else if len(accessibleIds) > 0 {
		// Partial access: intersect filterIds and accessibleIds
		if len(filterIds) > 0 {
			// Intersection of filterIds and accessibleIds
			filterIds = getIntersectionOfIdSlices(filterIds, accessibleIds)

			// If intersection is empty, return [-1] to indicate no access
			if len(filterIds) == 0 {
				filterIds = []*int64{new(int64(-1))}
			}
		} else {
			// No filterIds: use accessibleIds as filter
			filterIds = accessibleIds
		}
	}

	return filterIds
}

// GetIntersectionOfSlices returns the intersection of two slices of int64
// Example: slice1 = [1, 2, 3], slice2 = [2, 3, 4] => returns [2, 3]
func getIntersectionOfIdSlices(slice1 []*int64, slice2 []*int64) []*int64 {
	set := make(map[int64]struct{})

	for _, v := range slice1 {
		if v == nil {
			continue
		}

		set[*v] = struct{}{}
	}

	var intersection []*int64

	for _, v := range slice2 {
		if v == nil {
			continue
		}

		if _, found := set[*v]; found {
			intersection = append(intersection, v)
		}
	}

	return intersection
}

// Int64Ptr returns a pointer to the given int64 value.
//
//go:fix inline
func Int64Ptr(i int64) *int64 {
	return new(i)
}
