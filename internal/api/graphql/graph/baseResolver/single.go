// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/sirupsen/logrus"
)

// singleNodeByChildIds implements the shared control flow of every
// Single<Entity>BaseResolver: validate the parent, fetch the at-most-one element
// identified by the parent's ChildIds, and wrap it into its GraphQL model.
//
// Error construction stays with each caller: the (static) no-parent and
// multiple-results errors are passed in, and the list closure returns its own
// wrapped error. This keeps the per-resolver, client-observable error output
// byte-identical while removing the duplicated skeleton. The two coexisting
// error styles (NewResolverError vs ToGraphQLError) are therefore preserved
// as-is; unifying them is the separate, gated error-consolidation step.
func singleNodeByChildIds[Elem any, Node any](
	ctx context.Context,
	resolverName string,
	parent *model.NodeParent,
	noParentErr error,
	multipleErr error,
	list func(ctx context.Context, ids []*int64) ([]Elem, error),
	wrap func(Elem) Node,
) (*Node, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called " + resolverName)

	if parent == nil {
		return nil, noParentErr
	}

	elements, err := list(ctx, parent.ChildIds)
	if err != nil {
		return nil, err
	}

	// unexpected number of results (should at most be 1)
	if len(elements) > 1 {
		return nil, multipleErr
	}

	// not found
	if len(elements) < 1 {
		return nil, nil
	}

	node := wrap(elements[0])

	return &node, nil
}
