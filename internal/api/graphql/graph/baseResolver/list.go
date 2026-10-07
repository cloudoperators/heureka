// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

// buildEdges maps a slice of list-result elements into their GraphQL connection
// edges using the per-entity makeEdge closure. It replaces the identical
// `edges := []*model.XEdge{}; for ... { edges = append(...) }` loop duplicated
// across every list base resolver. Entity-specific concerns (node constructor,
// cursor, and the optional edge.Priority on Service/IssueRepository) stay in the
// closure, so behavior is preserved exactly.
//
// The returned slice is always non-nil (empty for empty input), matching the
// prior `[]*model.XEdge{}` initialization.
func buildEdges[Elem any, Edge any](elements []Elem, makeEdge func(result Elem) Edge) []Edge {
	edges := make([]Edge, 0, len(elements))

	for _, result := range elements {
		edges = append(edges, makeEdge(result))
	}

	return edges
}
