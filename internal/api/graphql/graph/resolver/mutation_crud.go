// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

// Generic helpers for the mechanical Create/Update/Delete mutation resolvers.
//
// This file is NOT a gqlgen schema-derived file, so `gqlgen generate` never
// writes or sweeps it: the generated resolver methods in mutation.go stay owned
// by gqlgen and simply delegate here. Error values are injected by each caller
// so the exact, client-observable op-tagged error strings are preserved; the
// underlying cause is intentionally discarded, matching the prior resolvers.

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
)

// createEntity marshals the GraphQL input into a domain entity, creates it via
// the app layer, and maps the result back to its GraphQL model. On failure it
// returns createErr verbatim.
func createEntity[Input any, Ent any, Model any](
	ctx context.Context,
	input Input,
	toEntity func(*Input) Ent,
	create func(context.Context, *Ent) (*Ent, error),
	toModel func(*Ent) Model,
	createErr error,
) (*Model, error) {
	e := toEntity(&input)

	created, err := create(ctx, &e)
	if err != nil {
		return nil, createErr
	}

	m := toModel(created)

	return &m, nil
}

// updateEntity parses the cursor id, marshals the input and stamps the id onto
// the entity, updates it, and maps the result back. Both the parse-failure and
// update-failure branches return updateErr verbatim, matching the current
// resolvers (which use one message for both).
func updateEntity[Input any, Ent any, Model any](
	ctx context.Context,
	id string,
	input Input,
	toEntity func(*Input) Ent,
	setID func(*Ent, int64),
	update func(context.Context, *Ent) (*Ent, error),
	toModel func(*Ent) Model,
	updateErr error,
) (*Model, error) {
	idInt, err := baseResolver.ParseCursor(&id)
	if err != nil {
		return nil, updateErr
	}

	e := toEntity(&input)
	setID(&e, *idInt)

	updated, err := update(ctx, &e)
	if err != nil {
		return nil, updateErr
	}

	m := toModel(updated)

	return &m, nil
}

// deleteEntity parses the cursor id and deletes the entity, echoing the id on
// success. parseErr and deleteErr are returned verbatim for their respective
// branches (usually identical, but a few resolvers tag them differently).
func deleteEntity(
	ctx context.Context,
	id string,
	del func(context.Context, int64) error,
	parseErr error,
	deleteErr error,
) (string, error) {
	idInt, err := baseResolver.ParseCursor(&id)
	if err != nil {
		return "", parseErr
	}

	if err := del(ctx, *idInt); err != nil {
		return "", deleteErr
	}

	return id, nil
}

// mutateRelation parses two cursor ids, invokes a relationship mutation on the
// app layer, and maps the returned entity to its model. All three failure
// branches (either parse, and the call) return relErr verbatim, matching the
// current relationship resolvers, which use one message throughout.
func mutateRelation[Ent any, Model any](
	ctx context.Context,
	idA string,
	idB string,
	call func(ctx context.Context, a int64, b int64) (*Ent, error),
	toModel func(*Ent) Model,
	relErr error,
) (*Model, error) {
	aInt, err := baseResolver.ParseCursor(&idA)
	if err != nil {
		return nil, relErr
	}

	bInt, err := baseResolver.ParseCursor(&idB)
	if err != nil {
		return nil, relErr
	}

	result, err := call(ctx, *aInt, *bInt)
	if err != nil {
		return nil, relErr
	}

	m := toModel(result)

	return &m, nil
}
