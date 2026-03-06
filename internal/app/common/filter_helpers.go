package common

// CombineFilterWithAccesibleIds combines filterIds and accessibleIds based on the following rules:
// - If accessibleIds is empty, return filterIds (full access)
// - If accessibleIds contains only -1, return [-1] (no access)
// - If filterIds is empty, return accessibleIds (use accessibleIds as filter)
// - Otherwise, calculate & return the intersection of filterIds and accessibleIds
func CombineFilterWithAccesibleIds(filterIds []*int64, accessibleIds []*int64) []*int64 {
	if len(accessibleIds) == 1 && accessibleIds[0] != nil && *accessibleIds[0] == -1 {
		filterIds = accessibleIds
	} else if len(accessibleIds) > 0 {
		// Partial access: intersect filterIds and accessibleIds
		if len(filterIds) > 0 {
			// Intersection of filterIds and accessibleIds
			filterIds = getIntersectionOfIdSlices(filterIds, accessibleIds)

			// If intersection is empty, return [-1] to indicate no access
			if len(filterIds) == 0 {
				filterIds = []*int64{Int64Ptr(-1)}
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
		set[*v] = struct{}{}
	}

	var intersection []*int64
	for _, v := range slice2 {
		if _, found := set[*v]; found {
			intersection = append(intersection, v)
		}
	}
	return intersection
}

// Int64Ptr returns a pointer to the given int64 value.
func Int64Ptr(i int64) *int64 {
	return &i
}
